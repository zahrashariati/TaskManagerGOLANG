# Go Learning Guide - Task Manager Project 🎓

Welcome! This guide will help you build a task manager step-by-step while learning Go. **You'll write the code yourself** - I'll just guide you!

## 🎯 Project Goals

Build a REST API task manager backend with:
- ✅ HTTP Server with routes and handlers
- ✅ Database storage (CockroachDB - PostgreSQL compatible)
- ✅ Caching layer (Redis)
- ✅ JSON request/response handling
- ✅ All core Go concepts

---

## 📚 Phase 1: Basics & Setup (Day 1)

### Step 1.0: Set Up Project Structure
**Your task:**
Create a proper layered architecture structure:

```
booking_ticket/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── handlers/                 # HTTP handlers
│   ├── service/                  # Business logic
│   ├── repository/               # Database access
│   ├── models/                   # Data models
│   └── config/                   # Configuration
├── go.mod
└── README.md
```

**Your implementation:**
1. Create folders: `mkdir -p cmd/api internal/{handlers,service,repository,models,config}`
2. Move your `main.go` to `cmd/api/main.go`
3. Read `ARCHITECTURE.md` to understand the layers

**Learning points:**
- Project organization
- Layered architecture
- Separation of concerns
- Professional structure

---

### Step 1.1: Project Setup
**Your task:**
1. Create `cmd/api/main.go` with a `main()` function
2. Print "Hello, Task Manager API!" to verify it works
3. Run: `go run cmd/api/main.go`

**Learning points:**
- Go program entry point: `main()` function
- Package declaration: `package main`
- Project structure
- Running Go programs

---

### Step 1.2: Define Your Data Structures
**Your task:**
Create `internal/models/task.go` with:

1. **Task struct** (response model):
   - `ID` (int) - `json:"id"`
   - `Title` (string) - `json:"title"`
   - `Description` (string) - `json:"description"`
   - `Completed` (bool) - `json:"completed"`
   - `CreatedAt` (time.Time) - `json:"created_at"`
   - `Priority` (string) - `json:"priority"` ("low", "medium", "high")

2. **CreateTaskRequest struct** (request model):
   - `Title` (string) - `json:"title"` (required)
   - `Description` (string) - `json:"description"` (optional)
   - `Priority` (string) - `json:"priority"` (default: "medium")

3. **UpdateTaskRequest struct** (for PUT requests):
   - `Title` (string) - `json:"title"` (optional)
   - `Description` (string) - `json:"description"` (optional)
   - `Completed` (bool) - `json:"completed"` (optional)
   - `Priority` (string) - `json:"priority"` (optional)

**File structure:**
```go
// internal/models/task.go
package models

import "time"

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    Priority    string    `json:"priority"`
}

type CreateTaskRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
}

type UpdateTaskRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Completed   *bool   `json:"completed,omitempty"`
    Priority    *string `json:"priority,omitempty"`
}
```

**Hints:**
- Use `type Task struct { ... }`
- JSON tags are CRITICAL for API - they control how data is serialized
- Import `time` package for `time.Time`
- Use pointers (`*string`, `*bool`) for optional fields in UpdateTaskRequest
- `omitempty` tag means field won't appear in JSON if empty/nil

**Learning points:**
- Structs define custom types
- JSON tags for API serialization (very important!)
- Request/Response models (DTOs)
- Optional fields with pointers
- Time handling in Go
- Package organization

---

### Step 1.3: Create Configuration Layer
**Your task:**
Create `internal/config/config.go` to manage configuration:

```go
package config

import "os"

type Config struct {
    DatabaseURL string
    RedisURL    string
    Port        string
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", ""),
        RedisURL:    getEnv("REDIS_URL", ""),
        Port:        getEnv("PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

**Learning points:**
- Configuration management
- Environment variables
- Default values
- Centralized config

---

### Step 1.4: Understanding the Layers
**Read `ARCHITECTURE.md`** to understand:
- Handler layer (HTTP requests/responses)
- Service layer (business logic)
- Repository layer (database access)
- Models layer (data structures)

**You'll build each layer step by step!**

**Learning points:**
- Layered architecture
- Separation of concerns
- Professional structure

---

### Step 1.5: Learn Error Handling Basics
**Read `ERROR_HANDLING.md`** - This is CRITICAL!

**Key concepts to understand:**
- Go errors are values (not exceptions)
- Always check errors: `if err != nil`
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return errors early (don't nest)
- Use appropriate HTTP status codes

**Practice these patterns:**
```go
// Pattern 1: Check and return
result, err := someFunction()
if err != nil {
    return err
}

// Pattern 2: Check and wrap
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Pattern 3: Check and handle
result, err := someFunction()
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

**You'll use error handling in EVERY layer!**

**Learning points:**
- Error handling patterns
- Error wrapping
- Error checking
- Professional error handling

---

## 📚 Phase 2: Database Integration with CockroachDB (Day 2-3)

### Step 2.1: Set Up CockroachDB

**Option A: CockroachDB Cloud (Free Tier) - Recommended for Learning**
1. Sign up at https://cockroachlabs.com/cloud/
2. Create a free cluster
3. Get your connection string (looks like: `postgresql://user:password@host:26257/defaultdb?sslmode=require`)

