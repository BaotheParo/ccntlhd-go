package integration

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"os"

	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func setupDB() *gorm.DB {
	dsn := os.Getenv("TEST_DSN")
	if dsn == "" {
		dsn = "host=localhost user=user password=password dbname=ticket_db port=5433 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func TestConcurrentOrderPlacement(t *testing.T) {
	db := setupDB()

	// Clean up schema to remove any zombie columns from previous runs
	db.Migrator().DropTable(&entity.OrderItem{}, &entity.Order{}, &entity.TicketType{})

	// Migration
	if err := db.AutoMigrate(&entity.TicketType{}, &entity.Order{}, &entity.OrderItem{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Clean up previous test data
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM ticket_types")
	db.Exec("DELETE FROM events")
	db.Exec("DELETE FROM users")

	// 1. Create User
	userID := uuid.New()
	if err := db.Exec("INSERT INTO users (id, username, email, password_hash) VALUES (?, ?, ?, ?)", userID, "testuser", "test@example.com", "hash").Error; err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	// 2. Create Event
	eventID := uuid.New()
	if err := db.Exec("INSERT INTO events (id, name, slug, start_time, end_time) VALUES (?, ?, ?, ?, ?)", eventID, "Test Event", "test-event", time.Now(), time.Now().Add(1*time.Hour)).Error; err != nil {
		t.Fatalf("Failed to seed event: %v", err)
	}

	// 3. Create TicketType
	ticketID := uuid.New()
	initialStock := 10
	price := decimal.NewFromFloat(100.00)

	ticket := entity.TicketType{
		ID:                ticketID,
		EventID:           eventID,
		Name:              "VIP Pass",
		Price:             price,
		InitialQuantity:   initialStock,
		RemainingQuantity: initialStock,
	}

	if err := db.Create(&ticket).Error; err != nil {
		t.Fatalf("Failed to seed ticket: %v", err)
	}

	// Initialize Service
	repo := repository.NewOrderRepository(db)
	svc := service.NewOrderService(db, repo)

	// Simulation: 20 concurrenct requests, each buying 1 ticket.
	// Only 10 should succeed.
	var wg sync.WaitGroup
	concurrentUsers := 20
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	fmt.Println("🚀 Starting concurrent order test...")

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ctx := context.Background()

			items := []service.RequestItem{
				{TicketTypeID: ticketID, Quantity: 1},
			}

			// Use the single userID for all requests (simulating one user or valid logic)
			// Ideally multiple users, but for locking logic verification on TicketType, this is sufficient.
			_, err := svc.PlaceOrder(ctx, userID, items)
			mu.Lock()
			if err != nil {
				failCount++
				fmt.Printf("Error placement: %v\n", err)
				// Expected error when out of stock
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Verification
	var updatedTicket entity.TicketType
	db.First(&updatedTicket, "id = ?", ticketID)

	fmt.Printf("✅ Test Complete:\n")
	fmt.Printf("   Initial Stock: %d\n", initialStock)
	fmt.Printf("   Concurrent Requests: %d\n", concurrentUsers)
	fmt.Printf("   Successful Orders: %d\n", successCount)
	fmt.Printf("   Failed Orders: %d\n", failCount)
	fmt.Printf("   Remaining Stock in DB: %d\n", updatedTicket.RemainingQuantity)

	if successCount != initialStock {
		t.Errorf("Expected %d successful orders, got %d", initialStock, successCount)
	}

	if updatedTicket.RemainingQuantity != 0 {
		t.Errorf("Expected remaining quantity 0, got %d", updatedTicket.RemainingQuantity)
	}

	var totalSold int64
	db.Model(&entity.TicketType{}).Where("id = ?", ticketID).Select("initial_quantity - remaining_quantity").Scan(&totalSold)
	if totalSold != int64(initialStock) {
		t.Errorf("DB Consistency check failed: sold %d != initial %d", totalSold, initialStock)
	}
}
