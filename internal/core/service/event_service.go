package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type eventService struct {
	eventRepo port.EventRepositoryPort
}

func NewEventService(eventRepo port.EventRepositoryPort) port.EventServicePort {
	return &eventService{
		eventRepo: eventRepo,
	}
}

func (s *eventService) CreateEvent(ctx context.Context, req entity.CreateEventRequest) (*entity.Event, error) {
	// Validate input
	if err := validateCreateEventRequest(req); err != nil {
		return nil, err
	}

	// Check if slug already exists
	existingEvent, _ := s.eventRepo.GetEventBySlug(ctx, req.Slug)
	if existingEvent != nil {
		return nil, errors.New("slug đã được sử dụng")
	}

	// Create event
	event := &entity.Event{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		Location:  req.Location,
		BannerURL: req.BannerURL,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    entity.EventStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save event to database
	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

// CreateEventWithTickets tạo event + ticket types (dùng cho test)
func (s *eventService) CreateEventWithTickets(ctx context.Context, eventReq entity.CreateEventRequest, ticketTypes []entity.CreateTicketTypeRequest) (*entity.Event, error) {
	// Tạo event trước
	event, err := s.CreateEvent(ctx, eventReq)
	if err != nil {
		return nil, err
	}

	// Sau đó tạo ticket types
	if len(ticketTypes) > 0 {
		tickets := make([]entity.TicketType, 0, len(ticketTypes))
		for _, tt := range ticketTypes {
			if err := validateCreateTicketTypeRequest(tt); err != nil {
				return nil, err
			}

			ticketType := entity.TicketType{
				ID:                uuid.New(),
				EventID:           event.ID,
				Name:              tt.Name,
				Price:             tt.Price,
				InitialQuantity:   tt.InitialQuantity,
				RemainingQuantity: tt.InitialQuantity,
			}
			tickets = append(tickets, ticketType)
		}

		// Save all ticket types
		if err := s.eventRepo.CreateTicketTypes(ctx, tickets); err != nil {
			return nil, err
		}
	}

	return event, nil
}

func (s *eventService) GetEvent(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	return s.eventRepo.GetEventByID(ctx, id)
}

func (s *eventService) GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error) {
	return s.eventRepo.GetEventBySlug(ctx, slug)
}

func (s *eventService) ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.eventRepo.ListEvents(ctx, limit, offset)
}

func validateCreateEventRequest(req entity.CreateEventRequest) error {
	if req.Name == "" {
		return errors.New("tên sự kiện không được để trống")
	}
	if req.Slug == "" {
		return errors.New("slug không được để trống")
	}
	if req.Location == "" {
		return errors.New("địa điểm không được để trống")
	}
	if req.StartTime.IsZero() {
		return errors.New("thời gian bắt đầu không được để trống")
	}
	if req.EndTime.IsZero() {
		return errors.New("thời gian kết thúc không được để trống")
	}
	if req.EndTime.Before(req.StartTime) {
		return errors.New("thời gian kết thúc phải sau thời gian bắt đầu")
	}
	return nil
}

func validateCreateTicketTypeRequest(req entity.CreateTicketTypeRequest) error {
	if req.Name == "" {
		return errors.New("tên loại vé không được để trống")
	}
	if req.Price.IsNegative() || req.Price.IsZero() {
		return errors.New("giá vé phải lớn hơn 0")
	}
	if req.InitialQuantity <= 0 {
		return errors.New("số lượng vé phải lớn hơn 0")
	}
	return nil
}
