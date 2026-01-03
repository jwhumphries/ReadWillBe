package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

func notificationCount(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		var readings []types.Reading
		tx := db.WithContext(c.Request().Context())
		tx.Where("plan_id IN (?) AND status != ?",
			tx.Table("plans").Select("id").Where("user_id = ?", user.ID),
			types.StatusCompleted,
		).Find(&readings)

		count := 0
		for _, reading := range readings {
			if reading.Status != types.StatusCompleted && reading.IsActiveToday() {
				count++
			}
		}

		return render(c, 200, views.NotificationBell(count))
	}
}

func notificationDropdown(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		var readings []types.Reading
		tx := db.WithContext(c.Request().Context())
		tx.Preload("Plan").Where("plan_id IN (?) AND status != ?",
			tx.Table("plans").Select("id").Where("user_id = ?", user.ID),
			types.StatusCompleted,
		).Find(&readings)

		todayReadings := []types.Reading{}
		for _, reading := range readings {
			if reading.Status == types.StatusCompleted {
				continue
			}

			if reading.IsActiveToday() {
				todayReadings = append(todayReadings, reading)
				if len(todayReadings) >= 5 {
					break
				}
			}
		}

		return render(c, 200, views.NotificationDropdown(todayReadings))
	}
}
