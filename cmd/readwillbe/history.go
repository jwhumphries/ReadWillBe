package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

func historyHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		var readings []types.Reading
		tx := db.WithContext(c.Request().Context())
		tx.Preload("Plan").
			Where("plan_id IN (?) AND status = ?",
				tx.Table("plans").Select("id").Where("user_id = ?", user.ID),
				types.StatusCompleted,
			).
			Order("completed_at DESC").
			Find(&readings)

		return render(c, 200, views.History(cfg, &user, readings))
	}
}
