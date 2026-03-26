package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/yourname/ticketing-system/docs"
)

// SetupRoutes tập trung tất cả định nghĩa API vào một chỗ
func SetupRoutes(app *fiber.App, authHandler *AuthHandler, eventHandler *EventHandler, orderHandler *OrderHandler, jwtSecret string) {
	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")
	// Swagger UI
	api.Get("/swagger/*", swagger.HandlerDefault)

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

	// Users routes (additional user endpoints)
	users := api.Group("/users", AuthMiddleware(jwtSecret))
	users.Get("/me/orders", orderHandler.GetUserOrders) // Get user's order history

	// Event routes
	events := api.Group("/events")
	events.Post("/", AuthMiddleware(jwtSecret), AdminMiddleware, eventHandler.CreateEvent) // Create event (admin only)
	events.Put("/:id", AuthMiddleware(jwtSecret), AdminMiddleware, eventHandler.UpdateEvent)
	events.Delete("/:id", AuthMiddleware(jwtSecret), AdminMiddleware, eventHandler.DeleteEvent)

	events.Get("/search", eventHandler.ListEventsAdvanced)
	events.Get("/:id", eventHandler.GetEvent)              // Get event by ID
	events.Get("/slug/:slug", eventHandler.GetEventBySlug) // Get event by slug
	events.Get("", eventHandler.ListEvents)                // List all events

	// Order routes
	orders := api.Group("/orders", AuthMiddleware(jwtSecret))
	orders.Post("/", orderHandler.PlaceOrder)            // Create order
	orders.Get("/", orderHandler.GetUserOrders)          // Get user's orders
	orders.Get("/:id", orderHandler.GetOrder)            // Get order by ID
	orders.Post("/:id/cancel", orderHandler.CancelOrder) // Cancel order
}
