// In-memory storage 
package storage

import (
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/zahrashariati/task-manager/internal/scheduler"
)

// Storage holds scheduled tasks in memory (thread-safe)
type Storage struct {
	mu     sync.RWMutex
	tasks  map[string]*scheduler.ScheduledTask // Key: event_id
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		tasks: make(map[string]*scheduler.ScheduledTask),
	}
}

// Add stores a scheduled task
func (s *Storage) Add(event models.TaskScheduledEvent, msg *kafka.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tasks[event.EventID] = &scheduler.ScheduledTask{
		Event:     event,
		Message:   msg,
		Processed: false,
	}
}

// GetAllDueTasks returns all tasks that are due (date and time) and not yet processed
func (s *Storage) GetAllDueTasks() []*scheduler.ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var dueTasks []*scheduler.ScheduledTask
	now := time.Now()
	
	for _, task := range s.tasks {
		// Task is due if due_date <= now (checks both date AND time) and not processed
		if !task.Processed && (task.Event.DueDate.Before(now) || task.Event.DueDate.Equal(now)) {
			dueTasks = append(dueTasks, task)
		}
	}
	
	return dueTasks
}

// Remove marks a task as processed (soft delete)
func (s *Storage) Remove(eventID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if task, exists := s.tasks[eventID]; exists {
		task.Processed = true
	}
}

// GetAll returns all tasks (for debugging)
func (s *Storage) GetAll() []*scheduler.ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*scheduler.ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
