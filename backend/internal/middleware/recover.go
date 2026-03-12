package middleware

import (
	"backend/internal/domain"
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v\n%s", r, debug.Stack())
				c.Status(fiber.StatusInternalServerError).JSON(
					domain.ErrorResponse("Internal Server Error", "An unexpected error occurred"),
				)
			}
		}()
		return c.Next()
	}
}
