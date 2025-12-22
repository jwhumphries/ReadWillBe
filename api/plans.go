//go:build !wasm

package api

import (
	"net/http"
	"strconv"

	"readwillbe/types"

	"github.com/charmbracelet/log"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CreatePlanRequest struct {
	Title string `json:"title"`
}

type RenamePlanRequest struct {
	Title string `json:"title"`
}

func GetPlans(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		var plans []types.Plan
		result := db.Preload("Readings").Where("user_id = ?", userID).Find(&plans)
		if result.Error != nil {
			log.Error("failed to fetch plans", "error", result.Error, "user_id", userID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching plans").Error())
		}

		return c.JSON(http.StatusOK, plans)
	}
}

func GetPlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid plan id")
		}

		var plan types.Plan
		result := db.Preload("Readings").Where("id = ? AND user_id = ?", planID, userID).First(&plan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "plan not found")
			}
			log.Error("failed to fetch plan", "error", result.Error, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching plan").Error())
		}

		return c.JSON(http.StatusOK, plan)
	}
}

func CreatePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)

		title := c.FormValue("title")
		if title == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "title is required")
		}

		file, err := c.FormFile("csv")
		if err != nil {
			log.Debug("csv file required", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest, "csv file required")
		}

		src, err := file.Open()
		if err != nil {
			log.Error("failed to open uploaded file", "error", err)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "opening file").Error())
		}
		defer src.Close()

		readings, err := ParseCSV(src)
		if err != nil {
			log.Debug("failed to parse csv", "error", err)
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		plan := types.Plan{
			Title:  title,
			UserID: userID,
		}

		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&plan).Error; err != nil {
				return errors.Wrap(err, "creating plan")
			}
			for i := range readings {
				readings[i].PlanID = plan.ID
			}
			if len(readings) > 0 {
				if err := tx.Create(&readings).Error; err != nil {
					return errors.Wrap(err, "creating readings")
				}
			}
			return nil
		})

		if txErr != nil {
			log.Error("failed to create plan", "error", txErr)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(txErr, "creating plan").Error())
		}

		db.Preload("Readings").First(&plan, plan.ID)
		log.Info("plan created", "plan_id", plan.ID, "user_id", userID, "readings", len(readings))
		return c.JSON(http.StatusCreated, plan)
	}
}

func UpdatePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid plan id")
		}

		var plan types.Plan
		result := db.Where("id = ? AND user_id = ?", planID, userID).First(&plan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "plan not found")
			}
			log.Error("failed to fetch plan for update", "error", result.Error, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching plan").Error())
		}

		var req CreatePlanRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.Title != "" {
			plan.Title = req.Title
		}

		if err := db.Save(&plan).Error; err != nil {
			log.Error("failed to update plan", "error", err, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "updating plan").Error())
		}

		return c.JSON(http.StatusOK, plan)
	}
}

func RenamePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid plan id")
		}

		var plan types.Plan
		result := db.Where("id = ? AND user_id = ?", planID, userID).First(&plan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "plan not found")
			}
			log.Error("failed to fetch plan for rename", "error", result.Error, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching plan").Error())
		}

		var req RenamePlanRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.Title == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "title is required")
		}

		plan.Title = req.Title
		if err := db.Save(&plan).Error; err != nil {
			log.Error("failed to rename plan", "error", err, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "renaming plan").Error())
		}

		log.Info("plan renamed", "plan_id", planID, "new_title", req.Title)
		return c.JSON(http.StatusOK, plan)
	}
}

func DeletePlan(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserIDFromContext(c)
		planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid plan id")
		}

		var plan types.Plan
		result := db.Where("id = ? AND user_id = ?", planID, userID).First(&plan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, "plan not found")
			}
			log.Error("failed to fetch plan for delete", "error", result.Error, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(result.Error, "fetching plan").Error())
		}

		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("plan_id = ?", planID).Delete(&types.Reading{}).Error; err != nil {
				return errors.Wrap(err, "deleting readings")
			}
			if err := tx.Delete(&plan).Error; err != nil {
				return errors.Wrap(err, "deleting plan")
			}
			return nil
		})

		if txErr != nil {
			log.Error("failed to delete plan", "error", txErr, "plan_id", planID)
			return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(txErr, "deleting plan").Error())
		}

		log.Info("plan deleted", "plan_id", planID, "user_id", userID)
		return c.NoContent(http.StatusNoContent)
	}
}
