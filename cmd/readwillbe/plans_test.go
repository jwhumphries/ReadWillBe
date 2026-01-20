package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"readwillbe/internal/cache"
	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Use a unique name for each test to ensure isolation and prevent flaky tests
	// caused by shared in-memory database state.
	// We sanitize the test name to be safe for URI usage.
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(&model.User{}, &model.Plan{}, &model.Reading{}, &model.PushSubscription{})
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func TestCreatePlan_BackgroundProcessing(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := model.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	db.Create(&user)

	// Setup Echo context
	e := echo.New()

	// Create CSV content
	csvContent := `date,reading
2025-01-01,Genesis 1
2025-01-02,Genesis 2`

	body := new(strings.Builder)
	writer := multipart.NewWriter(body)

	// Add title
	_ = writer.WriteField("title", "Test Plan")

	// Add CSV file
	part, _ := writer.CreateFormFile("csv", "readings.csv")
	_, _ = part.Write([]byte(csvContent))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/plans/create", strings.NewReader(body.String()))
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(mw.UserKey, user)

	// Invoke handler
	h := createPlan(afero.NewMemMapFs(), db)
	err := h(c)
	assert.NoError(t, err)

	// Verify redirect
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/plans", rec.Header().Get("Location"))

	// Verify plan status is initially "processing" or eventually "active"
	// Since it's a goroutine, it might be fast enough to be active immediately or still processing.
	// We'll poll for "active" status.

	var plan model.Plan
	assert.Eventually(t, func() bool {
		db.Preload("Readings").First(&plan, "title = ?", "Test Plan")
		return plan.Status == "active"
	}, 2*time.Second, 100*time.Millisecond, "Plan should eventually be active")

	assert.Equal(t, "Test Plan", plan.Title)
	assert.Len(t, plan.Readings, 2)
}

func TestCreatePlan_BackgroundProcessingFailure(t *testing.T) {
	db := setupTestDB(t)

	user := model.User{
		Email: "test@example.com",
		Name:  "Test User",
	}
	db.Create(&user)

	// Setup Echo context
	e := echo.New()

	// Create Invalid CSV content (missing columns)
	csvContent := `date,reading
invalid-row`

	body := new(strings.Builder)
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", "Failed Plan")
	part, _ := writer.CreateFormFile("csv", "readings.csv")
	_, _ = part.Write([]byte(csvContent))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/plans/create", strings.NewReader(body.String()))
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(mw.UserKey, user)

	// Invoke handler
	h := createPlan(afero.NewMemMapFs(), db)
	err := h(c)
	assert.NoError(t, err)

	// Verify plan status eventually "failed"
	var plan model.Plan
	assert.Eventually(t, func() bool {
		db.First(&plan, "title = ?", "Failed Plan")
		return plan.Status == "failed"
	}, 2*time.Second, 100*time.Millisecond, "Plan should eventually fail")

	assert.Contains(t, plan.ErrorMessage, "reading CSV")
}

func TestUserCache(t *testing.T) {
	cache := cache.NewUserCache(100*time.Millisecond, 200*time.Millisecond)
	user := model.User{
		Model: gorm.Model{ID: 1},
		Email: "cache@example.com",
	}

	// Set
	cache.Set(user)

	// Get Hit
	got, found := cache.Get(1)
	assert.True(t, found)
	assert.Equal(t, user.Email, got.Email)

	// Get Miss (wrong ID)
	_, found = cache.Get(2)
	assert.False(t, found)

	// Expiration
	time.Sleep(150 * time.Millisecond)
	_, found = cache.Get(1)
	assert.False(t, found, "Should have expired")
}

func TestDeletePlan(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")
	plan := createTestPlan(t, db, user, "Plan to Delete")

	t.Run("successfully delete plan", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.DELETE("/plans/:id", deletePlan(db))

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/plans/%d", plan.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		var deletedPlan model.Plan
		err := db.First(&deletedPlan, plan.ID).Error
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("delete non-existent plan", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.DELETE("/plans/:id", deletePlan(db))

		req := httptest.NewRequest("DELETE", "/plans/99999", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 404, rec.Code)
	})

	t.Run("unauthenticated request", func(t *testing.T) {
		e := echo.New()
		e.DELETE("/plans/:id", deletePlan(db))

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/plans/%d", plan.ID), nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 302, rec.Code)
	})
}

