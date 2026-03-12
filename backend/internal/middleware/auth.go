package middleware

import (
	"strings"

	"backend/internal/domain"
	"backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) JWTAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "Authorization header is required"),
			)
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "Invalid authorization header format"),
			)
		}

		tokenString := parts[1]

		claims, err := m.authService.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "Invalid or expired token"),
			)
		}

		// Get user from database
		user, err := m.authService.GetUserByID(c.Context(), claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "User not found"),
			)
		}

		// Set user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("user", user)

		return c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(
				domain.ErrorResponse("Forbidden", "Access denied"),
			)
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(
			domain.ErrorResponse("Forbidden", "Insufficient permissions"),
		)
	}
}

func GetUserID(c *fiber.Ctx) int64 {
	if id, ok := c.Locals("user_id").(int64); ok {
		return id
	}
	return 0
}

func GetEmail(c *fiber.Ctx) string {
	if email, ok := c.Locals("email").(string); ok {
		return email
	}
	return ""
}

func GetRole(c *fiber.Ctx) string {
	if role, ok := c.Locals("role").(string); ok {
		return role
	}
	return ""
}

func GetUser(c *fiber.Ctx) *domain.User {
	if user, ok := c.Locals("user").(*domain.User); ok {
		return user
	}
	return nil
}
