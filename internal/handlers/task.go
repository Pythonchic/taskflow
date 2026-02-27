// internal/handlers/task.go
package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"taskflow/internal/constants"
	"taskflow/internal/models"
	"taskflow/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	userRepo *repository.UserRepository
	taskRepo *repository.TaskRepository
}

func NewTaskHandler(userRepo *repository.UserRepository, taskRepo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{
		userRepo: userRepo,
		taskRepo: taskRepo,
	}
}

func (h *TaskHandler) getUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User not authenticated",
		})
		return 0, false
	}
	return userID.(uint), true
}

func (h *TaskHandler) getTaskID(c *gin.Context) (uint, error) {
	id := c.Param("id")
	idUint64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID format"})
		return 0, err
	}
	taskID := uint(idUint64)
	return taskID, nil
}

// GET /tasks
func (h *TaskHandler) TasksPage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	fmt.Printf("🔍 TasksPage: userID from context = %v (type: %T)\n", userID, userID)

	// Преобразуем в uint
	var uid uint
	switch v := userID.(type) {
	case uint:
		uid = v
	case float64:
		uid = uint(v)
	case int:
		uid = uint(v)
	default:
		fmt.Printf("❌ Unexpected type for userID: %T\n", v)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Invalid user ID type"})
		return
	}

	user, err := h.userRepo.GetByID(uid)
	if err != nil {
		fmt.Printf("❌ GetByID error: %v\n", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get user by id"})
		return
	}

	if user == nil {
		fmt.Printf("❌ User not found for ID: %d\n", uid)
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"FirstName": user.FirstName,
		"LastName":  user.LastName,
	})
}

// GET api/v1/tasks
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	// Получаем задачи из БД (модели Task)
	tasks, err := h.taskRepo.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to get tasks"})
		return
	}

	// Преобразуем Task → TaskResponse
	taskResponses := make([]models.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = models.TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Completed:   task.Completed,
			CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, models.TasksResponse{
		Tasks: taskResponses,
	})
}

// POST api/v1/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req models.CreateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Completed:   false,
	}

	// Сохраняем в БД
	if err := h.taskRepo.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to create task"})
		return
	}

	response := models.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(time.RFC3339), // форматируем дату
	}

	c.JSON(http.StatusCreated, response)
}

// PATCH /api/v1/tasks/:id {"title": "Новое название", "description": "Новое описание", "completed": true}
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	// Получаем ID
	id := c.Param("id")
	idUint64, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID format"})
		return
	}
	taskID := uint(idUint64)

	// Читаем запрос
	var req models.UpdateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем существующую задачу
	task, err := h.taskRepo.GetUserTask(userID, taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Обновляем только переданные поля
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Completed != nil {
		task.Completed = *req.Completed
	}

	// Всегда обновляем время
	task.UpdatedAt = time.Now()

	// Сохраняем
	if err := h.taskRepo.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to update task"})
		return
	}

	// Отвечаем
	c.JSON(http.StatusOK, models.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   task.CreatedAt.Format(constants.TimeFormat),
		UpdatedAt:   task.UpdatedAt.Format(constants.TimeFormat),
	})
}

// PUT /api/v1/tasks/:id/toggle
func (h *TaskHandler) ToggleTask(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID format"})
		return
	}
	taskID, err := h.getTaskID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	task, err := h.taskRepo.GetUserTask(userID, taskID)
	if err != nil {
		c.JSON(404, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Переключаем
	task.Completed = !task.Completed
	task.UpdatedAt = time.Now()

	h.taskRepo.Update(task)

	c.JSON(http.StatusOK, gin.H{
		"id":        task.ID,
		"completed": task.Completed,
		"message": fmt.Sprintf("Task marked as %s",
			map[bool]string{true: "completed", false: "pending"}[task.Completed]),
	})
}

// DELETE /api/v1/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid task ID format"})
		return
	}
	taskID, err := h.getTaskID(c)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	// Проверяем существование и принадлежность
	if _, err := h.taskRepo.GetUserTask(userID, uint(taskID)); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "Task not found"})
		return
	}

	if err := h.taskRepo.Delete(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete task",
		})
		return
	}
	c.JSON(http.StatusOK, models.MessageResponse{Message: "Task deleted successfully"})
}
