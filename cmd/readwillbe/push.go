package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"readwillbe/types"
)

const (
	NotificationCheckInterval = 1 * time.Minute
	MaxSubscriptionsPerUser   = 10
)

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

func validatePushSubscription(req PushSubscriptionRequest) error {
	if !strings.HasPrefix(req.Endpoint, "https://") {
		return fmt.Errorf("endpoint must use HTTPS")
	}

	if req.Keys.P256DH == "" || req.Keys.Auth == "" {
		return fmt.Errorf("missing encryption keys")
	}

	if _, err := base64.RawURLEncoding.DecodeString(req.Keys.P256DH); err != nil {
		if _, err := base64.StdEncoding.DecodeString(req.Keys.P256DH); err != nil {
			return fmt.Errorf("invalid P256DH key encoding")
		}
	}

	if _, err := base64.RawURLEncoding.DecodeString(req.Keys.Auth); err != nil {
		if _, err := base64.StdEncoding.DecodeString(req.Keys.Auth); err != nil {
			return fmt.Errorf("invalid Auth key encoding")
		}
	}

	return nil
}

func saveSubscription(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		var req PushSubscriptionRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		if err := validatePushSubscription(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		var count int64
		db.Model(&types.PushSubscription{}).Where("user_id = ?", user.ID).Count(&count)
		if count >= MaxSubscriptionsPerUser {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "maximum subscriptions reached"})
		}

		subscription := types.PushSubscription{
			UserID:   user.ID,
			Endpoint: req.Endpoint,
			P256DH:   req.Keys.P256DH,
			Auth:     req.Keys.Auth,
		}

		result := db.Where("user_id = ? AND endpoint = ?", user.ID, req.Endpoint).FirstOrCreate(&subscription)
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save subscription"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "subscribed"})
	}
}

func removeSubscription(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		var req struct {
			Endpoint string `json:"endpoint"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		result := db.Where("user_id = ? AND endpoint = ?", user.ID, req.Endpoint).Delete(&types.PushSubscription{})
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to remove subscription"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "unsubscribed"})
	}
}

func removeAllSubscriptions(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		result := db.Where("user_id = ?", user.ID).Delete(&types.PushSubscription{})
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to remove subscriptions"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "all unsubscribed"})
	}
}

func startNotificationWorker(cfg types.Config, db *gorm.DB) context.CancelFunc {
	pushEnabled := cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != ""
	emailEnabled := cfg.EmailEnabled()

	if !pushEnabled && !emailEnabled {
		logrus.Info("Neither VAPID keys nor email configured, notification worker not started")
		return func() {}
	}

	var emailService EmailService
	if emailEnabled {
		emailService = NewEmailService(cfg)
		logrus.Info("Email notifications enabled via " + cfg.EmailProvider)
	}

	if pushEnabled {
		logrus.Info("Push notifications enabled")
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(NotificationCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logrus.Info("Notification worker stopped")
				return
			case <-ticker.C:
				processNotifications(cfg, db, emailService, pushEnabled)
			}
		}
	}()

	logrus.Info("Notification worker started")
	return cancel
}

func processNotifications(cfg types.Config, db *gorm.DB, emailService EmailService, pushEnabled bool) {
	now := time.Now()
	currentTime := now.Format("15:04")

	var users []types.User
	err := db.Preload("PushSubscriptions").
		Where("notification_time = ?", currentTime).
		Where("notifications_enabled = ? OR email_notifications_enabled = ?", true, true).
		Find(&users).Error

	if err != nil {
		logrus.Errorf("Error fetching users for notifications: %v", err)
		return
	}

	for _, user := range users {
		var readings []types.Reading
		err := db.Preload("Plan").
			Where("plan_id IN (?)",
				db.Table("plans").Select("id").Where("user_id = ?", user.ID),
			).
			Where("status != ?", types.StatusCompleted).
			Find(&readings).Error

		if err != nil {
			logrus.Errorf("Error fetching readings for user %d: %v", user.ID, err)
			continue
		}

		var activeReadings []types.Reading
		for _, r := range readings {
			if r.IsActiveToday() || r.IsOverdue() {
				activeReadings = append(activeReadings, r)
			}
		}

		if len(activeReadings) == 0 {
			continue
		}

		if pushEnabled && user.NotificationsEnabled && len(user.PushSubscriptions) > 0 {
			sendPushNotification(cfg, db, user)
		}

		if emailService != nil && user.EmailNotificationsEnabled {
			if err := emailService.SendDailyDigest(user, activeReadings, cfg.Hostname); err != nil {
				logrus.Errorf("Error sending email to user %d: %v", user.ID, err)
			} else {
				logrus.Infof("Sent daily digest email to user %d", user.ID)
			}
		}
	}
}

func sendPushNotification(cfg types.Config, db *gorm.DB, user types.User) {
	payload := map[string]interface{}{
		"title": "ReadWillBe",
		"body":  "You have readings due today!",
		"icon":  fmt.Sprintf("https://%s/static/icon-192.png", cfg.Hostname),
		"badge": fmt.Sprintf("https://%s/static/badge-128.png", cfg.Hostname),
		"data": map[string]string{
			"url": "/",
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("Error marshaling payload: %v", err)
		return
	}

	for _, subscription := range user.PushSubscriptions {
		sub := &webpush.Subscription{
			Endpoint: subscription.Endpoint,
			Keys: webpush.Keys{
				P256dh: subscription.P256DH,
				Auth:   subscription.Auth,
			},
		}

		resp, err := webpush.SendNotification(payloadBytes, sub, &webpush.Options{
			Subscriber:      "mailto:noreply@readwillbe.app",
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			TTL:             60 * 60 * 24 * 7,
			Topic:           "daily-reading",
			Urgency:         webpush.UrgencyNormal,
		})

		if err != nil {
			logrus.Errorf("Error sending notification: %v", err)
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == 410 {
			if err := db.Delete(&subscription).Error; err != nil {
				logrus.Errorf("Error deleting stale subscription: %v", err)
			} else {
				logrus.Infof("Deleted stale subscription: %s", subscription.Endpoint)
			}
		}
	}
}
