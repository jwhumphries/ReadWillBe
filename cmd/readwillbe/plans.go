package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	csvservice "readwillbe/internal/service/csv"
	"readwillbe/internal/views"
)

const (
	MaxCSVFileSize   = 10 * 1024 * 1024 // 10MB
	MaxTitleLength   = 500
	MaxContentLength = 2000
)

func plansListHandler(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		var plans []model.Plan
		db.WithContext(c.Request().Context()).Preload("Readings").Where("user_id = ?", user.ID).Order("title ASC").Find(&plans)

		// Separate into in-progress and completed plans
		var inProgressPlans, completedPlans []model.Plan
		for _, plan := range plans {
			if plan.IsComplete() {
				completedPlans = append(completedPlans, plan)
			} else {
				inProgressPlans = append(inProgressPlans, plan)
			}
		}

		return render(c, 200, views.PlansList(cfg, &user, inProgressPlans, completedPlans))
	}
}

func createPlanForm(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		return render(c, 200, views.CreatePlanForm(cfg, &user, nil))
	}
}

const DraftTitleKey = "draft-plan-title"
const DraftReadingsKey = "draft-plan-readings"

func getDraftData(c echo.Context) (string, []views.ManualReading, error) {
	sess, err := session.Get(mw.SessionKey, c)
	if err != nil {
		logrus.Errorf("getDraftData: failed to get session: %v", err)
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
	logrus.Debugf("getDraftData: title=%q, readings=%d", title, len(readings))
	return title, readings, nil
}

func saveDraftData(c echo.Context, title string, readings []views.ManualReading) error {
	sess, err := session.Get(mw.SessionKey, c)
	if err != nil {
		logrus.Errorf("saveDraftData: failed to get session: %v", err)
		return err
	}
	sess.Values[DraftTitleKey] = title
	sess.Values[DraftReadingsKey] = readings
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		logrus.Errorf("saveDraftData: failed to save session: %v", err)
		return err
	}
	logrus.Debugf("saveDraftData: saved title=%q, readings=%d", title, len(readings))
	return nil
}

func clearDraftData(c echo.Context) error {
	sess, err := session.Get(mw.SessionKey, c)
	if err != nil {
		return err
	}
	delete(sess.Values, DraftTitleKey)
	delete(sess.Values, DraftReadingsKey)
	return sess.Save(c.Request(), c.Response())
}

func manualPlanForm(cfg model.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
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
		_, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		title := c.FormValue("title")
		if len(title) > MaxTitleLength {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Title exceeds maximum length of %d characters", MaxTitleLength))
		}
		if title != "" && csvservice.IsFormulaInjection(title) {
			return c.String(http.StatusBadRequest, "Title cannot start with formula characters (=, +, -, @)")
		}

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
		_, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		date := c.FormValue("date")
		content := c.FormValue("content")

		if date == "" || content == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		if len(content) > MaxContentLength {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Content exceeds maximum length of %d characters", MaxContentLength))
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
		_, ok := mw.GetSessionUser(c)
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
		_, ok := mw.GetSessionUser(c)
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
		_, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		id := c.Param("id")
		date := c.FormValue("date")
		content := c.FormValue("content")

		if date == "" || content == "" {
			return c.NoContent(http.StatusBadRequest)
		}

		if len(content) > MaxContentLength {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Content exceeds maximum length of %d characters", MaxContentLength))
		}

		title, readings, err := getDraftData(c)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		var updated *views.ManualReading
		for i, r := range readings {
			if r.ID == id {
				readings[i].Date = date
				readings[i].Content = content
				updated = &readings[i]
				break
			}
		}

		if updated == nil {
			return c.NoContent(http.StatusNotFound)
		}

		if err := saveDraftData(c, title, readings); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return render(c, 200, views.ManualReadingRow(*updated))
	}
}

func deleteDraftReading() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := mw.GetSessionUser(c)
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
		_, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		if err := clearDraftData(c); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.Redirect(http.StatusFound, "/plans")
	}
}

