package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/yourname/ticketing-system/pkg/config"
	"github.com/yourname/ticketing-system/pkg/logger" // Import our logger pkg
	"go.uber.org/zap"
)

func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		// If config fails, we might not have a logger yet, use panic or fmt
		panic(err)
	}

	// 2. Init Logger
	if err := logger.InitLogger(cfg.Server.Env); err != nil {
		panic(err)
	}
	// Verify logger works (flush buffered logs at the end)
	defer logger.Log.Sync()

	logger.Log.Info("Starting Ticketing System...", zap.String("env", cfg.Server.Env))

	// 3. Init Fiber
	app := fiber.New(fiber.Config{
		AppName: cfg.Server.ServiceName,
	})

	// 4. Middlewares
	app.Use(recover.New())
	app.Use(fiberLogger.New())

	// 5. Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"app":    cfg.Server.ServiceName,
		})
	})

	// 6. Graceful Shutdown
	// Create a channel to listen for Interrupt (Ctrl+C) or SIGTERM (Kubernetes/Docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Run server in a separate goroutine so it doesn't block the main thread
	go func() {
		logger.Log.Info("Server is listening on port " + cfg.Server.Port)
		if err := app.Listen(cfg.Server.Port); err != nil {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Block main thread until a signal is received
	<-quit
	logger.Log.Info("Graceful shutdown initiated...")

	// Shutdown Fiber app (waits for ongoing requests to finish)
	if err := app.Shutdown(); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exited successfully")
}
