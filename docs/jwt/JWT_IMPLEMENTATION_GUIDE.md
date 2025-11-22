# JWT RS256 Implementation Guide - Step by Step

## 📋 Overview

You'll implement JWT authentication with RS256 (RSA public/private keys) so that:
1. Users can register/login
2. Each task belongs to a specific user
3. Users can only see/modify their own tasks
4. Protected routes require valid JWT tokens

---

## 🔑 Step 1: Generate RSA Keys

### What are RSA keys?
RSA keys are a pair of cryptographic keys:
- **Private Key**: A secret key that only your server knows. Used to **sign** (create) JWT tokens.
- **Public Key**: A key that can be shared publicly. Used to **verify** (check) JWT tokens.

### Why do we need them?
- **Security**: RS256 uses asymmetric encryption (different keys for signing vs verifying)
- **Private key stays secret**: Only server can create tokens (can't be forged)
- **Public key can be shared**: Other services can verify tokens without seeing private key
- **Better than HS256**: More secure than symmetric keys (single shared secret)

### Where are they used?
- **Private key**: Used in `internal/auth/jwt.go` → `GenerateToken()` function (when user logs in)
- **Public key**: Used in `internal/auth/jwt.go` → `ValidateToken()` function (on every protected request)
- **In Docker**: Keys are mounted from your local `keys/` directory into container at `/app/keys/`

### How?

1. **Create keys directory:**
   ```bash
   mkdir -p keys
   ```

2. **Generate private key:**
   ```bash
   openssl genrsa -out keys/private.pem 2048
   ```

3. **Extract public key:**
   ```bash
   openssl rsa -in keys/private.pem -pubout -out keys/public.pem
   ```

4. **Set permissions (security):**
   ```bash
   chmod 600 keys/private.pem  # Only owner can read/write
   chmod 644 keys/public.pem   # Readable by all
   ```

5. **Add to .gitignore:**
   ```
   keys/private.pem
   ```

### ✅ Check:
- `keys/private.pem` exists
- `keys/public.pem` exists
- Private key is NOT in git

---

## 📁 Step 2: Update Config

### File: `internal/config/config.go`

### What does this do?
Adds two new configuration fields to store the file paths where your RSA keys are located.

### Why do we need this?
- **Flexibility**: Can change key paths via environment variables (useful for Docker, different environments)
- **Centralized config**: All configuration in one place
- **Default values**: If env vars not set, uses default paths (`keys/private.pem`, `keys/public.pem`)

### Where is this used?
- Used in `cmd/api/main.go` → `main()` function
- Passed to `auth.NewJWTService(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)`
- Allows JWT service to find and load the key files

### What to add:

1. **Add fields to `Config` struct:**
   ```go
   type Config struct {
       DatabaseURL      string
       RedisURL         string
       Port             string
       JWTPrivateKeyPath string  // NEW
       JWTPublicKeyPath  string  // NEW
   }
   ```

2. **Update `Load()` function:**
   ```go
   func Load() *Config {
       return &Config{
           DatabaseURL:      getEnv("DATABASE_URL", ""),
           RedisURL:         getEnv("REDIS_URL", ""),
           Port:             getEnv("PORT", "8080"),
           JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),  // NEW
           JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),      // NEW
       }
   }
   ```

### ✅ Check:
- Config struct has both JWT key paths
- Default paths point to `keys/` directory

---

## 🔐 Step 3: Create JWT Service

### File: `internal/auth/jwt.go` (NEW FILE)

### What does this do?
This is the core JWT service that:
1. **Loads RSA keys** from PEM files into memory
2. **Generates tokens** when user logs in (signs with private key)
3. **Validates tokens** on every protected request (verifies with public key)

### Why do we need this?
- **Centralized JWT logic**: All token operations in one place
- **Reusable**: Can be used by multiple handlers (login, register, middleware)
- **Security**: Handles all cryptographic operations safely
- **Separation of concerns**: JWT logic separate from business logic

### Where is this used?
- **Created in**: `cmd/api/main.go` → `main()` function
- **Used by**: 
  - `internal/handlers/auth_handler.go` → `Login()` and `Register()` (generate tokens)
  - `internal/auth/middleware.go` → `JWTMiddleware()` (validate tokens)
- **Flow**:
  1. User logs in → Handler calls `jwtService.GenerateToken()` → Returns token
  2. User makes request → Middleware calls `jwtService.ValidateToken()` → Extracts user info

### What to write:

1. **Package declaration and imports:**
   ```go
   package auth

   import (
       "crypto/rsa"
       "crypto/x509"
       "encoding/pem"
       "fmt"
       "os"
       "time"
       "github.com/golang-jwt/jwt/v5"
   )
   ```

2. **JWTService struct:**
   ```go
   type JWTService struct {
       privateKey *rsa.PrivateKey
       publicKey  *rsa.PublicKey
   }
   ```

3. **Claims struct (what goes in the token):**
   ```go
   type Claims struct {
       UserID   int    `json:"user_id"`    // Which user this token belongs to
       Username string `json:"username"`   // User's username (for convenience)
       jwt.RegisteredClaims                 // Standard JWT fields (expiration, issued at, etc.)
   }
   ```
   **What is Claims?** Claims are the data stored inside the JWT token. When you decode a token, you get these fields. `RegisteredClaims` includes expiration time, issued time, etc.

4. **NewJWTService function:**
   - Read private key file
   - Decode PEM
   - Parse RSA private key
   - Read public key file
   - Decode PEM
   - Parse RSA public key
   - Return JWTService with both keys

5. **GenerateToken function:**
   - Create Claims with userID, username, expiration (24 hours)
   - Create JWT token with RS256 method
   - Sign with private key
   - Return token string

6. **ValidateToken function:**
   - Parse token string
   - Verify signature with public key
   - Check expiration
   - Return Claims if valid

### ✅ Check:
- Can load keys from files
- Can generate tokens
- Can validate tokens

---

## 🛡️ Step 4: Create JWT Middleware

### File: `internal/auth/middleware.go` (NEW FILE)

### What does this do?
Middleware is a function that runs **before** your handlers. It:
1. Intercepts every HTTP request
2. Extracts JWT token from `Authorization` header
3. Validates the token using JWT service
4. If valid: Stores user info in context and allows request to continue
5. If invalid: Returns 401 Unauthorized and stops the request

### Why do we need this?
- **Protection**: Prevents unauthorized access to protected routes
- **Automatic**: Don't need to check token in every handler
- **Reusable**: Apply once, protects all routes in a group
- **Context**: Makes user info available to handlers via `c.Locals()`

### Where is this used?
- **Applied in**: `cmd/api/main.go` → `app.Group("/", auth.JWTMiddleware(jwtService))`
- **Protects**: All routes added to the `protected` group
- **Flow**:
  1. Request comes in → Middleware runs first
  2. Middleware checks token → If valid, sets `c.Locals("userID")` and `c.Locals("username")`
  3. Handler runs → Can access user info from context
  4. If token invalid → Handler never runs, returns 401 error

### What to write:

1. **JWTMiddleware function:**
   - Get `Authorization` header
   - Extract token (format: "Bearer <token>")
   - Validate token using JWTService
   - Store userID and username in context (`c.Locals()`)
   - Call `c.Next()` to continue

2. **Error handling:**
   - Missing header → 401 Unauthorized
   - Invalid format → 401 Unauthorized
   - Invalid token → 401 Unauthorized

### ✅ Check:
- Extracts token from header
- Validates token
- Sets userID in context

---

## 👤 Step 5: Create User Model

### File: `internal/models/user.go` (NEW FILE)

### What does this do?
Defines the data structures (structs) used for:
- **User**: Represents a user in the database
- **LoginRequest**: Data sent when user logs in
- **RegisterRequest**: Data sent when user registers
- **AuthResponse**: Data returned after login/register (includes token)

### Why do we need this?
- **Type safety**: Go structs ensure correct data types
- **JSON mapping**: Tags (`json:"username"`) map Go structs to JSON
- **Security**: `json:"-"` on Password prevents it from being returned in JSON responses
- **Consistency**: Same data structures used across handlers, services, and repository

### Where is this used?
- **User struct**: Used in repository (database), service (business logic), handlers (HTTP)
- **LoginRequest**: Used in `internal/handlers/auth_handler.go` → `Login()` handler
- **RegisterRequest**: Used in `internal/handlers/auth_handler.go` → `Register()` handler
- **AuthResponse**: Returned by Login and Register handlers to client

### What to write:

1. **User struct:**
   ```go
   type User struct {
       ID       int    `json:"id" db:"id"`
       Username string `json:"username" db:"username"`
       Email    string `json:"email" db:"email"`
       Password string `json:"-" db:"password"`  // Never return in JSON
   }
   ```

2. **LoginRequest struct:**
   ```go
   type LoginRequest struct {
       Username string `json:"username"`
       Password string `json:"password"`
   }
   ```

3. **RegisterRequest struct:**
   ```go
   type RegisterRequest struct {
       Username string `json:"username"`
       Email    string `json:"email"`
       Password string `json:"password"`
   }
   ```

4. **AuthResponse struct:**
   ```go
   type AuthResponse struct {
       Token string `json:"token"`
       User  User   `json:"user"`
   }
   ```

### ✅ Check:
- User struct has all fields
- Password has `json:"-"` tag (never returned)

---

## 🗄️ Step 6: Update Database Schema

### File: `internal/repo/db.go`

### What does this do?
Modifies the database schema to:
1. **Create `users` table**: Stores user accounts (username, email, password hash)
2. **Add `user_id` to `tasks` table**: Links each task to a user
3. **Add foreign key constraint**: Ensures `user_id` always references a valid user

### Why do we need this?
- **User storage**: Need somewhere to store user accounts
- **Task ownership**: Each task must belong to a user
- **Data integrity**: Foreign key prevents orphaned tasks (if user deleted, tasks deleted too via `ON DELETE CASCADE`)
- **Security**: Database enforces that tasks have valid owners

### Where is this used?
- **Executed**: When `repo.InitDB()` is called in `cmd/api/main.go`
- **Users table**: Used by `internal/repo/user_repository.go` for all user operations
- **Tasks table**: Used by `internal/repo/task_repo.go` - all queries filter by `user_id`
- **Foreign key**: Database automatically enforces relationship (can't create task with invalid user_id)

### What to change:

1. **Add users table creation:**
   ```go
   createUsersTable := `
       CREATE TABLE IF NOT EXISTS users (
           id SERIAL PRIMARY KEY,
           username TEXT NOT NULL UNIQUE,
           email TEXT NOT NULL UNIQUE,
           password TEXT NOT NULL,
           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
       )`
   if _, err := db.Exec(createUsersTable); err != nil {
       db.Close()
       return nil, fmt.Errorf("failed to create users table: %w", err)
   }
   ```

2. **Update tasks table:**
   ```go
   createTasksTable := `
       CREATE TABLE IF NOT EXISTS tasks ( 
           id SERIAL PRIMARY KEY,
           user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
           title TEXT NOT NULL,
           description TEXT,
           completed BOOLEAN DEFAULT FALSE,
           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
           priority TEXT DEFAULT 'medium'
       )`
   ```

### ✅ Check:
- Users table created
- Tasks table has `user_id` foreign key
- Foreign key references `users(id)`

---

## 📝 Step 7: Create User Repository

### File: `internal/repo/user_repository.go` (NEW FILE)

### What does this do?
Handles all database operations for users:
- **Create**: Insert new user into database
- **GetByUsername**: Find user by username (for login)
- **GetByID**: Find user by ID (for profile, etc.)

### Why do we need this?
- **Separation of concerns**: Database logic separate from business logic
- **Reusable**: Can be used by auth service, handlers, etc.
- **Consistent pattern**: Same structure as TaskRepository (you already know this pattern!)
- **Error handling**: Handles database errors (`sql.ErrNoRows`, etc.)

### Where is this used?
- **Created in**: `cmd/api/main.go` → `main()` function
- **Used by**: `internal/services/auth_service.go` → All methods use userRepo
- **Flow**:
  1. Auth service needs user data → Calls userRepo method
  2. UserRepo executes SQL → Returns user struct or error
  3. Auth service uses result → Returns to handler

### What to write:

1. **UserRepository struct:**
   ```go
   type UserRepository struct {
       db *sql.DB
   }
   ```

2. **NewUserRepository function:**
   - Return pointer to UserRepository

3. **Create method:**
   - Insert user (username, email, hashed password)
   - Return user with ID

4. **GetByUsername method:**
   - SELECT by username
   - Return user (including password hash)
   - Handle `sql.ErrNoRows`

5. **GetByID method:**
   - SELECT by ID
   - Return user (without password)
   - Handle `sql.ErrNoRows`

### ✅ Check:
- Can create users
- Can find users by username
- Can find users by ID

---

## 🔧 Step 8: Create Auth Service

### File: `internal/services/auth_service.go` (NEW FILE)

### What does this do?
Handles the business logic for authentication:
- **Register**: Validates input, hashes password, creates user
- **Login**: Validates credentials, compares password hash
- **GetUserByID**: Retrieves user information

### Why do we need this?
- **Password security**: Never store plain passwords! bcrypt hashes them
- **Business logic**: Validation rules, password requirements, etc.
- **Separation**: Auth logic separate from HTTP handlers and database
- **Reusable**: Can be used by multiple handlers or other services

### Where is this used?
- **Created in**: `cmd/api/main.go` → `main()` function
- **Used by**: `internal/handlers/auth_handler.go` → Login, Register, GetProfile handlers
- **Flow**:
  1. Handler receives request → Parses JSON body
  2. Handler calls authService → Passes credentials
  3. AuthService validates → Hashes password (register) or compares hash (login)
  4. AuthService calls userRepo → Database operations
  5. Returns user → Handler generates JWT token → Returns to client

### What to write:

1. **Import bcrypt:**
   ```go
   import "golang.org/x/crypto/bcrypt"
   ```

2. **AuthService struct:**
   ```go
   type AuthService struct {
       userRepo *repo.UserRepository
   }
   ```

3. **NewAuthService function:**
   - Return pointer to AuthService

4. **Register method:**
   - Validate input (username, email, password required)
   - Hash password with `bcrypt.GenerateFromPassword()`
   - Create user via repository
   - Return user

5. **Login method:**
   - Get user by username
   - Compare password with `bcrypt.CompareHashAndPassword()`
   - Return user if valid

6. **GetUserByID method:**
   - Get user from repository
   - Return user

### ✅ Check:
- Passwords are hashed
- Can register users
- Can login users
- Password comparison works

---

## 🎯 Step 9: Update Task Model

### File: `internal/models/task.go`

### What does this do?
Adds `UserID` field to the `Task` struct so each task knows which user owns it.

### Why do we need this?
- **Ownership**: Every task must belong to a user
- **Security**: Can filter tasks by user_id to show only user's tasks
- **Data model**: Matches database schema (tasks table has user_id column)
- **JSON response**: Clients can see which user owns each task

### Where is this used?
- **In database**: `user_id` column stores the owner
- **In handlers**: When creating task, `UserID` is set from JWT context
- **In repository**: All queries filter by `user_id`
- **In service**: Verifies task belongs to user before operations
- **In JSON**: Returned in API responses (`"user_id": 123`)

### What to change:

Add `UserID` field to `Task` struct:
```go
type Task struct {
    ID          int       `json:"id"`
    UserID      int       `json:"user_id" db:"user_id"`  // NEW
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    Priority    string    `json:"priority"`
}
```

### ✅ Check:
- Task struct has `UserID` field

---

## 🔄 Step 10: Update Task Repository

### File: `internal/repo/task_repo.go`

### What does this do?
Updates all repository methods to:
- Accept `userID` parameter
- Filter all queries by `user_id`
- Verify task ownership before operations
- Ensure users can only access their own tasks

### Why do we need this?
- **Security**: Prevents users from accessing other users' tasks
- **Data isolation**: Each user only sees their own data
- **Database level**: Queries filter at SQL level (efficient)
- **Defense in depth**: Even if service layer has bug, database enforces security

### Where is this used?
- **Called by**: `internal/services/task_service.go` → All task service methods
- **Flow**:
  1. Service receives `userID` from handler (via JWT)
  2. Service calls repo method → Passes `userID`
  3. Repo executes SQL → Includes `WHERE user_id = $1` or `AND user_id = $2`
  4. Database returns → Only tasks belonging to that user

### What to change:

1. **Create method:**
   - Add `user_id` to INSERT
   - Include `task.UserID` in VALUES

2. **GetByID method:**
   - Add `user_id` to SELECT
   - Scan `user_id` into struct

3. **GetAll method:**
   - Change signature: `GetAll(userID int, showCompleted bool)`
   - Add `WHERE user_id = $1` to query
   - Add `userID` parameter to `db.Query()`
   - Scan `user_id` in loop

4. **Update method:**
   - Change signature: `Update(id int, userID int, task *models.Task)`
   - Verify task belongs to user (check `existingTask.UserID == userID`)
   - Add `AND user_id = $5` to UPDATE query
   - Add `userID` parameter

5. **Delete method:**
   - Change signature: `Delete(id int, userID int)`
   - Add `AND user_id = $2` to DELETE query
   - Add `userID` parameter

6. **Complete method:**
   - Change signature: `Complete(id int, userID int)`
   - Verify task belongs to user
   - Add `AND user_id = $2` to UPDATE query
   - Add `userID` parameter

### ✅ Check:
- All methods include `user_id`
- All methods verify ownership
- Queries filter by `user_id`

---

## 🎨 Step 11: Update Task Service

### File: `internal/services/task_service.go`

### What does this do?
Updates all service methods to:
- Accept `userID` parameter
- Pass `userID` to repository methods
- Verify ownership before operations (extra security check)
- Set `userID` when creating tasks

### Why do we need this?
- **Bridge**: Connects handlers (HTTP layer) to repository (database layer)
- **Business logic**: Can add validation, caching, etc.
- **Security**: Additional ownership verification
- **Consistency**: All task operations require user context

### Where is this used?
- **Called by**: `internal/handlers/task_handler.go` → All task handlers
- **Flow**:
  1. Handler gets `userID` from JWT context (`c.Locals("userID")`)
  2. Handler calls service method → Passes `userID` and task data
  3. Service validates → Sets `task.UserID = userID` (for create)
  4. Service calls repo → Passes `userID` to repository
  5. Repository executes → Database filters by `user_id`

### What to change:

1. **GetAllTasks:**
   - Change signature: `GetAllTasks(userID int, showCompleted bool)`
   - Pass `userID` to `repo.GetAll()`

2. **CreateTask:**
   - Change signature: `CreateTask(userID int, task *models.Task)`
   - Set `task.UserID = userID`
   - Pass to repository

3. **GetTaskByID:**
   - Change signature: `GetTaskByID(id int, userID int)`
   - Verify `task.UserID == userID` after fetching
   - Return error if not owner

4. **UpdateTask:**
   - Change signature: `UpdateTask(id int, userID int, task *models.Task)`
   - Pass `userID` to `repo.Update()`

5. **DeleteTask:**
   - Change signature: `DeleteTask(id int, userID int)`
   - Pass `userID` to `repo.Delete()`

6. **CompleteTask:**
   - Change signature: `CompleteTask(id int, userID int)`
   - Pass `userID` to `repo.Complete()`

### ✅ Check:
- All methods accept `userID`
- Ownership verified before operations

---

## 🌐 Step 12: Create Auth Handler

### File: `internal/handlers/auth_handler.go` (NEW FILE)

### What does this do?
Creates HTTP handlers that:
- **Login**: Receives username/password, validates, returns JWT token
- **Register**: Receives user data, creates account, returns JWT token
- **GetProfile**: Returns current user's profile (protected route)

### Why do we need this?
- **HTTP layer**: Handles incoming requests, parses JSON, returns responses
- **Token generation**: After successful login/register, generates JWT token
- **Error handling**: Converts service errors to HTTP status codes
- **Public endpoints**: Login/Register don't require authentication

### Where is this used?
- **Registered in**: `cmd/api/main.go` → Route registration
- **Endpoints**:
  - `POST /auth/login` → Login handler
  - `POST /auth/register` → Register handler
  - `GET /auth/me` → GetProfile handler (protected, needs JWT)
- **Flow**:
  1. Client sends POST request → Handler receives it
  2. Handler parses JSON body → Converts to LoginRequest/RegisterRequest
  3. Handler calls authService → Validates credentials or creates user
  4. Handler calls jwtService → Generates token
  5. Handler returns JSON → Client receives token and user info

### What to write:

1. **AuthHandler struct:**
   ```go
   type AuthHandler struct {
       jwtService  *auth.JWTService
       authService *services.AuthService
   }
   ```

2. **NewAuthHandler function:**
   - Return pointer to AuthHandler

3. **Login handler:**
   - Parse `LoginRequest` from body
   - Validate input
   - Call `authService.Login()`
   - Generate JWT token with `jwtService.GenerateToken()`
   - Return `AuthResponse` with token and user

4. **Register handler:**
   - Parse `RegisterRequest` from body
   - Validate input
   - Call `authService.Register()`
   - Generate JWT token
   - Return `AuthResponse` with token and user

5. **GetProfile handler:**
   - Get `userID` from context (`c.Locals("userID")`)
   - Call `authService.GetUserByID()`
   - Return user JSON

### ✅ Check:
- Login works
- Register works
- GetProfile works

---

## 🔒 Step 13: Update Task Handlers

### File: `internal/handlers/task_handler.go`

### What does this do?
Updates all existing task handlers to:
- Extract `userID` from JWT context (set by middleware)
- Pass `userID` to service methods
- Ensures all task operations are user-scoped

### Why do we need this?
- **User context**: Handlers need to know which user is making the request
- **Security**: Prevents users from accessing other users' tasks
- **Data isolation**: Each user only sees/modifies their own tasks
- **JWT integration**: Connects JWT authentication to task operations

### Where is this used?
- **In every handler**: GetAllTasks, CreateTask, GetTaskByID, UpdateTask, DeleteTask, CompleteTask
- **Flow**:
  1. Request comes in → JWT middleware runs first
  2. Middleware validates token → Sets `c.Locals("userID")` and `c.Locals("username")`
  3. Handler runs → Gets `userID := c.Locals("userID").(int)`
  4. Handler calls service → Passes `userID` along with task data
  5. Service uses `userID` → Filters/verifies ownership

### What to change:

**In EVERY handler method, add at the top:**
```go
// Get user ID from JWT context (set by middleware)
userID := c.Locals("userID").(int)
```

**Then update service calls:**
- `GetAllTasks`: Pass `userID` to service
- `CreateTask`: Pass `userID` to service
- `GetTaskByID`: Pass `userID` to service
- `UpdateTask`: Pass `userID` to service
- `DeleteTask`: Pass `userID` to service
- `CompleteTask`: Pass `userID` to service

### ✅ Check:
- All handlers get `userID` from context
- All handlers pass `userID` to service

---

## 🚀 Step 14: Update Main.go

### File: `cmd/api/main.go`

### What does this do?
Wires all the pieces together:
1. Initializes JWT service (loads keys)
2. Creates user repository and auth service
3. Creates auth handler
4. Groups routes into public (no auth) and protected (requires JWT)
5. Applies middleware to protected routes

### Why do we need this?
- **Initialization**: Sets up all services and handlers
- **Dependency injection**: Passes dependencies to constructors
- **Route organization**: Separates public vs protected routes
- **Middleware application**: Protects routes automatically

### Where is this used?
- **Entry point**: `main()` function runs when server starts
- **Flow**:
  1. Server starts → `main()` runs
  2. Load config → Get database URL, Redis URL, JWT key paths
  3. Connect to DB/Redis → Initialize connections
  4. Create services → JWT service, auth service, task service
  5. Create handlers → Auth handler, task handler
  6. Register routes → Public routes (login/register) and protected routes (tasks)
  7. Apply middleware → JWT middleware protects task routes
  8. Start server → Listen for requests

### What to change:

1. **Add import:**
   ```go
   "booding_ticket/internal/auth"
   ```

2. **Validate JWT config:**
   ```go
   if cfg.JWTPrivateKeyPath == "" || cfg.JWTPublicKeyPath == "" {
       log.Fatal("JWT keys not configured")
   }
   ```

3. **Initialize JWT service:**
   ```go
   jwtService, err := auth.NewJWTService(cfg.JWTPrivateKeyPath, cfg.JWTPublicKeyPath)
   if err != nil {
       log.Fatal("Failed to initialize JWT service:", err)
   }
   ```

4. **Initialize user repository and auth service:**
   ```go
   userRepo := repo.NewUserRepository(db)
   authService := services.NewAuthService(userRepo)
   ```

5. **Initialize auth handler:**
   ```go
   authHandler := handlers.NewAuthHandler(jwtService, authService)
   ```

6. **Create route groups:**
   ```go
   // Public routes (no authentication)
   app.Post("/auth/login", authHandler.Login)
   app.Post("/auth/register", authHandler.Register)

   // Protected routes (require JWT)
   protected := app.Group("/", auth.JWTMiddleware(jwtService))
   protected.Get("/tasks", taskHandler.GetAllTasks)
   protected.Post("/tasks", taskHandler.CreateTask)
   // ... all other task routes
   protected.Get("/auth/me", authHandler.GetProfile)
   ```

### ✅ Check:
- JWT service initialized
- Auth service initialized
- Routes grouped correctly
- Middleware applied to protected routes

---

## 🐳 Step 15: Update Docker Configuration

### File: `docker-compose.yml`

### What does this do?
Configures Docker to:
1. Mount `keys/` directory into container (so container can read key files)
2. Set environment variables for key paths (so app knows where to find keys)
3. Update Dockerfile WORKDIR to `/app` (matches mount path)

### Why do we need this?
- **Key access**: Container needs to read key files to initialize JWT service
- **Volume mount**: Shares local `keys/` directory with container
- **Read-only**: `:ro` flag prevents container from modifying keys
- **Environment variables**: Allows overriding paths if needed

### Where is this used?
- **When**: Docker container starts up
- **Flow**:
  1. Docker starts → Mounts `./keys` to `/app/keys` in container
  2. Container runs → App reads `JWT_PRIVATE_KEY_PATH` env var
  3. App loads keys → Reads from `/app/keys/private.pem` and `/app/keys/public.pem`
  4. JWT service initialized → Can sign/verify tokens

### What to change:

**In `api` service, add:**
```yaml
volumes:
  - ./keys:/app/keys:ro  # Mount keys directory (read-only)

environment:
  # ... existing env vars ...
  - JWT_PRIVATE_KEY_PATH=/app/keys/private.pem
  - JWT_PUBLIC_KEY_PATH=/app/keys/public.pem
```

**Also update Dockerfile:**
```dockerfile
WORKDIR /app  # Change from /root/
```

### ✅ Check:
- Keys directory mounted
- Environment variables set
- Dockerfile WORKDIR is `/app`

---

## 📦 Step 16: Install Dependencies

### What does this do?
Downloads and adds two Go packages to your project:
- **`github.com/golang-jwt/jwt/v5`**: JWT library for creating and validating tokens
- **`golang.org/x/crypto/bcrypt`**: Password hashing library (secure way to store passwords)

### Why do we need these?
- **JWT library**: Provides functions to sign tokens (RS256) and verify them
- **bcrypt**: Hashes passwords so they're never stored in plain text (security requirement)

### Where are they used?
- **JWT library**: Used in `internal/auth/jwt.go` → `GenerateToken()` and `ValidateToken()`
- **bcrypt**: Used in `internal/services/auth_service.go` → `Register()` (hash) and `Login()` (compare)

### Run:
```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
```

### ✅ Check:
- Dependencies added to `go.mod`
- Can import packages in your code

---

## ✅ Final Checklist

- [ ] RSA keys generated (`keys/private.pem`, `keys/public.pem`)
- [ ] Config updated with JWT paths
- [ ] JWT service created (`internal/auth/jwt.go`)
- [ ] JWT middleware created (`internal/auth/middleware.go`)
- [ ] User model created (`internal/models/user.go`)
- [ ] Database schema updated (users table, user_id in tasks)
- [ ] User repository created (`internal/repo/user_repository.go`)
- [ ] Auth service created (`internal/services/auth_service.go`)
- [ ] Task model updated (UserID field)
- [ ] Task repository updated (all methods use user_id)
- [ ] Task service updated (all methods accept userID)
- [ ] Auth handler created (`internal/handlers/auth_handler.go`)
- [ ] Task handlers updated (get userID from context)
- [ ] Main.go updated (wire everything together)
- [ ] Docker config updated (mount keys, env vars)
- [ ] Dependencies installed

---

## 🧪 Testing

1. **Start server:**
   ```bash
   go run cmd/api/main.go
   ```

2. **Register user:**
   ```bash
   curl -X POST http://localhost:3000/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
   ```

3. **Login:**
   ```bash
   curl -X POST http://localhost:3000/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","password":"password123"}'
   ```
   Copy the `token` from response.

4. **Create task (with token):**
   ```bash
   curl -X POST http://localhost:3000/tasks \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_TOKEN_HERE" \
     -d '{"title":"My Task","description":"Test","priority":"high"}'
   ```

5. **Get tasks (with token):**
   ```bash
   curl http://localhost:3000/tasks \
     -H "Authorization: Bearer YOUR_TOKEN_HERE"
   ```

---

## 🎓 Key Concepts Explained

### **Where are keys used?**

**Private Key (`keys/private.pem`):**
- Used in `jwt.go` → `GenerateToken()` → Signs tokens
- Only server has this (keep secret!)

**Public Key (`keys/public.pem`):**
- Used in `jwt.go` → `ValidateToken()` → Verifies tokens
- Can be shared (for microservices)

**In Docker:**
- Keys mounted as volume: `./keys:/app/keys:ro`
- Paths set via environment variables
- Container reads keys from mounted directory

### **Why Auth Service?**

- **Separation of concerns**: Auth logic separate from handlers
- **Password hashing**: Handles bcrypt hashing/verification
- **Reusable**: Can be used by multiple handlers
- **Testable**: Easy to test business logic

### **Why user_id in tasks?**

- **Security**: Users can only see/modify their own tasks
- **Data isolation**: Each user has their own task list
- **Multi-tenancy**: Supports multiple users
- **Database integrity**: Foreign key ensures valid users

---

## 🐛 Common Issues

1. **"keys not found"**
   - Generate keys first (Step 1)
   - Check file paths in config

2. **"invalid token"**
   - Check Authorization header format: `Bearer <token>`
   - Token might be expired (24 hours)

3. **"task not found"**
   - User might not own the task
   - Check user_id matches

4. **"password hash error"**
   - Make sure bcrypt is imported
   - Check password length (bcrypt has limits)

---

Good luck! Take it step by step, and test after each major step! 🚀

