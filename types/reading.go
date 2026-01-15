package types

import (
	"time"

	"gorm.io/gorm"
)

type ReadingStatus string

const (
	StatusPending   ReadingStatus = "pending"
	StatusCompleted ReadingStatus = "completed"
	StatusOverdue   ReadingStatus = "overdue"
)

type DateType string

const (
	DateTypeDay   DateType = "day"
	DateTypeWeek  DateType = "week"
	DateTypeMonth DateType = "month"
)

type Reading struct {
	gorm.Model
	PlanID      uint
	Plan        Plan
	Date        time.Time
	DateType    DateType
	Content     string
	Status      ReadingStatus `gorm:"default:pending"`
	CompletedAt *time.Time
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   *time.Time `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time
}

type PlanGroup struct {
	Plan     Plan
	Readings []Reading
}

func (pg PlanGroup) HasOverdue() bool {
	for _, r := range pg.Readings {
		if r.IsOverdue() {
			return true
		}
	}
	return false
}

func (r Reading) IsOverdue() bool {
	now := time.Now()
	switch r.DateType {
	case DateTypeDay:
		return now.After(r.Date.AddDate(0, 0, 1)) && r.Status == StatusPending
	case DateTypeWeek:
		return now.After(r.Date.AddDate(0, 0, 7)) && r.Status == StatusPending
	case DateTypeMonth:
		return now.After(r.Date.AddDate(0, 1, 0)) && r.Status == StatusPending
	}
	return false
}

func (r Reading) IsActiveToday() bool {
	now := time.Now()
	switch r.DateType {
	case DateTypeDay:
		return r.Date.Year() == now.Year() && r.Date.YearDay() == now.YearDay()
	case DateTypeWeek:
		year, week := r.Date.ISOWeek()
		nowYear, nowWeek := now.ISOWeek()
		return year == nowYear && week == nowWeek
	case DateTypeMonth:
		return r.Date.Year() == now.Year() && r.Date.Month() == now.Month()
	}
	return false
}
