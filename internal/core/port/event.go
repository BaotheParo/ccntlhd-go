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

	CreateTicketType(ctx context.Context, ticketType *entity.TicketType) error
	CreateTicketTypes(ctx context.Context, ticketTypes []entity.TicketType) error
	GetTicketTypeByID(ctx context.Context, id uuid.UUID) (*entity.TicketType, error)
	UpdateTicketType(ctx context.Context, ticketType *entity.TicketType) error
	GetAllTicketTypeByIDEvent(ctx context.Context, id uuid.UUID) ([]entity.TicketType, error)

	ListEventsAdvanced(ctx context.Context, query entity.EventFilter) ([]entity.Event, int64, error)
}

type EventServicePort interface {
	CreateEvent(ctx context.Context, req entity.CreateEventRequest) (*entity.Event, error)
	CreateEventWithTickets(ctx context.Context, eventReq entity.CreateEventRequest, ticketTypes []entity.CreateTicketTypeRequest) (*entity.Event, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*entity.Event, error)
	GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error)
	UpdateEvent(ctx context.Context, id uuid.UUID, req entity.UpdateEventRequest) (*entity.Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID) error
	ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error)
	ListEventsAdvanced(ctx context.Context, req entity.ListEventRequest) ([]entity.Event, int64, error)

	GetAllTicketTypeByIDEvent(ctx context.Context, id uuid.UUID) ([]entity.TicketType, error)

	CreateTicketType(ctx context.Context, eventID uuid.UUID, req entity.CreateTicketTypeRequest) (*entity.TicketType, error)
	UpdateTicketType(ctx context.Context, ticketID uuid.UUID, req entity.UpdateTicketTypeRequest) (*entity.TicketType, error)
}
