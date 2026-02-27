package port

import (
	"context"

	"github.com/google/uuid"
)

// TicketCacheRepository định nghĩa các hàm thao tác với Redis cache
type TicketCacheRepository interface {
	// DeductStock kiểm tra và trừ tồn kho vé an toàn bằng Lua Script.
	// Trả về số lượng vé còn lại sau khi trừ, hoặc lỗi nếu hết vé/không đủ vé.
	DeductStock(ctx context.Context, ticketID uuid.UUID, quantity int) (int, error)

	// RollbackStock cộng lại số lượng vé vào Redis trong trường hợp
	// tạo đơn hàng ở DB (Postgres) bị lỗi.
	RollbackStock(ctx context.Context, ticketID uuid.UUID, quantity int) error

	// SetStock khởi tạo số lượng vé lên Redis (dùng cho Cache Warm-up)
	SetStock(ctx context.Context, ticketID uuid.UUID, quantity int) error
}
