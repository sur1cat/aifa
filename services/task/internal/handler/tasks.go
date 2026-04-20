package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/sur1cat/aifa/task-service/internal/domain"
	"github.com/sur1cat/aifa/task-service/internal/middleware"
	"github.com/sur1cat/aifa/task-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	defaultListLimit = 50
	maxListLimit     = 100
)

type TaskHandler struct {
	tasks *repository.TaskRepository
}

func NewTaskHandler(r *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{tasks: r}
}

type taskDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	CreatedAt   string `json:"created_at"`
}

func toDTO(t *domain.Task) taskDTO {
	return taskDTO{
		ID:          t.ID.String(),
		Title:       t.Title,
		IsCompleted: t.IsCompleted,
		Priority:    string(t.Priority),
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
	}
}

func (h *TaskHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	date := c.Query("date")
	limit, offset := paginationParams(c, defaultListLimit, maxListLimit)

	if date != "" {
		tasks, total, err := h.tasks.ListByUserAndDate(c.Request.Context(), userID, date, limit, offset)
		if err != nil {
			slog.Error("list tasks by date", "err", err, "user_id", userID, "date", date)
			respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get tasks")
			return
		}
		respondPaginated(c, toDTOs(tasks), paginationMeta{Limit: limit, Offset: offset, Total: total})
		return
	}

	tasks, err := h.tasks.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list tasks", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get tasks")
		return
	}
	respondPaginated(c, toDTOs(tasks), paginationMeta{Limit: limit, Offset: offset, Total: len(tasks)})
}

func toDTOs(tasks []*domain.Task) []taskDTO {
	dtos := make([]taskDTO, len(tasks))
	for i, t := range tasks {
		dtos[i] = toDTO(t)
	}
	return dtos
}

func paginationParams(c *gin.Context, defaultLimit, max int) (int, int) {
	limit, offset := defaultLimit, 0
	if s := c.Query("limit"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= max {
			limit = v
		}
	}
	if s := c.Query("offset"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 0 {
			offset = v
		}
	}
	return limit, offset
}

type createRequest struct {
	Title    string `json:"title" binding:"required"`
	Priority string `json:"priority"`
	DueDate  string `json:"due_date"`
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	priority := domain.Priority(req.Priority)
	if priority == "" {
		priority = domain.PriorityMedium
	}
	dueDate := req.DueDate
	if dueDate == "" {
		dueDate = time.Now().Format("2006-01-02")
	}

	t := &domain.Task{
		UserID:   userID,
		Title:    req.Title,
		Priority: priority,
		DueDate:  dueDate,
	}
	if err := h.tasks.Create(c.Request.Context(), t); err != nil {
		slog.Error("create task", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create task")
		return
	}
	respondCreated(c, toDTO(t))
}

func (h *TaskHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid task ID")
		return
	}

	t, err := h.tasks.GetOwnedByID(c.Request.Context(), taskID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load task")
		return
	}
	respondOK(c, toDTO(t))
}

type updateRequest struct {
	Title       string `json:"title"`
	IsCompleted *bool  `json:"is_completed"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
}

func (h *TaskHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid task ID")
		return
	}

	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	t, err := h.tasks.GetOwnedByID(c.Request.Context(), taskID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load task")
		return
	}

	applyUpdate(t, req)
	if err := h.tasks.Update(c.Request.Context(), t); err != nil {
		slog.Error("update task", "err", err, "task_id", t.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update task")
		return
	}
	respondOK(c, toDTO(t))
}

func applyUpdate(t *domain.Task, req updateRequest) {
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.IsCompleted != nil {
		t.IsCompleted = *req.IsCompleted
	}
	if req.Priority != "" {
		t.Priority = domain.Priority(req.Priority)
	}
	if req.DueDate != "" {
		t.DueDate = req.DueDate
	}
}

func (h *TaskHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid task ID")
		return
	}

	err = h.tasks.Delete(c.Request.Context(), taskID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Task not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete task")
		return
	}
	respondMessage(c, "Task deleted successfully")
}

func (h *TaskHandler) Toggle(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid task ID")
		return
	}

	t, err := h.tasks.ToggleCompleted(c.Request.Context(), taskID, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Task not found")
		return
	}
	if err != nil {
		slog.Error("toggle task", "err", err, "task_id", taskID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to toggle task")
		return
	}
	respondOK(c, toDTO(t))
}
