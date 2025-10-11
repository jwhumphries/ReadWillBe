package main

import (
	"fmt"
	"net/http"
	"strconv"

	"readwillbe/types"
	"readwillbe/views"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func plansListHandler(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		var plans []types.Plan
		db.Preload("Readings").Where("user_id = ?", user.ID).Find(&plans)

		return render(c, 200, views.PlansList(cfg, &user, plans))
	}
}

func createPlanForm(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.CreatePlanForm(cfg, &user, nil))
	}
}

func createPlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		title := c.FormValue("title")
		if title == "" {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("Plan title is required")))
		}

		file, err := c.FormFile("csv")
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file is required")))
		}

		src, err := file.Open()
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to open file")))
		}
		defer src.Close()

		readings, err := parseCSV(src)
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to parse CSV")))
		}

		plan := types.Plan{
			Title:  title,
			UserID: user.ID,
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&plan).Error; err != nil {
				return err
			}

			for i := range readings {
				readings[i].PlanID = plan.ID
			}

			if err := tx.Create(&readings).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to create plan")))
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func renamePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan types.Plan
		if err := db.First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		newTitle := c.FormValue("title")
		if newTitle == "" {
			return c.String(http.StatusBadRequest, "Title is required")
		}

		plan.Title = newTitle
		if err := db.Save(&plan).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update plan")
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func deletePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan types.Plan
		if err := db.First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		if err := db.Delete(&plan).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to delete plan")
		}

		return c.NoContent(http.StatusOK)
	}
}
