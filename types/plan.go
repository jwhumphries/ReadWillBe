package types

import (
	"time"

	"gorm.io/gorm"
)

type Plan struct {
	gorm.Model
	Title     string
	UserID    uint
	User      User
	Readings  []Reading
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time
}
