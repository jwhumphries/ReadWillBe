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

// serverComp is a placeholder component for server-side route matching.
// It ensures that the app handler serves the shell for these routes.
type serverComp struct {
	app.Compo
}

func (c *serverComp) Render() app.UI {
	return app.Div()
}

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

	// Performance: Enable WAL mode for better concurrency
	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		log.Warn("failed to enable WAL mode", "error", err)
	}

	// Performance: Configure connection pool
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
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
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())

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

	e.Group("/static", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
			return next(c)
		}
	}).StaticFS("/", static.FS)

	e.GET("/web/app.wasm", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/wasm")
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
		return c.File("web/app.wasm")
	})
	e.GET("/app.wasm", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/wasm")
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000")
		return c.File("web/app.wasm")
	})

	// Server-side routing for PWA
	// We register the same routes as the client to ensure the server
	// knows they exist and serves the shell instead of 404.
	// We use serverComp as a placeholder since the actual components
	// are not compiled into the server binary.
	app.Route("/", func() app.Composer { return &serverComp{} })
	app.Route("/dashboard", func() app.Composer { return &serverComp{} })
	app.Route("/auth/sign-in", func() app.Composer { return &serverComp{} })
	app.Route("/auth/sign-up", func() app.Composer { return &serverComp{} })
	app.Route("/history", func() app.Composer { return &serverComp{} })
	app.Route("/plans", func() app.Composer { return &serverComp{} })
	app.Route("/plans/create", func() app.Composer { return &serverComp{} })
	app.Route("/plans/{id}/edit", func() app.Composer { return &serverComp{} })
	app.Route("/account", func() app.Composer { return &serverComp{} })

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
