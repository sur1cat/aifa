package domain

import "testing"

func ptr(f float64) *float64 { return &f }

func TestGoalType_Valid(t *testing.T) {
	valid := []GoalType{GoalSavings, GoalDebt, GoalPurchase, GoalInvestment}
	for _, g := range valid {
		if !g.Valid() {
			t.Errorf("expected %q to be valid", g)
		}
	}
	invalid := []GoalType{"", "unknown", "SAVINGS"}
	for _, g := range invalid {
		if g.Valid() {
			t.Errorf("expected %q to be invalid", g)
		}
	}
}

func TestGoal_Progress(t *testing.T) {
	tests := []struct {
		name     string
		target   *float64
		current  float64
		expected float64
	}{
		{"nil target", nil, 100, 0},
		{"zero target", ptr(0), 50, 0},
		{"negative target", ptr(-10), 50, 0},
		{"50 percent", ptr(1000), 500, 0.5},
		{"100 percent", ptr(1000), 1000, 1},
		{"over 100 capped", ptr(1000), 1500, 1},
		{"negative current", ptr(1000), -100, 0},
		{"zero current", ptr(1000), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Goal{TargetAmount: tt.target, CurrentAmount: tt.current}
			got := g.Progress()
			if got != tt.expected {
				t.Errorf("Progress() = %v, want %v", got, tt.expected)
			}
		})
	}
}
