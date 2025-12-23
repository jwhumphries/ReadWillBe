//go:build !dev

package main

import (
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

func seedDatabase(db *gorm.DB, fs afero.Fs) error {
	return nil
}
