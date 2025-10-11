package main

import (
	"net/http"

	"readwillbe/types"
	"readwillbe/views"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func dashboardHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		var readings []types.Reading
		db.Preload("Plan").Where("plan_id IN (?)",
			db.Table("plans").Select("id").Where("user_id = ?", user.ID),
		).Find(&readings)

		todayReadings := []types.Reading{}
		overdueReadings := []types.Reading{}

		for _, reading := range readings {
			if reading.Status == types.StatusCompleted {
				continue
			}

			if reading.IsActiveToday() {
				todayReadings = append(todayReadings, reading)
			} else if reading.IsOverdue() {
				overdueReadings = append(overdueReadings, reading)
			}
		}

		return render(c, 200, views.Dashboard(cfg, &user, todayReadings, overdueReadings))
	}
}
