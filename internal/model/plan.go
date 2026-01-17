package model

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

func (p Plan) IsComplete() bool {
	if len(p.Readings) == 0 {
		return false
	}
	for _, r := range p.Readings {
		if r.Status != StatusCompleted {
			return false
		}
	}
	return true
}
