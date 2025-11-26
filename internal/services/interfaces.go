package services

import (
	"time"
	"github.com/zahrashariati/task-manager/internal/models"
)

// TaskRepositoryInterface defines the interface for task repository operations
// (moved from repo/interfaces.go - interfaces should be defined where they're used)
type TaskRepositoryInterface interface {
	Create(task *models.Task) (int, error)
	GetTaskByID(id int) (*models.Task, error)
	Update(id int, userID int, task *models.Task) error
	Delete(id int, userID int) error
	GetAll(userID int, showCompleted bool) ([]models.Task, error)
	Complete(id int, userID int) error
}

// UserRepositoryInterface defines the interface for user repository operations
// (moved from repo/interfaces.go - interfaces should be defined where they're used)
type UserRepositoryInterface interface {
	CreateUser(user *models.User) (int, error)
	GetByUsername(username string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(id int, user *models.User) error
	DeleteUser(id int) error
}

// RTKRepositoryInterface defines the interface for refresh token repository operations
// (moved from repo/interfaces.go - interfaces should be defined where they're used)
type RTKRepositoryInterface interface {
	CreateRefreshToken(userID int, token string, expiresAt time.Time) (int, error)
	GetRefreshTokenByToken(token string) (*models.RTK, error)
	RevokeRefreshToken(token string) error
	GetRefreshTokenByUserID(userID int) (*models.RTK, error)
	RevokeAllRefreshTokensForUser(userID int) error
}

// CacheInterface defines the interface for cache operations
// (moved from cache/interfaces.go - interfaces should be defined where they're used)
type CacheInterface interface {
	GetTasks(showCompleted bool) ([]*models.Task, error)
	SetTasks(showCompleted bool, tasks []*models.Task) error
	Invalidate() error
	GetTask(id int) (*models.Task, error)
	SetTask(id int, task *models.Task) error
	InvalidateTask(id int) error
}