func TestRenamePlan(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")
	plan := createTestPlan(t, db, user, "Original Title")

	t.Run("successfully rename plan", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.POST("/plans/:id/rename", renamePlan(db))

		form := fmt.Sprintf("title=%s", "New Title")
		req := httptest.NewRequest("POST", fmt.Sprintf("/plans/%d/rename", plan.ID), strings.NewReader(form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		var updated model.Plan
		db.First(&updated, plan.ID)
		assert.Equal(t, "New Title", updated.Title)
	})

	t.Run("rename with empty title", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.POST("/plans/:id/rename", renamePlan(db))

		form := "title="
		req := httptest.NewRequest("POST", fmt.Sprintf("/plans/%d/rename", plan.ID), strings.NewReader(form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 400, rec.Code)
	})

	t.Run("rename non-existent plan", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.POST("/plans/:id/rename", renamePlan(db))

		form := "title=New Title"
		req := httptest.NewRequest("POST", "/plans/99999/rename", strings.NewReader(form))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 404, rec.Code)
	})
}

func TestEditPlan(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test_edit@example.com", "password123")
	plan := createTestPlan(t, db, user, "Original Title")

	// Create initial readings with DateType
	readings := []model.Reading{
		{PlanID: plan.ID, Date: timeMustParse("2025-01-01"), DateType: model.DateTypeDay, Content: "Reading 1", Status: model.StatusPending},
		{PlanID: plan.ID, Date: timeMustParse("2025-01-02"), DateType: model.DateTypeDay, Content: "Reading 2", Status: model.StatusPending},
	}
	db.Create(&readings)

	t.Run("successfully update plan", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.POST("/plans/:id/edit", editPlan(model.Config{}, db))

		// Prepare JSON payload for readings
		// 1. Update Reading 1
		// 2. Remove Reading 2
		// 3. Add Reading 3

		readingsJSON := fmt.Sprintf(`[
			{"id": "%d", "date": "2025-01-01", "content": "Updated Reading 1"},
			{"id": "new-uuid", "date": "2025-01-03", "content": "Reading 3"}
		]`, readings[0].ID)

		form := make(url.Values)
		form.Set("title", "Updated Title")
		form.Set("readingsJSON", readingsJSON)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/plans/%d/edit", plan.ID), strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/plans", rec.Header().Get("Location"))

		// Verify updates in DB
		var updatedPlan model.Plan
		db.Preload("Readings").First(&updatedPlan, plan.ID)

		assert.Equal(t, "Updated Title", updatedPlan.Title)
		assert.Len(t, updatedPlan.Readings, 2)

		// Verify readings
		foundUpdated := false
		foundNew := false

		for _, r := range updatedPlan.Readings {
			if r.ID == readings[0].ID {
				assert.Equal(t, "Updated Reading 1", r.Content)
				assert.Equal(t, model.DateTypeDay, r.DateType, "DateType should be preserved on update")
				foundUpdated = true
			} else if r.Content == "Reading 3" {
				assert.Equal(t, "2025-01-03", r.Date.Format("2006-01-02"))
				assert.Equal(t, model.DateTypeDay, r.DateType, "DateType should be set on new reading")
				foundNew = true
			} else if r.ID == readings[1].ID {
				assert.Fail(t, "Reading 2 should have been deleted")
			}
		}

		assert.True(t, foundUpdated, "Existing reading was not updated")
		assert.True(t, foundNew, "New reading was not created")
	})

	t.Run("invalid readings JSON", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(mw.UserKey, *user)
				return next(c)
			}
		})
		e.POST("/plans/:id/edit", editPlan(model.Config{}, db))

		form := make(url.Values)
		form.Set("title", "Updated Title")
		form.Set("readingsJSON", "invalid-json")

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/plans/%d/edit", plan.ID), strings.NewReader(form.Encode()))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 422, rec.Code)
	})
}

// Helper to parse time for tests
func timeMustParse(value string) time.Time {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return t
}
