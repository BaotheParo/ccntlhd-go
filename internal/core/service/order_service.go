package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
)

// RequestItem là struct đơn giản để nhận input từ client (mobile/web).
// Ví dụ: {"ticket_type_id": "...", "quantity": 2}
type RequestItem struct {
	TicketTypeID uuid.UUID
	Quantity     int
}

type OrderService struct {
	db   *gorm.DB                    // connection DB để bắt đầu transaction
	repo *repository.OrderRepository // repo để gọi các hàm lock/trừ kho/tạo order
}

// NewOrderService tạo service mới, inject db và repo vào.
func NewOrderService(db *gorm.DB, repo *repository.OrderRepository) *OrderService {
	return &OrderService{
		db:   db,
		repo: repo,
	}
}

// PlaceOrder là hàm chính để user đặt vé.
// Nhận userID và list các loại vé muốn mua (có thể mua nhiều loại cùng lúc).
// Trả về order vừa tạo nếu thành công, hoặc lỗi nếu fail (hết vé, lỗi DB...).
func (s *OrderService) PlaceOrder(ctx context.Context, userID uuid.UUID, requestItems []RequestItem) (*entity.Order, error) {
	// Bắt đầu transaction – mọi thứ từ đây phải thành công hết, không thì rollback sạch
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Defer rollback phòng hờ: nếu panic hoặc có lỗi, tự động rollback.
	// Nếu commit thành công thì rollback này bị ignore (no-op).
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r) // ném panic lên trên để log lỗi
		} else if tx.Error != nil {
			tx.Rollback()
		}
	}()

	totalAmount := decimal.Zero // tổng tiền, dùng decimal để tránh lỗi float
	var orderItems []entity.OrderItem
	orderID := uuid.New() // sinh ID đơn hàng mới

	// Duyệt từng loại vé user muốn mua
	for _, item := range requestItems {
		// 1. Khóa vé lại (FOR UPDATE) để check và trừ kho an toàn
		ticketType, err := s.repo.GetTicketTypeForUpdate(ctx, tx, item.TicketTypeID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		// 2. Check xem còn đủ vé không
		if ticketType.RemainingQuantity < item.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("hết vé rồi bro! Loại vé: %s chỉ còn %d, bạn mua %d",
				ticketType.Name, ticketType.RemainingQuantity, item.Quantity)
		}

		// 3. Tính tiền cho loại vé này và cộng dồn tổng
		itemTotal := ticketType.Price.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)

		// 4. Trừ kho ngay (remaining_quantity -= quantity)
		if err := s.repo.DecreaseStock(ctx, tx, item.TicketTypeID, item.Quantity); err != nil {
			tx.Rollback()
			return nil, err
		}

		// 5. Tạo OrderItem (snapshot giá lúc mua, để sau này tính tiền không bị thay đổi)
		orderItems = append(orderItems, entity.OrderItem{
			ID:           uuid.New(),
			OrderID:      orderID,
			TicketTypeID: item.TicketTypeID,
			Quantity:     item.Quantity,
			UnitPrice:    ticketType.Price, // lưu giá lúc mua
		})
	}

	// 6. Tạo entity Order chính
	order := &entity.Order{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPending, // chờ thanh toán
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       orderItems, // gắn luôn list items vào order
	}

	// 7. Lưu order + order_items vào DB (GORM tự handle association)
	if err := s.repo.CreateOrder(ctx, tx, order); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 8. Commit transaction – nếu tới đây thì coi như thành công
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return order, nil
}
