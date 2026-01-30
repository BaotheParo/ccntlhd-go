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

	// Save to database
	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return nil, err
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
