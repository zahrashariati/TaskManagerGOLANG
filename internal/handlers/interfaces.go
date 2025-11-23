package handlers

import (
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/zahrashariati/task-manager/internal/models"
)

// TaskServiceInterface defines the interface for task service operations
// (moved from services/interfaces.go - interfaces should be defined where they're used)
type TaskServiceInterface interface {
	GetAllTasks(userID int, showCompleted bool) ([]*models.Task, error)
	CreateTask(userID int, task *models.Task) error
	GetTaskByID(id int, userID int) (*models.Task, error)
	UpdateTask(id int, userID int, task *models.Task) (*models.Task, error)
	DeleteTask(id int, userID int) error
	CompleteTask(id int, userID int) (*models.Task, error)
}

// AuthServiceInterface defines the interface for auth service operations
// (moved from services/interfaces.go - interfaces should be defined where they're used)
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

// ClaimsInterface defines the minimal interface for JWT claims
// (used to avoid circular dependency with auth package)
type ClaimsInterface interface {
	GetUserID() int
	GetUsername() string
}

// JWTServiceInterface defines the interface for JWT service operations
// (interfaces should be defined where they're used - handlers are the primary users)
type JWTServiceInterface interface {
	GenerateToken(userId int, username string, expiresAt time.Time) (string, error)
	GetTokenFromHeader(c *fiber.Ctx) (string, error)
	ValidateToken(tokenString string) (interface{}, error) // Returns ClaimsInterface (implemented by *auth.Claims)
}


