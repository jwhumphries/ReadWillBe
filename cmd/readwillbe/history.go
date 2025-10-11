package main

import (
	"net/http"

	"readwillbe/types"
	"readwillbe/views"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func historyHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		var readings []types.Reading
		db.Preload("Plan").
			Where("plan_id IN (?) AND status = ?",
				db.Table("plans").Select("id").Where("user_id = ?", user.ID),
				types.StatusCompleted,
			).
			Order("completed_at DESC").
			Find(&readings)

		return render(c, 200, views.History(cfg, &user, readings))
	}
}
