// each method corresponds to one endpoint
// Handler formats response = sends JSON
// Handlers always return error, not (*models.Task, error). The service returns the task; the handler formats the HTTP response.
package handlers

import (
	"errors"                                 //standard library errors (for errors.Is)
	"strconv"                                //for converting strings to integers
	"github.com/zahrashariati/task-manager/internal/models"           //task models
	"github.com/gofiber/fiber/v2" //Fiber web framework
)

type TaskHandler struct {
	service TaskServiceInterface //holds reference to TaskService interface
}

func NewTaskHandler(service TaskServiceInterface) *TaskHandler {
	return &TaskHandler{service: service}
}

// GetAllTasks handles GET /tasks
// Query param: ?showCompleted=true/false
func (h *TaskHandler) GetAllTasks(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Get query parameter
	showCompleted := c.Query("showCompleted") == "true"

	// Call service
	tasks, err := h.service.GetAllTasks(userID, showCompleted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return JSON response
	return c.JSON(tasks)
}

// CreateTask handles POST /tasks
// Body: JSON with title, description, priority
func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Parse JSON body into Task struct
	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	// Local error definitions
	var ErrTitleRequired = errors.New("title is required")

	// Call service
	if err := h.service.CreateTask(userID, &task); err != nil {
		// Check error type
		if errors.Is(err, ErrTitleRequired) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Title is required",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return created task
	return c.Status(fiber.StatusCreated).JSON(task)
}

// GetTaskByID handles GET /tasks/:id
// URL param: id
func (h *TaskHandler) GetTaskByID(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Get ID from URL parameter
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Local error definitions
	var ErrTaskNotFound = errors.New("task not found")

	// Call service
	task, err := h.service.GetTaskByID(id, userID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return task
	return c.JSON(task)
}

// UpdateTask handles PUT /tasks/:id
// URL param: id
// Body: JSON with fields to update
func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Get ID from URL parameter
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Parse JSON body
	var task models.Task
	if err := c.BodyParser(&task); err != nil { //pointer because it writes to the struct
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	// Local error definitions
	var ErrTaskNotFound = errors.New("task not found")

	// Call service (returns updated task with all fields)
	updatedTask, err := h.service.UpdateTask(id, userID, &task)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return updated task (with ID, created_at, etc. from database)
	return c.JSON(updatedTask)
}

// DeleteTask handles DELETE /tasks/:id
// URL param: id
func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Get ID from URL parameter
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Local error definitions
	var ErrTaskNotFound = errors.New("task not found")

	// Call service
	if err := h.service.DeleteTask(id, userID); err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Task deleted successfully",
	})
}

// CompleteTask handles PATCH /tasks/:id/complete
// URL param: id
func (h *TaskHandler) CompleteTask(c *fiber.Ctx) error {
	// Get user ID from JWT context (set by middleware)
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User not logged in",
		})
	}
	// Get ID from URL parameter
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid task ID",
		})
	}

	// Local error definitions
	var ErrTaskNotFound = errors.New("task not found")

	// Call service (service handles the business logic)
	completedTask, err := h.service.CompleteTask(id, userID)
	if err != nil {
		if errors.Is(err, ErrTaskNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return completed task
	return c.JSON(completedTask)
}
