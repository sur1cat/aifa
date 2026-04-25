package domain

import "testing"

func TestRecurring_NextDateFrom(t *testing.T) {
	tests := []struct {
		name      string
		freq      Frequency
		current   string
		expected  string
	}{
		{"weekly", FreqWeekly, "2026-04-21", "2026-04-28"},
		{"biweekly", FreqBiweekly, "2026-04-21", "2026-05-05"},
		{"monthly", FreqMonthly, "2026-04-21", "2026-05-21"},
		{"quarterly", FreqQuarterly, "2026-01-15", "2026-04-15"},
		{"yearly", FreqYearly, "2026-04-21", "2027-04-21"},
		{"monthly end of month", FreqMonthly, "2026-01-31", "2026-03-03"},
		{"yearly leap day", FreqYearly, "2024-02-29", "2025-03-01"},
		{"unknown frequency defaults to monthly", "custom", "2026-04-21", "2026-05-21"},
		{"invalid date returns input", FreqWeekly, "not-a-date", "not-a-date"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Recurring{Frequency: tt.freq}
			got := r.NextDateFrom(tt.current)
			if got != tt.expected {
				t.Errorf("NextDateFrom(%q) = %q, want %q", tt.current, got, tt.expected)
			}
		})
	}
}
