package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/yourname/ticketing-system/internal/adapter/handler"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
	"github.com/yourname/ticketing-system/pkg/config"
	redis_client "github.com/yourname/ticketing-system/pkg/redis"
)

// @title Ticketing System API
// @version 1.0
// @description API documentation for Flash Sale System
// @host localhost:8080
// @BasePath /api/v1
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
		db, err = gorm.Open(postgres.Open(dbConnStr), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Info),
		})
		if err == nil {
			break
		}
		log.Printf("Đang đợi DB... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Không thể kết nối Database: %v", err)
	}
	log.Println("✅ Connected to Database successfully!")

	// Tự động Migrate cấu trúc Database
	if err := db.AutoMigrate(&entity.User{}, &entity.Event{}, &entity.TicketType{}, &entity.Order{}, &entity.OrderItem{}); err != nil {
		log.Printf("⚠️ Lỗi Migrate Database: %v", err)
	} else {
		log.Println("✅ AutoMigrate Database successfully!")
	}

	// Tải cấu hình
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Không thể tải cấu hình: %v", err)
	}

	// 3. Khởi tạo các lớp (Dependency Injection)
	// Thứ tự: DB -> Repository -> Service -> Handler

	// Khởi tạo redisClient sớm để dùng cho TicketCacheRepo
	redisClient := redis_client.NewRedisClient(cfg)
	ticketCacheRepo := repository.NewRedisTicketRepository(redisClient)

	// User module
	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	// Event module
	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo, ticketCacheRepo)
	eventHandler := handler.NewEventHandler(eventService)

	// Statistics module
	statisticsRepo := repository.NewStatisticRepository(db)
	statisticsService := service.NewStatisticsService(statisticsRepo)
	statisticsHandler := handler.NewStatisticsHandler(statisticsService)

	// Order module
	orderRepo := repository.NewOrderRepository(db)

	// Khởi tạo Worker Pool parameters
	queueSize := 10000
	numWorkers := 50
	orderService := service.NewOrderService(orderRepo, eventRepo, ticketCacheRepo, queueSize, numWorkers)
	orderHandler := handler.NewOrderHandler(orderService)

	// 4. Khởi tạo Fiber
	app := fiber.New(fiber.Config{
		AppName: "Ticketing System v1",
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Middleware ghi log để bạn theo dõi trên Terminal khi Postman gọi tới
	app.Use(logger.New())

	// 5. GỌI ROUTER Ở ĐÂY
	handler.SetupRoutes(app, userRepo, authHandler, eventHandler, orderHandler, statisticsHandler, jwtSecret)

	// 6. Chạy Server và cấu hình Graceful Shutdown
	port := getEnv("SERVER_PORT", "8080")

	// Khởi chạy server trong một goroutine
	go func() {
		log.Printf("Starting server on port %s", port)
		if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi server: %v", err)
		}
	}()

	// Chờ tín hiệu từ OS (SIGTERM, SIGINT) để thực hiện Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("\nNhận tín hiệu dừng server, rục rịch tắt hệ thống...")

	// Dừng Fiber từ chối request mới
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Fiber Shutdown bị lỗi: %v", err)
	}

	// Chờ Worker Pool xử lý nốt đơn hàng đang tồn đọng
	if err := orderService.Shutdown(); err != nil {
		log.Fatalf("OrderService Shutdown bị lỗi: %v", err)
	}

	log.Println("Hệ thống đã tắt an toàn!")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
