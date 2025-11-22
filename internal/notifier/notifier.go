package notifier

import (
	"task_manager/internal/models"
	"fmt"
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
