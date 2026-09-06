// Package push runs the background notification worker that delivers
// Web Push and email notifications to ReadWillBe users.
package push

import (
	"context"
	"encoding/json"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"readwillbe/internal/model"
	"readwillbe/internal/service/email"
)

// NotificationCheckInterval is how often the worker scans for users due to
// receive their daily notification.
const NotificationCheckInterval = 1 * time.Minute

// StartNotificationWorker starts the background notification loop and returns
// a cancel function that stops it.
func StartNotificationWorker(cfg model.Config, db *gorm.DB) context.CancelFunc {
	pushEnabled := cfg.VAPIDPublicKey != "" && cfg.VAPIDPrivateKey != ""
	emailEnabled := cfg.EmailEnabled()

	if !pushEnabled && !emailEnabled {
		logrus.Info("Neither VAPID keys nor email configured, notification worker not started")
		return func() {}
	}

	var emailService email.Service
	if emailEnabled {
		svc, err := email.NewService(cfg)
		if err != nil {
			logrus.Errorf("Email notifications disabled: %v", err)
			emailEnabled = false
		} else {
			emailService = svc
			logrus.Info("Email notifications enabled via " + cfg.EmailProvider)
		}
	}

	if !pushEnabled && !emailEnabled {
		logrus.Info("No usable notification transport, notification worker not started")
		return func() {}
	}

	if pushEnabled {
		logrus.Info("Push notifications enabled")
	}

	if cfg.BaseURL() == "" {
		logrus.Warn("No hostname configured: links in notification email and push payloads will be relative and unusable (set READWILLBE_HOSTNAME)")
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

func processNotifications(cfg model.Config, db *gorm.DB, emailService email.Service, pushEnabled bool) {
	now := time.Now()
	currentTime := now.Format("15:04")

	var users []model.User
	err := db.Preload("PushSubscriptions").
		Where("notification_time = ?", currentTime).
		Where("notifications_enabled = ? OR email_notifications_enabled = ?", true, true).
		Find(&users).Error

	if err != nil {
		logrus.Errorf("Error fetching users for notifications: %v", err)
		return
	}

	for _, user := range users {
		var readings []model.Reading
		err := db.Preload("Plan").
			Where("plan_id IN (?)",
				db.Table("plans").Select("id").Where("user_id = ?", user.ID),
			).
			Where("status != ?", model.StatusCompleted).
			Find(&readings).Error

		if err != nil {
			logrus.Errorf("Error fetching readings for user %d: %v", user.ID, err)
			continue
		}

		var activeReadings []model.Reading
		for _, r := range readings {
			if r.IsActiveToday() || r.IsOverdue() {
				activeReadings = append(activeReadings, r)
			}
		}

		if len(activeReadings) == 0 {
			continue
		}

		if pushEnabled && user.NotificationsEnabled && len(user.PushSubscriptions) > 0 {
			SendPushNotification(cfg, db, user)
		}

		if emailService != nil && user.EmailNotificationsEnabled {
			if err := emailService.SendDailyDigest(user, activeReadings); err != nil {
				logrus.Errorf("Error sending email to user %d: %v", user.ID, err)
			} else {
				logrus.Infof("Sent daily digest email to user %d", user.ID)
			}
		}
	}
}

// SendPushNotification dispatches the daily-reading push payload to every
// stored subscription for user, deleting any subscriptions the push gateway
// reports as gone.
func SendPushNotification(cfg model.Config, db *gorm.DB, user model.User) {
	baseURL := cfg.BaseURL()
	payload := map[string]interface{}{
		"title": "ReadWillBe",
		"body":  "You have readings due today!",
		"icon":  baseURL + "/static/icon-192.png",
		"badge": baseURL + "/static/badge-128.png",
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
