// internal/repository/task_repo.go
package repository

import (
	"taskflow/internal/database"
	"taskflow/internal/models"
)

type TaskRepository struct{}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

// Создание задачи
func (r *TaskRepository) Create(task *models.Task) error {
	return database.DB.Create(task).Error
}

// Получение всех задач пользователя - 👈 ИСПРАВЛЕНО!
func (r *TaskRepository) GetByUserID(userID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&tasks).Error
	return tasks, err
}

// Получение одной задачи (с проверкой принадлежности пользователю)
func (r *TaskRepository) GetUserTask(userID, taskID uint) (*models.Task, error) {
	var task models.Task
	err := database.DB.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error
	return &task, err
}

// Обновление задачи
func (r *TaskRepository) Update(task *models.Task) error {
	return database.DB.Save(task).Error
}

// Удаление задачи
func (r *TaskRepository) Delete(taskID uint) error {
	return database.DB.Delete(&models.Task{}, taskID).Error
}
