// defines the data structures for the task manager API
package models

import (
	"time"
)

type Task struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id" db:"user_id"` // User who created the task
	Title       string     `json:"title"`                //always have value (even if empty "")
	Description string     `json:"description"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty" db:"due_date"` // Optional due date for scheduling
}

//Without tags, Go uses the field name directly (e.g., "ID" instead of "id").

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"` // Optional due date
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"` //If field is nil or empty, don't include it in JSON
	Completed   *bool      `json:"completed,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"` // Optional due date
} //`omitempty` tag means field won't appear in JSON if empty/nil - * for being optional
