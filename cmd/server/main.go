package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Khong the tai cau hinh: %v", err)
	}

	jwtSecret := getEnv("JWT_SECRET", "my-super-secret-key-2026")

	dbHost := getEnv("DB_HOST", cfg.Database.Host)
	dbPort := getEnv("DB_PORT", fmt.Sprintf("%d", cfg.Database.Port))
	dbUser := getEnv("DB_USER", cfg.Database.User)
	dbPass := getEnv("DB_PASS", cfg.Database.Password)
	dbName := getEnv("DB_NAME", cfg.Database.DBName)

	dbTargets := buildDBTargets(dbHost, dbPort, fmt.Sprintf("%d", cfg.Database.Port))
	db, err := connectDatabase(dbTargets, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("Khong the ket noi Database: %v", err)
	}
	log.Println("Connected to Database successfully")

	if err := db.AutoMigrate(&entity.User{}, &entity.Event{}, &entity.TicketType{}, &entity.Order{}, &entity.OrderItem{}); err != nil {
		log.Printf("Loi migrate Database: %v", err)
	} else {
		log.Println("AutoMigrate Database successfully")
	}

	redisClient := redis_client.NewRedisClient(cfg)
	ticketCacheRepo := repository.NewRedisTicketRepository(redisClient)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	eventRepo := repository.NewEventRepository(db)
	eventService := service.NewEventService(eventRepo, ticketCacheRepo)
	eventHandler := handler.NewEventHandler(eventService)

	statisticsRepo := repository.NewStatisticRepository(db)
	statisticsService := service.NewStatisticsService(statisticsRepo)
	statisticsHandler := handler.NewStatisticsHandler(statisticsService)

	orderRepo := repository.NewOrderRepository(db)
	queueSize := 10000
	numWorkers := 50
	orderService := service.NewOrderService(orderRepo, eventRepo, ticketCacheRepo, queueSize, numWorkers)
	orderHandler := handler.NewOrderHandler(orderService)

	app := fiber.New(fiber.Config{
		AppName: "Ticketing System v1",
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(logger.New())

	handler.SetupRoutes(app, userRepo, authHandler, eventHandler, orderHandler, statisticsHandler, jwtSecret)

	port := cfg.Server.Port
	if port == "" {
		port = ":8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	go func() {
		log.Printf("Starting server on port %s", port)
		if err := app.Listen(port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Loi server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Nhan tin hieu dung server, dang tat he thong...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Fiber Shutdown bi loi: %v", err)
	}

	if err := orderService.Shutdown(); err != nil {
		log.Fatalf("OrderService Shutdown bi loi: %v", err)
	}

	log.Println("He thong da tat an toan")
}

type dbTarget struct {
	Host string
	Port string
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func buildDBTargets(primaryHost string, primaryPort string, configPort string) []dbTarget {
	targets := []dbTarget{
		{Host: primaryHost, Port: primaryPort},
	}

	if primaryHost == "postgres" {
		targets = appendUniqueDBTarget(targets, dbTarget{Host: "localhost", Port: configPort})
		targets = appendUniqueDBTarget(targets, dbTarget{Host: "127.0.0.1", Port: configPort})
	}

	return targets
}

func appendUniqueDBTarget(targets []dbTarget, candidate dbTarget) []dbTarget {
	for _, target := range targets {
		if target.Host == candidate.Host && target.Port == candidate.Port {
			return targets
		}
	}

	return append(targets, candidate)
}

func connectDatabase(targets []dbTarget, user string, password string, dbName string) (*gorm.DB, error) {
	var lastErr error

	for i, target := range targets {
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			target.Host,
			target.Port,
			user,
			password,
			dbName,
		)
		log.Printf("Trying DB connection via %s:%s (%d/%d)", target.Host, target.Port, i+1, len(targets))

		db, err := openDBWithRetry(connStr, 5)
		if err == nil {
			return db, nil
		}

		lastErr = err
		log.Printf("DB connection via %s:%s failed: %v", target.Host, target.Port, err)
	}

	return nil, lastErr
}

func openDBWithRetry(connStr string, attempts int) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	for i := 0; i < attempts; i++ {
		db, err = gorm.Open(postgres.Open(connStr), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Info),
		})
		if err == nil {
			return db, nil
		}

		log.Printf("Dang doi DB... (%d/%d) - Connect String: %s", i+1, attempts, connStr)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}
