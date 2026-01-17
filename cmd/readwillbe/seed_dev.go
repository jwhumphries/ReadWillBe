//go:build dev

package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"readwillbe/internal/model"
	csvservice "readwillbe/internal/service/csv"
)

func seedDatabase(db *gorm.DB, fs afero.Fs) error {
	logrus.Info("Seeding database (dev environment)...")

	// 1. Ensure Test User Exists
	var testUser model.User
	var users []model.User
	err := db.Where("email = ?", "testy@testicular.test").Limit(1).Find(&users).Error
	if err != nil {
		return errors.Wrap(err, "checking for test user")
	}

	if len(users) > 0 {
		testUser = users[0]
		logrus.Info("Test user already exists, using existing user")
	} else {
		logrus.Info("Creating test user...")
		hash, err := bcrypt.GenerateFromPassword([]byte("password123"), BcryptCost)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}

		testUser = model.User{
			Name:      "Testy",
			Email:     "testy@testicular.test",
			Password:  string(hash),
			CreatedAt: time.Now(),
		}

		if err := db.Create(&testUser).Error; err != nil {
			return errors.Wrap(err, "creating test user")
		}
		logrus.Infof("✓ Created test user: %s (password: password123)", testUser.Email)
	}

	// 2. Seed Plans from test/ directory
	files, err := afero.ReadDir(fs, "test")
	if err != nil {
		if _, statErr := fs.Stat("test"); statErr != nil {
			logrus.Info("test directory not found, skipping plan seeding")
			return nil
		}
		return errors.Wrap(err, "reading test directory")
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
		if err := db.Model(&model.Plan{}).Where("user_id = ? AND title = ?", testUser.ID, planTitle).Count(&existingPlanCount).Error; err != nil {
			return errors.Wrapf(err, "checking existence of plan %s", planTitle)
		}

		if existingPlanCount > 0 {
			logrus.Infof("Plan '%s' already exists, skipping", planTitle)
			continue
		}

		logrus.Infof("Seeding plan '%s' from %s...", planTitle, filename)

		f, err := fs.Open(filepath.Join("test", filename))
		if err != nil {
			return errors.Wrapf(err, "opening file %s", filename)
		}

		readings, err := csvservice.ParseCSV(f)
		f.Close() // Close immediately after reading

		if err != nil {
			return errors.Wrapf(err, "parsing CSV %s", filename)
		}

		plan := model.Plan{
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
