//go:build !wasm

package api

import (
	"net/http"

	"readwillbe/types"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func GetHistory(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var readings []types.Reading
		result := db.Preload("Plan").
			Joins("JOIN plans ON plans.id = readings.plan_id").
			Where("plans.user_id = ? AND readings.status = ?", userID, types.StatusCompleted).
			Order("readings.completed_at DESC").
			Limit(50).
			Find(&readings)

		if result.Error != nil {
			log.Error("failed to fetch history", "error", result.Error, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching history").Error())
		}

		return c.JSON(http.StatusOK, readings)
	}
}
