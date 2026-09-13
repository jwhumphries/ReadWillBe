package push

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"readwillbe/internal/model"
)

type recordingEmailService struct {
	digests []string
}

func (r *recordingEmailService) SendDailyDigest(user model.User, _ []model.Reading) error {
	r.digests = append(r.digests, user.GetNotificationEmail())
	return nil
}

func (r *recordingEmailService) SendTestEmail(string) error { return nil }

func setupTestDB(t *testing.T) *gorm.DB {
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", safeName)
	db, err := gorm.Open(gormlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Plan{}, &model.Reading{}, &model.PushSubscription{}))

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// createUserWithReadingDueToday stores user along with a plan holding one
// reading scheduled for today, so the worker has something to notify about.
func createUserWithReadingDueToday(t *testing.T, db *gorm.DB, user *model.User) {
	t.Helper()

	require.NoError(t, db.Create(user).Error)

	plan := &model.Plan{Title: "Plan", UserID: user.ID, Status: "active"}
	require.NoError(t, db.Create(plan).Error)

	reading := &model.Reading{
		PlanID:   plan.ID,
		Content:  "Genesis 1",
		Date:     time.Now(),
		DateType: model.DateTypeDay,
		Status:   model.StatusPending,
	}
	require.NoError(t, db.Create(reading).Error)
}

func atClock(t *testing.T, hhmm string) time.Time {
	t.Helper()
	clock, err := time.ParseInLocation("15:04", hhmm, time.Local)
	require.NoError(t, err)
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), clock.Hour(), clock.Minute(), 0, 0, time.Local)
}

func TestProcessNotifications_SendsEmailAtEmailTime(t *testing.T) {
	db := setupTestDB(t)
	createUserWithReadingDueToday(t, db, &model.User{
		Email:                     "reader@example.com",
		NotificationsEnabled:      true,
		NotificationTime:          "07:00",
		EmailNotificationsEnabled: true,
		EmailNotificationTime:     "18:30",
	})

	emails := &recordingEmailService{}

	processNotifications(model.Config{}, db, emails, false, atClock(t, "07:00"))
	assert.Empty(t, emails.digests, "the push time must not trigger the email digest")

	processNotifications(model.Config{}, db, emails, false, atClock(t, "18:30"))
	assert.Equal(t, []string{"reader@example.com"}, emails.digests)
}

// A user who only uses email never set a push time, and must still be picked up.
func TestProcessNotifications_SendsEmailWithoutPushTime(t *testing.T) {
	db := setupTestDB(t)
	createUserWithReadingDueToday(t, db, &model.User{
		Email:                     "reader@example.com",
		EmailNotificationsEnabled: true,
		EmailNotificationTime:     "06:15",
	})

	emails := &recordingEmailService{}
	processNotifications(model.Config{}, db, emails, false, atClock(t, "06:15"))

	assert.Equal(t, []string{"reader@example.com"}, emails.digests)
}

func TestProcessNotifications_SkipsDisabledEmail(t *testing.T) {
	db := setupTestDB(t)
	createUserWithReadingDueToday(t, db, &model.User{
		Email:                 "reader@example.com",
		EmailNotificationTime: "06:15",
	})

	emails := &recordingEmailService{}
	processNotifications(model.Config{}, db, emails, false, atClock(t, "06:15"))

	assert.Empty(t, emails.digests)
}
