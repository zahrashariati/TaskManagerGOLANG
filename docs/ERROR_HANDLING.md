# Go Error Handling Guide 🚨

Error handling is **CRITICAL** in Go. This guide teaches you everything you need to know!

## 🎯 Why Error Handling Matters

- **Go doesn't have exceptions** - errors are values
- **Explicit error handling** - you MUST check errors
- **Prevents crashes** - handle errors gracefully
- **Better debugging** - know exactly what went wrong
- **Professional code** - proper error handling is essential

---

## 📚 Basics: The Error Type

### What is an Error?

```go
type error interface {
    Error() string
}
```

An error is just an interface with one method: `Error() string`

### Creating Errors

```go
import "errors"
import "fmt"

// Simple error
err := errors.New("something went wrong")

// Formatted error
err := fmt.Errorf("failed to connect: %s", reason)

// With values
err := fmt.Errorf("user %d not found", userID)
```

---

## 🔄 Common Patterns

### Pattern 1: Check and Return

```go
func doSomething() error {
    result, err := someFunction()
    if err != nil {
        return err  // Return error immediately
    }
    
    // Continue if no error
    useResult(result)
    return nil
}
```

### Pattern 2: Check and Handle

```go
func doSomething() {
    result, err := someFunction()
    if err != nil {
        log.Printf("Error: %v", err)
        return  // Exit early
    }
    
    useResult(result)
}
```

### Pattern 3: Check and Wrap

```go
func doSomething() error {
    result, err := someFunction()
    if err != nil {
        return fmt.Errorf("failed to do something: %w", err)
        // %w wraps the error (preserves original error)
    }
    
    return nil
}
```

---

## 🎨 Error Handling in Your REST API

### Handler Layer (HTTP)

```go
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        // Bad request - invalid ID format
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }
    
    task, err := h.service.GetTaskByID(id)
    if err != nil {
        // Check what kind of error
        if err == ErrTaskNotFound {
            http.Error(w, "Task not found", http.StatusNotFound)
        } else {
            // Internal server error
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            log.Printf("Error getting task: %v", err)  // Log for debugging
        }
        return
    }
    
    // Success
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}
```

### Service Layer (Business Logic)

```go
func (s *TaskService) CreateTask(title, desc, priority string) (*models.Task, error) {
    // Validate input
    if title == "" {
        return nil, ErrTitleRequired  // Custom error
    }
    
    if priority != "low" && priority != "medium" && priority != "high" {
        return nil, fmt.Errorf("invalid priority: %s", priority)
    }
    
    // Call repository
    id, err := s.repo.Create(title, desc, priority)
    if err != nil {
        return nil, fmt.Errorf("failed to create task: %w", err)
        // Wrap error with context
    }
    
    // Invalidate cache
    if err := s.cache.Invalidate(); err != nil {
        log.Printf("Warning: cache invalidation failed: %v", err)
        // Don't fail the whole operation, just log
    }
    
    // Get created task
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve created task: %w", err)
    }
    
    return task, nil
}
```

### Repository Layer (Database)

```go
func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
    var t models.Task
    err := r.db.QueryRow(
        "SELECT id, title, description, completed, created_at, priority FROM tasks WHERE id = $1",
        id,
    ).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
    
    // Check for specific database errors
    if err == sql.ErrNoRows {
        return nil, ErrTaskNotFound  // Custom error for not found
    }
    
    if err != nil {
        return nil, fmt.Errorf("database query failed: %w", err)
    }
    
    return &t, nil
}
```

---

## 🛠️ Custom Errors

### Define Custom Errors

```go
package errors

import "errors"

var (
    ErrTaskNotFound    = errors.New("task not found")
    ErrTitleRequired   = errors.New("title is required")
    ErrInvalidPriority = errors.New("invalid priority")
    ErrDatabaseError   = errors.New("database error")
)
```

### Using Custom Errors

```go
import "your-project/internal/errors"

func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
    // ... query code ...
    
    if err == sql.ErrNoRows {
        return nil, errors.ErrTaskNotFound
    }
    
    return &t, nil
}

// In handler:
if err == errors.ErrTaskNotFound {
    http.Error(w, "Task not found", http.StatusNotFound)
    return
}
```

### Error with Context

