// 1. Connect to Kafka server
// 2. Subscribe to `task-scheduled` topic
// 3. Read messages in a loop
// 4. For each message:
//    - Parse JSON
//    - Store for scheduler to process later
//    - Commit offset after processing

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/zahrashariati/task-manager/internal/scheduler"
)

type Consumer struct {
	consumer *kafka.Consumer
}

func NewConsumer(brokerURL string) *Consumer {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokerURL,
		"group.id":          "notifier",
		"auto.offset.reset": "earliest",
		"max.poll.interval.ms": "600000", // 10 minutes (default is 5 minutes)
		"enable.auto.commit": "false",      // Manual offset commit for reliability
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	
	// Subscribe to topic
	if err := c.SubscribeTopics([]string{"task_scheduled"}, nil); err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}
	
	return &Consumer{
		consumer: c,
	}
}

// Start reads messages from Kafka and stores them (doesn't process immediately)
// NOTE: ReadMessage(-1) blocks indefinitely waiting for messages. The select statement
// checks ctx.Done() but then immediately calls ReadMessage which blocks. If context is
// cancelled while waiting for a message, it won't respond until a message arrives.
// For true cancellation support, we'd need to use ReadMessage with a timeout and check
// context between reads, but this works for most cases since Kafka usually has messages.
func (c *Consumer) Start(ctx context.Context, taskStorage scheduler.StorageInterface) error {
	log.Println("Consumer started. Reading messages and storing for scheduler...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer")
			return ctx.Err()
		default:
			// Read message (blocking - waits until message arrives)
			// kafka.Message {
			//   TopicPartition: {Topic: "task_scheduled", Partition: 0, Offset: 123}
			//   Value: []byte(`{"event_id": "...", "task_id": 123}`)  // The actual data (JSON)
			//   Key: []byte("123")  // Optional: used for partitioning
			// }
			msg, err := c.consumer.ReadMessage(-1) // -1 means no timeout (blocking)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			
			var event models.TaskScheduledEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				// Skip bad messages but commit offset to avoid reprocessing
				c.CommitMessage(msg)
				continue
			}
			
			// Store event (scheduler will process it later)
			taskStorage.Add(event, msg)
			log.Printf("Stored event for task %d (due: %s)", event.TaskID, event.DueDate.Format("2006-01-02"))
			
			// Commit offset immediately after storing (prevents max poll interval error)
			// This tells Kafka we've processed the message, even though we haven't sent notification yet
			if err := c.CommitMessage(msg); err != nil {
				log.Printf("Warning: Failed to commit offset for task %d: %v", event.TaskID, err)
				// Continue anyway - scheduler will handle retry
			}
		}
	}
}

// Tells Kafka "I processed this message"
func (c *Consumer) CommitMessage(msg *kafka.Message) error {
	_, err := c.consumer.CommitMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to commit message: %v", err)
	}
	return nil
}

// CommitOffset commits a specific offset (for DB storage where we don't have Message object)
func (c *Consumer) CommitOffset(topic string, partition int32, offset int64) error {
	topicPartition := kafka.TopicPartition{
		Topic:     &topic,
		Partition: partition,
		Offset:    kafka.Offset(offset + 1), // Commit next offset (offset+1)
	}
	
	_, err := c.consumer.CommitOffsets([]kafka.TopicPartition{topicPartition})
	if err != nil {
		return fmt.Errorf("failed to commit offset: %v", err)
	}
	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
