//go:build dev

package main

import (
	"os"
	"path/filepath"
	"readwillbe/types"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedDatabase(db *gorm.DB) error {
	logrus.Info("Seeding database (dev environment)...")

	// 1. Ensure Test User Exists
	var testUser types.User
	err := db.Where("email = ?", "testy@testicular.test").First(&testUser).Error
	if err == nil {
		logrus.Info("Test user already exists, using existing user")
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Info("Creating test user...")
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}

		testUser = types.User{
			Name:      "Testy",
			Email:     "testy@testicular.test",
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if err := db.Create(&testUser).Error; err != nil {
			return errors.Wrap(err, "creating test user")
		}
		logrus.Infof("✓ Created test user: %s (password: password123)", testUser.Email)
	} else {
		return errors.Wrap(err, "checking for test user")
	}

	// 2. Seed Plans from @test/ directory
	files, err := os.ReadDir("@test")
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Info("@test directory not found, skipping plan seeding")
			return nil
		}
		return errors.Wrap(err, "reading @test directory")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasPrefix(filename, "seed_") || !strings.HasSuffix(filename, ".csv") {
			continue
		}

		// Derive plan title: remove "seed_" prefix and ".csv" suffix
		planTitle := strings.TrimSuffix(strings.TrimPrefix(filename, "seed_"), ".csv")

		// Check if plan already exists
		var existingPlanCount int64
		if err := db.Model(&types.Plan{}).Where("user_id = ? AND title = ?", testUser.ID, planTitle).Count(&existingPlanCount).Error; err != nil {
			return errors.Wrapf(err, "checking existence of plan %s", planTitle)
		}

		if existingPlanCount > 0 {
			logrus.Infof("Plan '%s' already exists, skipping", planTitle)
			continue
		}

		logrus.Infof("Seeding plan '%s' from %s...", planTitle, filename)

		f, err := os.Open(filepath.Join("@test", filename))
		if err != nil {
			return errors.Wrapf(err, "opening file %s", filename)
		}

		readings, err := parseCSV(f)
		f.Close() // Close immediately after reading

		if err != nil {
			return errors.Wrapf(err, "parsing CSV %s", filename)
		}

		plan := types.Plan{
			Title:    planTitle,
			UserID:   testUser.ID,
			Readings: readings,
		}

		if err := db.Create(&plan).Error; err != nil {
			return errors.Wrapf(err, "creating plan %s", planTitle)
		}

		logrus.Infof("✓ Created plan: %s with %d readings", planTitle, len(readings))
	}

	return nil
}
