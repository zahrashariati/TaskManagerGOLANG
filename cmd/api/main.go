//starts the server and wires all layers together
package main

import (
	"context" //for Redis Ping
	"fmt"
	"os"
	"strings" //for string manipulation
	"github.com/joho/godotenv" //load environment variables from .env file
	"log"                        //logging
	"github.com/gofiber/fiber/v2" //Fiber web framework
	"github.com/redis/go-redis/v9" //redis client
	"task_manager/internal/config" //configuration
	"task_manager/internal/repo"   //db operations
	"task_manager/internal/services" //logic layer
	"task_manager/internal/handlers" //handler layer
	"task_manager/internal/cache"    //cache operations
	"task_manager/internal/auth"    //JWT service
)

func main() {
	fmt.Println("Hello, Task Manager API!")
	
	// Load environment variables 
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Validate configuration
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL not set")
	}
	if cfg.RedisURL == "" {
		log.Fatal("REDIS_URL not set")
	}

	// Connect to database
	db, err := repo.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect to Redis
	// Strip redis:// prefix if present (Redis client expects just host:port)
	redisAddr := cfg.RedisURL
	if strings.HasPrefix(redisAddr, "redis://") {
		redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()
	
	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	if cfg.JWTPrivateKeyPath == "" || cfg.JWTPublicKeyPath == "" {
		log.Fatal("JWT keys not configured")
	}
	// Read private key file
	privateKeyBytes, err := os.ReadFile(cfg.JWTPrivateKeyPath)
	if err != nil {
		log.Fatal("Failed to read private key file:", err)
	}
	// Read public key file
	publicKeyBytes, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		log.Fatal("Failed to read public key file:", err)
	}
	jwtService, err := auth.NewJWTService(privateKeyBytes, publicKeyBytes)
	if err != nil {
		log.Fatal("Failed to initialize JWT service:", err)
	}

	// Initialize layers (each layer depends on the previous layer)
	taskRepo := repo.NewTaskRepository(db)              //create db operations
	userRepo := repo.NewUserRepository(db, "")          //create db operations (secretKey not needed for RS256)
	rtkRepo := repo.NewRTKRepository(db)                //create db operations
	cacheService := cache.NewCache(redisClient)        //create cache operations
	taskService := services.NewTaskService(taskRepo, cacheService) //logic layer
	taskHandler := handlers.NewTaskHandler(taskService) //create handler layer
	authService := services.NewAuthService(userRepo, rtkRepo) //logic layer
	authHandler := handlers.NewAuthHandler(authService, jwtService) //handler layer
	// Create Fiber app
	app := fiber.New()

	// // Register routes
	// app.Get("/tasks", taskHandler.GetAllTasks)           // GET /tasks
	// app.Post("/tasks", taskHandler.CreateTask)          // POST /tasks
	// app.Get("/tasks/:id", taskHandler.GetTaskByID)      // GET /tasks/:id
	// app.Put("/tasks/:id", taskHandler.UpdateTask)      // PUT /tasks/:id
	// app.Delete("/tasks/:id", taskHandler.DeleteTask)    // DELETE /tasks/:id
	// app.Patch("/tasks/:id/complete", taskHandler.CompleteTask)    // PATCH /tasks/:id/complete
	app.Post("/auth/login", authHandler.Login)                    // POST /auth/login
	app.Post("/auth/register", authHandler.Register)              // POST /auth/register
	app.Post("/auth/refresh", authHandler.RefreshToken)            // POST /auth/refresh

	// Protected routes (require JWT authentication)
	protected := app.Group("/", auth.JWTMiddleware(jwtService))

	protected.Get("/tasks", taskHandler.GetAllTasks)           // GET /tasks
	protected.Post("/tasks", taskHandler.CreateTask)          // POST /tasks
	protected.Get("/tasks/:id", taskHandler.GetTaskByID)      // GET /tasks/:id
	protected.Put("/tasks/:id", taskHandler.UpdateTask)      // PUT /tasks/:id
	protected.Delete("/tasks/:id", taskHandler.DeleteTask)    // DELETE /tasks/:id
	protected.Patch("/tasks/:id/complete", taskHandler.CompleteTask)    // PATCH /tasks/:id/complete
	protected.Get("/auth/me", authHandler.GetUser)                      // GET /auth/me
	protected.Put("/auth/me", authHandler.UpdateUser)                   // PUT /auth/me
	protected.Delete("/auth/me", authHandler.DeleteUser)                // DELETE /auth/me
	protected.Post("/auth/logout", authHandler.Logout)                    // POST /auth/logout


	// Start server
	addr := ":" + cfg.Port
	//creates address string for server to listen on
	log.Printf("Server starting on %s", addr)
	//starts the web server on that address
	//if server fails to start, log the error and exit
	log.Fatal(app.Listen(addr))
}
