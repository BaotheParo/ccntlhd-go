package entity

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
	EventStatusCancelled EventStatus = "CANCELLED"
	EventStatusEnded     EventStatus = "ENDED"
)

type Event struct {
	ID        uuid.UUID   `gorm:"type:uuid;primary_key;" json:"id"`
	Name      string      `gorm:"not null" json:"name"`
	Slug      string      `gorm:"uniqueIndex;not null" json:"slug"`
	Location  string      `gorm:"type:varchar(255)" json:"location"`
	BannerURL string      `gorm:"type:varchar(500)" json:"banner_url"`
	StartTime time.Time   `gorm:"not null" json:"start_time"`
	EndTime   time.Time   `gorm:"not null" json:"end_time"`
	Status    EventStatus `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

type TicketType struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key;" json:"id"`
	EventID           uuid.UUID       `gorm:"type:uuid;not null" json:"event_id"`
	Name              string          `gorm:"not null" json:"name"`
	Price             decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
	InitialQuantity   int             `gorm:"not null" json:"initial_quantity"`
	RemainingQuantity int             `gorm:"not null" json:"remaining_quantity"`
}

type CreateEventRequest struct {
	Name      string    `json:"name" validate:"required,min=3"`
	Slug      string    `json:"slug" validate:"required,min=3"`
	Location  string    `json:"location" validate:"required,min=3"`
	BannerURL string    `json:"banner_url"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

type CreateTicketTypeRequest struct {
	Name            string          `json:"name" validate:"required,min=3"`
	Price           decimal.Decimal `json:"price" validate:"required"`
	InitialQuantity int             `json:"initial_quantity" validate:"required,min=1"`
}
