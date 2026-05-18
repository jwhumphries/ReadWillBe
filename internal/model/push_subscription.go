package model

import "gorm.io/gorm"

// PushSubscription stores the browser-supplied data needed to deliver Web Push
// notifications to a user's device.
type PushSubscription struct {
	gorm.Model
	UserID   uint   `gorm:"index"`
	Endpoint string `gorm:"uniqueIndex"`
	P256DH   string
	Auth     string
}
