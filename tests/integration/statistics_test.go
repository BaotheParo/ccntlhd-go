package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
)

func TestDashboardStatistics(t *testing.T) {
	db := setupDB()

	// 1. Dọn dẹp dữ liệu cũ (Dùng cho môi trường Test)
	db.Exec("DELETE FROM order_items")
	db.Exec("DELETE FROM orders")
	db.Exec("DELETE FROM ticket_types")
	db.Exec("DELETE FROM events")

	statsRepo := repository.NewStatisticRepository(db)
	statsSvc := service.NewStatisticsService(statsRepo)

	// 2. Tạo dữ liệu mẫu: 1 Event, 1 TicketType
	eventID := uuid.New()
	event := entity.Event{
		ID:        eventID,
		Name:      "Test Statistics Event",
		Slug:      fmt.Sprintf("test-stats-%d", time.Now().UnixNano()),
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Status:    entity.EventStatusPublished,
	}
	db.Create(&event)

	ticketID := uuid.New()
	ticket := entity.TicketType{
		ID:                ticketID,
		EventID:           eventID,
		Name:              "VVIP",
		Price:             decimal.NewFromFloat(100.0),
		InitialQuantity:   100,
		RemainingQuantity: 100,
	}
	db.Create(&ticket)

	// 3. Tạo 2 đơn hàng PAID
	// Đơn 1: Mua 2 vé VVIP
	order1ID := uuid.New()
	order1 := entity.Order{
		ID:          order1ID,
		UserID:      uuid.New(), // User ảo
		Status:      entity.OrderStatusPaid,
		TotalAmount: decimal.NewFromFloat(200.0),
		CreatedAt:   time.Now(),
	}
	db.Create(&order1)
	db.Create(&entity.OrderItem{
		ID:           uuid.New(),
		OrderID:      order1ID,
		TicketTypeID: ticketID,
		Quantity:     2,
		UnitPrice:     decimal.NewFromFloat(100.0),
	})

	// Đơn 2: Mua 3 vé VVIP
	order2ID := uuid.New()
	order2 := entity.Order{
		ID:          order2ID,
		UserID:      uuid.New(),
		Status:      entity.OrderStatusPaid,
		TotalAmount: decimal.NewFromFloat(300.0),
		CreatedAt:   time.Now(),
	}
	db.Create(&order2)
	db.Create(&entity.OrderItem{
		ID:           uuid.New(),
		OrderID:      order2ID,
		TicketTypeID: ticketID,
		Quantity:     3,
		UnitPrice:     decimal.NewFromFloat(100.0),
	})

	// Đơn 3: Đơn CANCELLED (Không tệ vào thống kê)
	order3ID := uuid.New()
	order3 := entity.Order{
		ID:          order3ID,
		UserID:      uuid.New(),
		Status:      entity.OrderStatusCancelled,
		TotalAmount: decimal.NewFromFloat(100.0),
		CreatedAt:   time.Now(),
	}
	db.Create(&order3)

	// 4. Gọi Service tính toán thống kê
	ctx := context.Background()
	stats, err := statsSvc.GetDashboardStats(ctx)
	if err != nil {
		t.Fatalf("Failed to get dashboard stats: %v", err)
	}

	// 5. Kiểm tra kết quả
	// Mong đợi: 2 đơn PAID, 500.0 doanh thu, 5 vé bán ra.
	if stats.TotalOrders != 2 {
		t.Errorf("Expected 2 orders, got %d", stats.TotalOrders)
	}

	if stats.TotalRevenue != 500.0 {
		t.Errorf("Expected 500.0 revenue, got %f", stats.TotalRevenue)
	}

	if stats.TotalTicketsSold != 5 {
		t.Errorf("Expected 5 total tickets, got %d", stats.TotalTicketsSold)
	}

	fmt.Printf("✅ Test Dashboard Statistics Pass: Orders=%d, Revenue=%f, Tickets=%d\n", 
		stats.TotalOrders, stats.TotalRevenue, stats.TotalTicketsSold)
}
