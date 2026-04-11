package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Khởi tạo Redis client với Hồ chứa kết nối
	// Lưu ý: Đã đổi cổng thành 6380 để khớp với container đang chạy
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
		PoolSize: 10,
	})
	defer rdb.Close()

	ctx := context.Background()

	// Kiểm tra kết nối bằng lệnh PING
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Không thể kết nối Redis: %v", err)
	}
	fmt.Println("Kết nối thành công:", pong)

	// SET một khóa với thời gian tồn tại là 60 giây
	err = rdb.Set(ctx, "user:session:abc123", "user_id=42", 60*time.Second).Err()
	if err != nil {
		log.Fatalf("Lỗi SET: %v", err)
	}

	// GET giá trị của khóa vừa tạo
	val, err := rdb.Get(ctx, "user:session:abc123").Result()
	if err == redis.Nil {
		fmt.Println("Khóa không tồn tại")
	} else if err != nil {
		log.Fatalf("Lỗi GET: %v", err)
	} else {
		fmt.Println("GET kết quả:", val)
	}
}
