// Middleware is a function that runs **before** your handlers. It:
// 1. Intercepts every HTTP request
// 2. Extracts JWT token from `Authorization` header
// 3. Validates the token using JWT service
// 4. If valid: Stores user info in context and allows request to continue
// 5. If invalid: Returns 401 Unauthorized and stops the request

// - **Flow**:
//   1. Request comes in → Middleware runs first
//   2. Middleware checks token → If valid, sets `c.Locals("userID")` and `c.Locals("username")`
//   3. Handler runs → Can access user info from context
//   4. If token invalid → Handler never runs, returns 401 error

package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zahrashariati/task-manager/internal/handlers"
)

func JWTMiddleware(jwtService handlers.JWTServiceInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString, err := jwtService.GetTokenFromHeader(c)

		// Missing header
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header",
			})
		}

		// Invalid format
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header",
			})
		}

		claimsInterface, err := jwtService.ValidateToken(tokenString)

		// Invalid token
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Type assert to ClaimsInterface
		claims, ok := claimsInterface.(handlers.ClaimsInterface)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Valid token
		c.Locals("userID", claims.GetUserID())
		c.Locals("username", claims.GetUsername())
		return c.Next()
	}
}


