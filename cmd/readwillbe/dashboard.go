package main

import (
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	mw "readwillbe/internal/middleware"
	"readwillbe/internal/model"
	"readwillbe/internal/repository"
	"readwillbe/internal/views"
	"readwillbe/internal/views/partials"
)

func dashboardHandler(cfg model.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.Redirect(http.StatusFound, "/auth/sign-in")
		}

		// Fetch only relevant readings (active today or overdue, excluding future)
		// using optimized SQL query
		readings, err := repository.GetDashboardReadings(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load dashboard data")
		}

		// Filter to active/overdue readings and group by plan
		planGroups := groupReadingsByPlan(readings)

		weeklyCount, err := repository.GetWeeklyCompletedReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load weekly stats")
		}

		monthlyCount, err := repository.GetMonthlyCompletedReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load monthly stats")
		}

		return render(c, 200, views.Dashboard(cfg, &user, planGroups, weeklyCount, monthlyCount))
	}
}

func groupReadingsByPlan(readings []model.Reading) []model.PlanGroup {
	// Filter to only active today or overdue readings
	var activeReadings []model.Reading
	for _, r := range readings {
		if r.IsActiveToday() || r.IsOverdue() {
			activeReadings = append(activeReadings, r)
		}
	}

	// Group by PlanID
	groupMap := make(map[uint]*model.PlanGroup)
	for _, r := range activeReadings {
		if group, exists := groupMap[r.PlanID]; exists {
			group.Readings = append(group.Readings, r)
		} else {
			groupMap[r.PlanID] = &model.PlanGroup{
				Plan:     r.Plan,
				Readings: []model.Reading{r},
			}
		}
	}

	// Convert map to slice
	var groups []model.PlanGroup
	for _, g := range groupMap {
		// Sort readings within group by date (oldest first)
		sort.Slice(g.Readings, func(i, j int) bool {
			return g.Readings[i].Date.Before(g.Readings[j].Date)
		})
		groups = append(groups, *g)
	}

	// Sort groups by oldest reading date (most urgent first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Readings[0].Date.Before(groups[j].Readings[0].Date)
	})

	return groups
}

// dashboardStatsPartial returns just the stats section for HTMX updates
func dashboardStatsPartial(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, ok := mw.GetSessionUser(c)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		weeklyCount, err := repository.GetWeeklyCompletedReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load weekly stats")
		}

		monthlyCount, err := repository.GetMonthlyCompletedReadingsCount(db.WithContext(c.Request().Context()), user.ID)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Failed to load monthly stats")
		}

		return render(c, http.StatusOK, partials.DashboardStats(weeklyCount, monthlyCount))
	}
}
