package repo

import (
	"database/sql"
	"task_manager/internal/models"
)

type ScheduledNotificationRepository struct {
	db *sql.DB
}

func NewScheduledNotificationRepository(db *sql.DB) *ScheduledNotificationRepository {
	return &ScheduledNotificationRepository{db: db}
}

// Add stores a scheduled notification in the database
func (r *ScheduledNotificationRepository) Add(event models.TaskScheduledEvent, topic string, partition int32, offset int64) error {
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

// GetDueTasks returns all tasks that are due (today or in the past) and not yet notified
func (r *ScheduledNotificationRepository) GetDueTasks() ([]*models.ScheduledNotification, error) {
	query := `
		SELECT id, event_id, event_type, task_id, user_id, title, description, 
		       due_date, event_timestamp, kafka_topic, kafka_partition, kafka_offset, 
		       notified, created_at, processed_at
		FROM scheduled_notifications
		WHERE notified = FALSE
		AND DATE(due_date) <= CURRENT_DATE
		ORDER BY due_date ASC`
	
	rows, err := r.db.Query(query)
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

// MarkAsNotified marks a notification as sent and records the processing time
func (r *ScheduledNotificationRepository) MarkAsNotified(eventID string) error {
	query := `
		UPDATE scheduled_notifications
		SET notified = TRUE, processed_at = CURRENT_TIMESTAMP
		WHERE event_id = $1`
	
	_, err := r.db.Exec(query, eventID)
	return err
}

// Delete removes a notification from the database (after successful processing)
func (r *ScheduledNotificationRepository) Delete(eventID string) error {
	query := `DELETE FROM scheduled_notifications WHERE event_id = $1`
	_, err := r.db.Exec(query, eventID)
	return err
}
