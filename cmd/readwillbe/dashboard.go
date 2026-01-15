package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

func dashboardHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		// Fetch only relevant readings (active today or overdue, excluding future)
		// using optimized SQL query
		readings, err := GetDashboardReadings(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load dashboard data")
		}

		todayReadings := []types.Reading{}
		overdueReadings := []types.Reading{}

		for _, reading := range readings {
			if reading.IsActiveToday() {
				todayReadings = append(todayReadings, reading)
			} else if reading.IsOverdue() {
				overdueReadings = append(overdueReadings, reading)
			}
		}

		return render(c, 200, views.Dashboard(cfg, &user, todayReadings, overdueReadings))
	}
}
