package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	emailservice "readwillbe/internal/service/email"
	"readwillbe/internal/views"
)

var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func accountHandler(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.Account(cfg, &user))
	}
}

func updateSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
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

		// Email notification settings
		user.EmailNotificationsEnabled = c.FormValue("email_notifications_enabled") == "on"
		notificationEmail := strings.TrimSpace(c.FormValue("notification_email"))
		if notificationEmail != "" && !isValidEmail(notificationEmail) {
			return c.String(http.StatusBadRequest, "Invalid email address")
		}
		user.NotificationEmail = notificationEmail

		if err := db.WithContext(c.Request().Context()).Save(&user).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update settings")
		}

		return c.Redirect(http.StatusFound, "/account")
	}
}

func sendTestEmailHandler(cfg model.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !cfg.EmailEnabled() {
			return c.String(http.StatusServiceUnavailable, "Email not configured")
		}

		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		to := strings.TrimSpace(c.FormValue("email"))
		if to == "" {
			to = user.Email
		}
		if !isValidEmail(to) {
			return c.String(http.StatusBadRequest, "Invalid email address")
		}

		emailService := emailservice.NewService(cfg)
		if err := emailService.SendTestEmail(to, cfg.Hostname); err != nil {
			logrus.Errorf("Failed to send test email: %v", err)
			return c.String(http.StatusInternalServerError, "Failed to send test email: "+err.Error())
		}

		return c.String(http.StatusOK, "Test email sent successfully!")
	}
}
