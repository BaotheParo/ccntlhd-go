package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
)

// TestCheckDataInDatabase - Test xem dữ liệu có được lưu vào database thực không (không delete)
func TestCheckDataInDatabase(t *testing.T) {
	db := setupTestDB(t)
	eventRepo := repository.NewEventRepository(db)
	svc := service.NewEventService(eventRepo)
	ctx := context.Background()

	slug := fmt.Sprintf("verify-data-%d", time.Now().UnixNano())
	eventReq := entity.CreateEventRequest{
		Name:      "Verify Data Test",
		Slug:      slug,
		Location:  "Hà Nội",
		BannerURL: "https://example.com/test.jpg",
		StartTime: time.Date(2026, 10, 1, 19, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 10, 1, 23, 0, 0, 0, time.UTC),
	}

	ticketTypes := []entity.CreateTicketTypeRequest{
		{
			Name:            "VIP - Verify",
			Price:           decimal.NewFromInt(999000),
			InitialQuantity: 25,
		},
		{
			Name:            "Regular - Verify",
			Price:           decimal.NewFromInt(499000),
			InitialQuantity: 100,
		},
	}

	// Create event + ticket types
	event, err := svc.CreateEventWithTickets(ctx, eventReq, ticketTypes)
	if err != nil {
		t.Fatalf("CreateEventWithTickets failed: %v", err)
	}

	t.Logf("✅ Event created successfully!")
	t.Logf("   Event ID: %v", event.ID)
	t.Logf("   Event Name: %s", event.Name)
	t.Logf("   Event Slug: %s", event.Slug)
	t.Logf("   Status: %v", event.Status)

	// Verify event exists in database
	var eventFromDB entity.Event
	if err := db.First(&eventFromDB, "id = ?", event.ID).Error; err != nil {
		t.Fatalf("Event not found in database: %v", err)
	}

	t.Logf("✅ Event found in database!")
	t.Logf("   ID: %v", eventFromDB.ID)
	t.Logf("   Name: %s", eventFromDB.Name)
	t.Logf("   Slug: %s", eventFromDB.Slug)
	t.Logf("   Location: %s", eventFromDB.Location)
	t.Logf("   Status: %v", eventFromDB.Status)
	t.Logf("   Created At: %v", eventFromDB.CreatedAt)

	// Verify ticket types exist in database
	var ticketsFromDB []entity.TicketType
	if err := db.Where("event_id = ?", event.ID).Find(&ticketsFromDB).Error; err != nil {
		t.Fatalf("Failed to query ticket types: %v", err)
	}

	t.Logf("✅ Ticket types found in database: %d", len(ticketsFromDB))
	for i, tt := range ticketsFromDB {
		t.Logf("   Ticket %d:", i+1)
		t.Logf("      ID: %v", tt.ID)
		t.Logf("      Name: %s", tt.Name)
		t.Logf("      Price: %s", tt.Price.String())
		t.Logf("      Initial Quantity: %d", tt.InitialQuantity)
		t.Logf("      Remaining Quantity: %d", tt.RemainingQuantity)
	}

	// Count all events
	var totalEvents int64
	db.Model(&entity.Event{}).Count(&totalEvents)
	t.Logf("✅ Total events in database: %d", totalEvents)

	// Count all ticket types
	var totalTickets int64
	db.Model(&entity.TicketType{}).Count(&totalTickets)
	t.Logf("✅ Total ticket types in database: %d", totalTickets)
}

// TestDirectDatabaseQuery - Test truy vấn trực tiếp database để xác nhận dữ liệu
func TestDirectDatabaseQuery(t *testing.T) {
	db := setupTestDB(t)

	t.Logf("\n📊 DATABASE STATISTICS:")
	t.Logf("==================================================")

	// Count events
	var eventCount int64
	db.Model(&entity.Event{}).Count(&eventCount)
	t.Logf("Total Events: %d", eventCount)

	// List all events
	var events []entity.Event
	if err := db.Find(&events).Error; err != nil {
		t.Logf("Error querying events: %v", err)
	} else {
		t.Logf("\nAll Events in Database:")
		for _, e := range events {
			t.Logf("  - %s (ID: %v, Status: %v)", e.Name, e.ID, e.Status)
		}
	}

	// Count ticket types
	var ticketCount int64
	db.Model(&entity.TicketType{}).Count(&ticketCount)
	t.Logf("\nTotal Ticket Types: %d", ticketCount)

	// List all ticket types
	var tickets []entity.TicketType
	if err := db.Find(&tickets).Error; err != nil {
		t.Logf("Error querying ticket types: %v", err)
	} else {
		t.Logf("\nAll Ticket Types in Database:")
		for _, tt := range tickets {
			t.Logf("  - %s (Price: %s, Quantity: %d/%d)",
				tt.Name, tt.Price.String(), tt.RemainingQuantity, tt.InitialQuantity)
		}
	}

	t.Logf("==")
}
