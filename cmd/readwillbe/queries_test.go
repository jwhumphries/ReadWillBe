package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"readwillbe/types"
)

func TestGetDashboardReadings(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "dashboard@example.com", "password")
	plan := createTestPlan(t, db, user, "Test Plan")

	now := time.Now()
	// Create past reading (overdue)
	pastReading := createTestReading(t, db, plan, "Past Reading", now.AddDate(0, 0, -2))
	// Create today's reading (active)
	todayReading := createTestReading(t, db, plan, "Today Reading", now)
	// Create future reading (not active, not overdue) - should be filtered out
	futureReading := createTestReading(t, db, plan, "Future Reading", now.AddDate(0, 0, 5))
	// Create completed reading (should be filtered out)
	completedReading := createTestReading(t, db, plan, "Completed Reading", now.AddDate(0, 0, -1))
	completedReading.Status = types.StatusCompleted
	db.Save(completedReading)

	readings, err := GetDashboardReadings(db, user.ID)
	require.NoError(t, err)

	assert.Len(t, readings, 2)

	ids := make(map[uint]bool)
	for _, r := range readings {
		ids[r.ID] = true
	}
	assert.True(t, ids[pastReading.ID], "Past reading should be included")
	assert.True(t, ids[todayReading.ID], "Today reading should be included")
	assert.False(t, ids[futureReading.ID], "Future reading should be excluded")
	assert.False(t, ids[completedReading.ID], "Completed reading should be excluded")
}

func TestGetActiveReadingsCount(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "active@example.com", "password")
	plan := createTestPlan(t, db, user, "Test Plan")

	now := time.Now()

	// Active today
	createTestReading(t, db, plan, "Active 1", now)

	// Active today but different time
	// Active today but different time
	createTestReading(t, db, plan, "Active 2", now)

	// Past (not active today)
	createTestReading(t, db, plan, "Past", now.AddDate(0, 0, -2))

	// Future (not active today)
	createTestReading(t, db, plan, "Future", now.AddDate(0, 0, 2))

	count, err := GetActiveReadingsCount(db, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestGetActiveReadings(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "active_list@example.com", "password")
	plan := createTestPlan(t, db, user, "Test Plan")

	now := time.Now()
	createTestReading(t, db, plan, "Active 1", now)
	// Active today but different time
	createTestReading(t, db, plan, "Active 2", now)
	createTestReading(t, db, plan, "Past", now.AddDate(0, 0, -2))

	readings, err := GetActiveReadings(db, user.ID, 0)
	require.NoError(t, err)
	assert.Len(t, readings, 2)

	// Test limit
	readingsLimit, err := GetActiveReadings(db, user.ID, 1)
	require.NoError(t, err)
	assert.Len(t, readingsLimit, 1)
}

func TestGetWeeklyCompletedReadingsCount(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "weekly@example.com", "password")
	plan := createTestPlan(t, db, user, "Test Plan")

	now := time.Now()

	// Completed this week
	r1 := createTestReading(t, db, plan, "Completed 1", now)
	r1.Status = types.StatusCompleted
	completedAt1 := now
	r1.CompletedAt = &completedAt1
	db.Save(r1)

	// Completed but last week
	// Use -14 days to be safely outside the current week regardless of day of week
	r2 := createTestReading(t, db, plan, "Completed 2", now.AddDate(0, 0, -14))
	r2.Status = types.StatusCompleted
	completedAt2 := now.AddDate(0, 0, -14)
	r2.CompletedAt = &completedAt2
	db.Save(r2)

	// Pending
	createTestReading(t, db, plan, "Pending", now)

	count, err := GetWeeklyCompletedReadingsCount(db, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestGetMonthlyCompletedReadingsCount(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db, "monthly@example.com", "password")
	plan := createTestPlan(t, db, user, "Test Plan")

	now := time.Now()

	// Completed this month
	r1 := createTestReading(t, db, plan, "Completed 1", now)
	r1.Status = types.StatusCompleted
	completedAt1 := now
	r1.CompletedAt = &completedAt1
	db.Save(r1)

	// Completed but last month
	r2 := createTestReading(t, db, plan, "Completed 2", now.AddDate(0, -2, 0))
	r2.Status = types.StatusCompleted
	completedAt2 := now.AddDate(0, -2, 0)
	r2.CompletedAt = &completedAt2
	db.Save(r2)

	// Pending
	createTestReading(t, db, plan, "Pending", now)

	count, err := GetMonthlyCompletedReadingsCount(db, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestDateHelpers(t *testing.T) {
	// Fixed date: Monday Jan 1 2024
	fixedDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	startDay := getStartOfDay(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), startDay)

	endDay := getEndOfDay(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC), endDay)

	// Jan 1 2024 is Monday. Start of week is Jan 1.
	startWeek := getStartOfWeek(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), startWeek)

	// End of week is Sunday Jan 7.
	endWeek := getEndOfWeek(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC), endWeek)

	startMonth := getStartOfMonth(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), startMonth)

	endMonth := getEndOfMonth(fixedDate)
	assert.Equal(t, time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC), endMonth)
}
