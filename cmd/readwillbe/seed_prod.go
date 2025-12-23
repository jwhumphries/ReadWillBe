//go:build !dev

package main

import (
	"gorm.io/gorm"
)

func seedDatabase(db *gorm.DB) error {
	return nil
}
