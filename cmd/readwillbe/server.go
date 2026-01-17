package main

import (
	"fmt"
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

	"readwillbe/internal/cache"
	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	"readwillbe/internal/service/push"
	"readwillbe/static"

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

	cfg, err := model.ConfigFromViper()
	if err != nil {
		return errors.Wrap(err, "loading config from viper")
	}

	e := echo.New()

	// Disable caching for static assets in dev mode
	if !cfg.IsProduction() {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if strings.HasPrefix(c.Request().URL.Path, "/static/") {
					c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
					c.Response().Header().Set("Pragma", "no-cache")
					c.Response().Header().Set("Expires", "0")
				}
				return next(c)
			}
		})
	}

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
	e.Pre(mw.MethodOverride()) // Must use Pre() to run before routing

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
		CookieHTTPOnly: false, // Must be false so JavaScript can read the token for AJAX requests
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

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogMethod: true,
		Skipper: func(c echo.Context) bool {
			return c.Request().URL.Path == "/healthz"
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			fmt.Printf("method=%s, uri=%s, status=%d\n", v.Method, v.URI, v.Status)
			return nil
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

	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = db.AutoMigrate(&model.User{}, &model.Plan{}, &model.Reading{}, &model.PushSubscription{})
	if err != nil {
		return errors.Wrap(err, "failed to migrate")
	}

	if cfg.SeedDB {
		appFS := afero.NewOsFs()
		if err := seedDatabase(db, appFS); err != nil {
			return errors.Wrap(err, "seeding database")
		}
	}

	_ = push.StartNotificationWorker(cfg, db)

	store := sessions.NewCookieStore(cfg.CookieSecret)
	store.Options = mw.GetSecureSessionOptions(cfg)
	e.Use(session.Middleware(store))
	userCache := cache.NewUserCache(5*time.Minute, 10*time.Minute)
	e.Use(mw.UserMiddleware(db, userCache, cfg))

	appFS := afero.NewOsFs()

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
	e.GET("/partials/dashboard-stats", dashboardStatsPartial(db))
	e.GET("/history", historyHandler(cfg, db))
	e.GET("/plans", plansListHandler(cfg, db))
	e.GET("/plans/create", createPlanForm(cfg, db))
	e.POST("/plans/create", createPlan(appFS, db), generalRateLimiter)
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
	e.POST("/account/test-email", sendTestEmailHandler(cfg), generalRateLimiter)

	e.GET("/notifications/count", notificationCount(db))
	e.GET("/notifications/dropdown", notificationDropdown(db))

	// JSON API endpoints for React components
	e.GET("/api/notifications/count", apiNotificationCount(db))
	e.GET("/api/notifications/readings", apiNotificationReadings(db))
	e.GET("/api/plans/:id/status", apiPlanStatus(db))
	e.PUT("/plans/draft", apiSaveDraft(), generalRateLimiter)

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
	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(parsedLevel)
	}
}
