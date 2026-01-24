package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq" // Driver cho Postgres

	"github.com/yourname/ticketing-system/internal/adapter/handler"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func main() {
	// 1. Cấu hình các thông số (Ưu tiên lấy từ biến môi trường)
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
		log.Fatalf("Không thể kết nối Database: %v", err)
	}
	defer db.Close() // Đóng kết nối khi tắt server

	// Kiểm tra kết nối thực tế
	if err := db.Ping(); err != nil {
		log.Fatalf("Database không phản hồi: %v", err)
	}

	// 3. Khởi tạo các lớp (Dependency Injection)
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	// 4. Khởi tạo Fiber App
	app := fiber.New()

	// Thêm logger để theo dõi các request trên terminal
	app.Use(logger.New())

	// 5. Định nghĩa Routes
	api := app.Group("/api/v1")

	// --- Routes Công Khai ---
	authRoutes := api.Group("/auth")
	authRoutes.Post("/register", authHandler.Register)
	authRoutes.Post("/login", authHandler.Login)

	// --- Routes Bảo Mật (Cần Token) ---
	userRoutes := api.Group("/user", handler.AuthMiddleware(jwtSecret))

	userRoutes.Get("/me", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")

		return c.JSON(fiber.Map{
			"status": "success",
			"data": fiber.Map{
				"user_id": userID,
				"role":    role,
				"message": "Đây là khu vực riêng tư!",
			},
		})
	})

	port := getEnv("SERVER_PORT", "8080")
	fmt.Printf(port)
	log.Fatal(app.Listen(":" + port))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
