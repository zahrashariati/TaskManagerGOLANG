// Methods: Add, GetDueTasks, MarkAsNotified, Delete
package repo

import (
	"database/sql"
	"time"
	"github.com/zahrashariati/task-manager/internal/models"
)

// NewScheduledNotificationRepository creates a new scheduled notification repository
func NewScheduledNotificationRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Add stores a scheduled notification in the database
func (r *Repository) Add(event models.TaskScheduledEvent, topic string, partition int32, offset int64) error {
	query := `
		INSERT INTO scheduled_notifications 
		(event_id, event_type, task_id, user_id, title, description, due_date, event_timestamp, kafka_topic, kafka_partition, kafka_offset)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (event_id) DO NOTHING`
	
	_, err := r.db.Exec(
		query,
		event.EventID,
		event.EventType,
		event.TaskID,
		event.UserID,
		event.Title,
		event.Description,
		event.DueDate,
		event.Timestamp,
		topic,
		partition,
		offset,
	)
	return err
}

// GetDueTasks returns all tasks that are due (date and time) and not yet notified
// Uses UTC timezone for accurate comparison
func (r *Repository) GetDueTasks() ([]*models.ScheduledNotification, error) {
	// Use current UTC time for comparison (more reliable than database timezone functions)
	nowUTC := time.Now().UTC()
	
	query := `
		SELECT id, event_id, event_type, task_id, user_id, title, description, 
		       due_date, event_timestamp, kafka_topic, kafka_partition, kafka_offset, 
		       notified, created_at, processed_at
		FROM scheduled_notifications
		WHERE notified = FALSE
		AND due_date <= $1
		ORDER BY due_date ASC`
	
	rows, err := r.db.Query(query, nowUTC)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var tasks []*models.ScheduledNotification
	for rows.Next() {
		task := &models.ScheduledNotification{}
		var processedAt sql.NullTime
		
		err := rows.Scan(
			&task.ID,
			&task.EventID,
			&task.EventType,
			&task.TaskID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.DueDate,
			&task.EventTimestamp,
			&task.KafkaTopic,
			&task.KafkaPartition,
			&task.KafkaOffset,
			&task.Notified,
			&task.CreatedAt,
			&processedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if processedAt.Valid {
			task.ProcessedAt = &processedAt.Time
		}
		
		tasks = append(tasks, task)
	}
	
	return tasks, rows.Err()
}

// MarkAsNotified marks a notification as notified and records the processing time (soft delete)
func (r *Repository) MarkAsNotified(eventID string) error {
	query := `
		UPDATE scheduled_notifications
		SET notified = TRUE, processed_at = CURRENT_TIMESTAMP
		WHERE event_id = $1`
	
	_, err := r.db.Exec(query, eventID)
	return err
}

// DeleteNotification removes a notification from the database (hard delete - not recommended)
func (r *Repository) DeleteNotification(eventID string) error {
	query := `DELETE FROM scheduled_notifications WHERE event_id = $1`
	_, err := r.db.Exec(query, eventID)
	return err
}
