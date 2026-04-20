package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/domain"
	"github.com/sur1cat/aifa/habit-service/internal/middleware"
	"github.com/sur1cat/aifa/habit-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HabitHandler struct {
	habits *repository.HabitRepository
}

func NewHabitHandler(r *repository.HabitRepository) *HabitHandler {
	return &HabitHandler{habits: r}
}

type habitDTO struct {
	ID             string         `json:"id"`
	GoalID         *string        `json:"goal_id,omitempty"`
	Title          string         `json:"title"`
	Icon           string         `json:"icon"`
	Color          string         `json:"color"`
	Period         string         `json:"period"`
	CompletedDates []string       `json:"completed_dates"`
	TargetValue    *int           `json:"target_value,omitempty"`
	Unit           *string        `json:"unit,omitempty"`
	ProgressValues map[string]int `json:"progress_values"`
	Streak         int            `json:"streak"`
	CreatedAt      string         `json:"created_at"`
	ArchivedAt     *string        `json:"archived_at,omitempty"`
}

func toDTO(h *domain.Habit) habitDTO {
	completed := h.CompletedDates
	if completed == nil {
		completed = []string{}
	}
	progress := h.ProgressValues
	if progress == nil {
		progress = map[string]int{}
	}

	var goalID *string
	if h.GoalID != nil {
		s := h.GoalID.String()
		goalID = &s
	}
	var archivedAt *string
	if h.ArchivedAt != nil {
		s := h.ArchivedAt.Format(time.RFC3339)
		archivedAt = &s
	}

	return habitDTO{
		ID:             h.ID.String(),
		GoalID:         goalID,
		Title:          h.Title,
		Icon:           h.Icon,
		Color:          h.Color,
		Period:         string(h.Period),
		CompletedDates: completed,
		TargetValue:    h.TargetValue,
		Unit:           h.Unit,
		ProgressValues: progress,
		Streak:         calculateStreak(h, time.Now()),
		CreatedAt:      h.CreatedAt.Format(time.RFC3339),
		ArchivedAt:     archivedAt,
	}
}

// parseGoalID turns an optional string pointer into an optional UUID.
// Nil or empty string means "no goal"; malformed input is rejected at
// handler level instead of being silently dropped.
func parseGoalID(s *string) (*uuid.UUID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *HabitHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	habits, err := h.habits.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list habits", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get habits")
		return
	}

	dtos := make([]habitDTO, len(habits))
	for i, hb := range habits {
		dtos[i] = toDTO(hb)
	}
	respondOK(c, dtos)
}

type createRequest struct {
	Title       string  `json:"title" binding:"required"`
	GoalID      *string `json:"goal_id"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	Period      string  `json:"period"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
}

func (h *HabitHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	goalID, err := parseGoalID(req.GoalID)
	if err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid goal_id")
		return
	}

	icon := req.Icon
	if icon == "" {
		icon = "🎯"
	}
	color := req.Color
	if color == "" {
		color = "green"
	}
	period := domain.Period(req.Period)
	if period == "" {
		period = domain.PeriodDaily
	}

	habit := &domain.Habit{
		UserID:      userID,
		GoalID:      goalID,
		Title:       req.Title,
		Icon:        icon,
		Color:       color,
		Period:      period,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
	}
	if err := h.habits.Create(c.Request.Context(), habit); err != nil {
		slog.Error("create habit", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create habit")
		return
	}
	respondCreated(c, toDTO(habit))
}

func (h *HabitHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid habit ID")
		return
	}

	habit, err := h.habits.GetOwnedByID(c.Request.Context(), habitID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Habit not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load habit")
		return
	}
	respondOK(c, toDTO(habit))
}

