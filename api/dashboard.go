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

type DashboardResponse struct {
	TodayReadings   []types.Reading `json:"today_readings"`
	OverdueReadings []types.Reading `json:"overdue_readings"`
}

func GetDashboard(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var readings []types.Reading
		result := db.Preload("Plan").
			Joins("JOIN plans ON plans.id = readings.plan_id").
			Where("plans.user_id = ? AND readings.status != ?", userID, types.StatusCompleted).
			Find(&readings)

		if result.Error != nil {
			log.Error("failed to fetch dashboard readings", "error", result.Error, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching readings").Error())
		}

		todayReadings := make([]types.Reading, 0, len(readings))
		overdueReadings := make([]types.Reading, 0, len(readings))

		for _, reading := range readings {
			if reading.IsActiveToday() {
				todayReadings = append(todayReadings, reading)
			} else if reading.IsOverdue() {
				overdueReadings = append(overdueReadings, reading)
			}
		}

		return c.JSON(http.StatusOK, DashboardResponse{
			TodayReadings:   todayReadings,
			OverdueReadings: overdueReadings,
		})
	}
}

func GetCurrentUser(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var user types.User
		if err := db.First(&user, userID).Error; err != nil {
			log.Debug("user not found", "user_id", userID)
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		user.Password = ""
		return c.JSON(http.StatusOK, user)
	}
}
