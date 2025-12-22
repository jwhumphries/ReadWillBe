//go:build !wasm

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"readwillbe/types"

	"github.com/labstack/echo/v4"
)

func TestGetDashboard_Empty(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", user.ID)

	handler := GetDashboard(db)
	if err := handler(c); err != nil {
		t.Errorf("GetDashboard failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(response.TodayReadings) != 0 {
		t.Errorf("expected 0 today readings, got %d", len(response.TodayReadings))
	}
	if len(response.OverdueReadings) != 0 {
		t.Errorf("expected 0 overdue readings, got %d", len(response.OverdueReadings))
	}
}

func TestGetDashboard_WithTodayReadings(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	todayReading := types.Reading{
		PlanID:   plan.ID,
		Date:     time.Now(),
		DateType: types.DateTypeDay,
		Content:  "Today's reading",
		Status:   types.StatusPending,
	}
	if err := db.Create(&todayReading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", user.ID)

	handler := GetDashboard(db)
	if err := handler(c); err != nil {
		t.Errorf("GetDashboard failed: %v", err)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(response.TodayReadings) != 1 {
		t.Errorf("expected 1 today reading, got %d", len(response.TodayReadings))
	}
}

func TestGetDashboard_WithOverdueReadings(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	overdueReading := types.Reading{
		PlanID:   plan.ID,
		Date:     time.Now().AddDate(0, 0, -2),
		DateType: types.DateTypeDay,
		Content:  "Overdue reading",
		Status:   types.StatusPending,
	}
	if err := db.Create(&overdueReading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", user.ID)

	handler := GetDashboard(db)
	if err := handler(c); err != nil {
		t.Errorf("GetDashboard failed: %v", err)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(response.OverdueReadings) != 1 {
		t.Errorf("expected 1 overdue reading, got %d", len(response.OverdueReadings))
	}
}

func TestGetDashboard_ExcludesCompletedReadings(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	completedAt := time.Now()
	completedReading := types.Reading{
		PlanID:      plan.ID,
		Date:        time.Now(),
		DateType:    types.DateTypeDay,
		Content:     "Completed reading",
		Status:      types.StatusCompleted,
		CompletedAt: &completedAt,
	}
	if err := db.Create(&completedReading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", user.ID)

	handler := GetDashboard(db)
	if err := handler(c); err != nil {
		t.Errorf("GetDashboard failed: %v", err)
	}

	var response DashboardResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if len(response.TodayReadings) != 0 {
		t.Errorf("expected 0 today readings (completed excluded), got %d", len(response.TodayReadings))
	}
}

func TestGetCurrentUser_Success(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", user.ID)

	handler := GetCurrentUser(db)
	if err := handler(c); err != nil {
		t.Errorf("GetCurrentUser failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response types.User
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", response.Email)
	}
	if response.Password != "" {
		t.Error("password should be empty in response")
	}
}

func TestGetCurrentUser_NotFound(t *testing.T) {
	db := setupTestDB(t)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user_id", uint(9999))

	handler := GetCurrentUser(db)
	err := handler(c)
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, httpErr.Code)
	}
}
