package entity

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TicketType struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	EventID           uuid.UUID       `gorm:"type:uuid;not null" json:"event_id"`
	Name              string          `gorm:"not null" json:"name"`
	Price             decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	InitialQuantity   int             `gorm:"not null" json:"initial_quantity"`
	RemainingQuantity int             `gorm:"not null" json:"remaining_quantity"`
}
