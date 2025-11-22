// defines the data structures for the task manager API
package models

import (
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id" db:"user_id"` // User who created the task
	Title       string    `json:"title"`                //always have value (even if empty "")
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	Priority    string    `json:"priority"`
}

//Without tags, Go uses the field name directly (e.g., "ID" instead of "id").

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"` //If field is nil or empty, don't include it in JSON
	Completed   *bool   `json:"completed,omitempty"`
	Priority    *string `json:"priority,omitempty"`
} //`omitempty` tag means field won't appear in JSON if empty/nil - * for being optional
