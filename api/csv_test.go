//go:build !wasm

package api

import (
	"strings"
	"testing"
	"time"

	"readwillbe/types"
)

func TestParseCSV_ValidDayFormat(t *testing.T) {
	csv := `Date,Reading
2025-01-15,Read chapter 1
2025-01-16,Read chapter 2`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 readings, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeDay {
		t.Errorf("expected DateTypeDay, got %s", readings[0].DateType)
	}
	if readings[0].Content != "Read chapter 1" {
		t.Errorf("expected 'Read chapter 1', got '%s'", readings[0].Content)
	}
}

func TestParseCSV_ValidMonthFormat(t *testing.T) {
	csv := `Date,Reading
January 2025,Read book 1
February 2025,Read book 2`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 readings, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeMonth {
		t.Errorf("expected DateTypeMonth, got %s", readings[0].DateType)
	}
}

func TestParseCSV_ValidISOWeekFormat(t *testing.T) {
	csv := `Date,Reading
2025-W01,Read week 1 content
2025-W02,Read week 2 content`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 readings, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeWeek {
		t.Errorf("expected DateTypeWeek, got %s", readings[0].DateType)
	}
}

func TestParseCSV_ValidWeekFormat(t *testing.T) {
	csv := `Date,Reading
Week 1,Read week 1 content
Week 2,Read week 2 content`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 readings, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeWeek {
		t.Errorf("expected DateTypeWeek, got %s", readings[0].DateType)
	}
}

func TestParseCSV_USDateFormat(t *testing.T) {
	csv := `Date,Reading
01/15/2025,Read chapter 1`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 1 {
		t.Errorf("expected 1 reading, got %d", len(readings))
	}

	expected := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	if !readings[0].Date.Equal(expected) {
		t.Errorf("expected date %v, got %v", expected, readings[0].Date)
	}
}

func TestParseCSV_ShortMonthFormat(t *testing.T) {
	csv := `Date,Reading
Jan 2025,Read chapter 1`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 1 {
		t.Errorf("expected 1 reading, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeMonth {
		t.Errorf("expected DateTypeMonth, got %s", readings[0].DateType)
	}
}

func TestParseCSV_YearMonthFormat(t *testing.T) {
	csv := `Date,Reading
2025-01,Read chapter 1`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	if len(readings) != 1 {
		t.Errorf("expected 1 reading, got %d", len(readings))
	}

	if readings[0].DateType != types.DateTypeMonth {
		t.Errorf("expected DateTypeMonth, got %s", readings[0].DateType)
	}
}

func TestParseCSV_EmptyCSV(t *testing.T) {
	csv := `Date,Reading`

	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Error("expected error for empty CSV")
	}
}

func TestParseCSV_InvalidDateFormat(t *testing.T) {
	csv := `Date,Reading
invalid-date,Read chapter 1`

	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Error("expected error for invalid date format")
	}
}

func TestParseCSV_MissingContent(t *testing.T) {
	csv := `Date,Reading
2025-01-15,`

	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Error("expected error for missing content")
	}
}

func TestParseCSV_MissingDate(t *testing.T) {
	csv := `Date,Reading
,Read chapter 1`

	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Error("expected error for missing date")
	}
}

func TestParseCSV_InsufficientColumns(t *testing.T) {
	csv := `Date
2025-01-15`

	_, err := ParseCSV(strings.NewReader(csv))
	if err == nil {
		t.Error("expected error for insufficient columns")
	}
}

func TestParseISOWeek_ValidWeek(t *testing.T) {
	date, dateType, err := parseISOWeek("2025-W01")
	if err != nil {
		t.Fatalf("parseISOWeek failed: %v", err)
	}

	if dateType != types.DateTypeWeek {
		t.Errorf("expected DateTypeWeek, got %s", dateType)
	}

	if date.Weekday() != time.Monday {
		t.Errorf("expected Monday, got %s", date.Weekday())
	}
}

func TestParseISOWeek_InvalidWeekNumber(t *testing.T) {
	_, _, err := parseISOWeek("2025-W54")
	if err == nil {
		t.Error("expected error for week 54")
	}
}

func TestParseISOWeek_ZeroWeek(t *testing.T) {
	_, _, err := parseISOWeek("2025-W00")
	if err == nil {
		t.Error("expected error for week 0")
	}
}

func TestParseWeekFormat_Valid(t *testing.T) {
	date, dateType, err := parseWeekFormat("Week 1")
	if err != nil {
		t.Fatalf("parseWeekFormat failed: %v", err)
	}

	if dateType != types.DateTypeWeek {
		t.Errorf("expected DateTypeWeek, got %s", dateType)
	}

	if date.IsZero() {
		t.Error("expected non-zero date")
	}
}

func TestParseWeekFormat_InvalidWeekNumber(t *testing.T) {
	_, _, err := parseWeekFormat("Week 54")
	if err == nil {
		t.Error("expected error for week 54")
	}
}

func TestParseCSV_AllReadingsHavePendingStatus(t *testing.T) {
	csv := `Date,Reading
2025-01-15,Read chapter 1
2025-01-16,Read chapter 2`

	readings, err := ParseCSV(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("ParseCSV failed: %v", err)
	}

	for i, r := range readings {
		if r.Status != types.StatusPending {
			t.Errorf("reading %d: expected StatusPending, got %s", i, r.Status)
		}
	}
}
