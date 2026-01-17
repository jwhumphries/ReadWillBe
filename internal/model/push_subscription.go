package model

import "gorm.io/gorm"

type PushSubscription struct {
	gorm.Model
	UserID   uint   `gorm:"index"`
	Endpoint string `gorm:"uniqueIndex"`
	P256DH   string
	Auth     string
}
