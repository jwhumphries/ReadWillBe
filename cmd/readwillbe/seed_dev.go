//go:build dev

package main

import (
	"readwillbe/types"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func seedDatabase(db *gorm.DB) error {
	var count int64
	db.Model(&types.User{}).Count(&count)
	if count > 0 {
		logrus.Info("Database already has users, skipping seed")
		return nil
	}

	logrus.Info("Seeding database with test user...")

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	if err != nil {
		return errors.Wrap(err, "generating password hash")
	}

	testUser := types.User{
		Name:      "Testy",
		Email:     "testy@testicular.test",
		Password:  string(hash),
		CreatedAt: time.Now(),
	}

	if err := db.Create(&testUser).Error; err != nil {
		return errors.Wrap(err, "creating test user")
	}

	logrus.Infof("✓ Created test user: %s (password: password123)", testUser.Email)
	return nil
}
