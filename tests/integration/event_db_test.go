package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB kết nối đến database thực để test
func setupTestDB(t *testing.T) *gorm.DB {
	// Sử dụng biến môi trường hoặc mặc định
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Mặc định cho Docker
		dsn = "host=localhost user=user password=password dbname=ticket_db port=5433 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

// cleanupEvents xóa test data sau khi test xong
func cleanupEvents(t *testing.T, db *gorm.DB, slugs ...string) {
	for _, slug := range slugs {
		db.Where("slug = ?", slug).Delete(&entity.Event{})
	}
}

// TestCreateEvent_DB_Success - Test tạo event vào database thực
func TestCreateEvent_DB_Success(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	slug := fmt.Sprintf("concert-db-success-%d", time.Now().UnixNano())
	req := entity.CreateEventRequest{
		Name:      "Concert Database Test",
		Slug:      slug,
		Location:  "Hà Nội Arena",
		BannerURL: "https://example.com/concert.jpg",
		StartTime: time.Date(2026, 5, 1, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 5, 1, 23, 0, 0, 0, time.UTC),
	}

	defer cleanupEvents(t, db, slug)

	// Act
	event, err := svc.CreateEvent(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned")
	}

	// Verify data saved in database
	retrieved, err := svc.GetEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if retrieved.Name != req.Name {
		t.Errorf("Expected name %q, got %q", req.Name, retrieved.Name)
	}

	if retrieved.Slug != slug {
		t.Errorf("Expected slug %q, got %q", slug, retrieved.Slug)
	}

	if retrieved.Status != entity.EventStatusDraft {
		t.Errorf("Expected status DRAFT, got %v", retrieved.Status)
	}
}

// TestCreateEventWithTickets_DB_Success - Test tạo event + ticket types vào database
func TestCreateEventWithTickets_DB_Success(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	slug := fmt.Sprintf("festival-db-success-%d", time.Now().UnixNano())
	eventReq := entity.CreateEventRequest{
		Name:      "Music Festival 2026",
		Slug:      slug,
		Location:  "Sân vận động Mỹ Đình",
		BannerURL: "https://example.com/festival.jpg",
		StartTime: time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 6, 15, 23, 0, 0, 0, time.UTC),
	}

	ticketTypes := []entity.CreateTicketTypeRequest{
		{
			Name:            "VIP",
			Price:           decimal.NewFromInt(1000000),
			InitialQuantity: 50,
		},
		{
			Name:            "Thường",
			Price:           decimal.NewFromInt(500000),
			InitialQuantity: 200,
		},
		{
			Name:            "Student",
			Price:           decimal.NewFromInt(250000),
			InitialQuantity: 100,
		},
	}

	defer cleanupEvents(t, db, slug)

	// Act
	event, err := svc.CreateEventWithTickets(ctx, eventReq, ticketTypes)

	// Assert
	if err != nil {
		t.Fatalf("CreateEventWithTickets failed: %v", err)
	}

	if event == nil {
		t.Fatal("Expected event to be returned")
	}

	// Verify event in database
	retrieved, err := svc.GetEvent(ctx, event.ID)
	if err != nil {
		t.Fatalf("GetEvent failed: %v", err)
	}

	if retrieved.Name != eventReq.Name {
		t.Errorf("Expected name %q, got %q", eventReq.Name, retrieved.Name)
	}

	// Verify ticket types were created
	var ticketTypesFromDB []entity.TicketType
	if err := db.Where("event_id = ?", event.ID).Find(&ticketTypesFromDB).Error; err != nil {
		t.Fatalf("Failed to query ticket types: %v", err)
	}

	if len(ticketTypesFromDB) != 3 {
		t.Errorf("Expected 3 ticket types, got %d", len(ticketTypesFromDB))
	}

	// Verify ticket type details
	vipCount := 0
	thường := 0
	student := 0

	for _, tt := range ticketTypesFromDB {
		if tt.Name == "VIP" && tt.Price.Equal(decimal.NewFromInt(1000000)) && tt.InitialQuantity == 50 {
			vipCount++
		}
		if tt.Name == "Thường" && tt.Price.Equal(decimal.NewFromInt(500000)) && tt.InitialQuantity == 200 {
			thường++
		}
		if tt.Name == "Student" && tt.Price.Equal(decimal.NewFromInt(250000)) && tt.InitialQuantity == 100 {
			student++
		}
	}

	if vipCount != 1 {
		t.Errorf("Expected 1 VIP ticket type, got %d", vipCount)
	}
	if thường != 1 {
		t.Errorf("Expected 1 Thường ticket type, got %d", thường)
	}
	if student != 1 {
		t.Errorf("Expected 1 Student ticket type, got %d", student)
	}
}

