package services

import (
	"task_manager/internal/models"
)

// TaskServiceInterface defines the interface for task service operations
type TaskServiceInterface interface {
	GetAllTasks(userID int, showCompleted bool) ([]*models.Task, error)
	CreateTask(userID int, task *models.Task) error
	GetTaskByID(id int, userID int) (*models.Task, error)
	UpdateTask(id int, userID int, task *models.Task) (*models.Task, error)
	DeleteTask(id int, userID int) error
	CompleteTask(id int, userID int) (*models.Task, error)
}

// AuthServiceInterface defines the interface for auth service operations
type AuthServiceInterface interface {
	Register(user *models.User) (int, error)
	Login(user *models.User) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(id int, user *models.User) error
	DeleteUser(id int) error
	CreateRefreshToken(user_id int) (string, error)
	RevokeRefreshToken(user_id int, token string) error
	ValidateRefreshToken(token string) (int, error)
	RotateRefreshToken(oldToken string) (int, string, error)
}

