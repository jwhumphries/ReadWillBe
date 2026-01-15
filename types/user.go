package types

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name                 string
	Email                string `gorm:"uniqueIndex"`
	Password             string
	Plans                []Plan
	PushSubscriptions    []PushSubscription
	NotificationsEnabled bool
	NotificationTime     string
	CreatedAt            time.Time  `gorm:"autoCreateTime"`
	UpdatedAt            *time.Time `gorm:"autoUpdateTime"`
	DeletedAt            *time.Time

	// Email notifications (in addition to push)
	EmailNotificationsEnabled bool   `gorm:"default:false"`
	NotificationEmail         string // Empty = use user's primary Email
}

func (u User) IsSet() bool {
	return u.Email != ""
}

func (u User) GetNotificationEmail() string {
	if u.NotificationEmail != "" {
		return u.NotificationEmail
	}
	return u.Email
}