**Option B: Local CockroachDB**
1. Install CockroachDB: https://www.cockroachlabs.com/docs/stable/install-cockroachdb.html
2. Start a local cluster: `cockroach start-single-node --insecure`
3. Connection string: `postgresql://root@localhost:26257/defaultdb?sslmode=disable`

**Learning points:**
- Cloud vs local development
- Connection strings
- Database setup

---

### Step 2.2: Install PostgreSQL Driver
**Your task:**
1. Run: `go get github.com/lib/pq`
2. Update `go.mod` if needed

**Note:** CockroachDB uses PostgreSQL protocol, so we use the PostgreSQL driver!

**Learning points:**
- Go modules and dependencies
- External packages
- Database drivers

---

### Step 2.3: Database Setup Function (with Error Handling!)
**Your task:**
Create a function `initDB(connString string) (*sql.DB, error)` that:
1. Opens a database connection using `sql.Open("postgres", connString)`
2. **Check for errors** - if `sql.Open` fails, return error
3. Tests the connection with `db.Ping()` - **check for errors**
4. Creates a `tasks` table if it doesn't exist - **check for errors**
5. Returns the database connection

**Pattern with proper error handling:**
```go
func initDB(connString string) (*sql.DB, error) {
    // Open connection
    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    // Test connection
    if err := db.Ping(); err != nil {
        db.Close()  // Clean up on error
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    // Create table
    query := `CREATE TABLE IF NOT EXISTS tasks (...)`
    if _, err := db.Exec(query); err != nil {
        db.Close()  // Clean up on error
        return nil, fmt.Errorf("failed to create table: %w", err)
    }
    
    return db, nil
}
```

**Notice:** Every step checks for errors and wraps them with context!

**Learning points:**
- Database connections
- PostgreSQL SQL syntax
- Connection testing
- **Error handling at every step**
- Error wrapping with context
- Resource cleanup on error

**SQL Table Schema (PostgreSQL/CockroachDB):**
```sql
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    priority VARCHAR(20) DEFAULT 'medium'
)
```

**Hints:**
- Import: `database/sql` and `_ "github.com/lib/pq"`
- Use `db.Exec()` to run CREATE TABLE
- Use `db.Ping()` to verify connection
- Handle errors properly
- Use `defer db.Close()` in main (but keep connection open for app lifetime)

**Connection String Format:**
```go
connString := "postgresql://user:password@host:26257/database?sslmode=require"
```

**Learning points:**
- Database connections
- PostgreSQL SQL syntax
- Connection testing
- Error handling patterns

---

### Step 2.4: Add Task to Database (Error Handling Practice!)
**Your task:**
Create a method `(tm *TaskManager) AddTask(title, description, priority string) (int, error)` that:
1. Inserts a new task into the database
2. Uses `INSERT INTO tasks ... RETURNING id`
3. **Handles errors properly** at each step
4. Returns the new task ID and error

**Pattern with comprehensive error handling:**
```go
func (r *TaskRepository) Create(title, desc, priority string) (int, error) {
    // Validate input (business rule)
    if title == "" {
        return 0, errors.New("title cannot be empty")
    }
    
    // Execute query
    var id int
    err := r.db.QueryRow(
        "INSERT INTO tasks (title, description, priority) VALUES ($1, $2, $3) RETURNING id",
        title, desc, priority,
    ).Scan(&id)
    
    // Check for errors
    if err != nil {
        return 0, fmt.Errorf("failed to insert task '%s': %w", title, err)
    }
    
    return id, nil
}
```

**Error handling checklist:**
- [ ] Check input validation errors
- [ ] Check database query errors
- [ ] Wrap errors with context
- [ ] Return appropriate error messages

**Hints:**
- Use `db.QueryRow()` with `RETURNING id` clause (PostgreSQL feature!)
- Use placeholders: `$1, $2, $3` for PostgreSQL (not `?`)
- Scan the returned ID: `var id int; err := row.Scan(&id)`
- **Always check `err` after `Scan()`**

**Example SQL:**
```sql
INSERT INTO tasks (title, description, priority) 
VALUES ($1, $2, $3) 
RETURNING id
```

**Learning points:**
- Database inserts
- PostgreSQL placeholders ($1, $2, etc.)
- RETURNING clause
- **Comprehensive error handling**
- Input validation
- Error wrapping

---

### Step 2.5: List Tasks from Database (Error Handling Everywhere!)
**Your task:**
Create a method `(tm *TaskManager) ListTasks(showCompleted bool) ([]Task, error)` that:
1. Queries tasks from database
2. Filters by `completed` status if needed
3. Scans rows into Task structs
4. **Handles ALL errors** at each step
5. Returns slice of tasks

**Pattern with complete error handling:**
```go
func (r *TaskRepository) GetAll(showCompleted bool) ([]models.Task, error) {
    // Build query
    query := "SELECT id, title, description, completed, created_at, priority FROM tasks"
    var rows *sql.Rows
    var err error
    
    if showCompleted {
        rows, err = r.db.Query(query)
    } else {
        query += " WHERE completed = FALSE"
        rows, err = r.db.Query(query)
    }
    
    // Check query error
    if err != nil {
        return nil, fmt.Errorf("failed to query tasks: %w", err)
    }
    defer rows.Close()  // Always close, even on error
    
    // Scan rows
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
        if err != nil {
            return nil, fmt.Errorf("failed to scan task row: %w", err)
        }
        tasks = append(tasks, t)
    }
    
    // Check for iteration errors
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }
    
    return tasks, nil
}
```

