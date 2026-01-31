package handler

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes tập trung tất cả định nghĩa API vào một chỗ
func SetupRoutes(app *fiber.App, authHandler *AuthHandler, eventHandler *EventHandler, orderHandler *OrderHandler, jwtSecret string) {
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// User routes
	user := api.Group("/user", AuthMiddleware(jwtSecret))
	user.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

	// Event routes
	events := api.Group("/events")
	events.Post("/", AuthMiddleware(jwtSecret), AdminMiddleware, eventHandler.CreateEvent) // Create event (admin only)
	events.Get("/:id", eventHandler.GetEvent)                                              // Get event by ID
	events.Get("/slug/:slug", eventHandler.GetEventBySlug)                                 // Get event by slug
	events.Get("", eventHandler.ListEvents)                                                // List all events

	// Order routes
	orders := api.Group("/orders", AuthMiddleware(jwtSecret))
	orders.Post("/", orderHandler.PlaceOrder)
}
