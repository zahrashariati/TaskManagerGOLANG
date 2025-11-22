# Comprehensive Explanation - All Your Questions Answered

## 1. Where do we write to DB? Using CreateTask?

**YES!** The flow is:

```
HTTP POST /tasks 
  → Handler.CreateTask() 
    → Service.CreateTask() 
      → Repository.Create() 
        → Database INSERT ← WRITES HERE!
```

**The actual database write happens in `Repository.Create()`:**
```go
func (r *TaskRepository) Create(task *models.Task) (int, error) {
    query := "INSERT INTO tasks ..."  // ← SQL INSERT statement
    r.db.QueryRow(query, ...).Scan(&id)  // ← Executes INSERT
    return id, nil
}
```

**Other write operations:**
- `Update()` → `UPDATE` SQL
- `Delete()` → `DELETE` SQL
- `Complete()` → `UPDATE` SQL

---

## 2. Where is the for loop for Scan?

**It's in `GetAll()` method!** Here's how it works:

```go
func (r *TaskRepository) GetAll(showCompleted bool) ([]models.Task, error) {
    rows, err := r.db.Query(query)  // Get all rows
    defer rows.Close()
    
    var tasks []models.Task
    // THIS IS THE FOR LOOP!
    for rows.Next() {  // Loop through each row
        var task models.Task  // NEW variable each iteration
        rows.Scan(&task.ID, &task.Title, ...)  // Scan writes to THIS task
        tasks = append(tasks, task)  // Add to slice
    }
    return tasks, nil
}
```

**Why for loop?**
- `Query()` returns multiple rows
- Need to loop through each row
- Scan each row into a task struct
- Add each task to the slice

**Without loop:** Can only get one row!

---

## 3. showCompleted Cache Keys Explanation

**The Problem:**
- User requests: `GET /tasks?showCompleted=true` (all tasks)
- User requests: `GET /tasks?showCompleted=false` (only incomplete)
- **These are DIFFERENT data!**

**Solution - Different Cache Keys:**

```go
func (c *Cache) GetTasks(showCompleted bool) ([]models.Task, error) {
    key := "tasks:all"        // Cache key for ALL tasks
    if !showCompleted {
        key = "tasks:incomplete"  // Cache key for INCOMPLETE tasks
    }
    // Get from Redis using the correct key
}
```

**Why different keys?**

**Scenario 1:** User requests all tasks
- Cache key: `"tasks:all"`
- Stores: [task1, task2, task3] (all tasks)

**Scenario 2:** User requests incomplete tasks
- Cache key: `"tasks:incomplete"`
- Stores: [task1, task2] (only incomplete)

**Without different keys:**
- Both requests would use same cache key
- Wrong data returned!
- If cache has "all tasks", incomplete request gets wrong data

**With different keys:**
- Each request type has its own cache
- Correct data returned
- Better performance

---

## 4. What are Encoder/Decoder?

**Encoder** = Converts Go struct → JSON string
**Decoder** = Converts JSON string → Go struct

**In Handlers:**

**Decoder (Reading Request):**
```go
var task models.Task
json.NewDecoder(r.Body).Decode(&task)
// Reads JSON from HTTP request body
// Converts JSON → Go struct
// Example: {"title": "Learn Go"} → task.Title = "Learn Go"
```

**Encoder (Writing Response):**
```go
json.NewEncoder(w).Encode(tasks)
// Converts Go struct → JSON
// Writes JSON to HTTP response
// Example: tasks → [{"id":1,"title":"Learn Go"}]
```

**In Fiber (simpler):**
```go
c.BodyParser(&task)  // Decoder - reads JSON body
c.JSON(task)         // Encoder - writes JSON response
```

---

## 5. Handler Function Line-by-Line Explanation

### Example: `GetAllTasks`

```go
func (h *TaskHandler) GetAllTasks(c *fiber.Ctx) error {
```

**Line breakdown:**
- `func` = function definition
- `(h *TaskHandler)` = method on TaskHandler struct (pointer receiver)- `GetAllTasks` = function name
- `(c *fiber.Ctx)` = parameter - Fiber context (contains request/response)
- `error` = return type (can return error)

