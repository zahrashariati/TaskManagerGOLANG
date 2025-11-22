# All Your Questions Answered - Complete Guide

## 1. `addr := ":" + cfg.Port` - Syntax Explanation

**What it does:**
```go
addr := ":" + cfg.Port
```

**Breaking it down:**
- `addr` = variable name
- `:=` = short variable declaration (creates and assigns)
- `":"` = string literal (colon character)
- `+` = string concatenation operator
- `cfg.Port` = value from config (e.g., "8080")

**Result:**
- If `cfg.Port = "8080"`, then `addr = ":8080"`
- If `cfg.Port = "3000"`, then `addr = ":3000"`

**Why `":"` at the start?**
- `:8080` = listen on all network interfaces on port 8080
- `8080` = invalid (missing colon)
- `localhost:8080` = only localhost (not accessible from outside)
- `:8080` = all interfaces (accessible from anywhere)

**Example:**
```go
cfg.Port = "8080"
addr := ":" + cfg.Port
// addr = ":8080"

app.Listen(addr)  // Listens on :8080
```

---

## 2. Why Custom Errors + Fiber Status in Handlers?

**You ARE using custom errors!** Look at your code:

```go
if errors.Is(err, errors.ErrTaskNotFound) {  // ← Checking custom error!
    return c.Status(fiber.StatusNotFound).JSON(...)  // ← Using Fiber status
}
```

**Why both?**

**Custom Errors (from service):**
- Business logic errors
- Consistent error types
- Easy to check: `errors.Is(err, errors.ErrTaskNotFound)`

**Fiber Status (HTTP layer):**
- HTTP status codes (404, 500, etc.)
- Client needs HTTP status, not Go errors
- Converts Go errors → HTTP responses

**Flow:**
```
Service returns: errors.ErrTaskNotFound (custom error)
    ↓
Handler checks: errors.Is(err, errors.ErrTaskNotFound)
    ↓
Handler returns: c.Status(404).JSON(...) (HTTP response)
```

**Why not just custom errors?**
- HTTP clients don't understand Go errors
- Need HTTP status codes (404, 500, etc.)
- Custom errors = internal, HTTP status = external

---

## 3. What Happened to Encoding/Decoding in Handlers?

**They're still there, but Fiber makes it simpler!**

**Old way (net/http):**
```go
// Decoder
json.NewDecoder(r.Body).Decode(&task)  // Read JSON → Go struct

// Encoder
json.NewEncoder(w).Encode(task)  // Write Go struct → JSON
```

**New way (Fiber):**
```go
// Decoder (simpler!)
c.BodyParser(&task)  // Read JSON → Go struct

// Encoder (simpler!)
c.JSON(task)  // Write Go struct → JSON
```

**Fiber does the same thing, but simpler:**
- `BodyParser` = decoder (JSON → struct)
- `JSON` = encoder (struct → JSON)
- Less code, same functionality!

---

## 4. What is `Ctx` (Context)?

**`c *fiber.Ctx`** = Fiber Context

**What it contains:**
- Request data (body, headers, params, query)
- Response writer (to send response)
- URL parameters (`:id`)
- Query parameters (`?showCompleted=true`)
- Request body
- Response methods

**Think of it as:**
- Everything about the HTTP request/response
- Your way to read request and write response

**What you can do with it:**
```go
c.Params("id")           // Get URL param: /tasks/:id
c.Query("showCompleted") // Get query param: ?showCompleted=true
c.BodyParser(&task)      // Parse JSON body
c.JSON(task)             // Send JSON response
c.Status(404)            // Set HTTP status code
```

**Why pointer `*fiber.Ctx`?**
- Large struct (contains lots of data)
- Pointer = don't copy, share reference
- More efficient

---

## 5. Complete Flow: JSON → Struct → Service

**Complete request flow:**

### Step 1: Client sends JSON
```json
POST /tasks
{
  "title": "Learn Go",
  "description": "Study pointers",
  "priority": "high"
}
```

### Step 2: Handler receives request
```go
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
    var task models.Task  // Empty struct
    c.BodyParser(&task)  // JSON → Go struct
    // Now task.Title = "Learn Go"
    // task.Description = "Study pointers"
    // task.Priority = "high"
```

**What BodyParser does:**
- Reads JSON from `c` (request body)
- Converts JSON fields → struct fields
- Uses JSON tags: `json:"title"` → `Title` field

