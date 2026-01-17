package types

import (
	"time"

	"gorm.io/gorm"
)

type Plan struct {
	gorm.Model
	Title        string
	UserID       uint
	User         User
	Readings     []Reading
	Status       string `gorm:"default:'active'"`
	ErrorMessage string
	CreatedAt    time.Time  `gorm:"autoCreateTime"`
	UpdatedAt    *time.Time `gorm:"autoUpdateTime"`
	DeletedAt    *time.Time
}
