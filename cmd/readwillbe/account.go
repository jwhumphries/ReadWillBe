package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func accountHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.Account(cfg, &user))
	}
}

func updateSettings(db *gorm.DB, cache *UserCache) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		notificationsEnabled := c.FormValue("notifications_enabled") == "on"
		notificationTime := c.FormValue("notification_time")

		if notificationTime != "" && !timeFormatRegex.MatchString(notificationTime) {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid time format: %s (expected HH:MM)", notificationTime))
		}

		user.NotificationsEnabled = notificationsEnabled
		user.NotificationTime = notificationTime

		if err := db.WithContext(c.Request().Context()).Save(&user).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update settings")
		}

		cache.Invalidate(user.ID)

		return c.Redirect(http.StatusFound, "/account")
	}
}
