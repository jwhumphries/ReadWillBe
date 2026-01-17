package model

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

// FormattedDate returns a human-readable date string based on the reading's DateType.
// - Day: "Jan 2, 2006"
// - Week: "Jan 2-8, 2006" (or "Dec 30 - Jan 5" for month-spanning weeks)
// - Month: "Jan 2006"
func (r Reading) FormattedDate() string {
	switch r.DateType {
	case DateTypeWeek:
		return formatWeekRange(r.Date)
	case DateTypeMonth:
		return r.Date.Format("Jan 2006")
	case DateTypeDay:
		fallthrough
	default:
		return r.Date.Format("Jan 2, 2006")
	}
}

// formatWeekRange formats a week as a date range, e.g., "Jan 2-8, 2006"
// Handles weeks that span months: "Dec 30 - Jan 5"
func formatWeekRange(weekStart time.Time) string {
	weekEnd := weekStart.AddDate(0, 0, 6)

	if weekStart.Month() == weekEnd.Month() {
		// Same month: "Jan 2-8, 2006"
		return weekStart.Format("Jan 2") + "-" + weekEnd.Format("2, 2006")
	}

	// Different months: "Dec 30 - Jan 5"
	if weekStart.Year() == weekEnd.Year() {
		return weekStart.Format("Jan 2") + " - " + weekEnd.Format("Jan 2, 2006")
	}

	// Different years: "Dec 30, 2025 - Jan 5, 2026"
	return weekStart.Format("Jan 2, 2006") + " - " + weekEnd.Format("Jan 2, 2006")
}
