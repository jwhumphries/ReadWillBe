package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/repository"
	"readwillbe/internal/views"
)

func notificationCount(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		count, err := repository.GetActiveReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			// Log error but return 0 count
			return render(c, 200, views.NotificationBell(0))
		}

		return render(c, 200, views.NotificationBell(int(count)))
	}
}

func notificationDropdown(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		readings, err := repository.GetActiveReadings(db.WithContext(c.Request().Context()), user.ID, 5)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.NotificationDropdown(readings))
	}
}

// JSON API handlers for React components

func apiNotificationCount(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		count, err := repository.GetActiveReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.JSON(http.StatusOK, map[string]int{"count": 0})
		}

		return c.JSON(http.StatusOK, map[string]int64{"count": count})
	}
}

type apiReading struct {
	ID      uint     `json:"id"`
	Date    string   `json:"date"`
	Content string   `json:"content"`
	Plan    *apiPlan `json:"plan,omitempty"`
}

type apiPlan struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}

func apiNotificationReadings(db *gorm.DB) echo.HandlerFunc {
	return func(c *echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		readings, err := repository.GetActiveReadings(db.WithContext(c.Request().Context()), user.ID, 10)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch readings"})
		}

		apiReadings := make([]apiReading, len(readings))
		for i, r := range readings {
			apiReadings[i] = apiReading{
				ID:      r.ID,
				Date:    r.FormattedDate(),
				Content: r.Content,
			}
			if r.Plan.ID != 0 {
				apiReadings[i].Plan = &apiPlan{
					ID:    r.Plan.ID,
					Title: r.Plan.Title,
				}
			}
		}

		return c.JSON(http.StatusOK, map[string][]apiReading{"readings": apiReadings})
	}
}
