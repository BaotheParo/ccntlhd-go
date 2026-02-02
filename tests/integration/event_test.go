package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
	"gorm.io/gorm"
)

// Mock Repository cho testing
type mockEventRepository struct {
	events      map[uuid.UUID]*entity.Event
	ticketTypes map[uuid.UUID]*entity.TicketType
	slugs       map[string]uuid.UUID
}

func NewMockEventRepository() *mockEventRepository {
	return &mockEventRepository{
		events:      make(map[uuid.UUID]*entity.Event),
		ticketTypes: make(map[uuid.UUID]*entity.TicketType),
		slugs:       make(map[string]uuid.UUID),
	}
}

func (m *mockEventRepository) CreateEvent(ctx context.Context, event *entity.Event) error {
	m.events[event.ID] = event
	m.slugs[event.Slug] = event.ID
	return nil
}

func (m *mockEventRepository) GetEventByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	if event, ok := m.events[id]; ok {
		return event, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockEventRepository) GetEventBySlug(ctx context.Context, slug string) (*entity.Event, error) {
	if id, ok := m.slugs[slug]; ok {
		if event, ok := m.events[id]; ok {
			return event, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockEventRepository) ListEvents(ctx context.Context, limit int, offset int) ([]entity.Event, error) {
	events := make([]entity.Event, 0)
	for _, event := range m.events {
		events = append(events, *event)
	}
	return events, nil
}

func (m *mockEventRepository) UpdateEvent(ctx context.Context, event *entity.Event) error {
	m.events[event.ID] = event
	return nil
}

func (m *mockEventRepository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	delete(m.events, id)
	return nil
}

func (m *mockEventRepository) CreateTicketType(ctx context.Context, ticketType *entity.TicketType) error {
	m.ticketTypes[ticketType.ID] = ticketType
	return nil
}

func (m *mockEventRepository) CreateTicketTypes(ctx context.Context, ticketTypes []entity.TicketType) error {
	for i := range ticketTypes {
		m.ticketTypes[ticketTypes[i].ID] = &ticketTypes[i]
	}
	return nil
}

// Tests
func TestCreateEvent_Success(t *testing.T) {
	// Arrange
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	req := entity.CreateEventRequest{
		Name:      "Hòa nhạc lớn 2026",
		Slug:      "hoa-nhac-lon-2026",
		Location:  "Sân vận động Mỹ Đình",
		BannerURL: "https://example.com/banner.jpg",
		StartTime: time.Date(2026, 3, 15, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 3, 15, 23, 0, 0, 0, time.UTC),
	}

	// Act
	event, err := svc.CreateEvent(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned, got nil")
	}

	if event.Name != req.Name {
		t.Errorf("Expected name %q, got %q", req.Name, event.Name)
	}

	if event.Slug != req.Slug {
		t.Errorf("Expected slug %q, got %q", req.Slug, event.Slug)
	}

	if event.Location != req.Location {
		t.Errorf("Expected location %q, got %q", req.Location, event.Location)
	}

	if event.Status != entity.EventStatusDraft {
		t.Errorf("Expected status %v, got %v", entity.EventStatusDraft, event.Status)
	}
}

func TestCreateEvent_EmptyName(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	req := entity.CreateEventRequest{
		Name:      "",
		Slug:      "test-event",
		Location:  "Sân vận động",
		BannerURL: "https://example.com/banner.jpg",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
	}

	_, err := svc.CreateEvent(ctx, req)

	if err == nil {
		t.Fatal("Expected error for empty name")
	}

	if err.Error() != "tên sự kiện không được để trống" {
		t.Errorf("Expected empty name error, got: %v", err)
	}
}

func TestCreateEvent_InvalidTimeRange(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	startTime := time.Now().Add(25 * time.Hour)
	endTime := time.Now().Add(24 * time.Hour) // Earlier than startTime

	req := entity.CreateEventRequest{
		Name:      "Test Event",
		Slug:      "test-event",
		Location:  "Sân vận động",
		BannerURL: "https://example.com/banner.jpg",
		StartTime: startTime,
		EndTime:   endTime,
	}

	_, err := svc.CreateEvent(ctx, req)

	if err == nil {
		t.Fatal("Expected error for invalid time range")
	}

	if err.Error() != "thời gian kết thúc phải sau thời gian bắt đầu" {
		t.Errorf("Expected time range error, got: %v", err)
	}
}

func TestCreateEvent_DuplicateSlug(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	req1 := entity.CreateEventRequest{
		Name:      "Event 1",
		Slug:      "same-slug",
		Location:  "Location 1",
		BannerURL: "https://example.com/banner1.jpg",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
	}

	_, err := svc.CreateEvent(ctx, req1)
	if err != nil {
		t.Fatalf("First CreateEvent failed: %v", err)
	}

	req2 := entity.CreateEventRequest{
		Name:      "Event 2",
		Slug:      "same-slug",
		Location:  "Location 2",
		BannerURL: "https://example.com/banner2.jpg",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
	}

	_, err = svc.CreateEvent(ctx, req2)

	if err == nil {
		t.Fatal("Expected error for duplicate slug")
	}

	if err.Error() != "slug đã được sử dụng" {
		t.Errorf("Expected duplicate slug error, got: %v", err)
	}
}

func TestCreateEventWithTickets_Success(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	eventReq := entity.CreateEventRequest{
		Name:      "Concert 2026",
		Slug:      "concert-2026",
		Location:  "Hà Nội",
		BannerURL: "https://example.com/concert.jpg",
		StartTime: time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 20, 23, 0, 0, 0, time.UTC),
	}

	ticketTypes := []entity.CreateTicketTypeRequest{
		{
			Name:            "VIP",
			Price:           decimal.NewFromInt(500000),
			InitialQuantity: 100,
		},
		{
			Name:            "Standard",
			Price:           decimal.NewFromInt(250000),
			InitialQuantity: 500,
		},
		{
			Name:            "Budget",
			Price:           decimal.NewFromInt(100000),
			InitialQuantity: 1000,
		},
	}

	// Act
	event, err := svc.CreateEventWithTickets(ctx, eventReq, ticketTypes)

	// Assert
	if err != nil {
		t.Fatalf("CreateEventWithTickets failed: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned")
	}

	if event.Name != eventReq.Name {
		t.Errorf("Expected event name %q, got %q", eventReq.Name, event.Name)
	}

	if len(mockRepo.ticketTypes) != 3 {
		t.Errorf("Expected 3 ticket types to be created, got %d", len(mockRepo.ticketTypes))
	}

	// Verify ticket types
	vipCount := 0
	standardCount := 0
	budgetCount := 0

	for _, tt := range mockRepo.ticketTypes {
		if tt.EventID != event.ID {
			t.Errorf("Ticket type has wrong event ID")
		}

		if tt.Name == "VIP" && tt.Price.Equal(decimal.NewFromInt(500000)) {
			vipCount++
		}
		if tt.Name == "Standard" && tt.Price.Equal(decimal.NewFromInt(250000)) {
			standardCount++
		}
		if tt.Name == "Budget" && tt.Price.Equal(decimal.NewFromInt(100000)) {
			budgetCount++
		}
	}

	if vipCount != 1 {
		t.Errorf("Expected 1 VIP ticket type, got %d", vipCount)
	}
	if standardCount != 1 {
		t.Errorf("Expected 1 Standard ticket type, got %d", standardCount)
	}
	if budgetCount != 1 {
		t.Errorf("Expected 1 Budget ticket type, got %d", budgetCount)
	}
}

func TestCreateEventWithTickets_InvalidTicketPrice(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	eventReq := entity.CreateEventRequest{
		Name:      "Concert",
		Slug:      "concert",
		Location:  "Hà Nội",
		BannerURL: "https://example.com/concert.jpg",
		StartTime: time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 20, 23, 0, 0, 0, time.UTC),
	}

	ticketTypes := []entity.CreateTicketTypeRequest{
		{
			Name:            "Invalid",
			Price:           decimal.NewFromInt(0), // Invalid price
			InitialQuantity: 100,
		},
	}

	_, err := svc.CreateEventWithTickets(ctx, eventReq, ticketTypes)

	if err == nil {
		t.Fatal("Expected error for invalid ticket price")
	}

	if err.Error() != "giá vé phải lớn hơn 0" {
		t.Errorf("Expected price validation error, got: %v", err)
	}
}

