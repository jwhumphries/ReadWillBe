package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"readwillbe/static"
	"readwillbe/types"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	_ "github.com/ncruces/go-sqlite3/embed"
	sqlite "github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

const SessionKey = "session"
const UserKey = "session-user"
const SessionUserIDKey = "userid"

func render(ctx echo.Context, status int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(status)

	err := t.Render(ctx.Request().Context(), ctx.Response().Writer)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "failed to render response template")
	}

	return nil
}

func htmxRedirect(c echo.Context, url string) error {
	c.Response().Header().Set("HX-Redirect", url)
	return c.NoContent(http.StatusOK)
}

func main() {
	err := run()
	if err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	err := godotenv.Load(".env")
	if err != nil {
		logrus.Warn(errors.Wrap(err, "Failed to load .env"))
	}

	tz := os.Getenv("TZ")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return errors.Wrap(err, "failed to load timezone")
		}
		time.Local = loc
	}

	cfg, err := types.ConfigFromEnv()
	if err != nil {
		return errors.Wrap(err, "Loading config from env")
	}

	e := echo.New()

	e.StaticFS("/static", static.FS)

	origErrHandler := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		logrus.Error(err)
		origErrHandler(err, c)
	}

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

	e.Use(middleware.Secure())

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

	err = db.AutoMigrate(&types.User{}, &types.Plan{}, &types.Reading{})
	if err != nil {
		return errors.Wrap(err, "Failed to migrate")
	}

	store := sessions.NewCookieStore(cfg.CookieSecret)
	e.Use(session.Middleware(store))
	e.Use(UserMiddleware(db))

	e.GET("/", dashboardHandler(cfg, db))
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/auth/sign-in", signIn(cfg))
	e.POST("/auth/sign-in", signInWithEmailAndPassword(db, cfg))
	if cfg.AllowSignup {
		e.GET("/auth/sign-up", signUp(cfg))
		e.POST("/auth/sign-up", signUpWithEmailAndPassword(db, cfg))
	}
	e.POST("/auth/sign-out", signOut())

	e.GET("/dashboard", dashboardHandler(cfg, db))
	e.GET("/history", historyHandler(cfg, db))
	e.GET("/plans", plansListHandler(cfg, db))
	e.GET("/plans/create", createPlanForm(cfg, db))
	e.POST("/plans/create", createPlan(db))
	e.POST("/plans/:id/rename", renamePlan(db))
	e.DELETE("/plans/:id", deletePlan(db))
	e.GET("/account", accountHandler(cfg, db))
	e.POST("/account/settings", updateSettings(db))

	e.POST("/reading/:id/complete", completeReading(db))
	e.POST("/reading/:id/uncomplete", uncompleteReading(db))
	e.POST("/reading/:id/update", updateReading(db))

	return e.Start(cfg.Port)
}

func UserMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess, _ := session.Get(SessionKey, c)
			if sess.Values[SessionUserIDKey] != nil {
				userID := sess.Values[SessionUserIDKey].(uint)
				user, err := getUserByID(db, userID)
				if err != nil {
					return errors.Wrap(err, "getting user by id")
				}
				c.Set(UserKey, user)

				sess.Options = &sessions.Options{
					Path:     "/",
					MaxAge:   3600 * 24 * 365,
					HttpOnly: true,
				}

				sess.Values[SessionUserIDKey] = user.ID

				err = sess.Save(c.Request(), c.Response())
				if err != nil {
					return errors.Wrap(err, "saving session")
				}
			}
			return next(c)
		}
	}
}

func GetSessionUser(c echo.Context) (types.User, bool) {
	u := c.Get(UserKey)
	if u != nil {
		user := u.(types.User)
		logrus.Debugf("Found session user %s", user.Email)
		return user, true
	}
	return types.User{}, false
}

func requireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}
		return next(c)
	}
}
