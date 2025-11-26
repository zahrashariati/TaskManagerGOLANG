//connects to the database and creates the table if it doesn't exist

// 1. **Create `users` table**: Stores user accounts (username, email, password hash)
// 2. **Add `user_id` to `tasks` table**: Links each task to a user
// 3. **Add foreign key constraint**: Ensures `user_id` always references a valid user

package repo

import (
	"database/sql" //db interface
	"fmt"
	"log"
	_ "github.com/lib/pq" //postgres driver (blank import registers driver)
)

//sql.db is a struct that manages connection pool - pointer = share same connection pool with other parts of the code(better than copying)
//return error because connection could fail- returning so caller can handle it

func InitDB(connString string) (*sql.DB, error) { //Returns pointer to sql.DB and error
	// Open database connection
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Create users table
	createUsersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			is_active BOOLEAN DEFAULT TRUE,
			is_admin BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`
	if _, err := db.Exec(createUsersTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	// Create tasks table with user_id foreign key
	createTasksTable := `
		CREATE TABLE IF NOT EXISTS tasks ( 
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			priority TEXT DEFAULT 'medium',
			due_date TIMESTAMP
		)`
	if _, err := db.Exec(createTasksTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Add due_date column if it doesn't exist (for existing databases)
	if _, err := db.Exec("ALTER TABLE tasks ADD COLUMN IF NOT EXISTS due_date TIMESTAMP"); err != nil {
		// Ignore error if column already exists
		log.Printf("note: due_date column may already exist: %v", err)
	}


	createRTKTable := `	
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			revoked BOOLEAN DEFAULT FALSE
		)`

	if _,err := db.Exec(createRTKTable); err != nil {
		return nil, fmt.Errorf("failed to create refresh tokens table: %w", err)
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)"); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token)"); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Create scheduled_notifications table for Kafka events
	createScheduledNotificationsTable := `
		CREATE TABLE IF NOT EXISTS scheduled_notifications (
			id SERIAL PRIMARY KEY,
			event_id TEXT NOT NULL UNIQUE,
			event_type TEXT NOT NULL,
			task_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			due_date TIMESTAMP NOT NULL,
			event_timestamp TIMESTAMP NOT NULL,
			kafka_topic TEXT NOT NULL,
			kafka_partition INTEGER NOT NULL,
			kafka_offset BIGINT NOT NULL,
			notified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			processed_at TIMESTAMP
		)`
	if _, err := db.Exec(createScheduledNotificationsTable); err != nil {
		return nil, fmt.Errorf("failed to create scheduled_notifications table: %w", err)
	}

	// Create indexes for better query performance
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scheduled_notifications_due_date ON scheduled_notifications(due_date)"); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scheduled_notifications_notified ON scheduled_notifications(notified)"); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scheduled_notifications_event_id ON scheduled_notifications(event_id)"); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return db, nil
}
