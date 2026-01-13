package main

import (
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"

	"readwillbe/static"
	"readwillbe/types"

	_ "github.com/ncruces/go-sqlite3/embed"
	sqlite "github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func render(ctx echo.Context, status int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(status)

	err := t.Render(ctx.Request().Context(), ctx.Response().Writer)
	if err != nil {
		logrus.Errorf("Failed to render template: %v", err)
		return ctx.String(http.StatusInternalServerError, "failed to render response template")
	}

	return nil
}

func htmxRedirect(c echo.Context, url string) error {
	c.Response().Header().Set("HX-Redirect", url)
	return c.NoContent(http.StatusOK)
}

func runServer(cmd *cobra.Command, args []string) error {
	configureLogging()

	tz := viper.GetString("tz")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return errors.Wrap(err, "failed to load timezone")
		}
		time.Local = loc
	}

	cfg, err := types.ConfigFromViper()
	if err != nil {
		return errors.Wrap(err, "loading config from viper")
	}

	e := echo.New()

	e.StaticFS("/static", static.FS)
	e.GET("/serviceWorker.js", func(c echo.Context) error {
		data, readErr := fs.ReadFile(static.FS, "serviceWorker.js")
		if readErr != nil {
			return readErr
		}
		return c.Blob(http.StatusOK, "application/javascript", data)
	})

	origErrHandler := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		logrus.Error(err)
		origErrHandler(err, c)
	}

	e.Use(middleware.RequestID())
	e.Use(middleware.BodyLimit("11M"))

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogLevel:          log.ERROR,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logrus.Error(errors.Wrap(err, "recovered panic:"))
			for _, l := range strings.Split(string(stack), "\n") {
				logrus.Errorf("stack: %s", strings.ReplaceAll(l, "\t", "  "))
			}
			return nil
		},
		DisableErrorHandler: false,
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' https://unpkg.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))
	if cfg.IsProduction() {
		e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
			Level:     5,
			MinLength: 1400,
			Skipper:   middleware.DefaultSkipper,
		}))
	}

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:_csrf,header:X-CSRF-Token",
		CookiePath:     "/",
		CookieSecure:   cfg.IsProduction(),
		CookieHTTPOnly: false, // Must be false so JavaScript (HTMX) can read the token for AJAX requests
		CookieSameSite: http.SameSiteStrictMode,
		Skipper: func(c echo.Context) bool {
			return c.Path() == "/healthz"
		},
		ErrorHandler: func(err error, c echo.Context) error {
			if cfg.IsProduction() && c.Request().TLS == nil {
				logrus.Error("CSRF validation failed: secure cookies are enabled (GO_ENV=production) but request is not HTTPS. Either use HTTPS or set GO_ENV to something other than 'production' or 'prod'")
			}
			return echo.NewHTTPError(http.StatusForbidden, "invalid csrf token")
		},
	}))

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/healthz"
		},
	}))

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get underlying sql.DB")
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = db.AutoMigrate(&types.User{}, &types.Plan{}, &types.Reading{}, &types.PushSubscription{})
	if err != nil {
		return errors.Wrap(err, "failed to migrate")
	}

	if cfg.SeedDB {
		fs := afero.NewOsFs()
		if err := seedDatabase(db, fs); err != nil {
			return errors.Wrap(err, "seeding database")
		}
	}

	_ = startNotificationWorker(cfg, db)

	store := sessions.NewCookieStore(cfg.CookieSecret)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteStrictMode,
	}
	e.Use(session.Middleware(store))
	userCache := NewUserCache(5*time.Minute, 10*time.Minute)
	e.Use(UserMiddleware(db, userCache, cfg))

	e.GET("/", dashboardHandler(cfg, db))
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	authRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(5)))
	generalRateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(30)))

	e.GET("/auth/sign-in", signIn(cfg))
	e.POST("/auth/sign-in", signInWithEmailAndPassword(db, cfg), authRateLimiter)
	if cfg.AllowSignup {
		e.GET("/auth/sign-up", signUp(cfg))
		e.POST("/auth/sign-up", signUpWithEmailAndPassword(db, cfg), authRateLimiter)
	}
	e.POST("/auth/sign-out", signOut(), generalRateLimiter)

	e.GET("/dashboard", dashboardHandler(cfg, db))
	e.GET("/history", historyHandler(cfg, db))
	e.GET("/plans", plansListHandler(cfg, db))
	e.GET("/plans/create", createPlanForm(cfg, db))
	e.POST("/plans/create", createPlan(db), generalRateLimiter)
	e.GET("/plans/create-manual", manualPlanForm(cfg))
	e.POST("/plans/create-manual", createManualPlan(cfg, db), generalRateLimiter)
	e.POST("/plans/draft/title", updateDraftTitle(), generalRateLimiter)
	e.POST("/plans/draft/reading", addDraftReading(), generalRateLimiter)
	e.GET("/plans/draft/reading/:id", getDraftReading())
	e.GET("/plans/draft/reading/:id/edit", getDraftReadingEdit())
	e.PUT("/plans/draft/reading/:id", updateDraftReading(), generalRateLimiter)
	e.DELETE("/plans/draft/reading/:id", deleteDraftReading(), generalRateLimiter)
	e.DELETE("/plans/draft", deleteDraft(), generalRateLimiter)
	e.GET("/plans/:id/edit", editPlanForm(cfg, db))
	e.POST("/plans/:id/edit", editPlan(cfg, db), generalRateLimiter)
	e.POST("/plans/:id/rename", renamePlan(db), generalRateLimiter)
	e.DELETE("/plans/:id", deletePlan(db), generalRateLimiter)
	e.DELETE("/plans/:id/readings/:reading_id", deleteReading(db), generalRateLimiter)
	e.GET("/account", accountHandler(cfg, db))
	e.POST("/account/settings", updateSettings(db), generalRateLimiter)

	e.GET("/notifications/count", notificationCount(db))
	e.GET("/notifications/dropdown", notificationDropdown(db))

	e.POST("/push/subscribe", saveSubscription(db), generalRateLimiter)
	e.POST("/push/unsubscribe", removeSubscription(db), generalRateLimiter)
	e.POST("/push/unsubscribe-all", removeAllSubscriptions(db), generalRateLimiter)

	e.POST("/reading/:id/complete", completeReading(db), generalRateLimiter)
	e.POST("/reading/:id/uncomplete", uncompleteReading(db), generalRateLimiter)
	e.POST("/reading/:id/update", updateReading(db), generalRateLimiter)

	return e.Start(cfg.Port)
}

