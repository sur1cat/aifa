package handler

import (
	"fmt"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/domain"
)

const streakLookback = 365

// calculateStreak walks back from `now` one period at a time, counting
// how many consecutive periods the habit was completed. Progress-based
// habits count a period as met when the recorded value reaches the
// target; boolean habits use the completion set. Accepting `now` as a
// parameter keeps the function deterministic for tests.
func calculateStreak(h *domain.Habit, now time.Time) int {
	if len(h.CompletedDates) == 0 && len(h.ProgressValues) == 0 {
		return 0
	}

	completed, weekly, monthly := indexCompletions(h.CompletedDates)

	streak := 0
	cursor := now
	for i := 0; i < streakLookback; i++ {
		date := cursor.Format("2006-01-02")
		hit := false

		switch {
		case h.TargetValue != nil && *h.TargetValue > 0:
			if v, ok := h.ProgressValues[date]; ok && v >= *h.TargetValue {
				hit = true
			} else if completed[date] {
				hit = true
			}
		case h.Period == domain.PeriodWeekly:
			hit = weekly[weekKey(cursor)]
		case h.Period == domain.PeriodMonthly:
			hit = monthly[cursor.Format("2006-01")]
		default:
			hit = completed[date]
		}

		if hit {
			streak++
			cursor = stepBack(cursor, h.Period)
			continue
		}
		if streak > 0 {
			break
		}
		cursor = stepBack(cursor, h.Period)
	}
	return streak
}

func indexCompletions(dates []string) (day map[string]bool, week, month map[string]bool) {
	day = make(map[string]bool, len(dates))
	week = make(map[string]bool, len(dates))
	month = make(map[string]bool, len(dates))
	for _, d := range dates {
		day[d] = true
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		week[weekKey(t)] = true
		month[t.Format("2006-01")] = true
	}
	return
}

func weekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}

func stepBack(t time.Time, p domain.Period) time.Time {
	switch p {
	case domain.PeriodWeekly:
		return t.AddDate(0, 0, -7)
	case domain.PeriodMonthly:
		return t.AddDate(0, -1, 0)
	default:
		return t.AddDate(0, 0, -1)
	}
}
