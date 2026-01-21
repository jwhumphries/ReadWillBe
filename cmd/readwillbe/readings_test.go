package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"readwillbe/internal/model"
)

func TestCompleteReading(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	plan := createTestPlan(t, db, user, "Test Plan")
	reading := createTestReading(t, db, plan, "Genesis 1", time.Now())

	t.Run("successfully complete reading", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/complete", completeReading(db))

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/complete", reading.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		var updated model.Reading
		err := db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StatusCompleted, updated.Status)
		assert.NotNil(t, updated.CompletedAt)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		e.POST("/reading/:id/complete", completeReading(db))

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/complete", reading.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 302, rec.Code)
	})

	t.Run("invalid reading ID", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/complete", completeReading(db))

		req := httptest.NewRequest("POST", "/reading/99999/complete", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 404, rec.Code)
	})
}

func TestUncompleteReading(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	plan := createTestPlan(t, db, user, "Test Plan")
	reading := createTestReading(t, db, plan, "Genesis 1", time.Now())

	completedAt := time.Now()
	reading.Status = model.StatusCompleted
	reading.CompletedAt = &completedAt
	db.Save(&reading)

	t.Run("successfully uncomplete reading", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/uncomplete", uncompleteReading(db))

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/uncomplete", reading.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		var updated model.Reading
		err := db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StatusPending, updated.Status)
		assert.Nil(t, updated.CompletedAt)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		e.POST("/reading/:id/uncomplete", uncompleteReading(db))

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/uncomplete", reading.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 302, rec.Code)
	})
}

func TestUpdateReading(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	plan := createTestPlan(t, db, user, "Test Plan")
	reading := createTestReading(t, db, plan, "Genesis 1", time.Now())

	t.Run("successfully update reading content", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/update", updateReading(db))

		form := url.Values{}
		form.Set("content", "Genesis 1-3")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		var updated model.Reading
		err := db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Genesis 1-3", updated.Content)
	})

	t.Run("update with empty content", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/update", updateReading(db))

		form := url.Values{}
		form.Set("content", "")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 400, rec.Code)
	})

	t.Run("update non-existent reading", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(UserKey, *user)
				return next(c)
			}
		})
		e.POST("/reading/:id/update", updateReading(db))

		form := url.Values{}
		form.Set("content", "New content")

		req := httptest.NewRequest("POST", "/reading/99999/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 404, rec.Code)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		e.POST("/reading/:id/update", updateReading(db))

		form := url.Values{}
		form.Set("content", "New content")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 302, rec.Code)
	})
}
