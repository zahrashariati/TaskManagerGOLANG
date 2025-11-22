package models

import "time"

// TaskScheduledEvent represents an event sent to kafka when a task is scheduled
type TaskScheduledEvent struct {
	EventID     string    `json:"event_id"`
	EventType   string    `json:"event_type"`
	TaskID      int       `json:"task_id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Timestamp   time.Time `json:"timestamp"`
}
