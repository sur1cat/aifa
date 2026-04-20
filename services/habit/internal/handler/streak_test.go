package handler

import (
	"testing"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/domain"
)

func ptrInt(v int) *int { return &v }

func TestStreak_DailyBoolean(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	h := &domain.Habit{
		Period: domain.PeriodDaily,
		CompletedDates: []string{
			"2026-04-21", "2026-04-20", "2026-04-19", "2026-04-17",
		},
	}
	if got := calculateStreak(h, now); got != 3 {
		t.Fatalf("want streak=3 (today+2 back), got %d", got)
	}
}

func TestStreak_GapResets(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	h := &domain.Habit{
		Period:         domain.PeriodDaily,
		CompletedDates: []string{"2026-04-19"},
	}
	if got := calculateStreak(h, now); got != 1 {
		t.Fatalf("want streak=1 (today miss, then hit 2 days back), got %d", got)
	}
}

func TestStreak_ProgressTargetMet(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	target := 10
	h := &domain.Habit{
		Period:      domain.PeriodDaily,
		TargetValue: &target,
		ProgressValues: map[string]int{
			"2026-04-21": 12,
			"2026-04-20": 10,
			"2026-04-19": 5,
		},
	}
	if got := calculateStreak(h, now); got != 2 {
		t.Fatalf("want streak=2 (21 & 20 meet target, 19 below), got %d", got)
	}
}

func TestStreak_Weekly(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC) // Tuesday of week 17
	h := &domain.Habit{
		Period: domain.PeriodWeekly,
		CompletedDates: []string{
			"2026-04-20", // week 17
			"2026-04-15", // week 16 (Wed)
			"2026-04-06", // week 15
		},
	}
	if got := calculateStreak(h, now); got != 3 {
		t.Fatalf("want streak=3 consecutive weeks, got %d", got)
	}
}

func TestStreak_EmptyReturnsZero(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	h := &domain.Habit{Period: domain.PeriodDaily}
	if got := calculateStreak(h, now); got != 0 {
		t.Fatalf("want 0 for empty, got %d", got)
	}
}

// Compile-time assertion that ptrInt is referenced (keeps linter quiet if
// someone removes the only current usage).
var _ = ptrInt
