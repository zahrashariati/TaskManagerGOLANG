package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/zahrashariati/task-manager/internal/models"
)

type AuthHandler struct {
	authService AuthServiceInterface
	jwtService  JWTServiceInterface
}

func NewAuthHandler(authService AuthServiceInterface, jwtService JWTServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtService:  jwtService,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var request models.RegisterRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	user := &models.User{
		Username: request.Username,
		Email:    request.Email,
		Password: request.Password,
	}
	id, err := h.authService.Register(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to register user",
		})
	}
	
	// Set the ID returned from database
	user.ID = id
	
	// Register endpoint does NOT return tokens
	// User must login separately to get access token and refresh token
	// This follows the pattern: Register → redirects to Login page
	return c.Status(fiber.StatusCreated).JSON(models.RegisterResponse{
		Message: "User registered successfully. Please login to get access token.",
		User:     *user,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request models.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	user := &models.User{
		Username: request.Username,
		Password: request.Password,
	}
	user, err := h.authService.Login(user)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid credentials",
		})
	}
	
	// Generate JWT token
	ATKToken, err := h.jwtService.GenerateToken(user.ID, user.Username, time.Now().Add(5*time.Minute))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}

	rtkToken, err := h.authService.CreateRefreshToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create refresh token",
		})
	}
	
	// Ensure refresh token is not empty
	if rtkToken == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "refresh token generation failed",
		})
	}
	
	return c.Status(fiber.StatusOK).JSON(models.AuthResponse{
		Token:        ATKToken,
		RefreshToken: rtkToken,
		User:         *user,
	})
}

func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "user not logged in",
		})
	}
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	return c.JSON(user)
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "user not logged in",
		})
	}
	var request models.User
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	user := &models.User{
		Username: request.Username,
		Email: request.Email,
		Password: request.Password,
	}
	err := h.authService.UpdateUser(userID, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user updated successfully",
	})
}

func (h *AuthHandler) DeleteUser(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "user not logged in",
		})
	}
	err := h.authService.DeleteUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete user",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user deleted successfully",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "user not logged in",
		})
	}

	// Defines an inline struct to parse JSON
	// Expects {"refresh_token": "..."}
	// Parses the request body into request
	// Returns 400 if parsing fails
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	err := h.authService.RevokeRefreshToken(userID, request.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to revoke refresh token",
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "logged out successfully",
	})
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var request struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	
	// Rotate refresh token: validate old, revoke it, create new
	userID, newRefreshToken, err := h.authService.RotateRefreshToken(request.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}

	// Get user details to generate new access token
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get user",
		})
	}
	
	// Generate new access token
	ATKToken, err := h.jwtService.GenerateToken(user.ID, user.Username, time.Now().Add(5*time.Minute))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate token",
		})
	}
	
	// Return both new access token and new refresh token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"access_token":  ATKToken,
		"refresh_token": newRefreshToken,
	})
}
