# Quick Reference Guide - Go Syntax Cheat Sheet

## Database Operations (CockroachDB/PostgreSQL)

### Opening a Database
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

connString := "postgresql://user:password@host:26257/database?sslmode=require"
db, err := sql.Open("postgres", connString)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Test connection
err = db.Ping()
if err != nil {
    log.Fatal(err)
}
```

### Creating a Table
```go
query := `CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    priority VARCHAR(20) DEFAULT 'medium'
)`
_, err := db.Exec(query)
```

### Inserting Data (with RETURNING)
```go
var id int
err := db.QueryRow(
    "INSERT INTO tasks (title, description, priority) VALUES ($1, $2, $3) RETURNING id",
    title, description, priority,
).Scan(&id)
```

### Querying Data
```go
rows, err := db.Query("SELECT id, title, description, completed, created_at, priority FROM tasks WHERE completed = $1", false)
if err != nil {
    return nil, err
}
defer rows.Close()

var tasks []Task
for rows.Next() {
    var t Task
    err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
    if err != nil {
        return nil, err
    }
    tasks = append(tasks, t)
}
```

### Updating Data
```go
_, err := db.Exec("UPDATE tasks SET completed = $1 WHERE id = $2", true, id)
```

### Deleting Data
```go
_, err := db.Exec("DELETE FROM tasks WHERE id = $1", id)
```

**Note:** PostgreSQL uses `$1, $2, $3` placeholders (not `?`)

---

## Redis Caching

### Cache Structure
```go
import (
    "github.com/redis/go-redis/v9"
    "context"
)

type TaskManager struct {
    db          *sql.DB
    redisClient *redis.Client
}
```

### Connecting to Redis
```go
opt, err := redis.ParseURL("redis://default:password@host:port")
if err != nil {
    return nil, err
}
client := redis.NewClient(opt)

ctx := context.Background()
err = client.Ping(ctx).Err()
if err != nil {
    return nil, err
}
```

### Reading from Cache
```go
ctx := context.Background()
val, err := client.Get(ctx, "task:1").Result()
if err == redis.Nil {
    // Cache miss - not found
    return nil, fmt.Errorf("not in cache")
} else if err != nil {
    return nil, err
}

// Unmarshal JSON
var task Task
err = json.Unmarshal([]byte(val), &task)
```

### Writing to Cache
```go
ctx := context.Background()
data, err := json.Marshal(task)
if err != nil {
    return err
}

err = client.Set(ctx, "task:1", data, time.Minute*5).Err()
// Expires in 5 minutes
```

### Deleting from Cache
```go
ctx := context.Background()
err := client.Del(ctx, "task:1", "tasks:all", "tasks:incomplete").Err()
```

### Cache-Aside Pattern Example
```go
// Try cache first
key := "tasks:all"
val, err := client.Get(ctx, key).Result()
if err == nil {
    // Cache hit
    var tasks []Task
    json.Unmarshal([]byte(val), &tasks)
    return tasks, nil
}

// Cache miss - query database
tasks, err := tm.queryFromDB()
if err != nil {
    return nil, err
}

// Store in cache
data, _ := json.Marshal(tasks)
client.Set(ctx, key, data, time.Minute*5)
return tasks, nil
```

---

## REST API with HTTP Server

### Basic HTTP Server
```go
import (
    "net/http"
    "log"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello!"))
    })
    
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Handler Function
```go
func handleGetTasks(w http.ResponseWriter, r *http.Request) {
    // Check method
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // Set headers
    w.Header().Set("Content-Type", "application/json")
    
    // Send JSON response
    tasks := []Task{...}
    json.NewEncoder(w).Encode(tasks)
}
```

### Parse JSON Request Body
```go
func handleCreateTask(w http.ResponseWriter, r *http.Request) {
    var req CreateTaskRequest
    
    // Decode JSON body
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Process request...
}
```

