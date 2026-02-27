package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yourname/ticketing-system/internal/core/port"
)

var (
	ErrTicketNotFound    = errors.New("ticket type not found in cache")
	ErrInsufficientStock = errors.New("insufficient ticket stock")
)

// deductStockScript là Lua script dùng để trừ tồn kho an toàn (atomic).
// Giải thích logic Lua:
// 1. Nhận KEYS[1] là tên key chứa số lượng vé (VD: "ticket:123:stock").
// 2. Nhận ARGV[1] là số lượng user muốn mua.
// 3. redis.call("EXISTS", KEYS[1]): Kiểm tra xem vé có trong cache không.
//   - Nếu không (0), trả về -1 (Lỗi: Không tìm thấy vé).
//
// 4. redis.call("GET", KEYS[1]): Lấy số lượng hiện tại.
//   - Nếu số lượng hiện tại < số lượng muốn mua, trả về -2 (Lỗi: Hết vé/Đủ).
//
// 5. Nếu đủ điều kiện, gọi redis.call("DECRBY", KEYS[1], ARGV[1]) để trừ đi số vé.
// 6. Trả về số lượng vé còn lại sau khi trừ.
const deductStockScript = `
	local key = KEYS[1]
	local quantity = tonumber(ARGV[1])
	
	local exists = redis.call("EXISTS", key)
	if exists == 0 then
		return -1
	end
	
	local stock = tonumber(redis.call("GET", key))
	if stock < quantity then
		return -2
	end
	
	redis.call("DECRBY", key, quantity)
	return stock - quantity
`

type redisTicketRepository struct {
	client *redis.Client
}

func NewRedisTicketRepository(client *redis.Client) port.TicketCacheRepository {
	return &redisTicketRepository{
		client: client,
	}
}

func (r *redisTicketRepository) formatKey(ticketID uuid.UUID) string {
	return fmt.Sprintf("ticket:%s:stock", ticketID.String())
}

func (r *redisTicketRepository) DeductStock(ctx context.Context, ticketID uuid.UUID, quantity int) (int, error) {
	key := r.formatKey(ticketID)

	// Thực thi Lua Script. Eval nhận context, script string, mảng KEYS, mảng ARGV
	result, err := r.client.Eval(ctx, deductStockScript, []string{key}, quantity).Result()
	if err != nil {
		return 0, fmt.Errorf("redis eval error: %w", err)
	}

	remaining, ok := result.(int64)
	if !ok {
		return 0, errors.New("unexpected return type from lua script")
	}

	// Phân loại lỗi dựa trên giá trị trả về của Lua Script
	switch remaining {
	case -1:
		return 0, ErrTicketNotFound
	case -2:
		return 0, ErrInsufficientStock
	}

	return int(remaining), nil
}

func (r *redisTicketRepository) RollbackStock(ctx context.Context, ticketID uuid.UUID, quantity int) error {
	key := r.formatKey(ticketID)
	// Hoàn trả lại số lượng vé bằng lệnh INCRBY
	err := r.client.IncrBy(ctx, key, int64(quantity)).Err()
	if err != nil {
		return fmt.Errorf("failed to rollback stock in redis: %w", err)
	}
	return nil
}

func (r *redisTicketRepository) SetStock(ctx context.Context, ticketID uuid.UUID, quantity int) error {
	key := r.formatKey(ticketID)
	// Khởi tạo số lượng vé (chỉ dùng lúc khởi tạo application hoặc event)
	err := r.client.Set(ctx, key, quantity, 0).Err() // 0 = no expiration
	if err != nil {
		return fmt.Errorf("failed to set stock in redis: %w", err)
	}
	return nil
}
