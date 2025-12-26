package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
)

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256DH string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
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

		subscription := types.PushSubscription{
			UserID:   user.ID,
			Endpoint: req.Endpoint,
			P256DH:   req.Keys.P256DH,
			Auth:     req.Keys.Auth,
		}

		result := db.Where("endpoint = ?", req.Endpoint).FirstOrCreate(&subscription)
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

func startNotificationWorker(cfg types.Config, db *gorm.DB) {
	if cfg.VAPIDPublicKey == "" || cfg.VAPIDPrivateKey == "" {
		fmt.Println("VAPID keys not configured, notification worker not started")
		return
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			currentTime := now.Format("15:04")

			var users []types.User
			err := db.Preload("PushSubscriptions").
				Where("notifications_enabled = ?", true).
				Where("notification_time = ?", currentTime).
				Find(&users).Error

			if err != nil {
				fmt.Printf("Error fetching users for notifications: %v\n", err)
				continue
			}

			for _, user := range users {
				var readings []types.Reading
				db.Where("plan_id IN (?)",
					db.Table("plans").Select("id").Where("user_id = ?", user.ID),
				).Find(&readings)

				hasReadingsToday := false
				for _, reading := range readings {
					if reading.Status != types.StatusCompleted && reading.IsActiveToday() {
						hasReadingsToday = true
						break
					}
				}

				if hasReadingsToday {
					sendNotification(cfg, db, user)
				}
			}
		}
	}()

	fmt.Println("Notification worker started")
}

func sendNotification(cfg types.Config, db *gorm.DB, user types.User) {
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
		fmt.Printf("Error marshaling payload: %v\n", err)
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
			fmt.Printf("Error sending notification: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == 410 {
			db.Delete(&subscription)
			fmt.Printf("Deleted stale subscription: %s\n", subscription.Endpoint)
		}
	}
}