func configureLogging() {
	level := viper.GetString("log_level")

	switch strings.ToLower(level) {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func getSecureSessionOptions(cfg types.Config) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteStrictMode,
	}
}

func UserMiddleware(db *gorm.DB, cache *UserCache, cfg types.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				logrus.Warnf("Failed to get session: %v", err)
				return next(c)
			}
			if sess.Values[SessionUserIDKey] != nil {
				userID, ok := sess.Values[SessionUserIDKey].(uint)
				if !ok {
					return next(c)
				}

				user, found := cache.Get(userID)
				if !found {
					var err error
					user, err = getUserByID(db.WithContext(c.Request().Context()), userID)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							delete(sess.Values, SessionUserIDKey)
							_ = sess.Save(c.Request(), c.Response())
							return next(c)
						}
						return errors.Wrap(err, "getting user by id")
					}
					cache.Set(user)
				}

				c.Set(UserKey, user)

				// Only save session if the user ID has changed to avoid unnecessary Set-Cookie headers
				if sess.Values[SessionUserIDKey] != user.ID {
					sess.Options = getSecureSessionOptions(cfg)
					sess.Values[SessionUserIDKey] = user.ID

					err := sess.Save(c.Request(), c.Response())
					if err != nil {
						return errors.Wrap(err, "saving session")
					}
				}
			}
			return next(c)
		}
	}
}

func GetSessionUser(c echo.Context) (types.User, bool) {
	u := c.Get(UserKey)
	if u != nil {
		user, ok := u.(types.User)
		if !ok {
			return types.User{}, false
		}
		logrus.Debugf("Found session user ID=%d", user.ID)
		return user, true
	}
	return types.User{}, false
}
