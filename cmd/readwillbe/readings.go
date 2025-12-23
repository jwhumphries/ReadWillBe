package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
)

func completeReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading types.Reading
		if err := db.Preload("Plan").First(&reading, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		if reading.Plan.UserID != user.ID {
			return c.String(http.StatusForbidden, "Forbidden")
		}

		now := time.Now()
		reading.Status = types.StatusCompleted
		reading.CompletedAt = &now

		if err := db.Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return htmxRedirect(c, "/dashboard")
	}
}

func uncompleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading types.Reading
		if err := db.Preload("Plan").First(&reading, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		if reading.Plan.UserID != user.ID {
			return c.String(http.StatusForbidden, "Forbidden")
		}

		reading.Status = types.StatusPending
		reading.CompletedAt = nil

		if err := db.Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return htmxRedirect(c, "/history")
	}
}

func updateReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var reading types.Reading
		if err := db.Preload("Plan").First(&reading, id).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		if reading.Plan.UserID != user.ID {
			return c.String(http.StatusForbidden, "Forbidden")
		}

		content := c.FormValue("content")
		if content == "" {
			return c.String(http.StatusBadRequest, "Content is required")
		}

		reading.Content = content

		if err := db.Save(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update reading")
		}

		return htmxRedirect(c, "/plans")
	}
}
