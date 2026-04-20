package handler

import (
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HabitHandler struct {
	habitRepo *repository.HabitRepository
}

func NewHabitHandler(habitRepo *repository.HabitRepository) *HabitHandler {
	return &HabitHandler{habitRepo: habitRepo}
}

// Request/Response types
type CreateHabitRequest struct {
	Title       string  `json:"title" binding:"required"`
	GoalID      *string `json:"goal_id"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	Period      string  `json:"period"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
}

type UpdateHabitRequest struct {
	Title       string  `json:"title"`
	GoalID      *string `json:"goal_id"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	Period      string  `json:"period"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	ArchivedAt  *string `json:"archived_at"`
}

type ToggleCompletionRequest struct {
	Date  string `json:"date" binding:"required"` // "YYYY-MM-DD"
	Value *int   `json:"value"`                   // Progress value (optional, for habits with goals)
}

type HabitResponse struct {
	ID             string         `json:"id"`
	GoalID         *string        `json:"goal_id"`
	Title          string         `json:"title"`
	Icon           string         `json:"icon"`
	Color          string         `json:"color"`
	Period         string         `json:"period"`
	CompletedDates []string       `json:"completed_dates"`
	TargetValue    *int           `json:"target_value"`
	Unit           *string        `json:"unit"`
	ProgressValues map[string]int `json:"progress_values"`
	Streak         int            `json:"streak"`
	CreatedAt      string         `json:"created_at"`
	ArchivedAt     *string        `json:"archived_at"`
}

func toHabitResponse(h *domain.Habit) *HabitResponse {
	completedDates := h.CompletedDates
	if completedDates == nil {
		completedDates = []string{}
	}
	progressValues := h.ProgressValues
	if progressValues == nil {
		progressValues = map[string]int{}
	}

	var goalID *string
	if h.GoalID != nil {
		id := h.GoalID.String()
		goalID = &id
	}

	var archivedAt *string
	if h.ArchivedAt != nil {
		formatted := h.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		archivedAt = &formatted
	}

	streak := calculateStreak(h)

	return &HabitResponse{
		ID:             h.ID.String(),
		GoalID:         goalID,
		Title:          h.Title,
		Icon:           h.Icon,
		Color:          h.Color,
		Period:         string(h.Period),
		CompletedDates: completedDates,
		TargetValue:    h.TargetValue,
		Unit:           h.Unit,
		ProgressValues: progressValues,
		Streak:         streak,
		CreatedAt:      h.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ArchivedAt:     archivedAt,
	}
}

// calculateStreak calculates the current streak for a habit
func calculateStreak(h *domain.Habit) int {
	if len(h.CompletedDates) == 0 && len(h.ProgressValues) == 0 {
		return 0
	}

	// Create a set of completed dates for fast lookup
	completedSet := make(map[string]bool)
	for _, d := range h.CompletedDates {
		completedSet[d] = true
	}

	streak := 0
	checkDate := time.Now()

	for i := 0; i < 365; i++ {
		dateStr := checkDate.Format("2006-01-02")
		hasCompletion := false

		// For habits with goals, check progress values
		if h.TargetValue != nil && *h.TargetValue > 0 {
			if progress, ok := h.ProgressValues[dateStr]; ok && progress >= *h.TargetValue {
				hasCompletion = true
			} else if completedSet[dateStr] {
				hasCompletion = true
			}
		} else {
			// For simple habits, check completed dates
			switch h.Period {
			case domain.HabitPeriodDaily:
				hasCompletion = completedSet[dateStr]
			case domain.HabitPeriodWeekly:
				hasCompletion = isDateInWeek(completedSet, checkDate)
			case domain.HabitPeriodMonthly:
				hasCompletion = isDateInMonth(completedSet, checkDate)
			}
		}

		if hasCompletion {
			streak++
			checkDate = decrementDate(checkDate, h.Period)
		} else if streak > 0 {
			break
		} else {
			checkDate = decrementDate(checkDate, h.Period)
		}
	}

	return streak
}

func decrementDate(t time.Time, period domain.HabitPeriod) time.Time {
	switch period {
	case domain.HabitPeriodDaily:
		return t.AddDate(0, 0, -1)
	case domain.HabitPeriodWeekly:
		return t.AddDate(0, 0, -7)
	case domain.HabitPeriodMonthly:
		return t.AddDate(0, -1, 0)
	}
	return t.AddDate(0, 0, -1)
}

func isDateInWeek(completedSet map[string]bool, checkDate time.Time) bool {
	year, week := checkDate.ISOWeek()
	for dateStr := range completedSet {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			y, w := t.ISOWeek()
			if y == year && w == week {
				return true
			}
		}
	}
	return false
}

func isDateInMonth(completedSet map[string]bool, checkDate time.Time) bool {
	year, month := checkDate.Year(), checkDate.Month()
	for dateStr := range completedSet {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			if t.Year() == year && t.Month() == month {
				return true
			}
		}
	}
	return false
}

// ListHabits returns all habits for the current user
// GET /api/v1/habits
func (h *HabitHandler) ListHabits(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	habits, err := h.habitRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		respondInternalError(c, "Failed to get habits")
		return
	}

	response := make([]*HabitResponse, len(habits))
	for i, habit := range habits {
		response[i] = toHabitResponse(habit)
	}

	respondOK(c, response)
}