### Extract URL Parameters (gorilla/mux)
```go
import "github.com/gorilla/mux"

router := mux.NewRouter()
router.HandleFunc("/tasks/{id}", handleGetTask).Methods("GET")

// In handler:
func handleGetTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, _ := strconv.Atoi(vars["id"])
    // Use id...
}
```

### Extract Query Parameters
```go
func handleSearch(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("q")
    page := r.URL.Query().Get("page")
    // Use query parameters...
}
```

### HTTP Status Codes
```go
w.WriteHeader(http.StatusOK)        // 200
w.WriteHeader(http.StatusCreated)    // 201
w.WriteHeader(http.StatusBadRequest) // 400
w.WriteHeader(http.StatusNotFound)   // 404
w.WriteHeader(http.StatusInternalServerError) // 500
```

### CORS Middleware
```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

---

## Error Handling

### Basic Pattern
```go
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}
```

### Wrapping Errors
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### Check and Return Early
```go
result, err := step1()
if err != nil {
    return err  // Return immediately
}

result2, err := step2(result)
if err != nil {
    return err
}
```

### Custom Errors
```go
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid input")
)

// Check specific error
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

### HTTP Error Handling
```go
// Handler layer
task, err := service.GetTask(id)
if err != nil {
    if errors.Is(err, ErrNotFound) {
        http.Error(w, "Not found", http.StatusNotFound)
    } else {
        log.Printf("Error: %v", err)
        http.Error(w, "Internal error", http.StatusInternalServerError)
    }
    return
}
```

### Database Error Handling
```go
err := db.QueryRow(...).Scan(&id)
if err == sql.ErrNoRows {
    return nil, ErrNotFound
}
if err != nil {
    return nil, fmt.Errorf("query failed: %w", err)
}
```

**Read `ERROR_HANDLING.md` for comprehensive guide!**

---

## Concurrency

### Basic Goroutine
```go
go func() {
    // runs concurrently
    fmt.Println("Running in goroutine")
}()
```

### WaitGroup Pattern
```go
var wg sync.WaitGroup
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // work here
        fmt.Println("Worker", id)
    }(i)
}
wg.Wait() // wait for all goroutines
```

### Buffered Channel
```go
ch := make(chan string, 10) // buffer size 10
ch <- "send"        // send (non-blocking if buffer not full)
msg := <-ch         // receive
close(ch)           // close when done

// Range over channel
for msg := range ch {
    fmt.Println(msg)
}
```

### Select Statement (Multiple Channels)
```go
select {
case result := <-cacheChan:
    return result // Cache hit
case result := <-dbChan:
    return result // DB result
case err := <-errChan:
    return nil, err
case <-time.After(5 * time.Second):
    return nil, fmt.Errorf("timeout")
}
```

### Worker Pool Pattern
```go
const numWorkers = 5
jobs := make(chan int, 10)
results := make(chan string, 10)

// Start workers
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        for job := range jobs {
            results <- process(job)
        }
    }()
}

// Send jobs
for _, job := range jobList {
    jobs <- job
}
close(jobs)

// Wait and close results
go func() {
    wg.Wait()
    close(results)
}()

// Collect results
for result := range results {
    fmt.Println(result)
}
```

### Background Goroutine with Ticker
```go
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // Periodic task
        fmt.Println("Running periodic task")
    }
}()
```

### Context for Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

select {
case result := <-ch:
    return result
case <-ctx.Done():
    return nil, ctx.Err()
}
```

---

## Common Patterns

### Defer for Cleanup
```go
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close() // Always closes
```

### Pointer Receiver
```go
func (tm *TaskManager) AddTask() {
    // Can modify tm
}
```

### Value Receiver
```go
func (t Task) String() string {
    // Works on copy
}
```

---

## Testing

### Basic Test
```go
func TestAddTask(t *testing.T) {
    tm := NewTaskManager()
    err := tm.AddTask("Test")
    if err != nil {
        t.Fatalf("Failed: %v", err)
    }
}
```

### Table-Driven Test
```go
tests := []struct {
    name string
    input string
    want error
}{
    {"valid", "Task", nil},
    {"empty", "", ErrEmpty},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // test code
    })
}
```