**Error handling checklist:**
- [ ] Check `Query()` error
- [ ] Use `defer rows.Close()` for cleanup
- [ ] Check `Scan()` error for each row
- [ ] Check `rows.Err()` after iteration
- [ ] Wrap all errors with context

**Hints:**
- Use `db.Query()` for multiple rows
- Use `rows.Scan()` to read data
- **Always use `defer rows.Close()`** (even if error occurs)
- Build WHERE clause dynamically based on `showCompleted`
- Use PostgreSQL placeholders: `$1` instead of `?`
- **Check `rows.Err()` after loop** - catches iteration errors

**Example SQL:**
```sql
SELECT id, title, description, completed, created_at, priority 
FROM tasks 
WHERE completed = FALSE
```

**Learning points:**
- Database queries
- Row scanning
- Resource cleanup with `defer`
- PostgreSQL syntax
- **Comprehensive error handling**
- Error checking at every step

---

### Step 2.6: Update and Delete Tasks (Error Handling Critical!)
**Your task:**
Create methods with proper error handling:
- `CompleteTask(id int) error` - UPDATE tasks SET completed = TRUE WHERE id = $1
- `DeleteTask(id int) error` - DELETE FROM tasks WHERE id = $1

**Pattern with error handling:**
```go
func (r *TaskRepository) Update(id int, req *models.UpdateTaskRequest) error {
    // Check if task exists first
    _, err := r.GetByID(id)
    if err != nil {
        return fmt.Errorf("task %d not found: %w", id, err)
    }
    
    // Build dynamic UPDATE query
    // (You'll implement this fully later)
    if req.Completed != nil {
        result, err := r.db.Exec("UPDATE tasks SET completed = $1 WHERE id = $2", *req.Completed, id)
        if err != nil {
            return fmt.Errorf("failed to update task %d: %w", id, err)
        }
        
        // Check if any rows were affected
        rowsAffected, err := result.RowsAffected()
        if err != nil {
            return fmt.Errorf("failed to get rows affected: %w", err)
        }
        if rowsAffected == 0 {
            return errors.New("no rows updated")
        }
    }
    
    return nil
}

func (r *TaskRepository) Delete(id int) error {
    result, err := r.db.Exec("DELETE FROM tasks WHERE id = $1", id)
    if err != nil {
        return fmt.Errorf("failed to delete task %d: %w", id, err)
    }
    
    // Check if task was actually deleted
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return errors.New("task not found")
    }
    
    return nil
}
```

**Error handling checklist:**
- [ ] Check `Exec()` error
- [ ] Check `RowsAffected()` error
- [ ] Verify rows were actually affected (not found case)
- [ ] Wrap errors with context (include ID in message)

**Hints:**
- Use `db.Exec()` for UPDATE/DELETE
- Use `$1` placeholder for PostgreSQL
- Use `TRUE`/`FALSE` instead of `1`/`0` for booleans
- **Check `RowsAffected()`** to verify operation succeeded
- Return specific error if no rows affected (task not found)

**Learning points:**
- UPDATE and DELETE operations
- Parameterized queries
- PostgreSQL boolean values
- **Error handling for write operations**
- Verifying operation success
- Distinguishing "not found" from other errors

---

## 📚 Phase 3: Redis Caching Layer (Day 4)

### Step 3.1: Set Up Redis

**Option A: Redis Cloud (Free Tier) - Recommended**
1. Sign up at https://redis.com/try-free/
2. Create a free database
3. Get your connection URL (looks like: `redis://default:password@host:port`)

**Option B: Local Redis**
1. Install Redis: https://redis.io/docs/getting-started/installation/
2. Start Redis: `redis-server`
3. Connection: `redis://localhost:6379`

**Learning points:**
- Redis setup
- Cloud vs local services
- Connection strings

---

### Step 3.2: Install Redis Client
**Your task:**
1. Run: `go get github.com/redis/go-redis/v9`
2. Update `go.mod` if needed

**Learning points:**
- Go modules
- Redis client libraries

---

### Step 3.3: Connect to Redis
**Your task:**
Create a function `initRedis(redisURL string) (*redis.Client, error)` that:
1. Parses the Redis URL
2. Creates a Redis client: `redis.NewClient(&redis.Options{...})`
3. Tests the connection with `client.Ping(ctx)`
4. Returns the Redis client

**Hints:**
- Import: `github.com/redis/go-redis/v9` and `context`
- Parse URL: `redis.ParseURL(redisURL)`
- Use context: `ctx := context.Background()`
- Test: `client.Ping(ctx)`

**Example:**
```go
opt, err := redis.ParseURL(redisURL)
if err != nil {
    return nil, err
}
client := redis.NewClient(opt)
```

**Learning points:**
- Redis connections
- Context usage
- URL parsing
- Connection testing

---

