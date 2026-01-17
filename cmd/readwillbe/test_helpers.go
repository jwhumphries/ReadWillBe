package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"readwillbe/internal/model"
)

func createTestUser(t *testing.T, db *gorm.DB, email string, password string) *model.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	require.NoError(t, err)

	user := &model.User{
		Name:      "Test User",
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now(),
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	return user
}

func createTestPlan(t *testing.T, db *gorm.DB, user *model.User, title string) *model.Plan {
	plan := &model.Plan{
		Title:  title,
		UserID: user.ID,
		Status: "active",
	}

	err := db.Create(plan).Error
	require.NoError(t, err)

	return plan
}

func createTestReading(t *testing.T, db *gorm.DB, plan *model.Plan, content string, date time.Time) *model.Reading {
	reading := &model.Reading{
		PlanID:   plan.ID,
		Content:  content,
		Date:     date,
		DateType: model.DateTypeDay,
		Status:   model.StatusPending,
	}

	err := db.Create(reading).Error
	require.NoError(t, err)

	return reading
}
