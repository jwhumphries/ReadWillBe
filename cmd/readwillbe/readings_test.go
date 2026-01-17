package main

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
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
		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/complete", reading.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))
		c.Set(UserKey, *user)

		handler := completeReading(db)
		err := handler(c)
		require.NoError(t, err)

		var updated model.Reading
		err = db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StatusCompleted, updated.Status)
		assert.NotNil(t, updated.CompletedAt)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/complete", reading.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))

		handler := completeReading(db)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 302, rec.Code)
	})

	t.Run("invalid reading ID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("POST", "/reading/99999/complete", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("99999")
		c.Set(UserKey, *user)

		handler := completeReading(db)
		err := handler(c)
		require.NoError(t, err)
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
		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/uncomplete", reading.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))
		c.Set(UserKey, *user)

		handler := uncompleteReading(db)
		err := handler(c)
		require.NoError(t, err)

		var updated model.Reading
		err = db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.StatusPending, updated.Status)
		assert.Nil(t, updated.CompletedAt)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/uncomplete", reading.ID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))

		handler := uncompleteReading(db)
		err := handler(c)
		require.NoError(t, err)
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
		form := url.Values{}
		form.Set("content", "Genesis 1-3")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))
		c.Set(UserKey, *user)

		handler := updateReading(db)
		err := handler(c)
		require.NoError(t, err)

		var updated model.Reading
		err = db.First(&updated, reading.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Genesis 1-3", updated.Content)
	})

	t.Run("update with empty content", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("content", "")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))
		c.Set(UserKey, *user)

		handler := updateReading(db)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 400, rec.Code)
	})

	t.Run("update non-existent reading", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("content", "New content")

		req := httptest.NewRequest("POST", "/reading/99999/update", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("99999")
		c.Set(UserKey, *user)

		handler := updateReading(db)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 404, rec.Code)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		form := url.Values{}
		form.Set("content", "New content")

		req := httptest.NewRequest("POST", fmt.Sprintf("/reading/%d/update", reading.ID), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(fmt.Sprintf("%d", reading.ID))

		handler := updateReading(db)
		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, 302, rec.Code)
	})
}
