package main

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"readwillbe/types"
	"readwillbe/views"
)

const (
	MaxCSVFileSize   = 10 * 1024 * 1024 // 10MB
	MaxTitleLength   = 500
	MaxContentLength = 2000
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

const DraftTitleKey = "draft-plan-title"
const DraftReadingsKey = "draft-plan-readings"

func getDraftData(c echo.Context) (string, []views.ManualReading, error) {
	sess, err := session.Get(SessionKey, c)
	if err != nil {
		return "", nil, err
	}
	title, ok := sess.Values[DraftTitleKey].(string)
	if !ok {
		title = ""
	}
	readings, ok := sess.Values[DraftReadingsKey].([]views.ManualReading)
	if !ok || readings == nil {
		readings = []views.ManualReading{}
	}
	return title, readings, nil
}

func saveDraftData(c echo.Context, title string, readings []views.ManualReading) error {
	sess, err := session.Get(SessionKey, c)
	if err != nil {
		return err
	}
	sess.Values[DraftTitleKey] = title
	sess.Values[DraftReadingsKey] = readings
	return sess.Save(c.Request(), c.Response())
}

func clearDraftData(c echo.Context) error {
	sess, err := session.Get(SessionKey, c)
	if err != nil {
		return err
	}
	delete(sess.Values, DraftTitleKey)
	delete(sess.Values, DraftReadingsKey)
	return sess.Save(c.Request(), c.Response())
}

func manualPlanForm(cfg types.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		title, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return render(c, 200, views.ManualPlanCreate(cfg, &user, title, readings, nil))
	}
}

func updateDraftTitle() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		title := c.FormValue("title")
		_, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		if err := saveDraftData(c, title, readings); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
}

func addDraftReading() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		date := c.FormValue("date")
		content := c.FormValue("content")

		if date == "" || content == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		title, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		newReading := views.ManualReading{
			ID:      fmt.Sprintf("%d", len(readings)+1),
			Date:    date,
			Content: content,
		}
		readings = append(readings, newReading)

		if err := saveDraftData(c, title, readings); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.ManualPlanForm(title, readings, nil))
	}
}

func getDraftReading() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		id := c.Param("id")
		_, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		for _, r := range readings {
			if r.ID == id {
				return render(c, 200, views.ManualReadingRow(r))
			}
		}
		return c.NoContent(http.StatusNotFound)
	}
}

func getDraftReadingEdit() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		id := c.Param("id")
		_, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		for _, r := range readings {
			if r.ID == id {
				return render(c, 200, views.ManualReadingRowEdit(r))
			}
		}
		return c.NoContent(http.StatusNotFound)
	}
}

func updateDraftReading() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		id := c.Param("id")
		date := c.FormValue("date")
		content := c.FormValue("content")

		if date == "" || content == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		title, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		var updated views.ManualReading
		for i, r := range readings {
			if r.ID == id {
				readings[i].Date = date
				readings[i].Content = content
				updated = readings[i]
				break
			}
		}

		if err := saveDraftData(c, title, readings); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.ManualReadingRow(updated))
	}
}

func deleteDraftReading() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		id := c.Param("id")
		title, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		newReadings := make([]views.ManualReading, 0, len(readings))
		for _, r := range readings {
			if r.ID != id {
				newReadings = append(newReadings, r)
			}
		}

		if err := saveDraftData(c, title, newReadings); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.ManualPlanForm(title, newReadings, nil))
	}
}

func deleteDraft() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		if err := clearDraftData(c); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return htmxRedirect(c, "/plans")
	}
}

func createManualPlan(cfg types.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		title, draftReadings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		if title == "" {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, fmt.Errorf("plan title is required")))
		}

		if len(draftReadings) == 0 {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, fmt.Errorf("at least one reading is required")))
		}

		readings := make([]types.Reading, 0, len(draftReadings))
		for _, mr := range draftReadings {
			parsedDate, dateType, parseErr := parseDate(mr.Date)
			if err != nil {
				return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, errors.Wrap(parseErr, fmt.Sprintf("invalid date: %s", mr.Date))))
			}
			readings = append(readings, types.Reading{
				Date:     parsedDate,
				DateType: dateType,
				Content:  mr.Content,
				Status:   types.StatusPending,
			})
		}

		plan := types.Plan{
			Title:  title,
			UserID: user.ID,
			Status: "active",
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if txErr := tx.Create(&plan).Error; txErr != nil {
				return err
			}

			for i := range readings {
				readings[i].PlanID = plan.ID
			}

			if txErr := tx.Create(&readings).Error; txErr != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, errors.Wrap(err, "failed to create plan")))
		}

		_ = clearDraftData(c)
		return htmxRedirect(c, "/plans")
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
		if len(title) > MaxTitleLength {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("plan title must be less than %d characters", MaxTitleLength)))
		}

		file, err := c.FormFile("csv")
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file is required")))
		}

		if file.Size > MaxCSVFileSize {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file must be less than 10MB")))
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
			_ = src.Close()
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
				_ = f.Close()
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
				if txErr := tx.Create(&readings).Error; txErr != nil {
					return txErr
				}
				p.Status = "active"
				if txErr := tx.Save(&p).Error; txErr != nil {
					return txErr
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
		if len(newTitle) > MaxTitleLength {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Title must be less than %d characters", MaxTitleLength))
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
		if dbErr := db.Preload("Readings").First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; dbErr != nil {
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
			if txErr := tx.Save(&plan).Error; txErr != nil {
				return txErr
			}

			for _, reading := range plan.Readings {
				dateKey := fmt.Sprintf("readings[%d][date]", reading.ID)
				contentKey := fmt.Sprintf("readings[%d][content]", reading.ID)

				if dateStr, ok := params[dateKey]; ok && len(dateStr) > 0 {
					dateValue := dateStr[0]
					parsedDate, _, parseErr := parseDate(dateValue)
					if parseErr != nil {
						return errors.Wrap(parseErr, "Failed to parse date")
					}
					reading.Date = parsedDate
				}

				if content, ok := params[contentKey]; ok && len(content) > 0 {
					reading.Content = content[0]
				}

				if txErr := tx.Save(&reading).Error; txErr != nil {
					return txErr
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
