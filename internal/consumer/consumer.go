// // 1. Connect to Kafka
// // 2. Subscribe to `task-scheduled` topic
// // 3. Read messages in a loop
// // 4. For each message:
// //    - Parse JSON
// //    - Check if due_date matches today (or is in the past)
// //    - Print notification
// //    - Commit offset
package consumer

// package consumer

// import (
// 	"github.com/confluentinc/confluent-kafka-go/kafka"
// 	"context"
// 	"log"
// 	"encoding/json"
// 	"task_manager/internal/models"
// 	"task_manager/internal/storage"
// 	"fmt"
// )
// type Consumer struct {
// 	consumer *kafka.Consumer
// }

// func NewConsumer(brokerURL string) *Consumer {
// 	c, err := kafka.NewConsumer(&kafka.ConfigMap{
// 		"bootstrap.servers": brokerURL,
// 		"group.id":          "notifier",
// 		"auto.offset.reset": "earliest",
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to create consumer: %v", err)
// 	}
	
// 	// Subscribe to topic
// 	if err := c.SubscribeTopics([]string{"task_scheduled"}, nil); err != nil {
// 		log.Fatalf("Failed to subscribe to topic: %v", err)
// 	}
	
// 	return &Consumer{
// 		consumer: c,
// 	}
// }

// // Start reads messages from Kafka and stores them (doesn't process immediately)
// func (c *Consumer) Start(ctx context.Context, taskStorage storage.StorageInterface) error {
// 	log.Println("Consumer started. Reading messages and storing for scheduler...")
	
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("Context cancelled, stopping consumer")
// 			return ctx.Err()
// 		default:
// 			// Read message (blocking)
// 			msg, err := c.consumer.ReadMessage(-1) // -1 means no timeout (blocking)
// 			if err != nil {
// 				log.Printf("Error reading message: %v", err)
// 				continue
// 			}
			
// 			var event models.TaskScheduledEvent
// 			if err := json.Unmarshal(msg.Value, &event); err != nil {
// 				log.Printf("Error unmarshalling message: %v", err)
// 				// Skip bad messages but commit offset to avoid reprocessing
// 				c.CommitMessage(msg)
// 				continue
// 			}
			
// 			// Store event (scheduler will process it later)
// 			taskStorage.Add(event, msg)
// 			log.Printf("Stored event for task %d (due: %s)", event.TaskID, event.DueDate.Format("2006-01-02"))
// 		}
// 	}
// }

// func (c *Consumer) CommitMessage(msg *kafka.Message) error {
// 	_, err := c.consumer.CommitMessage(msg)
// 	if err != nil {
// 		return fmt.Errorf("failed to commit message: %v", err)
// 	}
// 	return nil
// }

// // CommitOffset commits a specific offset (for DB storage where we don't have Message object)
// func (c *Consumer) CommitOffset(topic string, partition int32, offset int64) error {
// 	topicPartition := kafka.TopicPartition{
// 		Topic:     &topic,
// 		Partition: partition,
// 		Offset:    kafka.Offset(offset + 1), // Commit next offset (offset+1)
// 	}
	
// 	_, err := c.consumer.CommitOffsets([]kafka.TopicPartition{topicPartition})
// 	if err != nil {
// 		return fmt.Errorf("failed to commit offset: %v", err)
// 	}
// 	return nil
// }

// func (c *Consumer) Close() error {
// 	return c.consumer.Close()
// }

