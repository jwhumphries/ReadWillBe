package model

import (
	"time"

	"gorm.io/gorm"
)

// User is an account that owns plans and receives notifications.
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

// IsSet reports whether the user has a non-empty email address, used as a
// proxy for "logged in" in template code.
func (u User) IsSet() bool {
	return u.Email != ""
}

// GetNotificationEmail returns the dedicated notification email if set,
// falling back to the primary user email.
func (u User) GetNotificationEmail() string {
	if u.NotificationEmail != "" {
		return u.NotificationEmail
	}
	return u.Email
}