func TestCreateEventWithTickets_InvalidTicketQuantity(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	eventReq := entity.CreateEventRequest{
		Name:      "Concert",
		Slug:      "concert",
		Location:  "Hà Nội",
		BannerURL: "https://example.com/concert.jpg",
		StartTime: time.Date(2026, 4, 20, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 20, 23, 0, 0, 0, time.UTC),
	}

	ticketTypes := []entity.CreateTicketTypeRequest{
		{
			Name:            "Invalid",
			Price:           decimal.NewFromInt(100000),
			InitialQuantity: 0, // Invalid quantity
		},
	}

	_, err := svc.CreateEventWithTickets(ctx, eventReq, ticketTypes)

	if err == nil {
		t.Fatal("Expected error for invalid ticket quantity")
	}

	if err.Error() != "số lượng vé phải lớn hơn 0" {
		t.Errorf("Expected quantity validation error, got: %v", err)
	}
}

func TestGetEvent_Success(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	// Create event first
	req := entity.CreateEventRequest{
		Name:      "Test Event",
		Slug:      "test-event",
		Location:  "Test Location",
		BannerURL: "https://example.com/banner.jpg",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
	}

	event, _ := svc.CreateEvent(ctx, req)

	// Get event
	retrieved, err := svc.GetEvent(ctx, event.ID)

	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if retrieved.ID != event.ID {
		t.Errorf("Expected event ID %v, got %v", event.ID, retrieved.ID)
	}
}

func TestGetEventBySlug_Success(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	// Create event first
	req := entity.CreateEventRequest{
		Name:      "Test Event",
		Slug:      "test-event-slug",
		Location:  "Test Location",
		BannerURL: "https://example.com/banner.jpg",
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
	}

	event, _ := svc.CreateEvent(ctx, req)

	// Get event by slug
	retrieved, err := svc.GetEventBySlug(ctx, "test-event-slug")

	if err != nil {
		t.Fatalf("GetEventBySlug failed: %v", err)
	}

	if retrieved.ID != event.ID {
		t.Errorf("Expected event ID %v, got %v", event.ID, retrieved.ID)
	}

	if retrieved.Slug != "test-event-slug" {
		t.Errorf("Expected slug %q, got %q", "test-event-slug", retrieved.Slug)
	}
}

func TestListEvents(t *testing.T) {
	mockRepo := NewMockEventRepository()
	svc := service.NewEventService(mockRepo)
	ctx := context.Background()

	// Create multiple events
	for i := 1; i <= 5; i++ {
		req := entity.CreateEventRequest{
			Name:      "Event " + string(rune(i)),
			Slug:      "event-" + string(rune(i)),
			Location:  "Location " + string(rune(i)),
			BannerURL: "https://example.com/banner.jpg",
			StartTime: time.Now().Add(24 * time.Hour),
			EndTime:   time.Now().Add(25 * time.Hour),
		}
		svc.CreateEvent(ctx, req)
	}

	// List events
	events, err := svc.ListEvents(ctx, 10, 0)

	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}
