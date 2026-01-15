package main

import (
	"net/http"
	"sort"

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

		// Filter to active/overdue readings and group by plan
		planGroups := groupReadingsByPlan(readings)

		return render(c, 200, views.Dashboard(cfg, &user, planGroups))
	}
}

func groupReadingsByPlan(readings []types.Reading) []types.PlanGroup {
	// Filter to only active today or overdue readings
	var activeReadings []types.Reading
	for _, r := range readings {
		if r.IsActiveToday() || r.IsOverdue() {
			activeReadings = append(activeReadings, r)
		}
	}

	// Group by PlanID
	groupMap := make(map[uint]*types.PlanGroup)
	for _, r := range activeReadings {
		if group, exists := groupMap[r.PlanID]; exists {
			group.Readings = append(group.Readings, r)
		} else {
			groupMap[r.PlanID] = &types.PlanGroup{
				Plan:     r.Plan,
				Readings: []types.Reading{r},
			}
		}
	}

	// Convert map to slice
	var groups []types.PlanGroup
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
