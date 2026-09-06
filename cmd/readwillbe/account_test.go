package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"readwillbe/internal/model"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// postSettings submits form values to updateSettings as the given user and
// returns the reloaded user record.
func postSettings(t *testing.T, db *gorm.DB, user *model.User, form url.Values) (*httptest.ResponseRecorder, model.User) {
	t.Helper()

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(UserKey, *user)
			return next(c)
		}
	})
	e.POST("/account/settings", updateSettings(db))

	req := httptest.NewRequest("POST", "/account/settings", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var reloaded model.User
	require.NoError(t, db.First(&reloaded, user.ID).Error)

	return rec, reloaded
}

// The push and email settings live in two separate forms on /account. Each
// form submits only its own fields, so the handler must leave the other
// group's fields alone rather than reading them as absent-and-therefore-off.
func TestUpdateSettings_PushFormLeavesEmailSettingsIntact(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	user.EmailNotificationsEnabled = true
	user.NotificationEmail = "digest@example.com"
	user.NotificationTime = "07:30"
	require.NoError(t, db.Save(user).Error)

	rec, updated := postSettings(t, db, user, url.Values{
		"section":               {"push"},
		"notifications_enabled": {"on"},
		"notification_time":     {"08:00"},
	})

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.True(t, updated.NotificationsEnabled)
	assert.Equal(t, "08:00", updated.NotificationTime)
	assert.True(t, updated.EmailNotificationsEnabled, "push form must not disable email notifications")
	assert.Equal(t, "digest@example.com", updated.NotificationEmail, "push form must not clear the notification email")
}

func TestUpdateSettings_EmailFormLeavesPushSettingsIntact(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	user.NotificationsEnabled = true
	user.NotificationTime = "07:30"
	require.NoError(t, db.Save(user).Error)

	rec, updated := postSettings(t, db, user, url.Values{
		"section":                     {"email"},
		"email_notifications_enabled": {"on"},
		"notification_email":          {"digest@example.com"},
	})

	assert.Equal(t, http.StatusFound, rec.Code)
	assert.True(t, updated.EmailNotificationsEnabled)
	assert.Equal(t, "digest@example.com", updated.NotificationEmail)
	assert.True(t, updated.NotificationsEnabled, "email form must not disable push notifications")
	assert.Equal(t, "07:30", updated.NotificationTime, "email form must not clear the notification time")
}

func TestUpdateSettings_EmailFormCanDisableEmailNotifications(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	user.EmailNotificationsEnabled = true
	user.NotificationEmail = "digest@example.com"
	require.NoError(t, db.Save(user).Error)

	_, updated := postSettings(t, db, user, url.Values{
		"section": {"email"},
	})

	assert.False(t, updated.EmailNotificationsEnabled)
	assert.Empty(t, updated.NotificationEmail)
}

func TestUpdateSettings_PushFormCanDisablePushNotifications(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")
	user.NotificationsEnabled = true
	user.NotificationTime = "07:30"
	require.NoError(t, db.Save(user).Error)

	_, updated := postSettings(t, db, user, url.Values{
		"section":           {"push"},
		"notification_time": {"07:30"},
	})

	assert.False(t, updated.NotificationsEnabled)
}

// A page cached before the section marker existed posts every field at once;
// it should keep updating every field rather than silently doing nothing.
func TestUpdateSettings_MissingSectionUpdatesEveryGroup(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")

	_, updated := postSettings(t, db, user, url.Values{
		"notifications_enabled":       {"on"},
		"notification_time":           {"09:15"},
		"email_notifications_enabled": {"on"},
		"notification_email":          {"digest@example.com"},
	})

	assert.True(t, updated.NotificationsEnabled)
	assert.Equal(t, "09:15", updated.NotificationTime)
	assert.True(t, updated.EmailNotificationsEnabled)
	assert.Equal(t, "digest@example.com", updated.NotificationEmail)
}

func TestUpdateSettings_RejectsInvalidNotificationTime(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")

	rec, updated := postSettings(t, db, user, url.Values{
		"section":           {"push"},
		"notification_time": {"25:00"},
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, updated.NotificationTime)
}

func TestUpdateSettings_RejectsInvalidNotificationEmail(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "reader@example.com", "password123")

	rec, updated := postSettings(t, db, user, url.Values{
		"section":                     {"email"},
		"email_notifications_enabled": {"on"},
		"notification_email":          {"not-an-email"},
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.False(t, updated.EmailNotificationsEnabled)
	assert.Empty(t, updated.NotificationEmail)
}
