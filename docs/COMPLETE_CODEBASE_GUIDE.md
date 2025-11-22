# Complete Codebase Guide 📚

Comprehensive explanation of the entire Task Manager API codebase - how everything works and why.

---

## 📋 Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Configuration Layer](#configuration-layer)
4. [Database Layer](#database-layer)
5. [Repository Layer](#repository-layer)
6. [Service Layer](#service-layer)
7. [Handler Layer](#handler-layer)
8. [Cache Layer](#cache-layer)
9. [Authentication Layer](#authentication-layer)
10. [Models Layer](#models-layer)
11. [Error Handling](#error-handling)
12. [Complete Request Flow](#complete-request-flow)
13. [Why This Architecture?](#why-this-architecture)

---

## 🏗️ Architecture Overview

### Layered Architecture Pattern

We use a **3-layer architecture** (also called **Clean Architecture** or **Layered Architecture**):

```
┌─────────────────────────────────────┐
│         HTTP Request                │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      HANDLER LAYER                  │  ← HTTP handling, request parsing
│  (internal/handlers/)               │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│      SERVICE LAYER                  │  ← Business logic, orchestration
│  (internal/services/)               │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│    REPOSITORY LAYER                 │  ← Database operations
│  (internal/repo/)                   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│         DATABASE                    │  ← CockroachDB (PostgreSQL)
└─────────────────────────────────────┘
```

**Plus:**
- **Cache Layer**: Redis for performance
- **Auth Layer**: JWT for authentication
- **Config Layer**: Environment variables
- **Models Layer**: Data structures

---

## 📁 Project Structure

```
task_manager/
├── cmd/
│   └── api/
│       └── main.go              # Entry point - wires everything together
├── internal/
│   ├── handlers/                # HTTP handlers (API endpoints)
│   │   ├── task_handler.go     # Task CRUD endpoints
│   │   └── auth_handler.go     # Auth endpoints (login, register)
│   ├── services/                # Business logic
│   │   ├── task_service.go     # Task business logic
│   │   └── auth_service.go     # Auth business logic
│   ├── repo/                    # Database operations
│   │   ├── db.go               # Database connection & table creation
│   │   ├── task_repo.go        # Task database queries
│   │   └── user_repo.go        # User database queries
│   ├── cache/                   # Redis caching
│   │   └── cache.go            # Cache operations
│   ├── auth/                    # JWT authentication
│   │   ├── jwt.go              # JWT service (generate/validate tokens)
│   │   └── middleware.go       # JWT middleware (protect routes)
│   ├── models/                  # Data structures
│   │   ├── task.go             # Task models
│   │   └── user.go             # User models
│   ├── config/                  # Configuration
│   │   └── config.go           # Load environment variables
│   └── errors/                  # Custom errors
│       └── errors.go           # Error definitions
├── keys/                        # RSA keys for JWT
│   ├── private.pem
│   └── public.pem
├── go.mod                       # Go dependencies
└── docker-compose.yml           # Docker services
```

---

## ⚙️ Configuration Layer

**File:** `internal/config/config.go`

### What It Does:
- Loads environment variables
- Provides default values
- Centralizes configuration

### Code:
```go
type Config struct {
    DatabaseURL      string
    RedisURL         string
    Port             string
    JWTPrivateKeyPath string
    JWTPublicKeyPath  string
}

func Load() *Config {
    return &Config{
        DatabaseURL:      getEnv("DATABASE_URL", ""),
        RedisURL:         getEnv("REDIS_URL", ""),
        Port:             getEnv("PORT", "8080"),
        JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
        JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
    }
}
```

### Why Return Pointer?
- **Efficiency**: Avoids copying large struct
- **Modification**: Allows modifying config if needed
- **Consistency**: Matches other layers (repositories, services)

### How It's Used:
```go
cfg := config.Load()
db, err := repo.InitDB(cfg.DatabaseURL)
```

---

## 🗄️ Database Layer

**File:** `internal/repo/db.go`

### What It Does:
1. Connects to CockroachDB (PostgreSQL-compatible)
2. Creates tables if they don't exist
3. Manages connection pool

### Database Schema:

**Users Table:**
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
```

**Tasks Table:**
```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    priority TEXT DEFAULT 'medium'
)
```

### Key Concepts:

**Foreign Key Constraint:**
```sql
user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
```
- **What**: Links tasks to users
- **Why**: Ensures data integrity (can't create task for non-existent user)
- **ON DELETE CASCADE**: If user deleted, their tasks are deleted too

**Connection Pool:**
```go
db, err := sql.Open("postgres", connString)
```
- **What**: `sql.DB` manages a pool of database connections
- **Why**: Reuses connections (faster than creating new ones each time)
- **Pointer**: We return `*sql.DB` to share the same pool across the app

### Code Flow:
```go
func InitDB(connString string) (*sql.DB, error) {
    // 1. Open connection
    db, err := sql.Open("postgres", connString)
    
    // 2. Test connection
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    // 3. Create tables
    db.Exec(createUsersTable)
    db.Exec(createTasksTable)
    
    return db, nil
}
```

---

## 📦 Repository Layer

**Files:** `internal/repo/task_repo.go`, `internal/repo/user_repo.go`

### What It Does:
- **Direct database operations**
- **SQL queries**
- **Data mapping** (database rows → Go structs)

### Task Repository:

#### Create Task:
```go
func (r *TaskRepository) Create(task *models.Task) (int, error) {
    query := "INSERT INTO tasks (user_id, title, description, priority) VALUES ($1, $2, $3, $4) RETURNING id"
    var id int
    err := r.db.QueryRow(query, task.UserID, task.Title, task.Description, task.Priority).Scan(&id)
    return id, nil
}
```

**Why `RETURNING id`?**
- PostgreSQL/CockroachDB feature
- Returns the generated ID in one query (no second query needed)
- More efficient than `SELECT id FROM tasks WHERE ...`

**Why `QueryRow().Scan()`?**
- `QueryRow`: For queries that return exactly one row
- `Scan`: Copies database values into Go variables
- `&id`: Pass address so `Scan` can write to it

#### Get All Tasks:
```go
func (r *TaskRepository) GetAll(userID int, showCompleted bool) ([]models.Task, error) {
    query := "SELECT id, user_id, title, description, completed, created_at, priority FROM tasks WHERE user_id = $1"
    if !showCompleted {
        query += " AND completed = FALSE"
    }
    
    rows, err := r.db.Query(query, userID)
    defer rows.Close()
    
    var tasks []models.Task
    for rows.Next() {
        var task models.Task
        rows.Scan(&task.ID, &task.UserID, ...)
        tasks = append(tasks, task)
    }
    return tasks, nil
}
```

**Why `rows.Next()`?**
- Iterates through result set row by row
- More memory-efficient than loading all rows at once
- `rows.Scan()` reads one row at a time

**Why `defer rows.Close()`?**
- Ensures database resources are freed
- Runs even if function returns early (error)
- Prevents connection leaks

#### Update Task:
```go
func (r *TaskRepository) Update(id int, userID int, task *models.Task) error {
    // 1. Check task exists and belongs to user
    existingTask, err := r.GetByID(id)
    if existingTask.UserID != userID {
        return errors.ErrTaskNotFound
    }
    
    // 2. Update task
    query := "UPDATE tasks SET title = $1, description = $2, priority = $3 WHERE id = $4 AND user_id = $5"
    _, err = r.db.Exec(query, task.Title, task.Description, task.Priority, id, userID)
    return err
}
```

**Why Check Ownership First?**
- **Security**: Prevents users from updating other users' tasks
- **Error Handling**: Returns "not found" instead of "unauthorized" (doesn't reveal task exists)

**Why `AND user_id = $5`?**
- **Double protection**: Even if check fails, SQL prevents unauthorized updates
- **Defense in depth**: Multiple layers of security

#### Delete Task:
```go
func (r *TaskRepository) Delete(id int, userID int) error {
    query := "DELETE FROM tasks WHERE id = $1 AND user_id = $2"
    result, err := r.db.Exec(query, id, userID)
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return errors.ErrTaskNotFound
    }
    return nil
}
```

**Why Check `RowsAffected()`?**
- Confirms deletion actually happened
- If 0 rows affected → task doesn't exist or doesn't belong to user
- Better error handling than just checking `err`

### User Repository:

#### Create User:
```go
func (r *UserRepository) CreateUser(user *models.User) (int, error) {
    query := "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id"
    var id int
    err := r.db.QueryRow(query, user.Username, user.Email, user.Password).Scan(&id)
    return id, nil
}
```

**Why `UNIQUE` constraint?**
- Database enforces uniqueness (prevents duplicate usernames/emails)
- Returns error if duplicate → service layer handles it

#### Get By Username:
```go
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
    query := "SELECT id, username, email, password FROM users WHERE username = $1"
    var user models.User
    err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
    if err == sql.ErrNoRows {
        return nil, errors.New("user not found")
    }
    return &user, nil
}
```

**Why Check `sql.ErrNoRows`?**
- `QueryRow` returns error when no rows found
- Distinguishes "not found" from "database error"
- Service layer can handle differently

---

## 🧠 Service Layer

**Files:** `internal/services/task_service.go`, `internal/services/auth_service.go`

### What It Does:
- **Business logic** (rules, validation)
- **Orchestration** (coordinates repository + cache)
- **Data transformation**

### Task Service:

#### GetAllTasks:
```go
func (s *TaskService) GetAllTasks(userID int, showCompleted bool) ([]models.Task, error) {
    // 1. Try cache first
    if tasks, err := s.cache.GetTasks(showCompleted); err == nil {
        return tasks, nil  // Cache hit!
    }
    
    // 2. Cache miss - get from database
    tasks, err := s.repo.GetAll(userID, showCompleted)
    
    // 3. Store in cache for next time
    s.cache.SetTasks(showCompleted, tasks)
    
    return tasks, nil
}
```

**Why Cache First?**
- **Performance**: Redis is much faster than database
- **Reduces load**: Fewer database queries
- **User experience**: Faster response times

**Cache Strategy:**
- **Cache-aside pattern**: Check cache → if miss, get from DB → store in cache
- **TTL**: 5 minutes (cache expires automatically)
- **Invalidation**: Clear cache when data changes (create/update/delete)

#### CreateTask:
```go
func (s *TaskService) CreateTask(userID int, task *models.Task) error {
    // 1. Validate title
    if task.Title == "" {
        return apperrors.ErrTitleRequired
    }
    
    // 2. Set user ID
    task.UserID = userID
    
    // 3. Create in database
    id, err := s.repo.Create(task)
    task.ID = id
    
    // 4. Invalidate cache
    s.cache.Invalidate()
    
    return nil
}
```

**Why Validate in Service?**
- **Business rule**: "Title is required" is business logic, not database constraint
- **Separation**: Repository handles data, service handles rules
- **Reusability**: Can be used by different handlers/APIs

**Why Invalidate Cache?**
- **Data consistency**: Cache now has stale data
- **User experience**: Next request gets fresh data
- **Simple**: Clear all cache (could be more granular)

#### GetTaskByID:
```go
func (s *TaskService) GetTaskByID(id int, userID int) (*models.Task, error) {
    // 1. Try cache
    if task, err := s.cache.GetTask(id); err == nil {
        if task.UserID == userID {
            return task, nil  // Cache hit!
        }
        // Wrong user's task - fetch from DB
    }
    
    // 2. Get from database
    task, err := s.repo.GetByID(id)
    
    // 3. Verify ownership
    if task.UserID != userID {
        return nil, errors.ErrTaskNotFound
    }
    
    // 4. Store in cache
    s.cache.SetTask(id, task)
    
    return task, nil
}
```

**Why Check Ownership in Service?**
- **Security**: Repository doesn't know about users
- **Business rule**: "Users can only see their own tasks"
- **Consistency**: Same check in all service methods

**Why Verify Cached Task?**
- **Security**: Cache might have another user's task
- **Data integrity**: Ensures user only sees their data
- **Defense**: Multiple layers of security

#### CompleteTask:
```go
func (s *TaskService) CompleteTask(id int, userID int) (*models.Task, error) {
    // 1. Update in database
    if err := s.repo.Complete(id, userID); err != nil {
        return nil, err
    }
    
    // 2. Invalidate cache
    s.cache.Invalidate()
    
    // 3. Fetch updated task
    completedTask, err := s.repo.GetByID(id)
    
    // 4. Verify ownership
    if completedTask.UserID != userID {
        return nil, errors.ErrTaskNotFound
    }
    
    return completedTask, nil
}
```

**Why Fetch After Update?**
- **Complete data**: Update only changes `completed` field
- **Response**: Handler needs full task (ID, created_at, etc.)
- **Consistency**: Returns same structure as other methods

### Auth Service:

#### Register:
```go
func (s *AuthService) Register(user *models.User) (int, error) {
    // 1. Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    user.Password = string(hashedPassword)
    
    // 2. Create user in database
    id, err := s.userRepo.CreateUser(user)
    
    return id, nil
}
```

**Why Hash Password?**
- **Security**: Never store plaintext passwords
- **bcrypt**: Industry-standard hashing algorithm
- **One-way**: Can't reverse hash → password

**Why `bcrypt.DefaultCost`?**
- **Balance**: Security vs. performance
- **Default**: Usually 10 rounds (good balance)
- **Configurable**: Can increase for more security

#### Login:
```go
func (s *AuthService) Login(user *models.User) (*models.User, error) {
    // 1. Get user from database
    dbUser, err := s.userRepo.GetByUsername(user.Username)
    
    // 2. Compare password hash
    err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password))
    if err != nil {
        return nil, errors.New("invalid credentials")
    }
    
    return dbUser, nil
}
```

**Why `CompareHashAndPassword`?**
- **Secure comparison**: Constant-time comparison (prevents timing attacks)
- **bcrypt**: Handles salt automatically (stored in hash)
- **One-way**: Can't reverse, only compare

**Why Return User?**
- **Handler needs**: User ID and username for JWT token
- **Complete data**: Handler returns user info to client
- **Consistency**: Same pattern as other services

---

## 🌐 Handler Layer

**Files:** `internal/handlers/task_handler.go`, `internal/handlers/auth_handler.go`

### What It Does:
- **HTTP handling** (request/response)
- **Request parsing** (JSON, URL params, query params)
- **Response formatting** (JSON)
- **Error handling** (HTTP status codes)

### Task Handler:

#### CreateTask:
```go
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
    // 1. Get user ID from JWT context
    userID := c.Locals("userID").(int)
    
    // 2. Parse JSON body
    var task models.Task
    if err := c.BodyParser(&task); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid JSON",
        })
    }
    
    // 3. Call service
    if err := h.service.CreateTask(userID, &task); err != nil {
        if errors.Is(err, apperrors.ErrTitleRequired) {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "Title is required",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    // 4. Return created task
    return c.Status(fiber.StatusCreated).JSON(task)
}
```

**Why Get UserID from Context?**
- **JWT Middleware**: Sets `userID` in context after validating token
- **Security**: User can only create tasks for themselves
- **No spoofing**: UserID comes from token, not request body

**Why Check Error Type?**
- **Different status codes**: Business errors (400) vs. server errors (500)
- **User-friendly**: "Title is required" vs. "Internal server error"
- **Security**: Don't leak internal errors

**HTTP Status Codes:**
- `200 OK`: Success (GET, PUT)
- `201 Created`: Resource created (POST)
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Not authenticated
- `403 Forbidden`: Authenticated but not authorized
- `404 Not Found`: Resource doesn't exist
- `500 Internal Server Error`: Server error

#### GetAllTasks:
```go
func (h *TaskHandler) GetAllTasks(c *fiber.Ctx) error {
    // 1. Get user ID
    userID := c.Locals("userID").(int)
    
    // 2. Get query parameter
    showCompleted := c.Query("showCompleted") == "true"
    
    // 3. Call service
    tasks, err := h.service.GetAllTasks(userID, showCompleted)
    
    // 4. Return JSON
    return c.JSON(tasks)
}
```

**Why Query Parameter?**
- **Flexibility**: Client chooses what data to get
- **Performance**: Can filter incomplete tasks (less data)
- **RESTful**: Query params for filtering

**Why `== "true"`?**
- **String comparison**: Query params are strings
- **Boolean conversion**: Converts "true" → `true`, anything else → `false`
- **Simple**: No need for complex parsing

### Auth Handler:

#### Register:
```go
func (h *AuthHandler) Register(c *fiber.Ctx) error {
    // 1. Parse request
    var request models.RegisterRequest
    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }
    
    // 2. Create user model
    user := &models.User{
        Username: request.Username,
        Email:    request.Email,
        Password: request.Password,
    }
    
    // 3. Register user
    id, err := h.authService.Register(user)
    user.ID = id
    
    // 4. Generate JWT token
    token, err := h.jwtService.GenerateToken(id, user.Username)
    
    // 5. Return token + user
    return c.Status(fiber.StatusCreated).JSON(models.AuthResponse{
        Token: token,
        User:  *user,
    })
}
```

**Why Generate Token Here?**
- **Convenience**: User logged in immediately after registration
- **UX**: No need for separate login step
- **Consistency**: Same response format as login

**Why Return User?**
- **Client needs**: User ID, username, email
- **Complete data**: Client can display user info
- **Consistency**: Same pattern as login

#### Login:
```go
func (h *AuthHandler) Login(c *fiber.Ctx) error {
    // 1. Parse request
    var request models.LoginRequest
    c.BodyParser(&request)
    
    // 2. Login user
    user, err := h.authService.Login(&models.User{
        Username: request.Username,
        Password: request.Password,
    })
    
    // 3. Generate JWT token
    token, err := h.jwtService.GenerateToken(user.ID, user.Username)
    
    // 4. Return token + user
    return c.Status(fiber.StatusOK).JSON(models.AuthResponse{
        Token: token,
        User:  *user,
    })
}
```

**Why Same Response as Register?**
- **Consistency**: Same API contract
- **Client simplicity**: Same handling for both endpoints
- **UX**: Seamless experience

---

## 💾 Cache Layer

**File:** `internal/cache/cache.go`

### What It Does:
- **Stores data in Redis** (in-memory cache)
- **Reduces database load**
- **Improves performance**

### Cache Operations:

#### GetTasks:
```go
func (c *Cache) GetTasks(showCompleted bool) ([]models.Task, error) {
    key := "tasks:all"
    if !showCompleted {
        key = "tasks:incomplete"
    }
    
    // Get from Redis
    val, err := c.client.Get(context.Background(), key).Result()
    
    // Unmarshal JSON
    var tasks []models.Task
    json.Unmarshal([]byte(val), &tasks)
    
    return tasks, nil
}
```

**Why Two Keys?**
- **Different data**: "all tasks" vs. "incomplete tasks"
- **Flexibility**: Client can choose what to cache
- **Performance**: Cache both separately (faster)

**Why JSON?**
- **Redis stores strings**: Can't store Go structs directly
- **JSON**: Standard format, easy to serialize/deserialize
- **Flexibility**: Can read from other languages too

#### SetTasks:
```go
func (c *Cache) SetTasks(showCompleted bool, tasks []models.Task) error {
    key := "tasks:all"
    if !showCompleted {
        key = "tasks:incomplete"
    }
    
    // Marshal to JSON
    data, err := json.Marshal(tasks)
    
    // Store in Redis with TTL
    err = c.client.Set(context.Background(), key, data, time.Minute*5).Err()
    
    return nil
}
```

**Why TTL (Time To Live)?**
- **Fresh data**: Cache expires after 5 minutes
- **Automatic cleanup**: Redis deletes expired keys
- **Balance**: Fresh enough, cached long enough

**Why 5 Minutes?**
- **Task data**: Doesn't change frequently
- **Performance**: Reduces database queries
- **Freshness**: Still relatively fresh

#### Invalidate:
```go
func (c *Cache) Invalidate() error {
    return c.client.Del(context.Background(), "tasks:all", "tasks:incomplete").Err()
}
```

**Why Invalidate?**
- **Data changed**: Cache has stale data
- **Consistency**: Next request gets fresh data
- **Simple**: Clear all cache (could be more granular)

**When to Invalidate?**
- **Create**: New task added
- **Update**: Task modified
- **Delete**: Task removed
- **Complete**: Task status changed

---

## 🔐 Authentication Layer

**Files:** `internal/auth/jwt.go`, `internal/auth/middleware.go`

### JWT Service:

#### GenerateToken:
```go
func (s *JWTService) GenerateToken(userId int, username string) (string, error) {
    claims := &Claims{
        UserID: userId,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt: jwt.NewNumericDate(time.Now()),
            Issuer: "task_manager",
            Subject: fmt.Sprintf("%d", userId),
            Audience: jwt.ClaimStrings{"task_manager"},
            NotBefore: jwt.NewNumericDate(time.Now()),
            ID: uuid.New().String(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    tokenString, err := token.SignedString(s.privateKey)
    
    return tokenString, nil
}
```

**Why RS256?**
- **Asymmetric**: Private key signs, public key verifies
- **Security**: Private key never leaves server
- **Scalability**: Multiple servers can verify (share public key)

**Why UUID for JTI?**
- **Unique**: Each token has unique ID
- **Tracking**: Can identify/revoke specific tokens
- **Security**: Prevents token reuse attacks

#### ValidateToken:
```go
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.publicKey, nil
    })
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

**Why Public Key?**
- **Verification**: Public key verifies signature (doesn't sign)
- **Security**: Can't generate tokens with public key
- **Distribution**: Can share public key safely

### JWT Middleware:

```go
func JWTMiddleware(jwtService *JWTService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // 1. Get token from header
        tokenString, err := jwtService.GetTokenFromHeader(c)
        
        // 2. Validate token
        claims, err := jwtService.ValidateToken(tokenString)
        
        // 3. Store user info in context
        c.Locals("userID", claims.UserID)
        c.Locals("username", claims.Username)
        
        // 4. Continue to handler
        return c.Next()
    }
}
```

**Why Middleware?**
- **DRY**: Don't repeat validation in every handler
- **Security**: Centralized authentication
- **Consistency**: Same validation for all protected routes

**Why `c.Locals()`?**
- **Context**: Stores data for this request only
- **Access**: Handlers can access user info
- **Security**: Can't be modified by client

**Why `c.Next()`?**
- **Chain**: Continues to next middleware/handler
- **Flow**: Request continues if token valid
- **Stop**: Returns error if token invalid (doesn't call `Next()`)

---

## 📊 Models Layer

**Files:** `internal/models/task.go`, `internal/models/user.go`

### What It Does:
- **Defines data structures**
- **Request/Response DTOs** (Data Transfer Objects)
- **JSON serialization**

### Task Model:
```go
type Task struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    Priority    string    `json:"priority"`
}
```

**Why JSON Tags?**
- **Serialization**: Controls how struct is converted to JSON
- **Naming**: `user_id` in JSON, `UserID` in Go
- **Consistency**: Matches API contract

**Why `UserID`?**
- **Ownership**: Links task to user
- **Security**: Users can only see their tasks
- **Database**: Foreign key constraint

### User Model:
```go
type User struct {
    ID       int    `json:"id"`
    Username string `json:"username"`
    Password string `json:"-"`  // Hidden in JSON
    Email    string `json:"email"`
}
```

**Why `json:"-"` for Password?**
- **Security**: Never send password in JSON response
- **Privacy**: Password hash stays on server
- **Best practice**: Never expose passwords

### Request Models:
```go
type RegisterRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
```

**Why Separate Request Models?**
- **Validation**: Different fields for different endpoints
- **Security**: Only expose needed fields
- **Clarity**: Clear API contract

---

## ⚠️ Error Handling

**File:** `internal/errors/errors.go`

### What It Does:
- **Defines custom errors**
- **Consistent error messages**
- **Error wrapping** (adds context)

### Custom Errors:
```go
var (
    ErrTaskNotFound = errors.New("task not found")
    ErrTitleRequired = errors.New("title is required")
    ErrDatabaseError = errors.New("database error")
    // ... more errors
)
```

**Why Package-Level Variables?**
- **Reusability**: Same error across codebase
- **Comparison**: Can use `errors.Is()` to check error type
- **Consistency**: Same error message everywhere

### Error Wrapping:
```go
func Wrap(err error, message string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", message, err)
}
```

**Why Wrap Errors?**
- **Context**: Adds information about where error occurred
- **Debugging**: Easier to trace error source
- **Chain**: Can unwrap to get original error

**Usage:**
```go
if err != nil {
    return apperrors.Wrap(err, "error creating user")
}
// Error: "error creating user: database connection failed"
```

**Why `%w`?**
- **Wrapping**: Preserves original error
- **Unwrapping**: Can use `errors.Unwrap()` or `errors.Is()`
- **Chain**: Maintains error chain for debugging

---

## 🔄 Complete Request Flow

### Example: Create Task

```
1. Client sends POST /tasks
   {
     "title": "Buy groceries",
     "description": "Milk, bread, eggs",
     "priority": "high"
   }
   Header: Authorization: Bearer <token>

2. JWT Middleware runs:
   - Extracts token from header
   - Validates token
   - Sets c.Locals("userID", 123)
   - Calls c.Next()

3. TaskHandler.CreateTask:
   - Gets userID from context: 123
   - Parses JSON body → Task struct
   - Calls service.CreateTask(123, task)

4. TaskService.CreateTask:
   - Validates title (not empty)
   - Sets task.UserID = 123
   - Calls repo.Create(task)

5. TaskRepository.Create:
   - Executes SQL: INSERT INTO tasks ...
   - Returns task ID: 456
   - Returns (456, nil)

6. TaskService.CreateTask:
   - Sets task.ID = 456
   - Invalidates cache
   - Returns nil

7. TaskHandler.CreateTask:
   - Returns HTTP 201 Created
   - JSON: { "id": 456, "title": "Buy groceries", ... }

8. Client receives response
```

### Example: Get All Tasks

```
1. Client sends GET /tasks?showCompleted=false
   Header: Authorization: Bearer <token>

2. JWT Middleware:
   - Validates token
   - Sets userID = 123

3. TaskHandler.GetAllTasks:
   - Gets userID = 123
   - Parses query param: showCompleted = false
   - Calls service.GetAllTasks(123, false)

4. TaskService.GetAllTasks:
   - Tries cache: cache.GetTasks(false)
   - Cache miss → Calls repo.GetAll(123, false)

5. TaskRepository.GetAll:
   - Executes SQL: SELECT ... WHERE user_id = 123 AND completed = FALSE
   - Returns []Task

6. TaskService.GetAllTasks:
   - Stores in cache: cache.SetTasks(false, tasks)
   - Returns tasks

7. TaskHandler.GetAllTasks:
   - Returns HTTP 200 OK
   - JSON: [{ "id": 1, ... }, { "id": 2, ... }]

8. Client receives response
```

---

## ✅ Why This Architecture?

### Separation of Concerns:
- **Handlers**: HTTP only
- **Services**: Business logic only
- **Repositories**: Database only
- **Each layer**: One responsibility

### Testability:
- **Mock dependencies**: Can mock repository for service tests
- **Unit tests**: Test each layer independently
- **Integration tests**: Test layers together

### Maintainability:
- **Easy to change**: Change one layer without affecting others
- **Clear structure**: Easy to find code
- **Scalability**: Can add features without breaking existing code

### Reusability:
- **Service layer**: Can be used by different handlers (REST, GraphQL, gRPC)
- **Repository**: Can be used by different services
- **Models**: Shared across layers

### Security:
- **Multiple layers**: Validation at handler, service, repository
- **JWT middleware**: Centralized authentication
- **User ownership**: Checked at service layer

### Performance:
- **Caching**: Redis reduces database load
- **Connection pooling**: Reuses database connections
- **Efficient queries**: Optimized SQL queries

---

## 🎯 Key Concepts Summary

### Dependency Injection:
- **What**: Pass dependencies as parameters (not create inside)
- **Why**: Makes code testable, flexible
- **Example**: `NewTaskService(repo, cache)` instead of creating inside

### Context (`c.Locals()`):
- **What**: Stores request-specific data
- **Why**: Share data between middleware and handlers
- **Example**: `c.Locals("userID", 123)` → `c.Locals("userID")`

### Error Handling:
- **What**: Return errors, don't panic
- **Why**: Graceful error handling, better UX
- **Example**: `if err != nil { return err }`

### SQL Injection Prevention:
- **What**: Use placeholders (`$1`, `$2`) instead of string concatenation
- **Why**: Prevents SQL injection attacks
- **Example**: `WHERE id = $1` not `WHERE id = " + id`

### Cache Strategy:
- **What**: Cache-aside pattern (check cache → DB → store in cache)
- **Why**: Reduces database load, improves performance
- **Example**: `GetAllTasks()` checks cache first

### Security:
- **What**: Multiple layers of validation
- **Why**: Defense in depth
- **Example**: Handler validates input, service validates business rules, repository validates ownership

---

## 📝 Conclusion

This codebase follows **professional backend architecture patterns**:

1. **Layered Architecture**: Clear separation of concerns
2. **Dependency Injection**: Testable, flexible code
3. **Error Handling**: Consistent, user-friendly errors
4. **Security**: Multiple layers of protection
5. **Performance**: Caching, connection pooling
6. **Maintainability**: Clear structure, easy to understand

Each layer has a **single responsibility**, making the code:
- **Easy to understand**
- **Easy to test**
- **Easy to maintain**
- **Easy to extend**

This is how **real-world production applications** are built! 🚀

