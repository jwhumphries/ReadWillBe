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

type UpdateSettingsRequest struct {
	NotificationsEnabled bool   `json:"notifications_enabled"`
	NotificationTime     string `json:"notification_time"`
}

func GetAccount(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var user types.User
		if err := db.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "user not found")
			}
			log.Error("failed to fetch user", "error", err, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "fetching user").Error())
		}

		user.Password = ""
		return c.JSON(http.StatusOK, user)
	}
}

func UpdateSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var user types.User
		if err := db.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "user not found")
			}
			log.Error("failed to fetch user for settings update", "error", err, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "fetching user").Error())
		}

		var req UpdateSettingsRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		user.NotificationsEnabled = req.NotificationsEnabled
		user.NotificationTime = req.NotificationTime

		if err := db.Save(&user).Error; err != nil {
			log.Error("failed to update settings", "error", err, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "updating settings").Error())
		}

		user.Password = ""
		log.Info("settings updated", "user_id", userID)
		return c.JSON(http.StatusOK, user)
	}
}
