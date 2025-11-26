package models

import "time"

// ScheduledNotification represents a stored notification in the database
type ScheduledNotification struct {
	ID              int
	EventID         string
	EventType       string
	TaskID          int
	UserID          int
	Title           string
	Description     string
	DueDate         time.Time
	EventTimestamp  time.Time
	KafkaTopic      string
	KafkaPartition  int32
	KafkaOffset     int64
	Notified        bool
	CreatedAt       time.Time
	ProcessedAt     *time.Time
}

// ConvertToEvent converts a ScheduledNotification back to TaskScheduledEvent
func (sn *ScheduledNotification) ConvertToEvent() TaskScheduledEvent {
	return TaskScheduledEvent{
		EventID:     sn.EventID,
		EventType:   sn.EventType,
		TaskID:      sn.TaskID,
		UserID:      sn.UserID,
		Title:       sn.Title,
		Description: sn.Description,
		DueDate:     sn.DueDate,
		Timestamp:   sn.EventTimestamp,
	}
}

