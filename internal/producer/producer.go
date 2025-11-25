package producer

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"

	"github.com/zahrashariati/task-manager/internal/models"
)

// ProducerInterface defines the interface for publishing events to Kafka
// This interface should be implemented by any Kafka producer
type ProducerInterface interface {
	PublishTaskScheduledEvent(task *models.Task) error
	Close()
}

// Producer handles publishing events to Kafka
type Producer struct {
	producer *kafka.Producer
	topic    string
}

// NewProducer creates a new Kafka producer
func NewProducer(brokerURL string, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokerURL,
		"client.id":         "task-manager-api",
		"acks":              "1", // Wait for leader acknowledgment (good balance)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

// PublishTaskScheduledEvent publishes a task scheduled event to Kafka
// This is fire-and-forget: errors are logged but don't fail the API request
func (p *Producer) PublishTaskScheduledEvent(task *models.Task) error {
	// Only publish if task has a due_date
	if task.DueDate == nil {
		return nil // No due date, nothing to schedule
	}

	// Create event
	event := models.TaskScheduledEvent{
		EventID:     uuid.New().String(),
		EventType:   "task_scheduled",
		TaskID:      task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		DueDate:     *task.DueDate,
		Timestamp:   time.Now(),
	}

	// Marshal to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Kafka (fire-and-forget)
	deliveryChan := make(chan kafka.Event)
	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Value: eventJSON,
		Key:   []byte(fmt.Sprintf("%d", task.ID)), // Use task ID as key for partitioning
	}, deliveryChan)

	if err != nil {
		log.Printf("failed to publish event to Kafka: %v", err)
		return err
	}

	// Check delivery status asynchronously (non-blocking)
	go func() {
		e := <-deliveryChan
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			log.Printf("failed to deliver message to Kafka: %v", m.TopicPartition.Error)
		} else {
			log.Printf("published event to Kafka: task_id=%d, event_id=%s, due_date=%s",
				task.ID, event.EventID, event.DueDate.Format("2006-01-02"))
		}
	}()

	return nil
}

// Close closes the producer
func (p *Producer) Close() {
	p.producer.Flush(15 * 1000) // Wait up to 15 seconds for pending messages
	p.producer.Close()
}

