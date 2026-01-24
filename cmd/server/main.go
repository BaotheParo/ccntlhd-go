package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq"

	"github.com/yourname/ticketing-system/internal/adapter/handler"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func main() {
	// 1. Lấy biến môi trường
	dbUser := getEnv("DATABASE_USER", "user")
	dbPass := getEnv("DATABASE_PASSWORD", "password")
	dbName := getEnv("DATABASE_DBNAME", "ticket_db")
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	jwtSecret := getEnv("JWT_SECRET", "my-super-secret-key-123")

	// 2. Kết nối Database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Lỗi cấu hình Database: %v", err)
	}
	defer db.Close()

	// Cơ chế Retry: Đợi Database sẵn sàng (Hữu ích khi chạy Docker)
	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Đang đợi Database... (Thử lại %d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Database không phản hồi sau nhiều lần thử: %v", err)
	}
	log.Println("Kết nối Database thành công!")

	// 3. Khởi tạo Dependency Injection
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// 4. Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Ticketing System v1.0",
	})

	app.Use(logger.New())

	// 5. Routes
	api := app.Group("/api/v1")

	// Public
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Private
	userRoutes := api.Group("/user", handler.AuthMiddleware(jwtSecret))
	userRoutes.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"user_id": c.Locals("user_id"),
				"role":    c.Locals("role"),
			},
		})
	})

	// 6. Chạy Server
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Server đang chạy tại cổng %s", port)
	log.Fatal(app.Listen(":" + port))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
