package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func LoggerMiddleware(c *fiber.Ctx) error {
	fmt.Printf("Co mot yeu cau %s gui den duong dan %s\n", c.Method(), c.Path())
	return c.Next()
}
func main() {
	app := fiber.New()
	app.Use(LoggerMiddleware)
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Xin chao, may chu Fiber dang hoat dong!",
		})
	})
	app.Listen(":3000")
}
