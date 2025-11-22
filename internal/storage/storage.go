package storage

import (
	"sync"
	"time"
	"task_manager/internal/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ScheduledTask stores an event with its Kafka message (for committing offset later)
type ScheduledTask struct {
	Event         models.TaskScheduledEvent
	Message       *kafka.Message  // For in-memory storage
	KafkaMetadata *KafkaMetadata   // For DB storage
}

// KafkaMetadata stores Kafka message metadata for committing offsets
type KafkaMetadata struct {
	Topic     string
	Partition int32
	Offset    int64
}

// Storage holds scheduled tasks in memory (thread-safe)
type Storage struct {
	mu     sync.RWMutex
	tasks  map[string]*ScheduledTask // Key: event_id
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		tasks: make(map[string]*ScheduledTask),
	}
}

// Add stores a scheduled task
func (s *Storage) Add(event models.TaskScheduledEvent, msg *kafka.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.tasks[event.EventID] = &ScheduledTask{
		Event:   event,
		Message: msg,
	}
}

// GetAllDueTasks returns all tasks that are due (date and time)
func (s *Storage) GetAllDueTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var dueTasks []*ScheduledTask
	now := time.Now()
	
	for _, task := range s.tasks {
		// Task is due if due_date <= now (checks both date AND time)
		if task.Event.DueDate.Before(now) || task.Event.DueDate.Equal(now) {
			dueTasks = append(dueTasks, task)
		}
	}
	
	return dueTasks
}

// Remove removes a task from storage (after processing)
func (s *Storage) Remove(eventID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, eventID)
}

// GetAll returns all tasks (for debugging)
func (s *Storage) GetAll() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
