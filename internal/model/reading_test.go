package model

import (
	"testing"
	"time"
)

func TestReading_IsOverdue(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -2)
	lastWeek := now.AddDate(0, 0, -8)
	lastMonth := now.AddDate(0, -2, 0)

	tests := []struct {
		name    string
		reading Reading
		want    bool
	}{
		{
			name: "day reading overdue",
			reading: Reading{
				Date:     yesterday,
				DateType: DateTypeDay,
				Status:   StatusPending,
			},
			want: true,
		},
		{
			name: "day reading not overdue",
			reading: Reading{
				Date:     now,
				DateType: DateTypeDay,
				Status:   StatusPending,
			},
			want: false,
		},
		{
			name: "completed reading not overdue",
			reading: Reading{
				Date:     yesterday,
				DateType: DateTypeDay,
				Status:   StatusCompleted,
			},
			want: false,
		},
		{
			name: "week reading overdue",
			reading: Reading{
				Date:     lastWeek,
				DateType: DateTypeWeek,
				Status:   StatusPending,
			},
			want: true,
		},
		{
			name: "month reading overdue",
			reading: Reading{
				Date:     lastMonth,
				DateType: DateTypeMonth,
				Status:   StatusPending,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reading.IsOverdue(); got != tt.want {
				t.Errorf("Reading.IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReading_IsActiveToday(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		reading Reading
		want    bool
	}{
		{
			name: "day reading active today",
			reading: Reading{
				Date:     now,
				DateType: DateTypeDay,
			},
			want: true,
		},
		{
			name: "day reading not active today",
			reading: Reading{
				Date:     yesterday,
				DateType: DateTypeDay,
			},
			want: false,
		},
		{
			name: "week reading active today",
			reading: Reading{
				Date:     now,
				DateType: DateTypeWeek,
			},
			want: true,
		},
		{
			name: "month reading active today",
			reading: Reading{
				Date:     now,
				DateType: DateTypeMonth,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reading.IsActiveToday(); got != tt.want {
				t.Errorf("Reading.IsActiveToday() = %v, want %v", got, tt.want)
			}
		})
	}
}
