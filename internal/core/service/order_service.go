package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type OrderService struct {
	repo      port.OrderRepositoryPort // repo để gọi các hàm lock/trừ kho/tạo order
	eventRepo port.EventRepositoryPort
	cacheRepo port.TicketCacheRepository // cache repo để trừ vé trên Redis
}

// NewOrderService tạo service mới, inject repo và cacheRepo vào.
func NewOrderService(repo port.OrderRepositoryPort, eventRepo port.EventRepositoryPort, cacheRepo port.TicketCacheRepository) *OrderService {
	return &OrderService{
		repo:      repo,
		eventRepo: eventRepo,
		cacheRepo: cacheRepo,
	}
}

// PlaceOrder tạo đơn hàng mới với 2-phase commit (Redis -> Postgres)
func (s *OrderService) PlaceOrder(ctx context.Context, userID uuid.UUID, items []entity.OrderItem) (*entity.Order, error) {
	// Validate input
	if len(items) == 0 {
		return nil, errors.New("đơn hàng phải có ít nhất một item")
	}

	// Calculate total amount and prepare items
	totalAmount := decimal.Zero
	for i, item := range items {
		if item.TicketTypeID == uuid.Nil {
			return nil, errors.New("ticket_type_id không được để trống")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("số lượng vé phải lớn hơn 0")
		}
		if item.UnitPrice.IsNegative() || item.UnitPrice.IsZero() {
			return nil, errors.New("giá vé phải lớn hơn 0")
		}

		itemTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)
		items[i].ID = uuid.New()
	}

	// 1. Phase 1 (Redis): Trừ tồn kho trên Redis trước
	for _, item := range items {
		_, err := s.cacheRepo.DeductStock(ctx, item.TicketTypeID, item.Quantity)
		if err != nil {
			// Nếu gặp lỗi khi trừ vé trên Redis (VD: hết vé), trả lỗi ngay lập tức
			return nil, fmt.Errorf("không thể giữ vé (loại %s): %w", item.TicketTypeID, err)
		}
	}

	// 2. Phase 2 (Postgres): Tạo Order trong DB (Với Transaction)
	order := &entity.Order{
		ID:          uuid.New(),
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       items,
	}

	for i := range items {
		items[i].OrderID = order.ID
	}

	// Gọi hàm lưu Order có kèm xử lý Transaction trong Repository
	// Giả định orderRepository.CreateOrderTransaction sẽ nhận Order và list OrderItem,
	// tạo order, giảm remaining_quantity trong DB và commit.
	// Nếu repository chưa có hàm transaction này, chúng ta sẽ implement nó sau hoặc dùng DB object trực tiếp.
	// Để giữ đúng Clean Arch, Repository nên handle DB transaction.
	err := s.repo.CreateOrderWithTransaction(ctx, order, items)

	// 3. Rollback Handling: Nếu DB fail, phải Rollback Redis
	if err != nil {
		// Hoàn lại vé vào Redis
		for _, item := range items {
			rollbackErr := s.cacheRepo.RollbackStock(ctx, item.TicketTypeID, item.Quantity)
			if rollbackErr != nil {
				// Cảnh báo: Lỗi nghiêm trọng, dữ liệu Redis và Postgres không đồng bộ!
				// Nên ghi log Critical hoặc gửi alert tới Slack/Monitoring system.
				fmt.Printf("CRITICAL: Failed to rollback Redis stock for ticket %s: %v\n", item.TicketTypeID, rollbackErr)
			}
		}
		return nil, fmt.Errorf("không thể tạo đơn hàng trong Database: %w", err)
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetOrdersByUserID(ctx, userID, limit, offset)
}

func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
	// Get order first
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	// Only PENDING orders can be cancelled
	if order.Status != entity.OrderStatusPending {
		return errors.New("chỉ có thể hủy đơn hàng ở trạng thái PENDING")
	}

	// Update status
	return s.repo.UpdateOrderStatus(ctx, id, entity.OrderStatusCancelled)
}
