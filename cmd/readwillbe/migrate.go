package main

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"readwillbe/internal/model"
)

// migrateDB brings the schema up to date and runs any one-time data backfills.
func migrateDB(db *gorm.DB) error {
	migrator := db.Migrator()

	// Email used to share the push notification time. When its own column is
	// first added, seed it from that time so existing digests keep arriving.
	backfillEmailTime := migrator.HasTable(&model.User{}) &&
		!migrator.HasColumn(&model.User{}, "EmailNotificationTime")

	if err := db.AutoMigrate(&model.User{}, &model.Plan{}, &model.Reading{}, &model.PushSubscription{}); err != nil {
		return errors.Wrap(err, "auto-migrating schema")
	}

	if backfillEmailTime {
		if err := db.Exec("UPDATE users SET email_notification_time = notification_time").Error; err != nil {
			return errors.Wrap(err, "backfilling email notification time")
		}
	}

	return nil
}
