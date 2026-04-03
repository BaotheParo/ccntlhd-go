package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
)

type OrderRepositoryPort interface {
	CreateOrder(ctx context.Context, order *entity.Order) error
	CreateOrderWithTransaction(ctx context.Context, order *entity.Order, items []entity.OrderItem) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error)
	ListOrdersAdvanced(ctx context.Context, page, limit int, status string, eventID string) ([]entity.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error
	DeleteOrder(ctx context.Context, id uuid.UUID) error
	CreateOrderItem(ctx context.Context, item *entity.OrderItem) error
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]entity.OrderItem, error)
}

type OrderServicePort interface {
	PlaceOrder(ctx context.Context, userID uuid.UUID, items []entity.OrderItem) (*entity.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error)
	ListAdminOrders(ctx context.Context, page, limit int, status string, eventID string) ([]entity.Order, int64, error)
	CancelOrder(ctx context.Context, id uuid.UUID) error
	UpdateOrderStatusAdmin(ctx context.Context, id uuid.UUID, newStatus string) error
	Shutdown() error
}
