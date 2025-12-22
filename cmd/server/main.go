//go:build !wasm

package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"readwillbe/api"
	"readwillbe/static"
	"readwillbe/types"

	"github.com/charmbracelet/log"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoLog "github.com/labstack/gommon/log"
	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"github.com/pkg/errors"

	_ "github.com/ncruces/go-sqlite3/embed"
	sqlite "github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		log.Fatal("server error", "error", err)
	}
}

func run() error {
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("failed to load .env file", "error", err)
	}

	tz := os.Getenv("TZ")
	if tz != "" {
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return errors.Wrap(err, "loading timezone")
		}
		time.Local = loc
	}

	cfg, err := types.ConfigFromEnv()
	if err != nil {
		return errors.Wrap(err, "loading config from env")
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return errors.Wrap(err, "connecting to database")
	}

	if err := db.AutoMigrate(&types.User{}, &types.Plan{}, &types.Reading{}); err != nil {
		return errors.Wrap(err, "migrating database")
	}

	origErrHandler := e.HTTPErrorHandler
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		log.Error("http error", "error", err, "path", c.Request().URL.Path)
		origErrHandler(err, c)
	}

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		Skipper:           middleware.DefaultSkipper,
		StackSize:         4 << 10,
		DisableStackAll:   false,
		DisablePrintStack: false,
		LogLevel:          echoLog.ERROR,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			log.Error("recovered panic", "error", err)
			for _, l := range strings.Split(string(stack), "\n") {
				log.Error("stack", "line", strings.ReplaceAll(l, "\t", "  "))
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

	store := sessions.NewCookieStore(cfg.CookieSecret)
	e.Use(session.Middleware(store))

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/api/auth/sign-in", api.SignIn(db, cfg))
	e.POST("/api/auth/sign-up", api.SignUp(db, cfg))

	apiGroup := e.Group("/api")
	apiGroup.Use(api.JWTMiddleware(cfg))

	apiGroup.POST("/auth/refresh", api.RefreshToken(db, cfg))
	apiGroup.GET("/user", api.GetCurrentUser(db))
	apiGroup.GET("/dashboard", api.GetDashboard(db))
	apiGroup.GET("/history", api.GetHistory(db))
	apiGroup.GET("/plans", api.GetPlans(db))
	apiGroup.POST("/plans", api.CreatePlan(db))
	apiGroup.GET("/plans/:id", api.GetPlan(db))
	apiGroup.PUT("/plans/:id", api.UpdatePlan(db))
	apiGroup.DELETE("/plans/:id", api.DeletePlan(db))
	apiGroup.POST("/plans/:id/rename", api.RenamePlan(db))
	apiGroup.POST("/readings/:id/complete", api.CompleteReading(db))
	apiGroup.POST("/readings/:id/uncomplete", api.UncompleteReading(db))
	apiGroup.PUT("/readings/:id", api.UpdateReading(db))
	apiGroup.DELETE("/readings/:id", api.DeleteReading(db))
	apiGroup.GET("/account", api.GetAccount(db))
	apiGroup.PUT("/account/settings", api.UpdateSettings(db))

	e.StaticFS("/static", static.FS)

	e.GET("/web/app.wasm", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/wasm")
		return c.File("web/app.wasm")
	})
	e.GET("/app.wasm", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/wasm")
		return c.File("web/app.wasm")
	})

	appHandler := &app.Handler{
		Name:        "ReadWillBe",
		ShortName:   "ReadWillBe",
		Description: "Track your daily reading progress",
		Styles: []string{
			"/static/css/style.min.css",
		},
		ThemeColor:      "#2E3440",
		BackgroundColor: "#ECEFF4",
		Icon: app.Icon{
			Default: "/static/icons/icon-192.png",
			Large:   "/static/icons/icon-512.png",
		},
	}

	e.GET("/*", echo.WrapHandler(appHandler))

	log.Info("starting server", "port", cfg.Port)
	return e.Start(cfg.Port)
}