### Step 3: Handler calls Service
```go
h.service.CreateTask(&task)  // Pass pointer to struct
```

**Service receives:**
```go
func (s *TaskService) CreateTask(task *models.Task) error {
    // task.Title = "Learn Go"  ← Already filled!
    // task.Description = "Study pointers"
    // task.Priority = "high"
    
    // Use struct fields directly
    if task.Title == "" {  // Check title
        return errors.ErrTitleRequired
    }
    
    // Pass to repository
    id, err := s.repo.Create(task)  // Uses task.Title, task.Description, etc.
```

### Step 4: Repository uses struct
```go
func (r *TaskRepository) Create(task *models.Task) (int, error) {
    // Access struct fields
    query := "INSERT INTO tasks (title, description, priority) VALUES ($1, $2, $3)"
    r.db.QueryRow(query, 
        task.Title,        // "Learn Go"
        task.Description,  // "Study pointers"
        task.Priority,     // "high"
    )
}
```

### Step 5: Response back
```go
// Service returns
task.ID = id  // Set ID from database
return nil    // Success

// Handler sends response
return c.Status(201).JSON(task)  // Struct → JSON
// Client receives:
// {
//   "id": 1,
//   "title": "Learn Go",
//   ...
// }
```

**Summary:**
```
JSON (client) 
  → BodyParser (decoder) 
    → Go struct (handler)
      → Service (uses struct fields)
        → Repository (uses struct fields)
          → Database (stores data)
            → Response (struct → JSON)
```

---

## 6. What is Blank Import?

**Blank import** = `_ "package"`

**Example:**
```go
import _ "github.com/lib/pq"  // Blank import
```

**What it does:**
- Imports package for side effects only
- Doesn't use package directly in code
- Package registers itself (driver registration)

**Why needed?**
- PostgreSQL driver needs to register itself
- `database/sql` needs to know driver exists
- Side effect: driver registers on import

**Without blank import:**
```go
import "github.com/lib/pq"  // Regular import
// Error: package not used!
```

**With blank import:**
```go
import _ "github.com/lib/pq"  // Blank import
// Driver registers itself ✅
// No "unused import" error ✅
```

**Common uses:**
- Database drivers (`_ "github.com/lib/pq"`)
- Image formats (`_ "image/png"`)
- Any package that needs to register itself

---

## 7. Connection Refused Error - Server Not Running!

**Error:** `ECONNREFUSED 127.0.0.1:8080`

**Meaning:** Server is not running!

**Solution:**

### Step 1: Check if server is running
```bash
# In terminal, run:
go run cmd/api/main.go
```

**You should see:**
```
Hello, Task Manager API!
Server starting on :8080
```

### Step 2: Check for errors
**Common issues:**

**Missing dependencies:**
```bash
go mod tidy  # Install all dependencies
```

**Database not connected:**
- Check `DATABASE_URL` in `.env`
- Make sure CockroachDB is running

**Redis not connected:**
- Check `REDIS_URL` in `.env`
- Make sure Redis is running: `redis-cli ping`

**Port already in use:**
```bash
# Check what's using port 8080
lsof -i :8080
# Or change PORT in .env
```

### Step 3: Test server is running
```bash
# In another terminal:
curl http://localhost:8080/tasks
# Should return JSON (even if empty array)
```

**If still not working:**
1. Check terminal output for errors
2. Verify `.env` file exists
3. Verify database/Redis are running
4. Check firewall settings

---

## 8. Cache Format - Keys and Values Example

**How data is stored in Redis:**

### Example 1: All Tasks
**Key:** `tasks:all`
**Value:** JSON string
```json
[
  {
    "id": 1,
    "title": "Learn Go",
    "description": "Study pointers",
    "completed": false,
    "created_at": "2024-01-15T10:00:00Z",
    "priority": "high"
  },
  {
    "id": 2,
    "title": "Build API",
    "description": "Create REST API",
    "completed": true,
    "created_at": "2024-01-14T09:00:00Z",
    "priority": "medium"
  }
]
```

**In Redis:**
```
Key: "tasks:all"
Value: "[{\"id\":1,\"title\":\"Learn Go\",...},{\"id\":2,...}]"
TTL: 5 minutes
```

### Example 2: Incomplete Tasks Only
**Key:** `tasks:incomplete`
**Value:** JSON string (only incomplete tasks)
```json
[
  {
    "id": 1,
    "title": "Learn Go",
    "completed": false,
    ...
  }
]
```

