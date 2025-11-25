package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/zahrashariati/task-manager/internal/config"
	"github.com/zahrashariati/task-manager/internal/consumer"
	"github.com/zahrashariati/task-manager/internal/notifier"
	"github.com/zahrashariati/task-manager/internal/repo"
	"github.com/zahrashariati/task-manager/internal/scheduler"
	"github.com/zahrashariati/task-manager/internal/storage"
)

func main() {
	log.Println("starting notifier service with scheduler")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	brokerURL := os.Getenv("KAFKA_BROKER_URL")
	if brokerURL == "" {
		brokerURL = "localhost:9092"
	}
//define type vars
	// Choose storage type: "memory" or "db"
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "memory" // Default to in-memory for simplicity
	}

	var taskStorage scheduler.StorageInterface
	if storageType == "db" {
		// Database storage (persistent, production-ready)
		cfg := config.Load()
		if cfg.DatabaseURL == "" {
			log.Fatal("DATABASE_URL not set (required for DB storage)")
		}
		
		db, err := repo.InitDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer db.Close()//If error happens here, db.Close() still runs!
		
		notificationRepo := repo.NewScheduledNotificationRepository(db)
		taskStorage = storage.NewDBStorage(notificationRepo)
		log.Println("using database storage (persistent)")
	} else {
		// In-memory storage (simple, for testing)
		taskStorage = storage.NewStorage()
		log.Println("using in-memory storage (data lost on restart)")
	}

	notifierService := notifier.NewNotifier()
	kafkaConsumer := consumer.NewConsumer(brokerURL)
	defer kafkaConsumer.Close()

	// Create callback function that processes due tasks
	callback := notifierService.Callback

	// Create scheduler (checks every 15 minutes)
	scheduler := scheduler.NewScheduler(
		taskStorage,
		callback,
		15*time.Minute, // Check every 15 minutes
	)

	// Create context that cancels on interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start consumer in goroutine (reads messages and stores them)
	go func() {
		if err := kafkaConsumer.Start(ctx, taskStorage); err != nil {
			log.Printf("consumer error: %v", err)
		}
	}()

	log.Println("notifier service started")
	log.Println("- consumer: reading messages from Kafka and storing them")
	log.Println("- scheduler: checking stored tasks every 15 minutes")

	// Start scheduler in main goroutine (blocks until ctx is cancelled)
	scheduler.Start(ctx)
	log.Println("shutting down notifier service...")
}
