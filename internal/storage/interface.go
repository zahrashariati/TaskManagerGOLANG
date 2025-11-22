package storage

import (
	"task_manager/internal/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// StorageInterface defines methods for storing scheduled tasks
type StorageInterface interface {
	Add(event models.TaskScheduledEvent, msg *kafka.Message)
	GetAllDueTasks() []*ScheduledTask
	Remove(eventID string)
}