### Step 3.4: Add Redis to TaskManager
**Your task:**
1. Add `redisClient *redis.Client` field to TaskManager
2. Update constructor to accept Redis client
3. Store Redis client in TaskManager

**Learning points:**
- Struct composition
- Dependency injection pattern

---

### Step 3.5: Implement Cache-Aside Pattern with Redis
**Your task:**
Modify `ListTasks()` to:
1. **Check Redis cache first** - try to get cached data
2. **If cache miss** - query database
3. **Store in Redis** - cache the results with expiration
4. Return results

**Pattern:**
```
1. Try to get from Redis (key: "tasks:all" or "tasks:incomplete")
2. If found, unmarshal JSON and return
3. If not found, query database
4. Marshal results to JSON
5. Store in Redis with expiration (e.g., 5 minutes)
6. Return results
```

**Hints:**
- Use `client.Get(ctx, key)` to read from Redis
- Use `client.Set(ctx, key, value, expiration)` to write
- JSON marshal/unmarshal: `json.Marshal()` and `json.Unmarshal()`
- Use context: `ctx := context.Background()`
- Set expiration: `time.Minute * 5`

**Key naming convention:**
- `tasks:all` - all tasks
- `tasks:incomplete` - incomplete tasks only
- `task:{id}` - single task by ID

**Learning points:**
- Cache-aside pattern
- Redis operations (GET/SET)
- JSON serialization
- Cache expiration
- Key naming strategies

---

### Step 3.6: Cache Individual Tasks
**Your task:**
Create a helper method `getTaskFromCache(id int) (*Task, error)` that:
1. Tries to get task from Redis using key `task:{id}`
2. Unmarshals JSON if found
3. Returns task or error if not found

**Learning points:**
- Helper methods
- Cache key patterns
- Error handling

---

### Step 3.7: Invalidate Cache on Updates
**Your task:**
When tasks are added/updated/deleted:
1. Update the database
2. **Invalidate Redis cache** - delete relevant keys

**Hints:**
- Use `client.Del(ctx, keys...)` to delete keys
- Delete both specific task key and list keys
- Example: `client.Del(ctx, "task:1", "tasks:all", "tasks:incomplete")`

**Pattern:**
```
After AddTask:
  - Delete "tasks:all" and "tasks:incomplete" keys

After CompleteTask/DeleteTask:
  - Delete "task:{id}" key
  - Delete "tasks:all" and "tasks:incomplete" keys
```

**Learning points:**
- Cache invalidation strategies
- Maintaining data consistency
- Redis DEL operation

---

## 📚 Phase 4: REST API with HTTP Server (Day 5)

### Step 4.1: Set Up HTTP Server
**Your task:**
Create a basic HTTP server that listens on port 8080:

**Pattern:**
```go
import (
    "net/http"
    "log"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, API!"))
    })
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Your implementation:**
1. Import `net/http` package
2. Create a handler function that takes `http.ResponseWriter` and `*http.Request`
3. Use `http.HandleFunc()` to register routes
4. Use `http.ListenAndServe(":8080", nil)` to start server
5. Test with: `curl http://localhost:8080`

**Learning points:**
- HTTP server basics
- Handler functions
- Request/Response objects
- Starting a server

---

### Step 4.2: Create Handler Layer
**Your task:**
Create `internal/handlers/task_handler.go` with handler struct and methods:

**Pattern:**
```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "your-project/internal/models"
    "your-project/internal/service"
    "github.com/gorilla/mux"
)

type TaskHandler struct {
    service *service.TaskService
}

func NewTaskHandler(s *service.TaskService) *TaskHandler {
    return &TaskHandler{service: s}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
    showCompleted := r.URL.Query().Get("completed") == "true"
    
    tasks, err := h.service.GetAllTasks(showCompleted)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    var req models.CreateTaskRequest
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Validate
    if req.Title == "" {
        http.Error(w, "Title is required", http.StatusBadRequest)
        return
    }
    
    task, err := h.service.CreateTask(req.Title, req.Description, req.Priority)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    task, err := h.service.GetTaskByID(id)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    var req models.UpdateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    task, err := h.service.UpdateTask(id, &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    if err := h.service.DeleteTask(id); err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    w.WriteHeader(http.StatusNoContent)
}
```

**Hints:**
- Create handler struct that holds service reference
- Each handler method corresponds to one endpoint
- Parse request, validate, call service, return response
- Use proper HTTP status codes
- Handle errors appropriately

**Learning points:**
- Handler layer structure
- Dependency injection (service in handler)
- HTTP methods (GET, POST, PUT, DELETE)
- JSON encoding/decoding
- HTTP headers and status codes
- Error handling in handlers

---

### Step 4.3: (Moved to Step 4.2 - Handler Layer)
**Your task:**
Read and parse JSON from request body in `handleCreateTask`:

**Pattern:**
```go
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Parse JSON body
    var req CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Validate input
    if req.Title == "" {
        http.Error(w, "Title is required", http.StatusBadRequest)
        return
    }
    
    // Create task (call TaskManager method)
    taskID, err := tm.AddTask(req.Title, req.Description, req.Priority)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return created task
    task, _ := tm.GetTaskByID(taskID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}
```

**Hints:**
- Use `json.NewDecoder(r.Body).Decode(&struct)` to parse JSON
- Always `defer r.Body.Close()`
- Validate input before processing
- Return appropriate status codes

