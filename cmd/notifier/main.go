package main

import (
	"context"
	"log"
	"github.com/zahrashariati/task-manager/internal/consumer"
	"github.com/zahrashariati/task-manager/internal/notifier"
	"github.com/zahrashariati/task-manager/internal/scheduler"
	"github.com/zahrashariati/task-manager/internal/storage"
	"github.com/zahrashariati/task-manager/internal/config"
	"github.com/zahrashariati/task-manager/internal/repo"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting notifier service with scheduler")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
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
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()//If error happens here, db.Close() still runs!
		
		notificationRepo := repo.NewScheduledNotificationRepository(db)
		taskStorage = storage.NewDBStorage(notificationRepo)
		log.Println("Using database storage (persistent)")
	} else {
		// In-memory storage (simple, for testing)
		taskStorage = storage.NewStorage()
		log.Println("Using in-memory storage (data lost on restart)")
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	// Start consumer in goroutine (reads messages and stores them)
	go func() {
		if err := kafkaConsumer.Start(ctx, taskStorage); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Start scheduler in goroutine (checks stored tasks every hour)
	go func() {
		scheduler.Start(ctx)
	}()

	log.Println("Notifier service started")
	log.Println("- Consumer: Reading messages from Kafka and storing them")
	log.Println("- Scheduler: Checking stored tasks every 15 minutes")
	log.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	<-signalChan
	log.Println("Shutting down notifier service...")
	
	// Cancel context to stop consumer and scheduler gracefully
	cancel()
	
	// Give services time to finish
	time.Sleep(2 * time.Second)
}