// CreateHabit creates a new habit
// POST /api/v1/habits
func (h *HabitHandler) CreateHabit(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Set defaults
	icon := req.Icon
	if icon == "" {
		icon = "🎯"
	}
	color := req.Color
	if color == "" {
		color = "green"
	}
	period := domain.HabitPeriod(req.Period)
	if period == "" {
		period = domain.HabitPeriodDaily
	}

	habit := &domain.Habit{
		UserID:      userID,
		Title:       req.Title,
		Icon:        icon,
		Color:       color,
		Period:      period,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
	}

	// Parse goal_id if provided
	if req.GoalID != nil && *req.GoalID != "" {
		goalID, err := uuid.Parse(*req.GoalID)
		if err == nil {
			habit.GoalID = &goalID
		}
	}

	if err := h.habitRepo.Create(c.Request.Context(), habit); err != nil {
		respondInternalError(c, "Failed to create habit")
		return
	}

	respondCreated(c, toHabitResponse(habit))
}

// GetHabit returns a single habit
// GET /api/v1/habits/:id
func (h *HabitHandler) GetHabit(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid habit ID")
		return
	}

	// Verify ownership
	owns, err := h.habitRepo.VerifyOwnership(c.Request.Context(), habitID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Habit not found")
		return
	}

	habit, err := h.habitRepo.GetByID(c.Request.Context(), habitID)
	if err != nil {
		respondNotFound(c, "Habit not found")
		return
	}

	respondOK(c, toHabitResponse(habit))
}

// UpdateHabit updates a habit
// PUT /api/v1/habits/:id
func (h *HabitHandler) UpdateHabit(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid habit ID")
		return
	}

	// Verify ownership
	owns, err := h.habitRepo.VerifyOwnership(c.Request.Context(), habitID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Habit not found")
		return
	}

	var req UpdateHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	habit, err := h.habitRepo.GetByID(c.Request.Context(), habitID)
	if err != nil {
		respondNotFound(c, "Habit not found")
		return
	}

	// Update fields if provided
	if req.Title != "" {
		habit.Title = req.Title
	}
	if req.Icon != "" {
		habit.Icon = req.Icon
	}
	if req.Color != "" {
		habit.Color = req.Color
	}
	if req.Period != "" {
		habit.Period = domain.HabitPeriod(req.Period)
	}
	if req.TargetValue != nil {
		habit.TargetValue = req.TargetValue
	}
	if req.Unit != nil {
		habit.Unit = req.Unit
	}
	// Handle goal_id - can be set or cleared
	if req.GoalID != nil {
		if *req.GoalID == "" {
			habit.GoalID = nil
		} else {
			goalID, err := uuid.Parse(*req.GoalID)
			if err == nil {
				habit.GoalID = &goalID
			}
		}
	}
	// Handle archived_at - can be set or cleared
	if req.ArchivedAt != nil {
		if *req.ArchivedAt == "" {
			habit.ArchivedAt = nil
		} else {
			parsedTime, err := time.Parse(time.RFC3339, *req.ArchivedAt)
			if err == nil {
				habit.ArchivedAt = &parsedTime
			}
		}
	}

	if err := h.habitRepo.Update(c.Request.Context(), habit); err != nil {
		respondInternalError(c, "Failed to update habit")
		return
	}

	respondOK(c, toHabitResponse(habit))
}

// DeleteHabit deletes a habit
// DELETE /api/v1/habits/:id
func (h *HabitHandler) DeleteHabit(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid habit ID")
		return
	}

	// Verify ownership
	owns, err := h.habitRepo.VerifyOwnership(c.Request.Context(), habitID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Habit not found")
		return
	}

	if err := h.habitRepo.Delete(c.Request.Context(), habitID); err != nil {
		respondInternalError(c, "Failed to delete habit")
		return
	}

	respondMessage(c, "Habit deleted successfully")
}

// ToggleCompletion toggles a habit completion for a date
// POST /api/v1/habits/:id/toggle
func (h *HabitHandler) ToggleCompletion(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid habit ID")
		return
	}

	// Verify ownership
	owns, err := h.habitRepo.VerifyOwnership(c.Request.Context(), habitID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Habit not found")
		return
	}

	var req ToggleCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Get current habit to check if date is already completed
	habit, err := h.habitRepo.GetByID(c.Request.Context(), habitID)
	if err != nil {
		respondNotFound(c, "Habit not found")
		return
	}

	// If habit has a target and value is provided, set progress value
	if habit.TargetValue != nil && req.Value != nil {
		if *req.Value > 0 {
			// Set progress value
			if err := h.habitRepo.SetProgressValue(c.Request.Context(), habitID, req.Date, *req.Value); err != nil {
				respondInternalError(c, "Failed to update progress")
				return
			}
			// Also mark as completed if progress >= target
			if *req.Value >= *habit.TargetValue {
				_ = h.habitRepo.AddCompletion(c.Request.Context(), habitID, req.Date)
			} else {
				// Remove from completed if below target
				_ = h.habitRepo.RemoveCompletion(c.Request.Context(), habitID, req.Date)
			}
		} else {
			// Remove progress
			_ = h.habitRepo.RemoveProgressValue(c.Request.Context(), habitID, req.Date)
			_ = h.habitRepo.RemoveCompletion(c.Request.Context(), habitID, req.Date)
		}
	} else {
		// Simple toggle for habits without goals
		// Check if already completed
		isCompleted := false
		for _, d := range habit.CompletedDates {
			if d == req.Date {
				isCompleted = true
				break
			}
		}

		if isCompleted {
			// Remove completion
			if err := h.habitRepo.RemoveCompletion(c.Request.Context(), habitID, req.Date); err != nil {
				respondInternalError(c, "Failed to update completion")
				return
			}
		} else {
			// Add completion
			if err := h.habitRepo.AddCompletion(c.Request.Context(), habitID, req.Date); err != nil {
				respondInternalError(c, "Failed to update completion")
				return
			}
		}
	}

	// Return updated habit
	habit, _ = h.habitRepo.GetByID(c.Request.Context(), habitID)
	respondOK(c, toHabitResponse(habit))
}