```go
type TaskError struct {
    Code    string
    Message string
    Err     error
}

func (e *TaskError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

// Usage:
return nil, &TaskError{
    Code:    "TASK_NOT_FOUND",
    Message: "Task not found",
    Err:     err,
}
```

---

## 🔍 Checking Error Types

### Using `errors.Is()`

```go
import "errors"

if errors.Is(err, sql.ErrNoRows) {
    // Handle "not found" case
    return nil, ErrTaskNotFound
}
```

### Using `errors.As()`

```go
import "errors"

var taskErr *TaskError
if errors.As(err, &taskErr) {
    // Handle TaskError specifically
    fmt.Printf("Error code: %s\n", taskErr.Code)
}
```

---

## 📝 Error Wrapping

### Why Wrap Errors?

Wrapping adds context while preserving the original error:

```go
// Without wrapping (loses context)
return nil, err

// With wrapping (adds context)
return nil, fmt.Errorf("failed to create task: %w", err)
```

### Unwrapping Errors

```go
// Get original error
originalErr := errors.Unwrap(err)

// Check if error chain contains specific error
if errors.Is(err, sql.ErrNoRows) {
    // This works even if error is wrapped!
}
```

### Example: Error Chain

```go
// Repository layer
err := r.db.QueryRow(...).Scan(...)
if err == sql.ErrNoRows {
    return nil, fmt.Errorf("task %d not found: %w", id, err)
}

// Service layer
task, err := r.repo.GetByID(id)
if err != nil {
    return nil, fmt.Errorf("failed to get task: %w", err)
}

// Handler layer
task, err := s.service.GetTaskByID(id)
if err != nil {
    // Error chain: "failed to get task: task 123 not found: sql: no rows in result set"
    log.Printf("Full error: %v", err)
    http.Error(w, "Task not found", http.StatusNotFound)
}
```

---

## 🎯 Best Practices

### 1. Always Check Errors

```go
// ❌ BAD - ignoring error
result, _ := someFunction()

// ✅ GOOD - checking error
result, err := someFunction()
if err != nil {
    return err
}
```

### 2. Return Errors Early

```go
// ❌ BAD - nested if statements
func doSomething() error {
    result, err := step1()
    if err == nil {
        result2, err2 := step2(result)
        if err2 == nil {
            return step3(result2)
        }
        return err2
    }
    return err
}

// ✅ GOOD - early returns
func doSomething() error {
    result, err := step1()
    if err != nil {
        return err
    }
    
    result2, err := step2(result)
    if err != nil {
        return err
    }
    
    return step3(result2)
}
```

### 3. Add Context to Errors

```go
// ❌ BAD - no context
if err != nil {
    return err
}

// ✅ GOOD - adds context
if err != nil {
    return fmt.Errorf("failed to create task with title '%s': %w", title, err)
}
```

### 4. Use Appropriate HTTP Status Codes

```go
// 400 Bad Request - client error (invalid input)
if req.Title == "" {
    http.Error(w, "Title is required", http.StatusBadRequest)
    return
}

// 404 Not Found - resource doesn't exist
if err == ErrTaskNotFound {
    http.Error(w, "Task not found", http.StatusNotFound)
    return
}

// 500 Internal Server Error - server error
if err != nil {
    log.Printf("Internal error: %v", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
```

### 5. Log Errors Appropriately

```go
// Log errors that shouldn't happen (bugs, unexpected errors)
if err != nil {
    log.Printf("Unexpected error: %v", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}

// Don't log expected errors (validation, not found)
if err == ErrTaskNotFound {
    http.Error(w, "Task not found", http.StatusNotFound)
    return  // No logging needed - this is expected
}
```

### 6. Don't Expose Internal Errors to Users

```go
// ❌ BAD - exposes database details
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}

// ✅ GOOD - generic message for users, detailed log for developers
if err != nil {
    log.Printf("Database error: %v", err)  // Detailed log
    http.Error(w, "Internal server error", http.StatusInternalServerError)  // Generic message
    return
}
```

---

## 🔄 Error Handling Patterns by Layer

### Handler Layer Pattern

```go
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req models.CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // 2. Validate input
    if req.Title == "" {
        http.Error(w, "Title is required", http.StatusBadRequest)
        return
    }
    
    // 3. Call service
    task, err := h.service.CreateTask(req.Title, req.Description, req.Priority)
    if err != nil {
        // Check error type
        if errors.Is(err, errors.ErrTitleRequired) {
            http.Error(w, "Title is required", http.StatusBadRequest)
        } else if errors.Is(err, errors.ErrInvalidPriority) {
            http.Error(w, "Invalid priority", http.StatusBadRequest)
        } else {
            // Unexpected error
            log.Printf("Error creating task: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }
    
    // 4. Success response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}
```

