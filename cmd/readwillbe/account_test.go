package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	sqlite "github.com/ncruces/go-sqlite3/gormlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"readwillbe/types"
)

func TestUpdateSettings(t *testing.T) {
	// Setup DB
	// Use unique DB name to prevent collisions
	dbPath := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&types.User{})
	require.NoError(t, err)

	// Create User
	user := createTestUser(t, db, "test@example.com", "password")

	// Setup Cache
	cache := NewUserCache(5*time.Minute, 10*time.Minute)
	cache.Set(*user) // Pre-populate cache to verify it gets removed

	// Setup Echo
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/account/settings", strings.NewReader(
		"notifications_enabled=on&notification_time=08:00",
	))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock Session User (needed for GetSessionUser)
	c.Set(UserKey, *user)

	// Call Handler
	h := updateSettings(db, cache)
	err = h(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/account", rec.Header().Get("Location"))

	// Verify DB Update
	var updatedUser types.User
	err = db.First(&updatedUser, user.ID).Error
	require.NoError(t, err)
	assert.True(t, updatedUser.NotificationsEnabled)
	assert.Equal(t, "08:00", updatedUser.NotificationTime)

	// Verify Cache Invalidation
	_, found := cache.Get(user.ID)
	assert.False(t, found, "Cache should be invalidated after update")
}
