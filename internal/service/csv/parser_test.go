package csv

import (
	"strings"
	"testing"

	"readwillbe/internal/model"
)

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
		wantErr bool
	}{
		{
			name: "valid CSV with day dates",
			input: `date,reading
2025-01-15,Read Chapter 1
2025-01-16,Read Chapter 2`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "valid CSV with month dates",
			input: `date,reading
January 2025,Read Book 1
February 2025,Read Book 2`,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty CSV",
			input:   `date,reading`,
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "missing columns",
			input: `date,reading
2025-01-15`,
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "invalid date format",
			input: `date,reading
invalid-date,Read Chapter 1`,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			readings, err := ParseCSV(r)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(readings) != tt.wantLen {
				t.Errorf("ParseCSV() got %d readings, want %d", len(readings), tt.wantLen)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name         string
		dateStr      string
		wantDateType model.DateType
		wantErr      bool
	}{
		{
			name:         "day format YYYY-MM-DD",
			dateStr:      "2025-01-15",
			wantDateType: model.DateTypeDay,
			wantErr:      false,
		},
		{
			name:         "day format MM/DD/YYYY",
			dateStr:      "01/15/2025",
			wantDateType: model.DateTypeDay,
			wantErr:      false,
		},
		{
			name:         "month format full",
			dateStr:      "January 2025",
			wantDateType: model.DateTypeMonth,
			wantErr:      false,
		},
		{
			name:         "month format short",
			dateStr:      "Jan 2025",
			wantDateType: model.DateTypeMonth,
			wantErr:      false,
		},
		{
			name:         "ISO week format W01",
			dateStr:      "2025-W01",
			wantDateType: model.DateTypeWeek,
			wantErr:      false,
		},
		{
			name:         "ISO week format W15",
			dateStr:      "2025-W15",
			wantDateType: model.DateTypeWeek,
			wantErr:      false,
		},
		{
			name:         "ISO week format W52",
			dateStr:      "2025-W52",
			wantDateType: model.DateTypeWeek,
			wantErr:      false,
		},
		{
			name:         "Week format simple",
			dateStr:      "Week 1",
			wantDateType: model.DateTypeWeek,
			wantErr:      false,
		},
		{
			name:         "Week format double digit",
			dateStr:      "Week 15",
			wantDateType: model.DateTypeWeek,
			wantErr:      false,
		},
		{
			name:    "invalid format",
			dateStr: "not-a-date",
			wantErr: true,
		},
		{
			name:    "week out of range",
			dateStr: "2025-W54",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, dateType, err := ParseDate(tt.dateStr)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if dateType != tt.wantDateType {
					t.Errorf("ParseDate() dateType = %v, want %v", dateType, tt.wantDateType)
				}
				if date.IsZero() {
					t.Errorf("ParseDate() returned zero time")
				}
			}
		})
	}
}
