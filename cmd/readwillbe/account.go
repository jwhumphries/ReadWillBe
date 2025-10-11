package main

import (
	"net/http"

	"readwillbe/types"
	"readwillbe/views"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func accountHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.Account(cfg, &user))
	}
}

func updateSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		notificationsEnabled := c.FormValue("notifications_enabled") == "on"
		notificationTime := c.FormValue("notification_time")

		user.NotificationsEnabled = notificationsEnabled
		user.NotificationTime = notificationTime

		if err := db.Save(&user).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update settings")
		}

		return c.Redirect(http.StatusFound, "/account")
	}
}
