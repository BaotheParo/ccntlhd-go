package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/yourname/ticketing-system/internal/adapter/handler"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func main() {
	// 1. Cấu hình (Lấy từ Environment hoặc mặc định)
	jwtSecret := getEnv("JWT_SECRET", "my-super-secret-key-2026")
	dbConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "postgres"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "user"),
		getEnv("DB_PASS", "password"),
		getEnv("DB_NAME", "ticket_db"),
	)

	// 2. Kết nối Database với GORM
	var db *gorm.DB
	var err error

	// Chờ DB sẵn sàng (Retry logic)
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dbConnStr), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Đang đợi DB... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Không thể kết nối Database: %v", err)
	}

	// 3. Khởi tạo các lớp (Dependency Injection)
	// Thứ tự: DB -> Repository -> Service -> Handler

	// User module
	sqlDB, _ := db.DB()
	userRepo := repository.NewUserRepository(sqlDB)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// Event module
	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventService)

	// 4. Khởi tạo Fiber
	app := fiber.New(fiber.Config{
		AppName: "Ticketing System v1",
	})

	// Middleware ghi log để bạn theo dõi trên Terminal khi Postman gọi tới
	app.Use(logger.New())

	// 5. GỌI ROUTER CỦA BẠN Ở ĐÂY
	handler.SetupRoutes(app, authHandler, eventHandler, jwtSecret)

	// 6. Chạy Server
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("Starting server on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
