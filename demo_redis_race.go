package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

func unsafeDecrement(rdb *redis.Client, ctx context.Context, key string, wg *sync.WaitGroup) {
	defer wg.Done()
	// Phương pháp sai: Đọc và ghi riêng lẻ tạo ra khoảng hở tranh chấp (Race Condition)
	val, _ := rdb.Get(ctx, key).Int()
	if val > 0 {
		// Giả lập xử lý mất thời gian để tăng khả năng xảy ra lỗi
		time.Sleep(time.Millisecond * 2)
		rdb.Set(ctx, key, val-1, 0)
	}
}

func safeDecrement(rdb *redis.Client, ctx context.Context, key string, wg *sync.WaitGroup) {
	defer wg.Done()
	// Phương pháp an toàn: Sử dụng lệnh nguyên tử của Redis (Decr)
	rdb.Decr(ctx, key)
}

func runTest(rdb *redis.Client, label string, fn func(*redis.Client, context.Context, string, *sync.WaitGroup)) {
	ctx := context.Background()
	// Reset stock về 100
	rdb.Set(ctx, "stock", 100, 0)

	var wg sync.WaitGroup
	fmt.Printf("[%s] Đang chạy 100 luồng đồng thời...\n", label)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go fn(rdb, ctx, "stock", &wg)
	}
	wg.Wait()

	final, _ := rdb.Get(ctx, "stock").Int()
	fmt.Printf(">>> KẾT QUẢ: Tồn kho còn lại: %d (Kỳ vọng: 0)\n", final)
	if final > 0 {
		fmt.Println("!!! CẢNH BÁO: Phát hiện sai lệch dữ liệu do Race Condition!")
	} else {
		fmt.Println("✅ CHÍNH XÁC: Dữ liệu được bảo toàn tuyệt đối.")
	}
	fmt.Println("--------------------------------------------------")
}

func main() {
	// Khởi tạo client (Cổng 6380 cho Docker)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6380"})

	fmt.Println("--- BẮT ĐẦU KIỂM CHỨNG TRANH CHẤP DỮ LIỆU (RACE CONDITION) ---")
	fmt.Println("Mục tiêu: 100 khách hàng cùng mua 100 món hàng đồng thời.")
	fmt.Println("")

	runTest(rdb, "PHƯƠNG PHÁP KHÔNG AN TOÀN (Get then Set)", unsafeDecrement)
	runTest(rdb, "PHƯƠNG PHÁP AN TOÀN (Atomic Decr)", safeDecrement)
}
