package cache

import (
	"task_manager/internal/models"
)

// CacheInterface defines the interface for cache operations
type CacheInterface interface {
	GetTasks(showCompleted bool) ([]*models.Task, error)
	SetTasks(showCompleted bool, tasks []*models.Task) error
	Invalidate() error
	GetTask(id int) (*models.Task, error)
	SetTask(id int, task *models.Task) error
	InvalidateTask(id int) error
}

