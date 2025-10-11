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
	layouts := []struct {
		layout   string
		dateType types.DateType
	}{
		{"2006-01-02", types.DateTypeDay},
		{"01/02/2006", types.DateTypeDay},
		{"January 2006", types.DateTypeMonth},
		{"Jan 2006", types.DateTypeMonth},
		{"2006-01", types.DateTypeMonth},
		{"2006-W01", types.DateTypeWeek},
	}

	for _, l := range layouts {
		t, err := time.Parse(l.layout, dateStr)
		if err == nil {
			return t, l.dateType, nil
		}
	}

	return time.Time{}, "", fmt.Errorf("invalid date format: %s", dateStr)
}
