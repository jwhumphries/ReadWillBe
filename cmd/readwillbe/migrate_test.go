package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"readwillbe/internal/model"
)

// legacyUser is the users table as it stood before email had its own time.
type legacyUser struct {
	ID                        uint `gorm:"primaryKey"`
	Email                     string
	NotificationTime          string
	EmailNotificationsEnabled bool
}

func (legacyUser) TableName() string { return "users" }

func openEmptyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// Existing users were emailed at their single notification time; adding a
// separate email time must not silently stop their digests.
func TestMigrateDB_BackfillsEmailTimeFromNotificationTime(t *testing.T) {
	db := openEmptyTestDB(t)
	require.NoError(t, db.AutoMigrate(&legacyUser{}))
	require.NoError(t, db.Create(&legacyUser{Email: "reader@example.com", NotificationTime: "07:30", EmailNotificationsEnabled: true}).Error)

	require.NoError(t, migrateDB(db))

	var user model.User
	require.NoError(t, db.First(&user, "email = ?", "reader@example.com").Error)
	assert.Equal(t, "07:30", user.EmailNotificationTime)
}

// The backfill runs only when the column is first added; later startups must
// leave a user's chosen email time alone.
func TestMigrateDB_DoesNotOverwriteEmailTimeOnLaterRuns(t *testing.T) {
	db := openEmptyTestDB(t)
	require.NoError(t, migrateDB(db))

	user := &model.User{Email: "reader@example.com", NotificationTime: "07:30", EmailNotificationTime: "18:00"}
	require.NoError(t, db.Create(user).Error)

	require.NoError(t, migrateDB(db))

	var reloaded model.User
	require.NoError(t, db.First(&reloaded, user.ID).Error)
	assert.Equal(t, "18:00", reloaded.EmailNotificationTime)
}
