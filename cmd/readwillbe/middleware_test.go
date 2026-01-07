package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stretchr/testify/assert"
)

func TestGzipConfiguration(t *testing.T) {
	e := echo.New()

	gzipConfig := middleware.GzipConfig{
		Level:     5,
		MinLength: 1400,
		Skipper:   middleware.DefaultSkipper,
	}
	e.Use(middleware.GzipWithConfig(gzipConfig))

	e.GET("/short", func(c echo.Context) error {
		return c.String(http.StatusOK, "short response")
	})

	e.GET("/long", func(c echo.Context) error {
		return c.String(http.StatusOK, strings.Repeat("a", 1500))
	})

	t.Run("Short response should not be compressed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/short", nil)
		req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "short response", rec.Body.String())
		assert.Empty(t, rec.Header().Get(echo.HeaderContentEncoding))
	})

	t.Run("Long response should be compressed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/long", nil)
		req.Header.Set(echo.HeaderAcceptEncoding, "gzip")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get(echo.HeaderContentEncoding))
		assert.True(t, rec.Body.Len() > 0)
		assert.True(t, rec.Body.Len() < 1500)
	})
}

func TestSecurityMiddlewares(t *testing.T) {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.BodyLimit("10M"))

	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.POST("/upload", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	t.Run("RequestID middleware adds header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))
	})

	t.Run("BodyLimit middleware allows small payloads", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader("small"))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
