package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) port.EventRepositoryPort {
	return &eventRepository{db: db}
}

func (r *eventRepository) CreateEvent(ctx context.Context, event *entity.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	var event entity.Event
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error) {
	var event entity.Event
	err := r.db.WithContext(ctx).First(&event, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error) {
	var events []entity.Event
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *eventRepository) UpdateEvent(ctx context.Context, event *entity.Event) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *eventRepository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Event{}, "id = ?", id).Error
}

func (r *eventRepository) CreateTicketType(ctx context.Context, ticketType *entity.TicketType) error {
	return r.db.WithContext(ctx).Create(ticketType).Error
}

func (r *eventRepository) CreateTicketTypes(ctx context.Context, ticketTypes []entity.TicketType) error {
	return r.db.WithContext(ctx).Create(ticketTypes).Error
}
