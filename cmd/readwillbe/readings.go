package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
)

// MaxContentLength is defined in plans.go

func completeReading(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading model.Reading
		if err := db.WithContext(c.Request().Context()).
			Preload("Plan").
			Joins("JOIN plans ON plans.id = readings.plan_id").
			Where("readings.id = ? AND plans.user_id = ?", id, user.ID).
			First(&reading).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		now := time.Now()
		reading.Status = model.StatusCompleted
		reading.CompletedAt = &now

		if err := db.WithContext(c.Request().Context()).Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return c.Redirect(http.StatusFound, "/dashboard")
	}
}

func uncompleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading model.Reading
		if err := db.WithContext(c.Request().Context()).
			Preload("Plan").
			Joins("JOIN plans ON plans.id = readings.plan_id").
			Where("readings.id = ? AND plans.user_id = ?", id, user.ID).
			First(&reading).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		reading.Status = model.StatusPending
		reading.CompletedAt = nil

		if err := db.WithContext(c.Request().Context()).Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return c.Redirect(http.StatusFound, "/history")
	}
}

func updateReading(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading model.Reading
		if err := db.WithContext(c.Request().Context()).
			Preload("Plan").
			Joins("JOIN plans ON plans.id = readings.plan_id").
			Where("readings.id = ? AND plans.user_id = ?", id, user.ID).
			First(&reading).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		content := c.FormValue("content")
		if content == "" {
			return c.String(http.StatusBadRequest, "Content is required")
		}

		if len(content) > MaxContentLength {
			return c.String(http.StatusBadRequest, "Content exceeds maximum length")
		}

		reading.Content = content

		if err := db.WithContext(c.Request().Context()).Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}
