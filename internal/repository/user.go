package repository

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"readwillbe/internal/model"
)

func GetUserByID(db *gorm.DB, id uint) (model.User, error) {
	var user model.User
	err := db.Preload("Plans").First(&user, "id = ?", id).Error

	return user, errors.Wrap(err, "Finding user")
}

func UserExists(email string, db *gorm.DB) bool {
	var user model.User
	err := db.First(&user, "email = ?", email).Error

	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func GetUserByEmail(db *gorm.DB, email string) (model.User, error) {
	var user model.User
	err := db.First(&user, "email = ?", email).Error

	return user, err
}

func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}
