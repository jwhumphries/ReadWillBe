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

func TestCompleteReading_Success(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	reading := types.Reading{
		PlanID:   plan.ID,
		Date:     time.Now(),
		DateType: types.DateTypeDay,
		Content:  "Test reading",
		Status:   types.StatusPending,
	}
	if err := db.Create(&reading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/readings/1/complete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user_id", user.ID)

	handler := CompleteReading(db)
	if err := handler(c); err != nil {
		t.Errorf("CompleteReading failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response types.Reading
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Status != types.StatusCompleted {
		t.Errorf("expected StatusCompleted, got %s", response.Status)
	}
	if response.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestCompleteReading_NotFound(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/readings/9999/complete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("9999")
	c.Set("user_id", user.ID)

	handler := CompleteReading(db)
	err := handler(c)
	if err == nil {
		t.Error("expected error for nonexistent reading")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, httpErr.Code)
	}
}

func TestCompleteReading_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	user1 := createTestUser(t, db, "user1@example.com", "password123")
	user2 := createTestUser(t, db, "user2@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user1.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	reading := types.Reading{
		PlanID:   plan.ID,
		Date:     time.Now(),
		DateType: types.DateTypeDay,
		Content:  "Test reading",
		Status:   types.StatusPending,
	}
	if err := db.Create(&reading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/readings/1/complete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user_id", user2.ID)

	handler := CompleteReading(db)
	err := handler(c)
	if err == nil {
		t.Error("expected error for forbidden access")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, httpErr.Code)
	}
}

func TestUncompleteReading_Success(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	completedAt := time.Now()
	reading := types.Reading{
		PlanID:      plan.ID,
		Date:        time.Now(),
		DateType:    types.DateTypeDay,
		Content:     "Test reading",
		Status:      types.StatusCompleted,
		CompletedAt: &completedAt,
	}
	if err := db.Create(&reading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/readings/1/uncomplete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user_id", user.ID)

	handler := UncompleteReading(db)
	if err := handler(c); err != nil {
		t.Errorf("UncompleteReading failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response types.Reading
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if response.Status != types.StatusPending {
		t.Errorf("expected StatusPending, got %s", response.Status)
	}
	if response.CompletedAt != nil {
		t.Error("expected CompletedAt to be nil")
	}
}

func TestDeleteReading_Success(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	plan := types.Plan{Title: "Test Plan", UserID: user.ID}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	reading := types.Reading{
		PlanID:   plan.ID,
		Date:     time.Now(),
		DateType: types.DateTypeDay,
		Content:  "Test reading",
		Status:   types.StatusPending,
	}
	if err := db.Create(&reading).Error; err != nil {
		t.Fatalf("failed to create reading: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/readings/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	c.Set("user_id", user.ID)

	handler := DeleteReading(db)
	if err := handler(c); err != nil {
		t.Errorf("DeleteReading failed: %v", err)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	var deletedReading types.Reading
	result := db.First(&deletedReading, reading.ID)
	if result.Error == nil {
		t.Error("expected reading to be deleted")
	}
}

func TestDeleteReading_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "test@example.com", "password123")

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/readings/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")
	c.Set("user_id", user.ID)

	handler := DeleteReading(db)
	err := handler(c)
	if err == nil {
		t.Error("expected error for invalid id")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Errorf("expected HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, httpErr.Code)
	}
}
