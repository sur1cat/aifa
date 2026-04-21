package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sur1cat/aifa/goal-service/internal/domain"
	"github.com/sur1cat/aifa/goal-service/internal/events"
	"github.com/sur1cat/aifa/goal-service/internal/middleware"
	"github.com/sur1cat/aifa/goal-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoalHandler struct {
	goals           *repository.GoalRepository
	publisher       *events.Publisher
	defaultCurrency string
}

func NewGoalHandler(r *repository.GoalRepository, pub *events.Publisher, defaultCurrency string) *GoalHandler {
	return &GoalHandler{goals: r, publisher: pub, defaultCurrency: defaultCurrency}
}

type goalDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Icon          string   `json:"icon"`
	GoalType      string   `json:"goal_type"`
	TargetAmount  *float64 `json:"target_amount,omitempty"`
	CurrentAmount float64  `json:"current_amount"`
	Currency      string   `json:"currency"`
	Progress      float64  `json:"progress"`
	Deadline      *string  `json:"deadline,omitempty"`
	ArchivedAt    *string  `json:"archived_at,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

func toDTO(g *domain.Goal) goalDTO {
	fmtTime := func(t *time.Time) *string {
		if t == nil {
			return nil
		}
		s := t.Format(time.RFC3339)
		return &s
	}
	return goalDTO{
		ID:            g.ID.String(),
		Title:         g.Title,
		Icon:          g.Icon,
		GoalType:      string(g.GoalType),
		TargetAmount:  g.TargetAmount,
		CurrentAmount: g.CurrentAmount,
		Currency:      g.Currency,
		Progress:      g.Progress(),
		Deadline:      fmtTime(g.Deadline),
		ArchivedAt:    fmtTime(g.ArchivedAt),
		CreatedAt:     g.CreatedAt.Format(time.RFC3339),
	}
}

func (h *GoalHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	goals, err := h.goals.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list goals", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get goals")
		return
	}
	dtos := make([]goalDTO, len(goals))
	for i, g := range goals {
		dtos[i] = toDTO(g)
	}
	respondOK(c, dtos)
}

type createRequest struct {
	Title         string   `json:"title" binding:"required"`
	Icon          string   `json:"icon"`
	GoalType      string   `json:"goal_type"`
	TargetAmount  *float64 `json:"target_amount"`
	CurrentAmount *float64 `json:"current_amount"`
	Currency      string   `json:"currency"`
	Deadline      *string  `json:"deadline"`
}

func (h *GoalHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	deadline, err := parseOptionalTime(req.Deadline)
	if err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid deadline")
		return
	}

	icon := req.Icon
	if icon == "" {
		icon = "🎯"
	}
	goalType := domain.GoalType(req.GoalType)
	if goalType == "" {
		goalType = domain.GoalSavings
	} else if !goalType.Valid() {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid goal_type (savings|debt|purchase|investment)")
		return
	}
	currency := req.Currency
	if currency == "" {
		currency = h.defaultCurrency
	}
	current := 0.0
	if req.CurrentAmount != nil {
		current = *req.CurrentAmount
	}

	g := &domain.Goal{
		UserID:        userID,
		Title:         req.Title,
		Icon:          icon,
		GoalType:      goalType,
		TargetAmount:  req.TargetAmount,
		CurrentAmount: current,
		Currency:      currency,
		Deadline:      deadline,
	}
	if err := h.goals.Create(c.Request.Context(), g); err != nil {
		slog.Error("create goal", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create goal")
		return
	}
	respondCreated(c, toDTO(g))
}

func (h *GoalHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid goal ID")
		return
	}

	g, err := h.goals.GetOwnedByID(c.Request.Context(), goalID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Goal not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load goal")
		return
	}
	respondOK(c, toDTO(g))
}

type updateRequest struct {
	Title         string   `json:"title"`
	Icon          string   `json:"icon"`
	GoalType      string   `json:"goal_type"`
	TargetAmount  *float64 `json:"target_amount"`
	CurrentAmount *float64 `json:"current_amount"`
	Currency      string   `json:"currency"`
	Deadline      *string  `json:"deadline"`
	ArchivedAt    *string  `json:"archived_at"`
}

func (h *GoalHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid goal ID")
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	g, err := h.goals.GetOwnedByID(c.Request.Context(), goalID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Goal not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load goal")
		return
	}

	if err := applyUpdate(g, req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if err := h.goals.Update(c.Request.Context(), g); err != nil {
		slog.Error("update goal", "err", err, "goal_id", g.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update goal")
		return
	}
	respondOK(c, toDTO(g))
}

func applyUpdate(g *domain.Goal, req updateRequest) error {
	if req.Title != "" {
		g.Title = req.Title
	}
	if req.Icon != "" {
		g.Icon = req.Icon
	}
	if req.GoalType != "" {
		gt := domain.GoalType(req.GoalType)
		if !gt.Valid() {
			return errors.New("invalid goal_type")
		}
		g.GoalType = gt
	}
	if req.TargetAmount != nil {
		g.TargetAmount = req.TargetAmount
	}
	if req.CurrentAmount != nil {
		g.CurrentAmount = *req.CurrentAmount
	}
	if req.Currency != "" {
		g.Currency = req.Currency
	}
	if req.Deadline != nil {
		d, err := parseOptionalTime(req.Deadline)
		if err != nil {
			return errors.New("invalid deadline")
		}
		g.Deadline = d
	}
	if req.ArchivedAt != nil {
		t, err := parseOptionalTime(req.ArchivedAt)
		if err != nil {
			return errors.New("invalid archived_at")
		}
		g.ArchivedAt = t
	}
	return nil
}

func (h *GoalHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid goal ID")
		return
	}

	err = h.goals.Delete(c.Request.Context(), goalID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Goal not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete goal")
		return
	}

	h.publisher.PublishGoalDeleted(events.GoalDeleted{GoalID: goalID.String(), UserID: userID.String()})
	respondMessage(c, "Goal deleted successfully")
}

// parseOptionalTime maps an optional RFC3339 string (or "") to an
// optional time.Time. Returns a typed nil for nil/"" input.
func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