### Service Layer Pattern

```go
func (s *TaskService) CreateTask(title, desc, priority string) (*models.Task, error) {
    // 1. Validate business rules
    if title == "" {
        return nil, errors.ErrTitleRequired
    }
    
    if priority == "" {
        priority = "medium"  // Default value
    }
    
    if priority != "low" && priority != "medium" && priority != "high" {
        return nil, fmt.Errorf("invalid priority '%s': %w", priority, errors.ErrInvalidPriority)
    }
    
    // 2. Call repository
    id, err := s.repo.Create(title, desc, priority)
    if err != nil {
        return nil, fmt.Errorf("failed to create task in database: %w", err)
    }
    
    // 3. Handle cache (non-critical, log but don't fail)
    if err := s.cache.Invalidate(); err != nil {
        log.Printf("Warning: failed to invalidate cache: %v", err)
    }
    
    // 4. Return result
    task, err := s.repo.GetByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve created task: %w", err)
    }
    
    return task, nil
}
```

### Repository Layer Pattern

```go
func (r *TaskRepository) GetByID(id int) (*models.Task, error) {
    var t models.Task
    err := r.db.QueryRow(
        "SELECT id, title, description, completed, created_at, priority FROM tasks WHERE id = $1",
        id,
    ).Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &t.CreatedAt, &t.Priority)
    
    // Handle specific database errors
    if err == sql.ErrNoRows {
        return nil, errors.ErrTaskNotFound
    }
    
    if err != nil {
        return nil, fmt.Errorf("database query failed for task %d: %w", id, err)
    }
    
    return &t, nil
}
```

---

## 🧪 Testing Errors

### Test Error Cases

```go
func TestGetTask_NotFound(t *testing.T) {
    repo := NewTaskRepository(mockDB)
    
    task, err := repo.GetByID(999)
    
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    
    if !errors.Is(err, errors.ErrTaskNotFound) {
        t.Errorf("expected ErrTaskNotFound, got %v", err)
    }
    
    if task != nil {
        t.Error("expected nil task")
    }
}
```

---

## 📋 Error Handling Checklist

For every function that can fail:

- [ ] Check if function returns error
- [ ] Handle error appropriately
- [ ] Add context if wrapping error
- [ ] Use appropriate HTTP status code (if in handler)
- [ ] Log unexpected errors
- [ ] Don't expose internal errors to users
- [ ] Return early on error (don't nest)

---

## 🎓 Common Mistakes to Avoid

### 1. Ignoring Errors

```go
// ❌ BAD
result, _ := someFunction()

// ✅ GOOD
result, err := someFunction()
if err != nil {
    return err
}
```

### 2. Generic Error Messages

```go
// ❌ BAD
if err != nil {
    return fmt.Errorf("error: %v", err)
}

// ✅ GOOD
if err != nil {
    return fmt.Errorf("failed to create task '%s': %w", title, err)
}
```

### 3. Not Checking Error Type

```go
// ❌ BAD
task, err := repo.GetByID(id)
if err != nil {
    http.Error(w, "Error", 500)  // Always 500?
    return
}

// ✅ GOOD
task, err := repo.GetByID(id)
if err != nil {
    if errors.Is(err, errors.ErrTaskNotFound) {
        http.Error(w, "Task not found", http.StatusNotFound)
    } else {
        http.Error(w, "Internal error", http.StatusInternalServerError)
    }
    return
}
```

---

## 💡 Quick Reference

```go
// Create error
err := errors.New("message")
err := fmt.Errorf("formatted: %s", value)

// Wrap error
return fmt.Errorf("context: %w", err)

// Check error
if err != nil {
    return err
}

// Check specific error
if errors.Is(err, targetErr) {
    // Handle specific error
}

// Extract error type
var customErr *CustomError
if errors.As(err, &customErr) {
    // Use customErr
}

// Unwrap error
original := errors.Unwrap(err)
```

---

**Remember: Error handling is not optional in Go - it's a core part of the language!** 🎯

Practice these patterns as you build your API, and your code will be robust and professional!

