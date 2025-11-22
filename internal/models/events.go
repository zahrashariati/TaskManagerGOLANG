package models

// import "time"

// // TaskScheduledEvent represents an event sent to kafka when a task is scheduled
// type TaskScheduledEvent struct {
// 	EventID string `json:"event_id"`
// 	EventType string `json:"event_type"`
// 	TaskID int `json:"task_id"`
// 	UserID int `json:"user_id"`
// 	Title string `json:"title"`
// 	Description string `json:"description"`
// 	DueDate time.Time `json:"due_date"`
// 	Timestamp time.Time `json:"timestamp"`
// }

// // ScheduledNotification represents a stored notification in the database
// // This model should be in models/ folder, not repo/ folder
// type ScheduledNotification struct {
// 	ID              int
// 	EventID         string
// 	EventType       string
// 	TaskID          int
// 	UserID          int
// 	Title           string
// 	Description     string
// 	DueDate         time.Time
// 	EventTimestamp  time.Time
// 	KafkaTopic      string
// 	KafkaPartition  int32
// 	KafkaOffset     int64
// 	Notified        bool
// 	CreatedAt       time.Time
// 	ProcessedAt     *time.Time
// }

