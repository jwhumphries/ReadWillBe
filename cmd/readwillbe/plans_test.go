package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"readwillbe/types"
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

	err = db.AutoMigrate(&types.User{}, &types.Plan{}, &types.Reading{}, &types.PushSubscription{})
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func TestCreatePlan_BackgroundProcessing(t *testing.T) {
	db := setupTestDB(t)

	// Create test user
	user := types.User{
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
	c.Set(UserKey, user)

	// Invoke handler
	h := createPlan(db)
	err := h(c)
	assert.NoError(t, err)

	// Verify redirect
	assert.Equal(t, http.StatusOK, rec.Code) // HTMX redirect returns 200 with HX-Redirect header
	assert.Equal(t, "/plans", rec.Header().Get("HX-Redirect"))

	// Verify plan status is initially "processing" or eventually "active"
	// Since it's a goroutine, it might be fast enough to be active immediately or still processing.
	// We'll poll for "active" status.

	var plan types.Plan
	assert.Eventually(t, func() bool {
		db.Preload("Readings").First(&plan, "title = ?", "Test Plan")
		return plan.Status == "active"
	}, 2*time.Second, 100*time.Millisecond, "Plan should eventually be active")

	assert.Equal(t, "Test Plan", plan.Title)
	assert.Len(t, plan.Readings, 2)
}

func TestCreatePlan_BackgroundProcessingFailure(t *testing.T) {
	db := setupTestDB(t)

	user := types.User{
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
	c.Set(UserKey, user)

	// Invoke handler
	h := createPlan(db)
	err := h(c)
	assert.NoError(t, err)

	// Verify plan status eventually "failed"
	var plan types.Plan
	assert.Eventually(t, func() bool {
		db.First(&plan, "title = ?", "Failed Plan")
		return plan.Status == "failed"
	}, 2*time.Second, 100*time.Millisecond, "Plan should eventually fail")

	assert.Contains(t, plan.ErrorMessage, "reading CSV")
}

func TestUserCache(t *testing.T) {
	cache := NewUserCache(100 * time.Millisecond)
	user := types.User{
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
