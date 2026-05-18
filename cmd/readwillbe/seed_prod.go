//go:build !dev

package main

import (
	"github.com/spf13/afero"
	"gorm.io/gorm"
)

func seedDatabase(_ *gorm.DB, _ afero.Fs) error {
	return nil
}
