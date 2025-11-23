// Runs periodic checks (every 15 minutes)
// Gets all due tasks from Storage
// Calls Notifier to send notifications
// Removes successfully processed tasks from Storage

package scheduler

import (
	"context"
	"log"
	"time"
	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ScheduledTask stores an event with its Kafka message (for committing offset later)
type ScheduledTask struct {
	Event         models.TaskScheduledEvent
	Message       *kafka.Message  // For in-memory storage
	KafkaMetadata *KafkaMetadata   // For DB storage
	Processed     bool             // Soft delete flag
}

// KafkaMetadata stores Kafka message metadata for committing offsets
type KafkaMetadata struct {
	Topic     string
	Partition int32
	Offset    int64
}

// StorageInterface defines methods for storing scheduled tasks
// (interfaces should be defined where they're used - scheduler is the primary user)
type StorageInterface interface {
	Add(event models.TaskScheduledEvent, msg *kafka.Message)
	GetAllDueTasks() []*ScheduledTask
	Remove(eventID string)
}

// CallbackFunc defines the callback function type for processing due tasks
// (interfaces should be defined where they're used)
type CallbackFunc func(storage StorageInterface)

// Scheduler runs periodic checks on stored tasks
type Scheduler struct {
	storage      StorageInterface
	callback     CallbackFunc
	checkInterval time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler(storage StorageInterface, callback CallbackFunc, interval time.Duration) *Scheduler {
	return &Scheduler{
		storage:       storage,
		callback:      callback,
		checkInterval: interval,
	}
}

// Start starts the scheduler (runs checks periodically)
// Uses channels with select to wait for EITHER cancellation OR ticker events
// This is non-blocking in the sense that we can respond to cancellation immediately
// even if we're waiting for the next tick. Without channels/select, we couldn't
// handle both events simultaneously.
func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()
	
	log.Printf("Scheduler started. Checking every %v", s.checkInterval)
	
	// Run initial check immediately
	s.callback(s.storage)
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		case <-ticker.C:
			s.callback(s.storage)
		}
	}
}
