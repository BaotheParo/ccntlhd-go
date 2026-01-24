package handler

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes tập trung tất cả định nghĩa API vào một chỗ
func SetupRoutes(app *fiber.App, authHandler *AuthHandler, jwtSecret string) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	user := api.Group("/user", AuthMiddleware(jwtSecret))
	user.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

}