**Learning points:**
- Reading request body
- JSON decoding
- Input validation
- HTTP status codes (201 Created, 400 Bad Request, etc.)

---

### Step 4.4: (Moved to Step 4.2 - Handler Layer)
**Your task:**
Extract task ID from URL path (e.g., `/tasks/123`):

**Pattern:**
```go
import (
    "strings"
    "strconv"
)

func handleGetTask(w http.ResponseWriter, r *http.Request) {
    // Extract ID from URL: /tasks/123
    path := strings.TrimPrefix(r.URL.Path, "/tasks/")
    id, err := strconv.Atoi(path)
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    // Get task
    task, err := tm.GetTaskByID(id)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}
```

**Better approach - Using a router (recommended):**
```go
import "github.com/gorilla/mux"  // go get github.com/gorilla/mux

router := mux.NewRouter()
router.HandleFunc("/tasks/{id}", handleGetTask).Methods("GET")

// In handler:
vars := mux.Vars(r)
id, _ := strconv.Atoi(vars["id"])
```

**Hints:**
- Parse URL path manually OR use a router library
- `gorilla/mux` is popular and makes routing easier
- Convert string to int: `strconv.Atoi()`
- Handle errors when parsing

**Learning points:**
- URL path parsing
- String to int conversion
- Router libraries
- Path parameters

---

### Step 4.3: Create Service Layer (Business Logic)
**Your task:**
Create `internal/service/task_service.go`:

**Pattern:**
```go
package service

import (
    "your-project/internal/models"
    "your-project/internal/repository"
    "your-project/internal/cache"
)

type TaskService struct {
    repo  *repository.TaskRepository
    cache *cache.Cache
}

func NewTaskService(repo *repository.TaskRepository, cache *cache.Cache) *TaskService {
    return &TaskService{
        repo:  repo,
        cache: cache,
    }
}

func (s *TaskService) GetAllTasks(showCompleted bool) ([]models.Task, error) {
    // Check cache first
    if tasks, err := s.cache.GetTasks(showCompleted); err == nil {
        return tasks, nil
    }
    
    // If cache miss, get from database
    tasks, err := s.repo.GetAll(showCompleted)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    s.cache.SetTasks(showCompleted, tasks)
    
    return tasks, nil
}

func (s *TaskService) CreateTask(title, desc, priority string) (*models.Task, error) {
    // Validate priority
    if priority == "" {
        priority = "medium"
    }
    if priority != "low" && priority != "medium" && priority != "high" {
        return nil, errors.New("invalid priority")
    }
    
    // Create in database
    id, err := s.repo.Create(title, desc, priority)
    if err != nil {
        return nil, err
    }
    
    // Invalidate cache
    s.cache.Invalidate()
    
    // Return created task
    return s.repo.GetByID(id)
}

func (s *TaskService) GetTaskByID(id int) (*models.Task, error) {
    // Check cache first
    if task, err := s.cache.GetTask(id); err == nil {
        return task, nil
    }
    
    // Get from database
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // Cache it
    s.cache.SetTask(id, task)
    
    return task, nil
}

func (s *TaskService) UpdateTask(id int, req *models.UpdateTaskRequest) (*models.Task, error) {
    // Update in database
    err := s.repo.Update(id, req)
    if err != nil {
        return nil, err
    }
    
    // Invalidate cache
    s.cache.Invalidate()
    
    // Return updated task
    return s.repo.GetByID(id)
}

func (s *TaskService) DeleteTask(id int) error {
    err := s.repo.Delete(id)
    if err != nil {
        return err
    }
    
    // Invalidate cache
    s.cache.Invalidate()
    
    return nil
}
```

**Learning points:**
- Service layer (business logic)
- Cache management
- Validation
- Error handling
- Orchestration between repository and cache

---

### Step 4.4: Create Repository Layer (Database Access)
**Your task:**
Create `internal/repository/task_repository.go`:

**Pattern:**
```go
package repository

import (
    "database/sql"
    "your-project/internal/models"
)

type TaskRepository struct {
    db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
    return &TaskRepository{db: db}
}

func (r *TaskRepository) GetAll(showCompleted bool) ([]models.Task, error) {
    query := "SELECT id, title, description, completed, created_at, priority FROM tasks"
    if !showCompleted {
        query += " WHERE completed = FALSE"
    }
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var t models.Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }
    
    return tasks, nil
}

func (r *TaskRepository) Create(title, desc, priority string) (int, error) {
    var id int
    err := r.db.QueryRow(
        "INSERT INTO tasks (title, description, priority) VALUES ($1, $2, $3) RETURNING id",
        title, desc, priority,
    ).Scan(&id)
    return id, err
}

func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
    var t models.Task
    err := r.db.QueryRow(
        "SELECT id, title, description, completed, created_at, priority FROM tasks WHERE id = $1",
        id,
    ).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
    
    if err == sql.ErrNoRows {
        return nil, errors.New("task not found")
    }
    if err != nil {
        return nil, err
    }
    
    return &t, nil
}

func (r *TaskRepository) Update(id int, req *models.UpdateTaskRequest) error {
    // Build dynamic UPDATE query based on what's provided
    // This is more complex - you'll build it step by step
    // For now, simple version:
    if req.Completed != nil {
        _, err := r.db.Exec("UPDATE tasks SET completed = $1 WHERE id = $2", *req.Completed, id)
        return err
    }
    return nil
}

func (r *TaskRepository) Delete(id int) error {
    result, err := r.db.Exec("DELETE FROM tasks WHERE id = $1", id)
    if err != nil {
        return err
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return errors.New("task not found")
    }
    
    return nil
}
```

