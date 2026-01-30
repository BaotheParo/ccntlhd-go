package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderRepository là thằng trung gian nói chuyện với database cho phần Order.
// Mọi thao tác với DB đều đi qua đây để dễ quản lý và test.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository tạo mới một repo, truyền vào db connection là xong.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// GetTicketTypeForUpdate lấy thông tin loại vé theo ID, kèm theo khóa bi quan (FOR UPDATE).
// Dùng để "khóa" dòng vé lại, không cho thằng khác đụng vào trong lúc mình đang trừ số lượng.
// LƯU Ý: Phải gọi hàm này bên trong transaction (tx), chứ không dùng db trực tiếp.
func (r *OrderRepository) GetTicketTypeForUpdate(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*entity.TicketType, error) {
	var ticketType entity.TicketType

	// Dùng Clauses(clause.Locking{Strength: "UPDATE"}) để thêm FOR UPDATE vào query
	// Nghĩa là: "Ê PostgreSQL, khóa dòng này cho tao, đừng cho ai đụng tới trong lúc tao xử lý!"
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&ticketType, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ticketType, nil
}

// DecreaseStock trừ số lượng vé còn lại (remaining_quantity) đi một lượng quantity.
// Cũng phải gọi trong transaction (tx) để đảm bảo an toàn khi nhiều người đặt cùng lúc.
func (r *OrderRepository) DecreaseStock(ctx context.Context, tx *gorm.DB, id uuid.UUID, quantity int) error {
	// Dùng gorm.Expr để viết câu update kiểu "remaining_quantity = remaining_quantity - ?"
	// Tránh race condition kiểu thằng A đọc 10, thằng B đọc 10, cả hai trừ 5 → còn 5 thay vì 0.
	return tx.WithContext(ctx).
		Model(&entity.TicketType{}).
		Where("id = ?", id).
		Update("remaining_quantity", gorm.Expr("remaining_quantity - ?", quantity)).
		Error
}

// CreateOrder tạo mới một đơn hàng kèm theo các OrderItem bên trong.
// Gọi trong transaction để đảm bảo: hoặc tạo hết, hoặc không tạo cái nào cả (atomic).
func (r *OrderRepository) CreateOrder(ctx context.Context, tx *gorm.DB, order *entity.Order) error {
	// GORM tự động tạo cả order và các order_items liên quan (nếu có association)
	return tx.WithContext(ctx).Create(order).Error
}
