package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"readwillbe/types"
	"github.com/pkg/errors"
)

func parseCSV(r io.Reader) ([]types.Reading, error) {
	reader := csv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, "reading CSV")
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have at least a header and one data row")
	}

	header := records[0]
	if len(header) < 2 {
		return nil, fmt.Errorf("CSV must have at least 2 columns: date and reading")
	}

	var readings []types.Reading
	for i, record := range records[1:] {
		if len(record) < 2 {
			return nil, fmt.Errorf("row %d: insufficient columns", i+2)
		}

		dateStr := strings.TrimSpace(record[0])
		content := strings.TrimSpace(record[1])

		if dateStr == "" || content == "" {
			return nil, fmt.Errorf("row %d: date and reading content are required", i+2)
		}

		date, dateType, err := parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i+2, err)
		}

		reading := types.Reading{
			Date:     date,
			DateType: dateType,
			Content:  content,
			Status:   types.StatusPending,
		}

		readings = append(readings, reading)
	}

	return readings, nil
}

func parseDate(dateStr string) (time.Time, types.DateType, error) {
	if strings.HasPrefix(dateStr, "Week ") {
		return parseWeekFormat(dateStr)
	}

	if len(dateStr) >= 8 && dateStr[4] == '-' && dateStr[5] == 'W' {
		return parseISOWeek(dateStr)
	}

	layouts := []struct {
		layout   string
		dateType types.DateType
	}{
		{"2006-01-02", types.DateTypeDay},
		{"01/02/2006", types.DateTypeDay},
		{"January 2006", types.DateTypeMonth},
		{"Jan 2006", types.DateTypeMonth},
		{"2006-01", types.DateTypeMonth},
	}

	for _, l := range layouts {
		t, err := time.Parse(l.layout, dateStr)
		if err == nil {
			return t, l.dateType, nil
		}
	}

	return time.Time{}, "", fmt.Errorf("invalid date format: %s (supported: YYYY-MM-DD, MM/DD/YYYY, Month YYYY, YYYY-MM, YYYY-Wnn, Week n)", dateStr)
}

func parseISOWeek(dateStr string) (time.Time, types.DateType, error) {
	var year, week int
	_, err := fmt.Sscanf(dateStr, "%d-W%d", &year, &week)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid ISO week format: %s (expected YYYY-Wnn)", dateStr)
	}

	if week < 1 || week > 53 {
		return time.Time{}, "", fmt.Errorf("week number must be between 1 and 53, got %d", week)
	}

	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	weekday := int(jan4.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	mondayOfWeek1 := jan4.AddDate(0, 0, 1-weekday)
	weekStart := mondayOfWeek1.AddDate(0, 0, 7*(week-1))

	return weekStart, types.DateTypeWeek, nil
}

func parseWeekFormat(dateStr string) (time.Time, types.DateType, error) {
	var week int
	_, err := fmt.Sscanf(dateStr, "Week %d", &week)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid week format: %s (expected 'Week n')", dateStr)
	}

	if week < 1 || week > 53 {
		return time.Time{}, "", fmt.Errorf("week number must be between 1 and 53, got %d", week)
	}

	currentYear := time.Now().Year()
	return parseISOWeek(fmt.Sprintf("%d-W%02d", currentYear, week))
}
