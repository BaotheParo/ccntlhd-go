package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) port.OrderRepositoryPort {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) CreateOrderWithTransaction(ctx context.Context, order *entity.Order, items []entity.OrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Tạo order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 2. Tạo logic trừ kho trong DB Postgres (sync từ Redis)
		for _, item := range items {
			// Lưu ý: Cần xử lý logic trừ kho trong bảng ticket_types ở đây
			// Decrement DB stock (GORM)
			result := tx.Exec("UPDATE ticket_types SET remaining_quantity = remaining_quantity - ? WHERE id = ? AND remaining_quantity >= ?", item.Quantity, item.TicketTypeID, item.Quantity)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errors.New("không đủ số lượng vé trong database")
			}
		}

		return nil
	})
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetOrdersByUserID(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error) {
	var orders []entity.Order
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *orderRepository) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Order{}, "id = ?", id).Error
}

func (r *orderRepository) CreateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *orderRepository) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]entity.OrderItem, error) {
	var items []entity.OrderItem
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Find(&items).Error
	return items, err
}
