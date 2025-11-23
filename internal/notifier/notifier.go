package notifier

import (
	"github.com/zahrashariati/task-manager/internal/models"
	"github.com/zahrashariati/task-manager/internal/scheduler"
	"fmt"
	"log"
	"time"
)

type Notifier struct {}

func NewNotifier() *Notifier {
	return &Notifier{}
}

//process event check if task is due and send notification
func (n *Notifier) ProcessEvent(event models.TaskScheduledEvent) error {
	now := time.Now()
	
	// Compare full timestamp (date AND time) - task is due if due_date <= now
	if event.DueDate.Before(now) || event.DueDate.Equal(now) {
		return n.sendNotification(event)
	} else {
		fmt.Printf("Task %d is not due yet (due: %s, now: %s)\n", 
			event.TaskID, 
			event.DueDate.Format(time.RFC3339), 
			now.Format(time.RFC3339))
		return nil
	}
}

func (n *Notifier) sendNotification(event models.TaskScheduledEvent) error {
	message := fmt.Sprintf("Notification: Task %d (%s) is due today: %s", event.TaskID, event.Title, event.DueDate.Format(time.RFC3339))
	fmt.Println(message)
	return nil
}

// Callback processes all due tasks from storage, sends notifications, and removes successfully processed tasks
func (n *Notifier) Callback(storage scheduler.StorageInterface) {
	log.Println("Running scheduled check...")
	
	// Get all due tasks
	dueTasks := storage.GetAllDueTasks()
	
	if len(dueTasks) == 0 {
		log.Println("No tasks due at this time")
		return
	}
	
	log.Printf("Found %d due task(s)", len(dueTasks))
	
	// Process each due task
	for _, task := range dueTasks {
		// Send notification
		if err := n.ProcessEvent(task.Event); err != nil {
			log.Printf("Error sending notification for task %d: %v", task.Event.TaskID, err)
			// Don't remove from storage if notification failed (will retry next check)
			continue
		}
		
		// Remove from storage (successfully processed)
		// Note: Kafka offset was already committed when consumer stored the message
		storage.Remove(task.Event.EventID)
		log.Printf("Task %d processed and removed from storage", task.Event.TaskID)
	}
}
