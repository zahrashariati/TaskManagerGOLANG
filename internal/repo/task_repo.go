//CRUD operations for the task manager API
package repo

import (
	"database/sql"
	"errors"
	"task_manager/internal/models"
)

type TaskRepository struct {
	db *sql.DB //holds db connection- pointer to sql.DB struct
}

func NewTaskRepository(db *sql.DB) *TaskRepository { //pointer to TaskRepository struct
	return &TaskRepository{db: db}//pointer so methods can access and modify the db connection
}


// r means method of TaskRepository struct - r is like self 
// output int(new task id) and error
func (r *TaskRepository) Create(task *models.Task) (int, error) {
	// 	// Step 1: Insert task
	// query := "INSERT INTO tasks (user_id, title, description, priority) VALUES ($1, $2, $3, $4)"
	// _, err := r.db.Exec(query, task.UserID, task.Title, task.Description, task.Priority)

	// // Step 2: Get the ID (requires another query!)
	// var id int
	// err = r.db.QueryRow(
	// 	"SELECT id FROM tasks WHERE user_id = $1 AND title = $2 AND created_at = (SELECT MAX(created_at) FROM tasks WHERE user_id = $1)",
	// 	task.UserID, task.Title
	// ).Scan(&id)

	//INSERT INTO tasks (user_id, title, description, priority) VALUES ($1, $2, $3, $4) RETURNING i
	// Local error definitions
	var ErrDatabaseQueryFailed = errors.New("failed to query database")

	query := "INSERT INTO tasks (user_id, title, description, priority) VALUES ($1, $2, $3, $4) RETURNING id"
	var id int
	err := r.db.QueryRow(query, task.UserID, task.Title, task.Description, task.Priority).Scan(&id)
	if err != nil {
		return 0, ErrDatabaseQueryFailed
	}
	return id, nil
}

func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
	// Local error definitions
	var (
		ErrTaskNotFound       = errors.New("task not found")
		ErrDatabaseQueryFailed = errors.New("failed to query database")
	)

	query := "SELECT id, user_id, title, description, completed, created_at, priority FROM tasks WHERE id = $1" //first parameter 
	var task models.Task 
	err := r.db.QueryRow(query, id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.Priority) //scan writes to variables and needs addresses to write
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTaskNotFound
		}
		return nil, ErrDatabaseQueryFailed
	}
	return &task, nil //returns address of task struct- more efficient than returning the struct itself
}

func (r *TaskRepository) Update(id int, userID int, task *models.Task) error {
	// Local error definitions
	var (
		ErrTaskNotFound      = errors.New("task not found")
		ErrDatabaseExecFailed = errors.New("failed to execute database query")
	)

	// First check if task exists and belongs to user
	existingTask, err := r.GetByID(id)
	if err != nil {
		return ErrTaskNotFound
	}
	// Verify task belongs to user
	if existingTask.UserID != userID {
		return ErrTaskNotFound // Don't reveal task exists for other users
	}
	
	query := "UPDATE tasks SET title = $1, description = $2, priority = $3 WHERE id = $4 AND user_id = $5"//safe way to pass values using placeholders
	_, err = r.db.Exec(query, task.Title, task.Description, task.Priority, id, userID)
	if err != nil {
		return ErrDatabaseExecFailed
	}
	return nil
}

func (r *TaskRepository) Delete(id int, userID int) error {
	// Local error definitions
	var (
		ErrTaskNotFound      = errors.New("task not found")
		ErrDatabaseExecFailed = errors.New("failed to execute database query")
	)

	query := "DELETE FROM tasks WHERE id = $1 AND user_id = $2"
	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		return ErrDatabaseExecFailed
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return ErrDatabaseExecFailed
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}
func (r *TaskRepository) GetAll(userID int, showCompleted bool) ([]models.Task, error) {
	// Local error definitions
	var ErrDatabaseQueryFailed = errors.New("failed to query database")

	query := "SELECT id, user_id, title, description, completed, created_at, priority FROM tasks WHERE user_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, ErrDatabaseQueryFailed
	}
	defer rows.Close()
	if !showCompleted {
		query += " AND completed = FALSE"
		rows, err = r.db.Query(query, userID)
		if err != nil {
			return nil, ErrDatabaseQueryFailed
		}
	}
	defer rows.Close()

	var tasks []models.Task
	// Loop through each row and scan into task struct
	for rows.Next() {
		var task models.Task  // NEW variable each iteration
		err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Completed, &task.CreatedAt, &task.Priority)
		if err != nil {
			return nil, ErrDatabaseQueryFailed
		}
		tasks = append(tasks, task)  // Add to slice
	}
	
	// Check for errors from iteration
	if err = rows.Err(); err != nil {
		return nil, ErrDatabaseQueryFailed
	}
	
	return tasks, nil
}

func (r *TaskRepository) Complete(id int, userID int) error {
	// Local error definitions
	var (
		ErrTaskNotFound      = errors.New("task not found")
		ErrDatabaseExecFailed = errors.New("failed to execute database query")
	)

	// First check if task exists and belongs to user
	existingTask, err := r.GetByID(id)
	if err != nil {
		return ErrTaskNotFound
	}
	// Verify task belongs to user
	if existingTask.UserID != userID {
		return ErrTaskNotFound // Don't reveal task exists for other users
	}
	
	query := "UPDATE tasks SET completed = TRUE WHERE id = $1 AND user_id = $2"
	_, err = r.db.Exec(query, id, userID)
	if err != nil {
		return ErrDatabaseExecFailed
	}
	return nil
}