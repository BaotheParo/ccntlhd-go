package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Kịch bản Lua chạy nguyên tử trên máy chủ Redis
// Giúp tránh tình trạng Race Condition khi nhiều người cùng đặt vé
const decrementScript = `
local key      = KEYS[1]
local amount   = tonumber(ARGV[1])

if redis.call('EXISTS', key) == 0 then
    return -1
end
local current = tonumber(redis.call('GET', key))
if current < amount then
    return -2
end
local new_val = current - amount
redis.call('SET', key, new_val)
return new_val
`

func AtomicDecrement(rdb *redis.Client, ctx context.Context, key string, amount int) (int, error) {
	result, err := rdb.Eval(ctx, decrementScript, []string{key}, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("lỗi thực thi Lua: %w", err)
	}

	remaining := result.(int64)
	if remaining == -1 {
		return 0, fmt.Errorf("khóa không tồn tại")
	}
	if remaining == -2 {
		return 0, fmt.Errorf("không đủ số lượng để trừ (Over-selling protection)")
	}

	return int(remaining), nil
}

func main() {
	// 1. Khởi tạo client (Cổng 6380 cho Docker)
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6380"})
	ctx := context.Background()

	// 2. Thiết lập kho vé ban đầu (ví dụ: 10 vé)
	stockKey := "ticket:stock:event123"
	rdb.Set(ctx, stockKey, 10, 0)
	fmt.Println("--- BẮT ĐẦU DEMO LUA SCRIPT (ATOMIC DECREMENT) ---")
	fmt.Printf("Kho vé ban đầu: 10\n\n")

	// 3. Mô phỏng các lượt mua vé
	testCases := []int{4, 4, 3} // Tổng là 11, lượt cuối sẽ thất bại

	for i, qty := range testCases {
		fmt.Printf("Lượt mua #%d: Yêu cầu %d vé...\n", i+1, qty)
		remaining, err := AtomicDecrement(rdb, ctx, stockKey, qty)
		if err != nil {
			fmt.Printf(">>> THẤT BẠI: %v\n", err)
		} else {
			fmt.Printf(">>> THÀNH CÔNG! Số vé còn lại: %d\n", remaining)
		}
		fmt.Println("--------------------------------------------------")
	}
}