**Learning points:**
- Repository layer (data access)
- SQL queries
- Row scanning
- Error handling
- Database operations

---

### Step 4.5: Set Up Routes with Router
**Your task:**
Create routes in `cmd/api/main.go` or separate router file:

**Pattern with gorilla/mux:**
```go
import (
    "github.com/gorilla/mux"
    "your-project/internal/handlers"
    "your-project/internal/service"
    "your-project/internal/repository"
)

func setupRoutes(db *sql.DB, redisClient *redis.Client) *mux.Router {
    // Initialize layers (bottom to top)
    repo := repository.NewTaskRepository(db)
    cache := cache.NewCache(redisClient)
    svc := service.NewTaskService(repo, cache)
    handler := handlers.NewTaskHandler(svc)
    
    router := mux.NewRouter()
    
    // Routes
    router.HandleFunc("/tasks", handler.GetTasks).Methods("GET")
    router.HandleFunc("/tasks", handler.CreateTask).Methods("POST")
    router.HandleFunc("/tasks/{id}", handler.GetTask).Methods("GET")
    router.HandleFunc("/tasks/{id}", handler.UpdateTask).Methods("PUT")
    router.HandleFunc("/tasks/{id}", handler.DeleteTask).Methods("DELETE")
    
    // Health check
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
    
    return router
}
```

**Pattern with built-in (no external deps):**
```go
func setupRoutes(tm *TaskManager) {
    http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            handleGetTasks(w, r, tm)
        case http.MethodPost:
            handleCreateTask(w, r, tm)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })
    
    http.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
        // Extract ID and route to appropriate handler
        handleTaskByID(w, r, tm)
    })
}
```

**Your implementation:**
1. Install router: `go get github.com/gorilla/mux` (recommended)
2. Create `setupRoutes()` function
3. Register all endpoints
4. Return router and use in `main()`

**Learning points:**
- Route organization
- HTTP method routing
- Router libraries
- Clean code structure

---

### Step 4.6: Wire Everything Together in main.go
**Your task:**
In `cmd/api/main.go`, wire all layers together:

**Pattern:**
```go
package main

import (
    "log"
    "net/http"
    
    "your-project/internal/config"
    "your-project/internal/handlers"
    "your-project/internal/service"
    "your-project/internal/repository"
    "your-project/internal/cache"
    "github.com/gorilla/mux"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    if cfg.DatabaseURL == "" {
        log.Fatal("DATABASE_URL not set")
    }
    if cfg.RedisURL == "" {
        log.Fatal("REDIS_URL not set")
    }
    
    // Initialize database
    db, err := initDB(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Initialize Redis
    redisClient, err := initRedis(cfg.RedisURL)
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer redisClient.Close()
    
    // Initialize layers (bottom to top)
    repo := repository.NewTaskRepository(db)
    cacheService := cache.NewCache(redisClient)
    svc := service.NewTaskService(repo, cacheService)
    handler := handlers.NewTaskHandler(svc)
    
    // Set up routes
    router := setupRoutes(handler)
    
    // Start server
    log.Printf("Server starting on :%s", cfg.Port)
    log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func setupRoutes(handler *handlers.TaskHandler) *mux.Router {
    router := mux.NewRouter()
    
    // Task routes
    router.HandleFunc("/tasks", handler.GetTasks).Methods("GET")
    router.HandleFunc("/tasks", handler.CreateTask).Methods("POST")
    router.HandleFunc("/tasks/{id}", handler.GetTask).Methods("GET")
    router.HandleFunc("/tasks/{id}", handler.UpdateTask).Methods("PUT")
    router.HandleFunc("/tasks/{id}", handler.DeleteTask).Methods("DELETE")
    
    // Health check
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    })
    
    return router
}
```

**Hints:**
- Initialize layers in order: Repository → Cache → Service → Handler
- Use dependency injection (pass dependencies to constructors)
- Always `defer` close connections
- Use config for all settings

**Learning points:**
- Dependency injection
- Layer initialization order
- Program structure
- Environment variables
- Error handling in main
- Resource cleanup
- HTTP server lifecycle
- Complete layered architecture

---

### Step 4.7: Add CORS Middleware (Optional but Important)
**Your task:**
Add CORS headers so your API can be called from browsers:

