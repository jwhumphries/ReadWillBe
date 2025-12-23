package main

import (
	"fmt"
	"mime/multipart"
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
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("plan title is required")))
		}

		file, err := c.FormFile("csv")
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file is required")))
		}

		src, err := file.Open()
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to open file")))
		}

		// Create plan immediately in processing state
		plan := types.Plan{
			Title:  title,
			UserID: user.ID,
			Status: "processing",
		}

		if err := db.Create(&plan).Error; err != nil {
			src.Close()
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to create plan record")))
		}

		// Process CSV in background
		go func(p types.Plan, f multipart.File, d *gorm.DB) {
			defer func() {
				if r := recover(); r != nil {
					p.Status = "failed"
					p.ErrorMessage = fmt.Sprintf("Panic during processing: %v", r)
					d.Save(&p)
				}
				f.Close()
			}()

			readings, err := parseCSV(f)
			if err != nil {
				p.Status = "failed"
				p.ErrorMessage = fmt.Sprintf("Failed to parse CSV: %v", err)
				d.Save(&p)
				return
			}

			err = d.Transaction(func(tx *gorm.DB) error {
				for i := range readings {
					readings[i].PlanID = p.ID
				}
				if err := tx.Create(&readings).Error; err != nil {
					return err
				}
				p.Status = "active"
				if err := tx.Save(&p).Error; err != nil {
					return err
				}
				return nil
			})

			if err != nil {
				p.Status = "failed"
				p.ErrorMessage = fmt.Sprintf("Failed to save readings: %v", err)
				d.Save(&p)
			}
		}(plan, src, db)

		return htmxRedirect(c, "/plans")
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

		return htmxRedirect(c, "/plans")
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

		return htmxRedirect(c, "/plans")
	}
}

func editPlanForm(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
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
		if err := db.Preload("Readings").First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		return render(c, 200, views.EditPlan(cfg, &user, plan, nil))
	}
}

func editPlan(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
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
		if err := db.Preload("Readings").First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		title := c.FormValue("title")
		if title == "" {
			return render(c, 422, views.EditPlan(cfg, &user, plan, fmt.Errorf("plan title is required")))
		}

		plan.Title = title

		params, err := c.FormParams()
		if err != nil {
			return render(c, 422, views.EditPlan(cfg, &user, plan, fmt.Errorf("failed to parse form data")))
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&plan).Error; err != nil {
				return err
			}

			for _, reading := range plan.Readings {
				dateKey := fmt.Sprintf("readings[%d][date]", reading.ID)
				contentKey := fmt.Sprintf("readings[%d][content]", reading.ID)

				if dateStr, ok := params[dateKey]; ok && len(dateStr) > 0 {
					dateValue := dateStr[0]
					parsedDate, _, err := parseDate(dateValue)
					if err != nil {
						return errors.Wrap(err, "Failed to parse date")
					}
					reading.Date = parsedDate
				}

				if content, ok := params[contentKey]; ok && len(content) > 0 {
					reading.Content = content[0]
				}

				if err := tx.Save(&reading).Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return render(c, 422, views.EditPlan(cfg, &user, plan, errors.Wrap(err, "Failed to update plan")))
		}

		return htmxRedirect(c, "/plans")
	}
}

func deleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		readingID, err := strconv.ParseUint(c.Param("reading_id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid reading ID")
		}

		var plan types.Plan
		if err := db.First(&plan, "id = ? AND user_id = ?", planID, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		var reading types.Reading
		if err := db.First(&reading, "id = ? AND plan_id = ?", readingID, planID).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		if err := db.Delete(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to delete reading")
		}

		return c.NoContent(http.StatusOK)
	}
}
