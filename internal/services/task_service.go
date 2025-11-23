//orchestrates between repository and cache, handles business logic
package services

import (
	"errors" //standard library errors (for errors.Is)
	"github.com/zahrashariati/task-manager/internal/models" //task structs
	"github.com/zahrashariati/task-manager/internal/producer" //Kafka producer interface
	"log" //for logging
)

type TaskService struct { //share same instances of repo and cache across different methods
	repo     TaskRepositoryInterface //holds reference to TaskRepository interface
	cache    CacheInterface         //holds pointer to Cache interface
	producer producer.ProducerInterface   //holds reference to Kafka producer (optional)
}

//Constructor: returns pointer to TaskService struct
//DI: Dependency Injection - allows flexibility in how dependencies are provided
//creates service with those dependencies
func NewTaskService(repo TaskRepositoryInterface, cache CacheInterface, producer producer.ProducerInterface) *TaskService {
	return &TaskService{
		repo:     repo,
		cache:    cache,
		producer: producer,
	}
}

func (s *TaskService) GetAllTasks(userID int, showCompleted bool) ([]*models.Task, error) {
	//try cache first
	if tasks, err := s.cache.GetTasks(showCompleted); err == nil {
		log.Printf("Cache hit %d tasks for GetAllTasks\n", len(tasks))
		return tasks, nil //cache hit - return cached data
	}

	//cache miss - get from database
	log.Println("Cache miss for GetAllTasks")
	tasks, err := s.repo.GetAll(userID, showCompleted)
	if err != nil {
		return nil, err //return error if database operation fails
	}
	// Convert []models.Task to []*models.Task
	taskPointers := make([]*models.Task, len(tasks))
	for i := range tasks {
		taskPointers[i] = &tasks[i]
	}
	// Store in cache for next time
	log.Printf("retrieved %d tasks from db, Storing %d tasks in cache for GetAllTasks\n", len(tasks), len(tasks))
	if err := s.cache.SetTasks(showCompleted, taskPointers); err != nil {
		log.Printf("⚠️ WARNING: Failed to store tasks in cache: %v\n", err)
		// Continue anyway - cache is optional
	} else {
		keyName := "all"
		if !showCompleted {
			keyName = "incomplete"
		}
		log.Printf("✅ Successfully stored %d tasks in cache (key: tasks:%s)\n", len(tasks), keyName)
	}
	return taskPointers, nil //return tasks
}

func (s *TaskService) CreateTask(userID int, task *models.Task) error {
	// Local error definitions
	var ErrTitleRequired = errors.New("title is required")

	// Validate title
	if task.Title == "" {
		return ErrTitleRequired
	}
	
	// Set user ID
	task.UserID = userID
	
	// Create in database
	id, err := s.repo.Create(task)
	if err != nil {
		return err
	}
	task.ID = id // Set the ID returned from database
	s.cache.Invalidate() //invalidate cache

	// Publish event to Kafka if task has due_date (fire-and-forget)
	if s.producer != nil && task.DueDate != nil {
		if err := s.producer.PublishTaskScheduledEvent(task); err != nil {
			// Log error but don't fail the request (fire-and-forget)
			log.Printf("⚠️ Failed to publish task scheduled event: %v", err)
		}
	}

	return nil
}

func (s *TaskService) GetTaskByID(id int, userID int) (*models.Task, error) {
	// Local error definitions
	var (
		ErrTaskNotFound  = errors.New("task not found")
		ErrDatabaseError = errors.New("database error")
	)

	// Try cache first
	if task, err := s.cache.GetTask(id); err == nil {
		// Verify cached task belongs to user
		if task.UserID == userID {
			log.Printf("Cache hit for GetTaskByID %d\n", id)
			return task, nil
		}
		// Cache has wrong user's task, fetch from DB
	}
	
	// Cache miss - get from database
	log.Printf("Cache miss for GetTaskByID %d\n", id)
	task, err := s.repo.GetTaskByID(id)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, ErrDatabaseError
	}
	
	// Verify task belongs to user
	if task.UserID != userID {
		return nil, ErrTaskNotFound // Don't reveal task exists
	}
	
	// Store in cache
	s.cache.SetTask(id, task) //store in cache
	return task, nil
}

func (s *TaskService) UpdateTask(id int, userID int, task *models.Task) (*models.Task, error) {
	// Local error definitions
	var (
		ErrTaskNotFound    = errors.New("task not found")
		ErrTaskUpdateFailed = errors.New("failed to update task")
	)

	if err := s.repo.Update(id, userID, task); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, ErrTaskUpdateFailed
	}
	s.cache.Invalidate() //invalidate cache
	
	// Fetch updated task from database to get all fields (ID, created_at, etc.)
	updatedTask, err := s.repo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	// Verify task belongs to user
	if updatedTask.UserID != userID {
		return nil, ErrTaskNotFound
	}

	// Publish event to Kafka if task has due_date (fire-and-forget)
	if s.producer != nil && updatedTask.DueDate != nil {
		if err := s.producer.PublishTaskScheduledEvent(updatedTask); err != nil {
			// Log error but don't fail the request (fire-and-forget)
			log.Printf("⚠️ Failed to publish task scheduled event: %v", err)
		}
	}

	return updatedTask, nil
}

func (s *TaskService) DeleteTask(id int, userID int) error {
	// Local error definitions
	var (
		ErrTaskNotFound      = errors.New("task not found")
		ErrTaskDeletionFailed = errors.New("failed to delete task")
	)

	if err := s.repo.Delete(id, userID); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return ErrTaskNotFound
		}
		return ErrTaskDeletionFailed
	}
	s.cache.Invalidate() //invalidate cache
	return nil
}

func (s *TaskService) CompleteTask(id int, userID int) (*models.Task, error) {
	// Local error definitions
	var (
		ErrTaskNotFound    = errors.New("task not found")
		ErrTaskUpdateFailed = errors.New("failed to update task")
	)

	// Update task in database (sets completed = TRUE)
	if err := s.repo.Complete(id, userID); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, ErrTaskUpdateFailed
	}
	// Invalidate cache (data changed)
	s.cache.Invalidate()
	// Fetch updated task from database to get all fields
	completedTask, err := s.repo.GetTaskByID(id)
	if err != nil {
		return nil, err
	}
	// Verify task belongs to user
	if completedTask.UserID != userID {
		return nil, ErrTaskNotFound
	}
	return completedTask, nil
}
