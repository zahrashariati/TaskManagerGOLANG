package scheduler

// import (
// 	"context"
// 	"log"
// 	"time"
// 	"task_manager/internal/storage"
// 	"task_manager/internal/notifier"
// 	"task_manager/internal/consumer"
// )

// // Scheduler runs periodic checks on stored tasks
// type Scheduler struct {
// 	storage      storage.StorageInterface
// 	notifier     *notifier.Notifier
// 	consumer     *consumer.Consumer
// 	checkInterval time.Duration
// }

// // NewScheduler creates a new scheduler
// func NewScheduler(storage storage.StorageInterface, notifier *notifier.Notifier, consumer *consumer.Consumer, interval time.Duration) *Scheduler {
// 	return &Scheduler{
// 		storage:       storage,
// 		notifier:      notifier,
// 		consumer:      consumer,
// 		checkInterval: interval,
// 	}
// }

// // Start starts the scheduler (runs checks periodically)
// func (s *Scheduler) Start(ctx context.Context) {
// 	ticker := time.NewTicker(s.checkInterval)
// 	defer ticker.Stop()
	
// 	log.Printf("Scheduler started. Checking every %v", s.checkInterval)
	
// 	// Run initial check immediately
// 	s.checkAndNotify()
	
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Println("Scheduler stopped")
// 			return
// 		case <-ticker.C:
// 			s.checkAndNotify()
// 		}
// 	}
// }

// // checkAndNotify checks stored tasks and sends notifications for due ones
// func (s *Scheduler) checkAndNotify() {
// 	log.Println("Running scheduled check...")
	
// 	// Get all due tasks
// 	dueTasks := s.storage.GetAllDueTasks()
	
// 	if len(dueTasks) == 0 {
// 		log.Println("No tasks due at this time")
// 		return
// 	}
	
// 	log.Printf("Found %d due task(s)", len(dueTasks))
	
// 	// Process each due task
// 	for _, task := range dueTasks {
// 		// Send notification
// 		if err := s.notifier.ProcessEvent(task.Event); err != nil {
// 			log.Printf("Error sending notification for task %d: %v", task.Event.TaskID, err)
// 			// Don't remove from storage if notification failed (will retry next hour)
// 			continue
// 		}
		
// 		// Commit Kafka message offset (mark as processed)
// 		var commitErr error
// 		if task.Message != nil {
// 			// In-memory storage: commit using Message object
// 			commitErr = s.consumer.CommitMessage(task.Message)
// 		} else if task.KafkaMetadata != nil {
// 			// DB storage: commit using partition/offset
// 			commitErr = s.consumer.CommitOffset(
// 				task.KafkaMetadata.Topic,
// 				task.KafkaMetadata.Partition,
// 				task.KafkaMetadata.Offset,
// 			)
// 		}
		
// 		if commitErr != nil {
// 			log.Printf("Error committing message for task %d: %v", task.Event.TaskID, commitErr)
// 			// Don't remove from storage if commit failed (will retry next hour)
// 			continue
// 		}
		
// 		// Remove from storage (successfully processed)
// 		s.storage.Remove(task.Event.EventID)
// 		log.Printf("Task %d processed and removed from storage", task.Event.TaskID)
// 	}
// }