**Pattern:**
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Usage:
router := mux.NewRouter()
router.Use(corsMiddleware)
```

**Learning points:**
- Middleware pattern
- CORS headers
- HTTP OPTIONS method
- Request/response modification

---

## 📚 Phase 5: Testing (Day 6)

### Step 5.1: Write Unit Tests
**Your task:**
Create `main_test.go` with tests for:
- Adding tasks
- Listing tasks
- Completing tasks
- Deleting tasks
- Cache functionality (Redis)

**Hints:**
- Use `testing` package
- Test functions: `func TestXxx(t *testing.T)`
- Use test database (separate CockroachDB database for tests)
- Use test Redis instance or mock Redis
- Clean up after tests (delete test data)
- Use `t.Skip()` if Redis/DB not available

**For testing:**
- Set up test database: `testdb` database in CockroachDB
- Use test Redis: separate Redis database (db 1 instead of 0)
- Or use `redis.NewClient()` with test-specific config

**Learning points:**
- Go testing framework
- Test organization
- Test databases
- Test isolation
- Skipping tests

---

## 📚 Phase 6: Advanced Features - Concurrency (Day 7)

### Step 6.1: Understanding Goroutines
**What are goroutines?**
- Lightweight threads managed by Go runtime
- Start with `go` keyword: `go function()`
- Much cheaper than OS threads
- Perfect for concurrent operations

**When to use goroutines in your task manager:**
- ✅ Batch operations (adding multiple tasks)
- ✅ Concurrent reads (fetching multiple tasks)
- ✅ Background operations (cache warming)
- ✅ Parallel operations (cache + DB simultaneously)

**Learning points:**
- Goroutines basics
- When concurrency helps
- Performance benefits

---

### Step 6.2: Basic Goroutine - Batch Task Creation
**Your task:**
Add a `demo` command that creates multiple tasks concurrently:

**Pattern:**
```go
var wg sync.WaitGroup
results := make(chan int, 10) // Buffer for 10 results

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        taskID, _ := tm.AddTask(...)
        results <- taskID
    }(i)
}

wg.Wait()
close(results)
```

**Your implementation:**
1. Create a `demo` command
2. Use `sync.WaitGroup` to wait for all goroutines
3. Use a buffered channel to collect task IDs
4. Launch 5-10 goroutines to add tasks concurrently
5. Print results from the channel

**Hints:**
- Import `sync` package
- `wg.Add(1)` before each goroutine
- `defer wg.Done()` inside goroutine
- `wg.Wait()` to wait for all
- Close channel when done: `close(results)`

**Learning points:**
- Starting goroutines
- WaitGroups for coordination
- Buffered channels
- Collecting results

---

### Step 6.3: Concurrent Reads - Fetch Multiple Tasks
**Your task:**
Create a method `GetTasksByIDs(ids []int) ([]Task, error)` that:
1. Fetches multiple tasks concurrently by their IDs
2. Uses goroutines for each ID lookup
3. Collects results via channel
4. Returns all tasks found

**Pattern:**
```go
results := make(chan Task, len(ids))
var wg sync.WaitGroup

for _, id := range ids {
    wg.Add(1)
    go func(taskID int) {
        defer wg.Done()
        task, err := tm.GetTaskByID(taskID)
        if err == nil {
            results <- task
        }
    }(id)
}

wg.Wait()
close(results)

// Collect from channel
var tasks []Task
for task := range results {
    tasks = append(tasks, task)
}
```

**Why this helps:**
- If each DB query takes 50ms, fetching 10 tasks sequentially = 500ms
- With goroutines, all queries run in parallel ≈ 50ms total!

**Learning points:**
- Concurrent database reads
- Performance improvement
- Error handling in goroutines
- Channel collection patterns

---

### Step 6.4: Cache Warming with Goroutines
**Your task:**
Add a background cache warming function that:
1. Runs in a goroutine
2. Pre-loads frequently accessed data
3. Runs periodically (every few minutes)

**Pattern:**
```go
func (tm *TaskManager) StartCacheWarming() {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            // Pre-load all tasks into cache
            tm.ListTasks(true)
            fmt.Println("Cache warmed!")
        }
    }()
}
```

**Usage:**
- Call `tm.StartCacheWarming()` in `main()`
- Runs in background forever
- Keeps cache fresh automatically

**Learning points:**
- Background goroutines
- Tickers for periodic tasks
- Long-running processes
- Resource cleanup

---

### Step 6.5: Parallel Cache and DB Lookup
**Your task:**
Optimize `GetTaskByID()` to check cache and DB simultaneously:
1. Launch two goroutines: one for cache, one for DB
2. Use `select` to take whichever responds first
3. If cache responds first, cancel DB query (optional)

**Pattern:**
```go
cacheChan := make(chan Task, 1)
dbChan := make(chan Task, 1)
errChan := make(chan error, 2)

// Try cache
go func() {
    task, err := tm.getFromCache(id)
    if err == nil {
        cacheChan <- task
    } else {
        errChan <- err
    }
}()

// Try database
go func() {
    task, err := tm.queryFromDB(id)
    if err == nil {
        dbChan <- task
    } else {
        errChan <- err
    }
}()

// Take whichever responds first
select {
case task := <-cacheChan:
    return task, nil // Cache hit - fastest!
case task := <-dbChan:
    return task, nil // DB result
case err := <-errChan:
    // Handle error
}
```

**Why this helps:**
- If cache is slow (network delay), DB might respond first
- User gets fastest response possible
- Better user experience

**Learning points:**
- Parallel operations
- `select` statement
- Multiple channels
- Race conditions (good kind!)

---

### Step 6.6: Worker Pool Pattern (Advanced)
**Your task (bonus):**
Implement a worker pool for processing multiple tasks:
1. Create N worker goroutines (e.g., 5 workers)
2. Send tasks to workers via channel
3. Workers process tasks concurrently
4. Collect results

**Pattern:**
```go
const numWorkers = 5
taskChan := make(chan int, 10) // Task IDs to process
resultChan := make(chan Task, 10)

