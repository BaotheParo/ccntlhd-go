package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
)

type EventRepositoryPort interface {
	CreateEvent(ctx context.Context, event *entity.Event) error
	GetEventByID(ctx context.Context, id uuid.UUID) (*entity.Event, error)
	GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error)
	ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error)
	UpdateEvent(ctx context.Context, event *entity.Event) error
	DeleteEvent(ctx context.Context, id uuid.UUID) error
}

type EventServicePort interface {
	CreateEvent(ctx context.Context, req entity.CreateEventRequest) (*entity.Event, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*entity.Event, error)
	GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error)
	ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error)
}
