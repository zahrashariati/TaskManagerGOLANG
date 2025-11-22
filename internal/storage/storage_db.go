package storage

import (
	"task_manager/internal/models"
	"task_manager/internal/repo"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// DBStorage implements Storage interface using database
type DBStorage struct {
	repo *repo.ScheduledNotificationRepository
}

// NewDBStorage creates a new database-backed storage
func NewDBStorage(repo *repo.ScheduledNotificationRepository) *DBStorage {
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
func (s *DBStorage) GetAllDueTasks() []*ScheduledTask {
	dbTasks, err := s.repo.GetDueTasks()
	if err != nil {
		// Log error and return empty slice
		return []*ScheduledTask{}
	}
	
	// Convert database tasks to ScheduledTask format
	tasks := make([]*ScheduledTask, 0, len(dbTasks))
	for _, dbTask := range dbTasks {
		// Reconstruct Kafka message metadata (we can't recreate the full Message object,
		// but we store partition/offset for committing)
		task := &ScheduledTask{
			Event: dbTask.ConvertToEvent(),
			// Store Kafka metadata in a way we can use for committing
			KafkaMetadata: &KafkaMetadata{
				Topic:     dbTask.KafkaTopic,
				Partition: dbTask.KafkaPartition,
				Offset:    dbTask.KafkaOffset,
			},
		}
		tasks = append(tasks, task)
	}
	
	return tasks
}

// Remove removes a task from storage (marks as notified in DB)
func (s *DBStorage) Remove(eventID string) {
	// Mark as notified instead of deleting (for audit trail)
	if err := s.repo.MarkAsNotified(eventID); err != nil {
		// Log error
		return
	}
	
	// Optionally delete after marking (uncomment if you want to delete)
	// s.repo.Delete(eventID)
}

// GetAll returns all tasks (for debugging) - not implemented for DB
func (s *DBStorage) GetAll() []*ScheduledTask {
	// Not needed for DB storage, but implement if needed
	return []*ScheduledTask{}
}