func createManualPlan(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		// Get title from form
		title := c.FormValue("title")

		// Get readings from JSON (React form submission)
		readingsJSON := c.FormValue("readingsJSON")
		var formReadings []struct {
			ID      string `json:"id"`
			Date    string `json:"date"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(readingsJSON), &formReadings); err != nil {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, nil, fmt.Errorf("invalid readings data")))
		}

		// Convert to ManualReading for error display
		draftReadings := make([]views.ManualReading, len(formReadings))
		for i, r := range formReadings {
			draftReadings[i] = views.ManualReading{ID: r.ID, Date: r.Date, Content: r.Content}
		}

		if title == "" {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, fmt.Errorf("plan title is required")))
		}

		if csvservice.IsFormulaInjection(title) {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, fmt.Errorf("title cannot start with formula characters (=, +, -, @)")))
		}

		if len(formReadings) == 0 {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, fmt.Errorf("at least one reading is required")))
		}

		readings := make([]model.Reading, 0, len(formReadings))
		for _, mr := range formReadings {
			parsedDate, dateType, parseErr := csvservice.ParseDate(mr.Date)
			if parseErr != nil {
				return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, errors.Wrap(parseErr, fmt.Sprintf("invalid date: %s", mr.Date))))
			}
			readings = append(readings, model.Reading{
				Date:     parsedDate,
				DateType: dateType,
				Content:  mr.Content,
				Status:   model.StatusPending,
			})
		}

		plan := model.Plan{
			Title:  title,
			UserID: user.ID,
			Status: "active",
		}

		txErr := db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
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

		if txErr != nil {
			return render(c, 422, views.ManualPlanCreate(cfg, &user, title, draftReadings, errors.Wrap(txErr, "failed to create plan")))
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func createPlan(fs afero.Fs, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
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
		if csvservice.IsFormulaInjection(title) {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("title cannot start with formula characters (=, +, -, @)")))
		}

		file, err := c.FormFile("csv")
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file is required")))
		}

		if file.Size > MaxCSVFileSize {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("CSV file must be less than 10MB")))
		}

		contentType := file.Header.Get("Content-Type")
		validCSVTypes := map[string]bool{
			"text/csv":                 true,
			"application/csv":          true,
			"text/plain":               true,
			"application/vnd.ms-excel": true,
			"application/octet-stream": true,
			"":                         true, // Allow empty content-type from multipart forms
		}
		if !validCSVTypes[contentType] {
			return render(c, 422, views.CreatePlanFormError(fmt.Errorf("invalid file type: must be a CSV file")))
		}

		src, err := file.Open()
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to open file")))
		}
		defer func() { _ = src.Close() }()

		// Create a temp file to store the CSV
		tempFile, err := afero.TempFile(fs, "", "plan-upload-*.csv")
		if err != nil {
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to create temp file")))
		}
		tempPath := tempFile.Name()
		// We don't defer fs.Remove here because the goroutine needs it.
		// The goroutine is responsible for cleanup.

		if _, err := io.Copy(tempFile, src); err != nil {
			_ = tempFile.Close()
			_ = fs.Remove(tempPath)
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to save CSV")))
		}
		_ = tempFile.Close()

		// Create plan immediately in processing state
		plan := model.Plan{
			Title:  title,
			UserID: user.ID,
			Status: "processing",
		}

		if err := db.WithContext(c.Request().Context()).Create(&plan).Error; err != nil {
			_ = fs.Remove(tempPath)
			return render(c, 422, views.CreatePlanFormError(errors.Wrap(err, "Failed to create plan record")))
		}

		// Process CSV in background
		go func(p model.Plan, filePath string, fileSys afero.Fs, d *gorm.DB) {
			defer func() {
				if r := recover(); r != nil {
					p.Status = "failed"
					p.ErrorMessage = fmt.Sprintf("Panic during processing: %v", r)
					d.Save(&p)
				}
				_ = fileSys.Remove(filePath)
			}()

			f, err := fileSys.Open(filePath)
			if err != nil {
				p.Status = "failed"
				p.ErrorMessage = fmt.Sprintf("Failed to open CSV file: %v", err)
				d.Save(&p)
				return
			}
			defer func() { _ = f.Close() }()

			readings, err := csvservice.ParseCSV(f)
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
		}(plan, tempPath, fs, db)

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func renamePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan model.Plan
		if err := db.WithContext(c.Request().Context()).First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		newTitle := c.FormValue("title")
		if newTitle == "" {
			return c.String(http.StatusBadRequest, "Title is required")
		}
		if len(newTitle) > MaxTitleLength {
			return c.String(http.StatusBadRequest, fmt.Sprintf("Title must be less than %d characters", MaxTitleLength))
		}
		if csvservice.IsFormulaInjection(newTitle) {
			return c.String(http.StatusBadRequest, "Title cannot start with formula characters (=, +, -, @)")
		}

		plan.Title = newTitle
		if err := db.WithContext(c.Request().Context()).Save(&plan).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to update plan")
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func deletePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan model.Plan
		if err := db.WithContext(c.Request().Context()).First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		if err := db.WithContext(c.Request().Context()).Delete(&plan).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to delete plan")
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

func editPlanForm(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan model.Plan
		if err := db.WithContext(c.Request().Context()).Preload("Readings").First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		return render(c, 200, views.EditPlan(cfg, &user, plan, nil))
	}
}

func editPlan(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid plan ID")
		}

		var plan model.Plan
		if dbErr := db.WithContext(c.Request().Context()).Preload("Readings").First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; dbErr != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		title := c.FormValue("title")
		if title == "" {
			return render(c, 422, views.EditPlan(cfg, &user, plan, fmt.Errorf("plan title is required")))
		}
		if csvservice.IsFormulaInjection(title) {
			return render(c, 422, views.EditPlan(cfg, &user, plan, fmt.Errorf("title cannot start with formula characters (=, +, -, @)")))
		}

		plan.Title = title

		readingsJSON := c.FormValue("readingsJSON")
		var formReadings []struct {
			ID      string `json:"id"`
			Date    string `json:"date"`
			Content string `json:"content"`
		}
		if jsonErr := json.Unmarshal([]byte(readingsJSON), &formReadings); jsonErr != nil {
			return render(c, 422, views.EditPlan(cfg, &user, plan, fmt.Errorf("invalid readings data")))
		}

		err = db.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
			if txErr := tx.Save(&plan).Error; txErr != nil {
				return txErr
			}

			processedIDs := make(map[uint]bool)

			for _, fr := range formReadings {
				parsedDate, dateType, parseErr := csvservice.ParseDate(fr.Date)
				if parseErr != nil {
					return errors.Wrap(parseErr, fmt.Sprintf("invalid date: %s", fr.Date))
				}

				// Try to parse ID as uint to check if it's an existing reading
				readingID, parseIDErr := strconv.ParseUint(fr.ID, 10, 32)
				var existingReading *model.Reading

				// If ID is valid uint, check if it belongs to this plan
				if parseIDErr == nil {
					for i := range plan.Readings {
						if plan.Readings[i].ID == uint(readingID) {
							existingReading = &plan.Readings[i]
							break
						}
					}
				}

				if existingReading != nil {
					// Update existing
					existingReading.Date = parsedDate
					existingReading.DateType = dateType
					existingReading.Content = fr.Content
					if saveErr := tx.Save(existingReading).Error; saveErr != nil {
						return saveErr
					}
					processedIDs[existingReading.ID] = true
				} else {
					// Create new
					newReading := model.Reading{
						PlanID:   plan.ID,
						Date:     parsedDate,
						DateType: dateType,
						Content:  fr.Content,
						Status:   model.StatusPending,
					}
					if createErr := tx.Create(&newReading).Error; createErr != nil {
						return createErr
					}
				}
			}

			// Delete readings that were not in the form data
			for _, r := range plan.Readings {
				if !processedIDs[r.ID] {
					if delErr := tx.Delete(&r).Error; delErr != nil {
						return delErr
					}
				}
			}

			return nil
		})

		if err != nil {
			return render(c, 422, views.EditPlan(cfg, &user, plan, errors.Wrap(err, "Failed to update plan")))
		}

		return c.Redirect(http.StatusFound, "/plans")
	}
}

// JSON API endpoint for plan status (used by React polling)
func apiPlanStatus(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid plan ID"})
		}

		var plan model.Plan
		if err := db.WithContext(c.Request().Context()).First(&plan, "id = ? AND user_id = ?", id, user.ID).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "plan not found"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": plan.Status})
	}
}

// JSON API endpoint for saving draft (used by React PlanEditor)
func apiSaveDraft() echo.HandlerFunc {
	return func(c echo.Context) error {
		_, ok := mw.GetSessionUser(c)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		var body struct {
			Title    string `json:"title"`
			Readings []struct {
				Date    string `json:"date"`
				Content string `json:"content"`
			} `json:"readings"`
		}

		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		}

		title, readings, err := getDraftData(c)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get draft"})
		}

		// Update title if provided
		if body.Title != "" {
			if len(body.Title) > MaxTitleLength {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("title exceeds maximum length of %d characters", MaxTitleLength)})
			}
			if csvservice.IsFormulaInjection(body.Title) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "title cannot start with formula characters"})
			}
			title = body.Title
		}

		// Update readings if provided
		if body.Readings != nil {
			readings = make([]views.ManualReading, len(body.Readings))
			for i, r := range body.Readings {
				if len(r.Content) > MaxContentLength {
					return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("reading content exceeds maximum length of %d characters", MaxContentLength)})
				}
				readings[i] = views.ManualReading{
					ID:      fmt.Sprintf("%d", i+1),
					Date:    r.Date,
					Content: r.Content,
				}
			}
		}

		if err := saveDraftData(c, title, readings); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save draft"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

func deleteReading(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
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

		var plan model.Plan
		if err := db.WithContext(c.Request().Context()).First(&plan, "id = ? AND user_id = ?", planID, user.ID).Error; err != nil {
			return c.String(http.StatusNotFound, "Plan not found")
		}

		var reading model.Reading
		if err := db.WithContext(c.Request().Context()).First(&reading, "id = ? AND plan_id = ?", readingID, planID).Error; err != nil {
			return c.String(http.StatusNotFound, "Reading not found")
		}

		if err := db.WithContext(c.Request().Context()).Delete(&reading).Error; err != nil {
			return c.String(http.StatusInternalServerError, "Failed to delete reading")
		}

		return c.NoContent(http.StatusOK)
	}
}
