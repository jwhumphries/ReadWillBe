package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readwillbe/types"
)

func createTestUser(t *testing.T, db *gorm.DB, email string, password string) *types.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	require.NoError(t, err)

	user := &types.User{
		Name:      "Test User",
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	return user
}

func createTestPlan(t *testing.T, db *gorm.DB, user *types.User, title string) *types.Plan {
	plan := &types.Plan{
		Title:  title,
		UserID: user.ID,
		Status: "active",
	}

	err := db.Create(plan).Error
	require.NoError(t, err)

	return plan
}

func createTestReading(t *testing.T, db *gorm.DB, plan *types.Plan, content string, date time.Time) *types.Reading {
	reading := &types.Reading{
		PlanID:   plan.ID,
		Content:  content,
		Date:     date,
		DateType: types.DateTypeDay,
		Status:   types.StatusPending,
	}

	err := db.Create(reading).Error
	require.NoError(t, err)

	return reading
}
