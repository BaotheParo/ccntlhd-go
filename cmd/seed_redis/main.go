package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/pkg/config"
	redis_client "github.com/yourname/ticketing-system/pkg/redis"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Khong the tai cau hinh: %v", err)
	}

	// 1. Connect Postgres
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Khong the ket noi Postgres: %v", err)
	}

	// 2. Connect Redis
	redisClient := redis_client.NewRedisClient(cfg)

	// 3. Fetch all Ticket Types
	var ticketTypes []entity.TicketType
	db.Find(&ticketTypes)

	fmt.Println("--- Seeding Redis Cache ---")
	for _, tt := range ticketTypes {
		key := fmt.Sprintf("ticket:%s:stock", tt.ID.String())
		err := redisClient.Set(ctx, key, tt.RemainingQuantity, 0).Err()
		if err != nil {
			fmt.Printf("❌ Loi khi seed ve %s: %v\n", tt.Name, err)
		} else {
			fmt.Printf("✅ Da seed ve %s (SL: %d) vao Redis\n", tt.Name, tt.RemainingQuantity)
		}
	}
}
