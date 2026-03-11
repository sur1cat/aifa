package handler

import (
	"strconv"
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	taskRepo *repository.TaskRepository
}

func NewTaskHandler(taskRepo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{taskRepo: taskRepo}
}

type CreateTaskRequest struct {
	Title    string `json:"title" binding:"required"`
	Priority string `json:"priority"`
	DueDate  string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	IsCompleted *bool  `json:"is_completed"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
}

type TaskResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	IsCompleted bool   `json:"is_completed"`
	Priority    string `json:"priority"`
	DueDate     string `json:"due_date"`
	CreatedAt   string `json:"created_at"`
}

func toTaskResponse(t *domain.Task) *TaskResponse {
	return &TaskResponse{
		ID:          t.ID.String(),
		Title:       t.Title,
		IsCompleted: t.IsCompleted,
		Priority:    string(t.Priority),
		DueDate:     t.DueDate,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *TaskHandler) ListTasks(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	date := c.Query("date")

	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var tasks []*domain.Task
	var total int
	var err error

	if date != "" {
		tasks, total, err = h.taskRepo.GetByUserIDAndDate(c.Request.Context(), userID, date, limit, offset)
	} else {
		tasks, err = h.taskRepo.GetByUserID(c.Request.Context(), userID)
		total = len(tasks)
	}

	if err != nil {
		respondInternalError(c, "Failed to get tasks")
		return
	}

	response := make([]*TaskResponse, len(tasks))
	for i, task := range tasks {
		response[i] = toTaskResponse(task)
	}

	respondPaginated(c, response, PaginationMeta{Limit: limit, Offset: offset, Total: total})
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	priority := domain.TaskPriority(req.Priority)
	if priority == "" {
		priority = domain.TaskPriorityMedium
	}
	dueDate := req.DueDate
	if dueDate == "" {
		dueDate = time.Now().Format("2006-01-02")
	}

	task := &domain.Task{
		UserID:      userID,
		Title:       req.Title,
		IsCompleted: false,
		Priority:    priority,
		DueDate:     dueDate,
	}

	if err := h.taskRepo.Create(c.Request.Context(), task); err != nil {
		respondInternalError(c, "Failed to create task")
		return
	}

	respondCreated(c, toTaskResponse(task))
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid task ID")
		return
	}

	owns, err := h.taskRepo.VerifyOwnership(c.Request.Context(), taskID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Task not found")
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), taskID)
	if err != nil {
		respondNotFound(c, "Task not found")
		return
	}

	respondOK(c, toTaskResponse(task))
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid task ID")
		return
	}

	owns, err := h.taskRepo.VerifyOwnership(c.Request.Context(), taskID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Task not found")
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	task, err := h.taskRepo.GetByID(c.Request.Context(), taskID)
	if err != nil {
		respondNotFound(c, "Task not found")
		return
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.IsCompleted != nil {
		task.IsCompleted = *req.IsCompleted
	}
	if req.Priority != "" {
		task.Priority = domain.TaskPriority(req.Priority)
	}
	if req.DueDate != "" {
		task.DueDate = req.DueDate
	}

	if err := h.taskRepo.Update(c.Request.Context(), task); err != nil {
		respondInternalError(c, "Failed to update task")
		return
	}

	respondOK(c, toTaskResponse(task))
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid task ID")
		return
	}

	owns, err := h.taskRepo.VerifyOwnership(c.Request.Context(), taskID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Task not found")
		return
	}

	if err := h.taskRepo.Delete(c.Request.Context(), taskID); err != nil {
		respondInternalError(c, "Failed to delete task")
		return
	}

	respondMessage(c, "Task deleted successfully")
}

func (h *TaskHandler) ToggleTask(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid task ID")
		return
	}

	owns, err := h.taskRepo.VerifyOwnership(c.Request.Context(), taskID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Task not found")
		return
	}

	task, err := h.taskRepo.ToggleCompleted(c.Request.Context(), taskID)
	if err != nil {
		respondInternalError(c, "Failed to toggle task")
		return
	}

	respondOK(c, toTaskResponse(task))
}
