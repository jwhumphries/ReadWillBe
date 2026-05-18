package repository

import (
	"readwillbe/internal/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetUserByID returns the user with the given primary key, preloading their plans.
func GetUserByID(db *gorm.DB, id uint) (model.User, error) {
	var user model.User
	err := db.Preload("Plans").First(&user, "id = ?", id).Error

	return user, errors.Wrap(err, "Finding user")
}

// UserExists reports whether a user with the given email is present in db.
func UserExists(email string, db *gorm.DB) bool {
	var user model.User
	err := db.First(&user, "email = ?", email).Error

	return !errors.Is(err, gorm.ErrRecordNotFound)
}

// GetUserByEmail returns the user with the given email address.
func GetUserByEmail(db *gorm.DB, email string) (model.User, error) {
	var user model.User
	err := db.First(&user, "email = ?", email).Error

	return user, err
}

// CreateUser inserts user into db.
func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}
