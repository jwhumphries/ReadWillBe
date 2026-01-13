package main

import (
	"time"

	"gorm.io/gorm"
	"readwillbe/types"
)

func getStartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func getEndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, 999999999, t.Location())
}

func getStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := weekday - 1
	return getStartOfDay(t.AddDate(0, 0, -offset))
}

func getEndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := 7 - weekday
	return getEndOfDay(t.AddDate(0, 0, offset))
}

func getStartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func getEndOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return getEndOfDay(time.Date(y, m+1, 0, 0, 0, 0, 0, t.Location()))
}

// GetDashboardReadings fetches relevant readings for the dashboard.
// It filters out future readings that are not yet active.
func GetDashboardReadings(tx *gorm.DB, userID uint) ([]types.Reading, error) {
	now := time.Now() // Uses time.Local if set
	endDay := getEndOfDay(now)
	endWeek := getEndOfWeek(now)
	endMonth := getEndOfMonth(now)

	var readings []types.Reading
	err := tx.Preload("Plan").
		Where("plan_id IN (?)", tx.Model(&types.Plan{}).Select("id").Where("user_id = ?", userID)).
		Where("status != ?", types.StatusCompleted).
		Where(
			tx.Where("date_type = ? AND date <= ?", types.DateTypeDay, endDay).
				Or("date_type = ? AND date <= ?", types.DateTypeWeek, endWeek).
				Or("date_type = ? AND date <= ?", types.DateTypeMonth, endMonth),
		).
		Find(&readings).Error

	return readings, err
}

// GetActiveReadingsCount fetches the count of readings active today.
func GetActiveReadingsCount(tx *gorm.DB, userID uint) (int64, error) {
	now := time.Now()
	startDay, endDay := getStartOfDay(now), getEndOfDay(now)
	startWeek, endWeek := getStartOfWeek(now), getEndOfWeek(now)
	startMonth, endMonth := getStartOfMonth(now), getEndOfMonth(now)

	var count int64
	err := tx.Model(&types.Reading{}).
		Joins("JOIN plans ON plans.id = readings.plan_id").
		Where("plans.user_id = ? AND readings.status != ?", userID, types.StatusCompleted).
		Where(
			tx.Where("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeDay, startDay, endDay).
				Or("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeWeek, startWeek, endWeek).
				Or("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeMonth, startMonth, endMonth),
		).
		Count(&count).Error

	return count, err
}

// GetActiveReadings fetches readings active today.
func GetActiveReadings(tx *gorm.DB, userID uint, limit int) ([]types.Reading, error) {
	now := time.Now()
	startDay, endDay := getStartOfDay(now), getEndOfDay(now)
	startWeek, endWeek := getStartOfWeek(now), getEndOfWeek(now)
	startMonth, endMonth := getStartOfMonth(now), getEndOfMonth(now)

	var readings []types.Reading
	q := tx.Preload("Plan").
		Joins("JOIN plans ON plans.id = readings.plan_id").
		Where("plans.user_id = ? AND readings.status != ?", userID, types.StatusCompleted).
		Where(
			tx.Where("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeDay, startDay, endDay).
				Or("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeWeek, startWeek, endWeek).
				Or("readings.date_type = ? AND readings.date >= ? AND readings.date <= ?", types.DateTypeMonth, startMonth, endMonth),
		)

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Find(&readings).Error
	return readings, err
}