**Why `*fiber.Ctx`?**
- `fiber.Ctx` = Context struct (holds request, response, params, etc.)
- `*fiber.Ctx` = pointer (more efficient)
- Contains: request body, URL params, query params, response writer

**Inside function:**

```go
// Line 1: Get query parameter
showCompleted := c.Query("showCompleted") == "true"
// c.Query() = gets query param from URL
// Example: /tasks?showCompleted=true → returns "true"
// == "true" → converts to boolean

// Line 2: Call service
tasks, err := h.service.GetAllTasks(showCompleted)
// Calls business logic layer
// Returns tasks slice and error

// Line 3-6: Handle error
if err != nil {
    return c.Status(fiber.StatusInternalServerError).JSON(...)
    // If error, return HTTP 500 with error message
}

// Line 7: Return JSON response
return c.JSON(tasks)
// Converts tasks to JSON and sends to client
// Returns HTTP 200 OK
```

---

## 6. Each Handler Method Explained

### `GetAllTasks` - GET /tasks
```go
func (h *TaskHandler) GetAllTasks(c *fiber.Ctx) error {
    showCompleted := c.Query("showCompleted") == "true"  // Get query param
    tasks, err := h.service.GetAllTasks(showCompleted)   // Call service
    if err != nil { return error }                        // Handle error
    return c.JSON(tasks)                                  // Return JSON
}
```
**Purpose:** Get all tasks (optionally filter completed)

### `CreateTask` - POST /tasks
```go
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
    var task models.Task                                    // Create empty task
    c.BodyParser(&task)                                     // Parse JSON body
    if err := h.service.CreateTask(&task); err != nil {    // Call service
        return error                                        // Handle error
    }
    return c.Status(201).JSON(task)                         // Return created task
}
```
**Purpose:** Create new task from JSON body

### `GetTaskByID` - GET /tasks/:id
```go
func (h *TaskHandler) GetTaskByID(c *fiber.Ctx) error {
    id := c.Params("id")                                    // Get ID from URL
    idInt, _ := strconv.Atoi(id)                           // Convert to int
    task, err := h.service.GetTaskByID(idInt)               // Call service
    if err != nil { return error }                         // Handle error
    return c.JSON(task)                                    // Return JSON
}
```
**Purpose:** Get one task by ID

### `UpdateTask` - PUT /tasks/:id
```go
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
    id := c.Params("id")                                    // Get ID
    var task models.Task                                    // Create task
    c.BodyParser(&task)                                     // Parse JSON
    if err := h.service.UpdateTask(id, &task); err != nil { // Call service
        return error                                        // Handle error
    }
    return c.JSON(task)                                    // Return updated task
}
```
**Purpose:** Update existing task

### `DeleteTask` - DELETE /tasks/:id
```go
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
    id := c.Params("id")                                    // Get ID
    if err := h.service.DeleteTask(id); err != nil {       // Call service
        return error                                        // Handle error
    }
    return c.JSON({"message": "deleted"})                  // Return success
}
```
**Purpose:** Delete task by ID

---

## 7. Method Signatures Explained

**Method signature** = Function name + parameters + return types

**Example:**
```go
func (r *TaskRepository) Create(task *models.Task) (int, error)
//     ↑                  ↑       ↑                    ↑      ↑
//  Receiver          Name    Parameter          Return types
```

**Breaking it down:**

**`(r *TaskRepository)`** = Receiver
- Method belongs to TaskRepository struct
- `r` = name (like `self` in Python)
- `*TaskRepository` = pointer receiver

**`Create`** = Method name

**`(task *models.Task)`** = Parameter
- Input: pointer to Task struct

**`(int, error)`** = Return types
- Output: task ID (int) and error

---

## 8. Why `*models.Task` in Different Places?

### In `Update(id int, task *models.Task)`:
```go
func (r *TaskRepository) Update(id int, task *models.Task) error
//                                ↑       ↑
//                            Input parameter
```
**Why pointer?**
- Input parameter (what you pass IN)
- Pointer = don't copy large struct
- More efficient
- Can modify if needed

