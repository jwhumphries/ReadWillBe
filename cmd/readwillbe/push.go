package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
)

const MaxSubscriptionsPerUser = 10

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
		user, ok := mw.GetSessionUser(c)
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
		db.Model(&model.PushSubscription{}).Where("user_id = ?", user.ID).Count(&count)
		if count >= MaxSubscriptionsPerUser {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "maximum subscriptions reached"})
		}

		subscription := model.PushSubscription{
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
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		var req struct {
			Endpoint string `json:"endpoint"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		result := db.Where("user_id = ? AND endpoint = ?", user.ID, req.Endpoint).Delete(&model.PushSubscription{})
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to remove subscription"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "unsubscribed"})
	}
}

func removeAllSubscriptions(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		result := db.Where("user_id = ?", user.ID).Delete(&model.PushSubscription{})
		if result.Error != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to remove subscriptions"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "all unsubscribed"})
	}
}