// Start workers
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func(workerID int) {
        defer wg.Done()
        for taskID := range taskChan {
            task, _ := tm.GetTaskByID(taskID)
            resultChan <- task
        }
    }(i)
}

// Send work
for _, id := range taskIDs {
    taskChan <- id
}
close(taskChan)

// Wait for workers
go func() {
    wg.Wait()
    close(resultChan)
}()
```

**Use case:**
- Processing bulk operations
- Rate limiting (control concurrency)
- Resource management

**Learning points:**
- Worker pool pattern
- Controlling concurrency
- Channel closing
- Advanced patterns

---

### Step 6.7: Add Concurrent Endpoint (Bonus)
**Your task:**
Create a `/tasks/batch` endpoint that creates multiple tasks concurrently:
1. Accept JSON array of tasks in request body
2. Use goroutines to create all tasks in parallel
3. Return all created tasks
4. Measure and return performance stats

**Pattern:**
```go
func handleBatchCreateTasks(w http.ResponseWriter, r *http.Request) {
    var requests []CreateTaskRequest
    json.NewDecoder(r.Body).Decode(&requests)
    
    start := time.Now()
    results := make(chan Task, len(requests))
    var wg sync.WaitGroup
    
    for _, req := range requests {
        wg.Add(1)
        go func(cr CreateTaskRequest) {
            defer wg.Done()
            id, _ := tm.AddTask(cr.Title, cr.Description, cr.Priority)
            task, _ := tm.GetTaskByID(id)
            results <- task
        }(req)
    }
    
    wg.Wait()
    close(results)
    
    var tasks []Task
    for task := range results {
        tasks = append(tasks, task)
    }
    
    duration := time.Since(start)
    
    response := map[string]interface{}{
        "tasks": tasks,
        "count": len(tasks),
        "duration_ms": duration.Milliseconds(),
    }
    
    json.NewEncoder(w).Encode(response)
}
```

**Learning points:**
- API endpoint with concurrency
- Performance measurement
- Batch operations
- HTTP response with metadata

---

### Step 6.8: Add Search Endpoint
**Your task:**
Add a `GET /tasks/search?q=keyword` endpoint that:
1. Searches tasks by title/description
2. Uses SQL LIKE query
3. Returns matching tasks as JSON

**Pattern:**
```go
func handleSearchTasks(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    if query == "" {
        http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
        return
    }
    
    tasks, err := tm.SearchTasks(query)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}
```

**Learning points:**
- Query parameters
- SQL LIKE queries
- String matching
- Filtering data
- URL query parsing

---

## 🎯 Bonus Challenges

1. **Pagination**: Add `?page=1&limit=10` to GET /tasks endpoint
2. **Filtering**: Add `?completed=true` or `?priority=high` query parameters
3. **Authentication**: Add JWT token authentication
4. **Validation**: Add input validation middleware
5. **Logging**: Add request logging middleware
6. **Statistics**: Add GET /stats endpoint with completion rates
7. **Due Dates**: Add due date field and reminders
8. **Rate Limiting**: Prevent too many requests
9. **Swagger Docs**: Add API documentation
10. **Docker**: Containerize your API

---

## 📖 Key Go Concepts You'll Learn

### Basics
- ✅ Variables and types
- ✅ Functions and methods
- ✅ Structs and interfaces
- ✅ Slices and maps
- ✅ Pointers

### Advanced
- ✅ Error handling
- ✅ HTTP servers and routing
- ✅ JSON encoding/decoding
- ✅ REST API design
- ✅ Goroutines and channels
- ✅ Mutexes and synchronization
- ✅ Database operations
- ✅ Testing

### Best Practices
- ✅ Code organization
- ✅ Error wrapping
- ✅ Resource cleanup
- ✅ Thread safety
- ✅ Package structure

---

## 🚀 Getting Started Right Now

**Start here:**
1. Create `main.go` with empty `main()` function
2. Print "Hello, Task Manager!"
3. Run it: `go run main.go`
4. Then follow Phase 1, Step 1.2

**Remember:**
- Write code yourself (don't copy-paste)
- Read error messages carefully
- Use `go doc` to read documentation
- Experiment and break things!
- Ask questions when stuck

---

## 💡 Tips for Learning

1. **Read the errors** - Go errors are usually helpful
2. **Use `go doc`** - `go doc package.Function` shows docs
3. **Experiment** - Try changing things to see what happens
4. **Test often** - Run your code frequently
5. **Google is your friend** - Search "golang [your question]"

---

## 🆘 When You Get Stuck

1. Read the error message carefully
2. Check the Go documentation
3. Search Stack Overflow
4. Ask me specific questions!
5. Take a break and come back fresh

---

**Good luck! You've got this! 🎉**

Remember: The goal is to **learn by doing**, not to finish quickly. Take your time, understand each concept, and build something you're proud of!

