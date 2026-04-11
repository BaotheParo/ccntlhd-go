package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()

func init() {
	// Lưu ý: Đã đổi cổng thành 6380 để khớp với container đang chạy
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6380"})
}

func SetValue(key, value string, ttl time.Duration) error {
	return rdb.Set(ctx, key, value, ttl).Err()
}

func GetValue(key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

func main() {
	if err := SetValue("my:counter", "100", 10*time.Second); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Đã lưu dữ liệu thành công.")

	val, _ := GetValue("my:counter")
	fmt.Println("Kết quả đọc ngay lập tức:", val)

	fmt.Println("Chờ 11 giây để dữ liệu hết hạn...")
	time.Sleep(11 * time.Second)

	val2, err2 := GetValue("my:counter")
	if err2 == redis.Nil {
		fmt.Println("Khóa đã hết hạn (redis.Nil) - Cơ chế hoạt động đúng đắn!")
	} else {
		fmt.Println("Vẫn còn giá trị:", val2)
	}
}
