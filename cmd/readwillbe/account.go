package main

import (
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	emailservice "readwillbe/internal/service/email"
	"readwillbe/internal/views"
)

var timeFormatRegex = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

// Settings sections. Each form on /account posts the section it owns so that
// updateSettings only touches that group's fields.
const (
	settingsSectionPush  = "push"
	settingsSectionEmail = "email"
)

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func accountHandler(cfg model.Config, _ *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.Account(cfg, &user))
	}
}

func updateSettings(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		// The push and email settings are two separate forms on /account, and a
		// form submits nothing for an unchecked box. Without the section marker
		// the handler cannot tell "left this box off" from "this form does not
		// own that box", and saving one card would silently clear the other.
		// An absent marker updates every group, so a page cached before the
		// marker existed keeps working.
		section := c.FormValue("section")
		updatePush := section == "" || section == settingsSectionPush
		updateEmail := section == "" || section == settingsSectionEmail

		// Only the submitted section's columns are written. The session user is
		// served from a TTL cache, so persisting the whole struct would revert
		// any column another request changed while that snapshot was cached.
		updates := map[string]any{}

		if updatePush {
			notificationTime := c.FormValue("notification_time")
			if notificationTime != "" && !timeFormatRegex.MatchString(notificationTime) {
				return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid time format: %s (expected HH:MM)", notificationTime))
			}

			updates["notifications_enabled"] = c.FormValue("notifications_enabled") == "on"
			updates["notification_time"] = notificationTime
		}

		if updateEmail {
			notificationEmail := strings.TrimSpace(c.FormValue("notification_email"))
			if notificationEmail != "" && !isValidEmail(notificationEmail) {
				return c.String(http.StatusBadRequest, "Invalid email address")
			}

			updates["email_notifications_enabled"] = c.FormValue("email_notifications_enabled") == "on"
			updates["notification_email"] = notificationEmail
		}

		if len(updates) > 0 {
			if err := db.WithContext(c.Request().Context()).Model(&user).Updates(updates).Error; err != nil {
				return c.String(http.StatusInternalServerError, "Failed to update settings")
			}
		}

		return c.Redirect(http.StatusFound, "/account")
	}
}

func sendTestEmailHandler(cfg model.Config) echo.HandlerFunc {
	return func(c *echo.Context) error {
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

		emailService, err := emailservice.NewService(cfg)
		if err != nil {
			logrus.Errorf("Failed to build email service: %v", err)
			return c.String(http.StatusServiceUnavailable, "Email not configured")
		}

		// The provider error can name the mail host and account, so it is
		// logged rather than returned to the browser.
		if err := emailService.SendTestEmail(to); err != nil {
			logrus.Errorf("Failed to send test email: %v", err)
			return c.String(http.StatusInternalServerError, "Failed to send test email. Check the server logs for details.")
		}

		return c.String(http.StatusOK, "Test email sent successfully!")
	}
}
