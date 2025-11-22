package repo

import (
	"time"
	"task_manager/internal/models"
)

// TaskRepositoryInterface defines the interface for task repository operations
type TaskRepositoryInterface interface {
	Create(task *models.Task) (int, error)
	GetByID(id int) (*models.Task, error)
	Update(id int, userID int, task *models.Task) error
	Delete(id int, userID int) error
	GetAll(userID int, showCompleted bool) ([]models.Task, error)
	Complete(id int, userID int) error
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	CreateUser(user *models.User) (int, error)
	GetByUsername(username string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	UpdateUser(id int, user *models.User) error
	DeleteUser(id int) error
}

// RTKRepositoryInterface defines the interface for refresh token repository operations
type RTKRepositoryInterface interface {
	CreateRefreshToken(userID int, token string, expiresAt time.Time) (int, error)
	GetRefreshTokenByToken(token string) (*models.RTK, error)
	RevokeRefreshToken(token string) error
	GetRefreshTokenByUserID(userID int) (*models.RTK, error)
	RevokeAllRefreshTokensForUser(userID int) error
}

