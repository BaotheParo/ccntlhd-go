package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPaid      OrderStatus = "PAID"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status      OrderStatus     `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	Items       []OrderItem     `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE;" json:"items"`
}

type OrderItem struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	OrderID      uuid.UUID       `gorm:"type:uuid;not null" json:"order_id"`
	TicketTypeID uuid.UUID       `gorm:"type:uuid;not null" json:"ticket_type_id"`
	Quantity     int             `gorm:"not null" json:"quantity"`
	UnitPrice    decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null" json:"unit_price"`
}