### In `GetByID(id int) (*models.Task, error)`:
```go
func (r *TaskRepository) GetByID(id int) (*models.Task, error)
//                                    ↑              ↑
//                                  Return type
```
**Why pointer?**
- Return type (what you get OUT)
- Pointer = don't copy large struct
- More efficient
- Caller gets reference, not copy

**Summary:**
- **Input `*models.Task`** = Pass pointer IN (efficient)
- **Return `*models.Task`** = Return pointer OUT (efficient)
- **Both use pointers** = Avoid copying large structs

---

## 9. Method Signature Issues Fixed

**Issue 1: Service.CreateTask expected different input**

**Before:**
```go
// Service
func CreateTask(task models.Task) error  // Value, not pointer

// Repository  
func Create(task *models.Task) (int, error)  // Pointer
```

**After (Fixed):**
```go
// Service
func CreateTask(task *models.Task) error  // Now matches!

// Repository
func Create(task *models.Task) (int, error)  // Matches!
```

**Issue 2: Service.UpdateTask expected different input**

**Before:**
```go
// Service
func UpdateTask(id int, task models.Task) error  // Value

// Repository
func Update(id int, task *models.Task) error  // Pointer
```

**After (Fixed):**
```go
// Service
func UpdateTask(id int, task *models.Task) error  // Now matches!

// Repository
func Update(id int, task *models.Task) error  // Matches!
```

---

## 10. DeleteTask Undefined Task Fixed

**The Problem:**
```go
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
    // ... delete logic ...
    json.NewEncoder(w).Encode(task)  // ❌ task undefined!
}
```

**Why undefined?**
- `DeleteTask` doesn't return a task
- No `task` variable exists
- Can't encode undefined variable

**Fixed:**
```go
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
    // ... delete logic ...
    return c.JSON(fiber.Map{
        "message": "Task deleted successfully",  // ✅ Success message
    })
}
```

---

## 11. Custom Errors Usage

**All errors now use custom errors:**

**Repository:**
```go
if err == sql.ErrNoRows {
    return nil, errors.ErrTaskNotFound  // ✅ Custom error
}
return nil, errors.ErrDatabaseQueryFailed  // ✅ Custom error
```

**Service:**
```go
if errors.Is(err, errors.ErrTaskNotFound) {
    return nil, errors.ErrTaskNotFound  // ✅ Check and return custom error
}
```

**Handler:**
```go
if errors.Is(err, errors.ErrTaskNotFound) {
    return c.Status(404).JSON(fiber.Map{
        "error": "Task not found",  // ✅ User-friendly message
    })
}
```

**Benefits:**
- Consistent error messages
- Easy to check error type
- Better HTTP status codes
- User-friendly messages

---

## 12. Fiber vs net/http

**Fiber (what we're using now):**
```go
app := fiber.New()
app.Get("/tasks", handler.GetAllTasks)
app.Listen(":8080")
```

**net/http (old way):**
```go
http.HandleFunc("/tasks", handler.GetAllTasks)
http.ListenAndServe(":8080", nil)
```

**Why Fiber?**
- Simpler syntax
- Better performance
- Built-in JSON handling
- Better error handling
- More features

---

## 13. Local CockroachDB Setup

**Connection string format:**
```
postgresql://username:password@localhost:26257/database?sslmode=disable
```

**For local:**
- Host: `localhost`
- Port: `26257` (default CockroachDB port)
- SSL: `sslmode=disable` (for local)

**In .env:**
```
DATABASE_URL=postgresql://root@localhost:26257/taskmanager?sslmode=disable
REDIS_URL=localhost:6379
PORT=8080
```

---

## Summary

✅ **DB writes:** Repository methods (Create, Update, Delete)
✅ **For loop:** In GetAll() to scan multiple rows
✅ **Cache keys:** Different keys for different data types
✅ **Encoder/Decoder:** Convert JSON ↔ Go structs
✅ **Handlers:** Each method = one endpoint
✅ **Method signatures:** Fixed to match between layers
✅ **Custom errors:** Used everywhere
✅ **Fiber:** Converted from net/http
✅ **Local DB:** Connection string updated

**Next Steps:**
1. Run `go mod tidy` to install Fiber
2. Set up local CockroachDB
3. Set environment variables
4. Run `go run cmd/api/main.go`
5. Test endpoints!

