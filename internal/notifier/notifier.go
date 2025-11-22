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
	
	// Compare dates only (ignore time) - check if task is due today or overdue
	dueDateOnly := time.Date(event.DueDate.Year(), event.DueDate.Month(), event.DueDate.Day(), 0, 0, 0, 0, event.DueDate.Location())
	todayOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	// Task is due if due_date is today or in the past
	if dueDateOnly.Before(todayOnly) || dueDateOnly.Equal(todayOnly) {
		return n.sendNotification(event)
	} else {
		fmt.Printf("Task %d is not due yet (due: %s, today: %s)\n", 
			event.TaskID, 
			dueDateOnly.Format("2006-01-02"), 
			todayOnly.Format("2006-01-02"))
		return nil
	}
}

func (n *Notifier) sendNotification(event models.TaskScheduledEvent) error {
	message := fmt.Sprintf("Notification: Task %d (%s) is due today: %s", event.TaskID, event.Title, event.DueDate.Format(time.RFC3339))
	fmt.Println(message)
	return nil
}