// TestGetEventBySlug_DB - Test lấy event theo slug từ database
func TestGetEventBySlug_DB(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	slug := fmt.Sprintf("cinema-night-%d", time.Now().UnixNano())
	req := entity.CreateEventRequest{
		Name:      "Cinema Night",
		Slug:      slug,
		Location:  "Tòa nhà Chiếu phim 1",
		BannerURL: "https://example.com/cinema.jpg",
		StartTime: time.Date(2026, 7, 10, 20, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 10, 22, 30, 0, 0, time.UTC),
	}

	defer cleanupEvents(t, db, slug)

	// Create event
	event, err := svc.CreateEvent(ctx, req)
	if err != nil {
		t.Fatalf("CreateEvent failed: %v", err)
	}

	// Get by slug
	retrieved, err := svc.GetEventBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("GetEventBySlug failed: %v", err)
	}

	if retrieved.ID != event.ID {
		t.Errorf("Expected event ID %v, got %v", event.ID, retrieved.ID)
	}

	if retrieved.Slug != slug {
		t.Errorf("Expected slug %q, got %q", slug, retrieved.Slug)
	}
}

// TestListEvents_DB - Test liệt kê events từ database
func TestListEvents_DB(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	// Create 3 events
	slugs := make([]string, 3)
	for i := 0; i < 3; i++ {
		slug := fmt.Sprintf("test-list-%d-%d", i, time.Now().UnixNano())
		slugs[i] = slug

		req := entity.CreateEventRequest{
			Name:      fmt.Sprintf("List Test Event %d", i),
			Slug:      slug,
			Location:  fmt.Sprintf("Location %d", i),
			BannerURL: "https://example.com/banner.jpg",
			StartTime: time.Date(2026, 8, 1, 19, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 8, 1, 23, 0, 0, 0, time.UTC),
		}

		_, err := svc.CreateEvent(ctx, req)
		if err != nil {
			t.Fatalf("CreateEvent failed for event %d: %v", i, err)
		}
	}

	defer cleanupEvents(t, db, slugs...)

	// List events
	events, err := svc.ListEvents(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListEvents failed: %v", err)
	}

	if len(events) == 0 {
		t.Fatal("Expected events to be returned")
	}

	// Verify our created events are in the list
	count := 0
	for _, event := range events {
		for _, slug := range slugs {
			if event.Slug == slug {
				count++
			}
		}
	}

	if count < 3 {
		t.Errorf("Expected to find at least 3 created events, found %d", count)
	}
}

// TestCreateEvent_DB_DuplicateSlug - Test không thể tạo event với slug trùng
func TestCreateEvent_DB_DuplicateSlug(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	slug := fmt.Sprintf("duplicate-slug-%d", time.Now().UnixNano())

	req1 := entity.CreateEventRequest{
		Name:      "Event 1",
		Slug:      slug,
		Location:  "Location 1",
		BannerURL: "https://example.com/banner1.jpg",
		StartTime: time.Date(2026, 9, 1, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 9, 1, 23, 0, 0, 0, time.UTC),
	}

	defer cleanupEvents(t, db, slug)

	// Create first event
	_, err := svc.CreateEvent(ctx, req1)
	if err != nil {
		t.Fatalf("First CreateEvent failed: %v", err)
	}

	// Try to create second event with same slug
	req2 := entity.CreateEventRequest{
		Name:      "Event 2",
		Slug:      slug,
		Location:  "Location 2",
		BannerURL: "https://example.com/banner2.jpg",
		StartTime: time.Date(2026, 9, 2, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 9, 2, 23, 0, 0, 0, time.UTC),
	}

	_, err = svc.CreateEvent(ctx, req2)

	if err == nil {
		t.Fatal("Expected error for duplicate slug, but got none")
	}

	if err.Error() != "slug đã được sử dụng" {
		t.Errorf("Expected 'slug đã được sử dụng' error, got: %v", err)
	}
}
