package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/lib/pq" // Driver kết nối Postgres

	"github.com/yourname/ticketing-system/internal/adapter/handler"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func main() {
	// 1. Cấu hình (Lấy từ Environment hoặc mặc định)
	jwtSecret := getEnv("JWT_SECRET", "my-super-secret-key-2026")
	dbConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "user"),
		getEnv("DB_PASS", "password"),
		getEnv("DB_NAME", "ticket_db"),
	)

	// 2. Kết nối Database
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Lỗi cấu hình DB: %v", err)
	}
	defer db.Close()

	// Chờ DB sẵn sàng (Retry logic)
	for i := 0; i < 5; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("Đang đợi DB... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Không thể kết nối Database!")
	}

	// 3. Khởi tạo các lớp (Dependency Injection)
	// Thứ tự: DB -> Repository -> Service -> Handler
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Khởi tạo Fiber
	app := fiber.New(fiber.Config{
		AppName: "Ticketing System v1",
	})

	// Middleware ghi log để bạn theo dõi trên Terminal khi Postman gọi tới
	app.Use(logger.New())

	// 5. GỌI ROUTER CỦA BẠN Ở ĐÂY
	handler.SetupRoutes(app, authHandler, jwtSecret)

	// 6. Chạy Server
	port := getEnv("SERVER_PORT", "8080")
	log.Printf("🚀 Server đang chạy tại: http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
