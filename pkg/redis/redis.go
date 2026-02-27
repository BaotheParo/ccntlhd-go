package redis_client

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourname/ticketing-system/pkg/config"
)

// NewRedisClient khởi tạo và trả về một Redis client
func NewRedisClient(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: 100, // Tối ưu cho high concurrency
	})

	// Retry connect
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		err = rdb.Ping(ctx).Err()
		if err == nil {
			log.Println("✅ Connected to Redis successfully!")
			return rdb
		}
		log.Printf("⏳ Đang đợi Redis... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Không thể kết nối Redis: %v", err)
	}
	return rdb
}
