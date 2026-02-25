package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type orderService struct {
	orderRepository port.OrderRepositoryPort
	eventRepository port.EventRepositoryPort
}

func NewOrderService(
	orderRepository port.OrderRepositoryPort,
	eventRepository port.EventRepositoryPort,
) port.OrderServicePort {
	return &orderService{
		orderRepository: orderRepository,
		eventRepository: eventRepository,
	}
}

// PlaceOrder tạo đơn hàng mới
func (s *orderService) PlaceOrder(ctx context.Context, userID uuid.UUID, items []entity.OrderItem) (*entity.Order, error) {
	// Validate input
	if len(items) == 0 {
		return nil, errors.New("đơn hàng phải có ít nhất một item")
	}

	// Validate mỗi item
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

		// Tính tổng tiền
		itemTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)

		items[i].ID = uuid.New()
	}

	// Tạo order
	order := &entity.Order{
		ID:          uuid.New(),
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       items,
	}

	// Lưu order vào database
	if err := s.orderRepository.CreateOrder(ctx, order); err != nil {
		return nil, err
	}

	// Lưu order items
	for _, item := range items {
		item.OrderID = order.ID
		if err := s.orderRepository.CreateOrderItem(ctx, &item); err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	return s.orderRepository.GetOrderByID(ctx, id)
}

func (s *orderService) GetUserOrders(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.orderRepository.GetOrdersByUserID(ctx, userID, limit, offset)
}

func (s *orderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
	// Get order first
	order, err := s.orderRepository.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	// Only PENDING orders can be cancelled
	if order.Status != entity.OrderStatusPending {
		return errors.New("chỉ có thể hủy đơn hàng ở trạng thái PENDING")
	}

	// Update status
	return s.orderRepository.UpdateOrderStatus(ctx, id, entity.OrderStatusCancelled)
}