type updateRequest struct {
	Title       string  `json:"title"`
	GoalID      *string `json:"goal_id"`
	Icon        string  `json:"icon"`
	Color       string  `json:"color"`
	Period      string  `json:"period"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	ArchivedAt  *string `json:"archived_at"`
}

func (h *HabitHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid habit ID")
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	habit, err := h.habits.GetOwnedByID(c.Request.Context(), habitID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Habit not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load habit")
		return
	}

	if err := applyUpdate(habit, req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if err := h.habits.Update(c.Request.Context(), habit); err != nil {
		slog.Error("update habit", "err", err, "habit_id", habit.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update habit")
		return
	}
	respondOK(c, toDTO(habit))
}

func applyUpdate(h *domain.Habit, req updateRequest) error {
	if req.Title != "" {
		h.Title = req.Title
	}
	if req.Icon != "" {
		h.Icon = req.Icon
	}
	if req.Color != "" {
		h.Color = req.Color
	}
	if req.Period != "" {
		h.Period = domain.Period(req.Period)
	}
	if req.TargetValue != nil {
		h.TargetValue = req.TargetValue
	}
	if req.Unit != nil {
		h.Unit = req.Unit
	}
	if req.GoalID != nil {
		if *req.GoalID == "" {
			h.GoalID = nil
		} else {
			goalID, err := parseGoalID(req.GoalID)
			if err != nil {
				return errors.New("invalid goal_id")
			}
			h.GoalID = goalID
		}
	}
	if req.ArchivedAt != nil {
		if *req.ArchivedAt == "" {
			h.ArchivedAt = nil
		} else {
			t, err := time.Parse(time.RFC3339, *req.ArchivedAt)
			if err != nil {
				return errors.New("invalid archived_at")
			}
			h.ArchivedAt = &t
		}
	}
	return nil
}

func (h *HabitHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid habit ID")
		return
	}

	err = h.habits.Delete(c.Request.Context(), habitID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Habit not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete habit")
		return
	}
	respondMessage(c, "Habit deleted successfully")
}

type toggleRequest struct {
	Date  string `json:"date" binding:"required"`
	Value *int   `json:"value"`
}

func (h *HabitHandler) Toggle(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid habit ID")
		return
	}

	var req toggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	habit, err := h.habits.GetOwnedByID(c.Request.Context(), habitID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Habit not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load habit")
		return
	}

	if err := h.applyToggle(c, habit, req); err != nil {
		slog.Error("toggle habit", "err", err, "habit_id", habit.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update completion")
		return
	}

	respondOK(c, toDTO(habit))
}

// applyToggle mutates both the DB and the in-memory habit so that the
// response reflects the new state without a second fetch.
func (h *HabitHandler) applyToggle(c *gin.Context, habit *domain.Habit, req toggleRequest) error {
	ctx := c.Request.Context()

	if habit.TargetValue != nil && req.Value != nil {
		if *req.Value <= 0 {
			if err := h.habits.RemoveProgress(ctx, habit.ID, req.Date); err != nil {
				return err
			}
			if err := h.habits.RemoveCompletion(ctx, habit.ID, req.Date); err != nil {
				return err
			}
			delete(habit.ProgressValues, req.Date)
			removeDate(&habit.CompletedDates, req.Date)
			return nil
		}
		if err := h.habits.SetProgress(ctx, habit.ID, req.Date, *req.Value); err != nil {
			return err
		}
		if habit.ProgressValues == nil {
			habit.ProgressValues = map[string]int{}
		}
		habit.ProgressValues[req.Date] = *req.Value
		if *req.Value >= *habit.TargetValue {
			if err := h.habits.AddCompletion(ctx, habit.ID, req.Date); err != nil {
				return err
			}
			addDateIfMissing(&habit.CompletedDates, req.Date)
		} else {
			if err := h.habits.RemoveCompletion(ctx, habit.ID, req.Date); err != nil {
				return err
			}
			removeDate(&habit.CompletedDates, req.Date)
		}
		return nil
	}

	for _, d := range habit.CompletedDates {
		if d == req.Date {
			if err := h.habits.RemoveCompletion(ctx, habit.ID, req.Date); err != nil {
				return err
			}
			removeDate(&habit.CompletedDates, req.Date)
			return nil
		}
	}
	if err := h.habits.AddCompletion(ctx, habit.ID, req.Date); err != nil {
		return err
	}
	addDateIfMissing(&habit.CompletedDates, req.Date)
	return nil
}

func addDateIfMissing(dates *[]string, date string) {
	for _, d := range *dates {
		if d == date {
			return
		}
	}
	// Repository returns dates in DESC order; prepend to preserve that contract.
	*dates = append([]string{date}, *dates...)
}

func removeDate(dates *[]string, date string) {
	for i, d := range *dates {
		if d == date {
			*dates = append((*dates)[:i], (*dates)[i+1:]...)
			return
		}
	}
}
