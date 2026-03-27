package repository

import (
	"context"
	"time"

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
	err := r.db.WithContext(ctx).Where("deleted_at is null").Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *eventRepository) UpdateEvent(ctx context.Context, event *entity.Event) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *eventRepository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Event{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error

}

func (r *eventRepository) CreateTicketType(ctx context.Context, ticketType *entity.TicketType) error {
	return r.db.WithContext(ctx).Create(ticketType).Error
}

func (r *eventRepository) CreateTicketTypes(ctx context.Context, ticketTypes []entity.TicketType) error {
	return r.db.WithContext(ctx).Create(ticketTypes).Error
}

func (r *eventRepository) ListEventsAdvanced(ctx context.Context, f entity.EventFilter) ([]entity.Event, int64, error) {
	var events []entity.Event
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Event{}).Where("deleted_at is null")

	if f.Search != "" {
		query = query.Where("name ILIKE ? ", "%"+f.Search+"%")
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	} else {
		query = query.Where("status <> ?", entity.EventStatusDraft)
	}

	if f.FromTime != nil {
		query = query.Where("start_time >= ?", *f.FromTime)
	}
	if f.ToTime != nil {
		query = query.Where("start_time <= ?", *f.ToTime)
	}

	//count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(f.Limit).Offset(f.Offset).Order("start_time DESC").Find(&events).Error; err != nil {
		return nil, 0, err
	}

	return events, total, nil
}