**In Redis:**
```
Key: "tasks:incomplete"
Value: "[{\"id\":1,\"title\":\"Learn Go\",...}]"
TTL: 5 minutes
```

### Example 3: Single Task
**Key:** `task:1` (for task ID 1)
**Value:** JSON string (single task)
```json
{
  "id": 1,
  "title": "Learn Go",
  "description": "Study pointers",
  "completed": false,
  "created_at": "2024-01-15T10:00:00Z",
  "priority": "high"
}
```

**In Redis:**
```
Key: "task:1"
Value: "{\"id\":1,\"title\":\"Learn Go\",...}"
TTL: 5 minutes
```

### Visual Representation:
```
Redis Cache:
├── tasks:all          → [all tasks JSON]
├── tasks:incomplete    → [incomplete tasks JSON]
├── task:1             → {task 1 JSON}
├── task:2             → {task 2 JSON}
└── task:3             → {task 3 JSON}
```

**Key naming pattern:**
- `tasks:all` = all tasks
- `tasks:incomplete` = incomplete tasks only
- `task:{id}` = single task by ID

---

## 9. Why No Key for Completed Tasks?

**Good question!** You could add it, but here's why it's not needed:

**Current keys:**
- `tasks:all` = includes completed AND incomplete
- `tasks:incomplete` = only incomplete

**Why not `tasks:completed`?**

**Reason 1: Less common use case**
- Users usually want "all" or "incomplete"
- Completed tasks are less frequently accessed
- Can filter from `tasks:all` if needed

**Reason 2: Cache efficiency**
- Fewer keys = less memory
- `tasks:all` already contains completed tasks
- Can extract completed from `tasks:all` if needed

**If you want to add it:**

```go
func (c *Cache) GetCompletedTasks() ([]models.Task, error) {
    key := "tasks:completed"
    // ... same logic
}

func (c *Cache) SetCompletedTasks(tasks []models.Task) error {
    key := "tasks:completed"
    // ... same logic
}
```

**But it's optional!** Current design works fine.

---

## 10. Why Pointer in BodyParser but Not in JSON?

**Great observation!**

### BodyParser uses pointer:
```go
var task models.Task
c.BodyParser(&task)  // ← Pointer!
```

**Why pointer?**
- BodyParser **WRITES** to the struct
- Needs address to modify the variable
- Without `&`, can't fill the struct

**What happens:**
```go
var task models.Task  // Empty struct
c.BodyParser(&task)  // Fills task with JSON data
// Now task.Title = "Learn Go" ✅
```

**Without pointer:**
```go
var task models.Task
c.BodyParser(task)  // ❌ Can't modify! task stays empty
```

### JSON doesn't need pointer:
```go
c.JSON(task)  // ← No pointer needed!
```

**Why no pointer?**
- JSON **READS** from the struct
- Doesn't need to modify
- Can read value or pointer (both work)

**Both work:**
```go
c.JSON(task)    // ✅ Works (reads value)
c.JSON(&task)   // ✅ Also works (reads from pointer)
```

**Summary:**
- **BodyParser** = writes (needs pointer `&task`)
- **JSON** = reads (pointer optional, but value works fine)

---

## Quick Reference

| Question | Answer |
|----------|--------|
| `addr := ":" + cfg.Port` | String concatenation: `:8080` |
| Custom errors in handlers? | Yes! Using `errors.Is()` to check |
| Encoding/decoding? | `BodyParser` (decoder), `JSON` (encoder) |
| What is Ctx? | Fiber context (request/response data) |
| JSON flow? | JSON → BodyParser → struct → service → DB |
| Blank import? | `_ "package"` for side effects (driver registration) |
| Connection refused? | Server not running - run `go run cmd/api/main.go` |
| Cache format? | Key: `tasks:all`, Value: JSON string |
| Completed tasks key? | Not needed - can get from `tasks:all` |
| Pointer in BodyParser? | Yes - needs to write to struct |
| Pointer in JSON? | No - only reads, doesn't modify |

---

## Next Steps

1. **Fix the errors** in `task_repo.go` (Update and Complete methods)
2. **Add missing cache errors** to `errors.go`
3. **Run server:** `go run cmd/api/main.go`
4. **Test in Postman:** Make sure server is running first!


