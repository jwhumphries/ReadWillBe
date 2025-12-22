//go:build !wasm

package api

import (
	"net/http"
	"strconv"
	"time"

	"readwillbe/types"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UpdateReadingRequest struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

func CompleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		readingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid reading id")
		}

		var reading types.Reading
		result := db.Preload("Plan").First(&reading, readingID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "reading not found")
			}
			log.Error("failed to fetch reading", "error", result.Error, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching reading").Error())
		}

		if reading.Plan.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "not authorized")
		}

		now := time.Now()
		reading.Status = types.StatusCompleted
		reading.CompletedAt = &now

		if err := db.Save(&reading).Error; err != nil {
			log.Error("failed to complete reading", "error", err, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "completing reading").Error())
		}

		log.Info("reading completed", "reading_id", readingID, "user_id", userID)
		return c.JSON(http.StatusOK, reading)
	}
}

func UncompleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		readingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid reading id")
		}

		var reading types.Reading
		result := db.Preload("Plan").First(&reading, readingID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "reading not found")
			}
			log.Error("failed to fetch reading", "error", result.Error, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching reading").Error())
		}

		if reading.Plan.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "not authorized")
		}

		reading.Status = types.StatusPending
		reading.CompletedAt = nil

		if err := db.Save(&reading).Error; err != nil {
			log.Error("failed to uncomplete reading", "error", err, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "uncompleting reading").Error())
		}

		log.Info("reading uncompleted", "reading_id", readingID, "user_id", userID)
		return c.JSON(http.StatusOK, reading)
	}
}

func UpdateReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		readingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid reading id")
		}

		var reading types.Reading
		result := db.Preload("Plan").First(&reading, readingID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "reading not found")
			}
			log.Error("failed to fetch reading", "error", result.Error, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching reading").Error())
		}

		if reading.Plan.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "not authorized")
		}

		var req UpdateReadingRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.Date != "" {
			date, err := time.Parse("2006-01-02", req.Date)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
			}
			reading.Date = date
		}

		if req.Content != "" {
			reading.Content = req.Content
		}

		if err := db.Save(&reading).Error; err != nil {
			log.Error("failed to update reading", "error", err, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "updating reading").Error())
		}

		log.Info("reading updated", "reading_id", readingID, "user_id", userID)
		return c.JSON(http.StatusOK, reading)
	}
}

func DeleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		readingID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid reading id")
		}

		var reading types.Reading
		result := db.Preload("Plan").First(&reading, readingID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "reading not found")
			}
			log.Error("failed to fetch reading", "error", result.Error, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching reading").Error())
		}

		if reading.Plan.UserID != userID {
			return echo.NewHTTPError(http.StatusForbidden, "not authorized")
		}

		if err := db.Delete(&reading).Error; err != nil {
			log.Error("failed to delete reading", "error", err, "reading_id", readingID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "deleting reading").Error())
		}

		log.Info("reading deleted", "reading_id", readingID, "user_id", userID)
		return c.NoContent(http.StatusNoContent)
	}
}
