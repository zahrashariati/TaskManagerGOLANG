// Database storage in scheduled_notification_repo.go
//Add, GetAllDueTasks, Remove

package storage

import (
	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/zahrashariati/task-manager/internal/scheduler"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ScheduledNotificationRepositoryInterface defines the interface for scheduled notification repository operations
// (interfaces should be defined where they're used - storage package uses this)
type ScheduledNotificationRepositoryInterface interface {
	Add(event models.TaskScheduledEvent, topic string, partition int32, offset int64) error
	GetDueTasks() ([]*models.ScheduledNotification, error)
	MarkAsNotified(eventID string) error
}

// DBStorage implements Storage interface using database
type DBStorage struct {
	repo ScheduledNotificationRepositoryInterface
}

// NewDBStorage creates a new database-backed storage
func NewDBStorage(repo ScheduledNotificationRepositoryInterface) *DBStorage {
	return &DBStorage{repo: repo}
}

// Add stores a scheduled task in the database
func (s *DBStorage) Add(event models.TaskScheduledEvent, msg *kafka.Message) {
	// Extract Kafka metadata
	topic := *msg.TopicPartition.Topic
	partition := msg.TopicPartition.Partition
	offset := int64(msg.TopicPartition.Offset)
	
	// Store in database
	if err := s.repo.Add(event, topic, partition, offset); err != nil {
		// Log error but don't fail (could be duplicate event_id)
		// In production, you might want to handle this differently
		return
	}
}

// GetAllDueTasks returns all tasks that are due (from database)
func (s *DBStorage) GetAllDueTasks() []*scheduler.ScheduledTask {
	dbTasks, err := s.repo.GetDueTasks()
	if err != nil {
		// Log error and return empty slice
		return []*scheduler.ScheduledTask{}
	}
	
	// Convert database tasks to ScheduledTask format
	tasks := make([]*scheduler.ScheduledTask, 0, len(dbTasks))
	for _, dbTask := range dbTasks {
		// Reconstruct Kafka message metadata (we can't recreate the full Message object,
		// but we store partition/offset for committing)
		task := &scheduler.ScheduledTask{
			Event: dbTask.ConvertToEvent(),
			// Store Kafka metadata in a way we can use for committing
			KafkaMetadata: &scheduler.KafkaMetadata{
				Topic:     dbTask.KafkaTopic,
				Partition: dbTask.KafkaPartition,
				Offset:    dbTask.KafkaOffset,
			},
			Processed: dbTask.Notified,
		}
		tasks = append(tasks, task)
	}
	
	return tasks
}

// Remove removes a task from storage (marks as notified in DB - soft delete)
func (s *DBStorage) Remove(eventID string) {
	// Mark as notified instead of deleting (for audit trail)
	if err := s.repo.MarkAsNotified(eventID); err != nil {
		// Log error
		return
	}
}
