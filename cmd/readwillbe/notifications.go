package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/views"
)

func notificationCount(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		count, err := GetActiveReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			// Log error but return 0 count
			return render(c, 200, views.NotificationBell(0))
		}

		return render(c, 200, views.NotificationBell(int(count)))
	}
}

func notificationDropdown(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		readings, err := GetActiveReadings(db.WithContext(c.Request().Context()), user.ID, 5)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.NotificationDropdown(readings))
	}
}
