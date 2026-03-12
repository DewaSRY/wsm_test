package handler

import (
	"backend/internal/domain"
	"backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Invalid request body"),
		)
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Email, password, and name are required"),
		)
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Password must be at least 6 characters"),
		)
	}

	user, err := h.authService.Register(c.Context(), &req)
	if err != nil {
		if err == service.ErrUserExists {
			return c.Status(fiber.StatusConflict).JSON(
				domain.ErrorResponse("Conflict", "User with this email already exists"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", err.Error()),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(
		domain.SuccessResponse("User registered successfully", user),
	)
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Invalid request body"),
		)
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Email and password are required"),
		)
	}

	response, err := h.authService.Login(c.Context(), &req)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "Invalid email or password"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		domain.SuccessResponse("Login successful", response),
	)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req domain.RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Invalid request body"),
		)
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			domain.ErrorResponse("Bad Request", "Refresh token is required"),
		)
	}

	response, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrInvalidToken {
			return c.Status(fiber.StatusUnauthorized).JSON(
				domain.ErrorResponse("Unauthorized", "Invalid or expired refresh token"),
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			domain.ErrorResponse("Internal Server Error", err.Error()),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		domain.SuccessResponse("Token refreshed successfully", response),
	)
}

// GetMe returns the current authenticated user
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(*domain.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(
			domain.ErrorResponse("Unauthorized", "User not found in context"),
		)
	}

	return c.Status(fiber.StatusOK).JSON(
		domain.SuccessResponse("User retrieved successfully", user),
	)
}
