package integration

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/adapter/repository"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/service"
	"testing"
	"time"
)

func TestOrderAdminStatusUpdateAndRollback(t *testing.T) {
	db := setupDB()

	// Clean up schema explicitly
	db.Migrator().DropTable(&entity.OrderItem{}, &entity.Order{}, &entity.TicketType{}, &entity.Event{}, &entity.User{})
	_ = db.AutoMigrate(&entity.User{}, &entity.Event{}, &entity.TicketType{}, &entity.Order{}, &entity.OrderItem{})

	// 1. Create User
	userID := uuid.New()
	email := fmt.Sprintf("admin-%d@example.com", time.Now().UnixNano())
	username := fmt.Sprintf("admintest-%d", time.Now().UnixNano())
	testUser := entity.User{
		ID:           userID,
		Username:     username,
		Email:        email,
		PasswordHash: "hash",
		Role:         entity.RoleAdmin,
	}
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}

	// 2. Create Event
	eventID := uuid.New()
	eventSlug := fmt.Sprintf("admin-test-event-%d", time.Now().UnixNano())
	testEvent := entity.Event{
		ID:        eventID,
		Name:      "Admin Test Event",
		Slug:      eventSlug,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(1 * time.Hour),
		Status:    entity.EventStatusDraft,
	}
	if err := db.Create(&testEvent).Error; err != nil {
		t.Fatalf("Failed to seed event: %v", err)
	}

	// 3. Create TicketType
	ticketID := uuid.New()
	initialStock := 10
	price := decimal.NewFromFloat(500.00)

	ticket := entity.TicketType{
		ID:                ticketID,
		EventID:           eventID,
		Name:              "General Pass",
		Price:             price,
		InitialQuantity:   initialStock,
		RemainingQuantity: initialStock,
	}

	if err := db.Create(&ticket).Error; err != nil {
		t.Fatalf("Failed to seed ticket: %v", err)
	}

	// Setup Redis
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6380"})
	cacheRepo := repository.NewRedisTicketRepository(rdb)

	ctx := context.Background()
	rdb.FlushDB(ctx)
	if err := cacheRepo.SetStock(ctx, ticketID, initialStock); err != nil {
		t.Fatalf("Failed to seed redis stock: %v", err)
	}

	// Initialize Services
	orderRepo := repository.NewOrderRepository(db)
	eventRepo := repository.NewEventRepository(db)
	orderSvc := service.NewOrderService(orderRepo, eventRepo, cacheRepo, 100, 2)

	// ============================================
	// BƯỚC 1: PLACE ORDER (Mua 2 vé)
	// ============================================
	items := []entity.OrderItem{
		{TicketTypeID: ticketID, Quantity: 2, UnitPrice: price},
	}
	order, err := orderSvc.PlaceOrder(ctx, userID, items)
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}

	// Dừng hệ thống Worker 1 giây để worker kịp xử lý DB insert
	time.Sleep(1 * time.Second)

	// Kiểm tra Redis Stock phải còn 8
	stockStr, err := rdb.Get(ctx, fmt.Sprintf("ticket:%s:stock", ticketID.String())).Result()
	if err != nil || stockStr != "8" {
		t.Fatalf("Expected Redis stock to be 8, got %s. Err: %v", stockStr, err)
	}
	fmt.Printf("✅ Mua 2 vé thành công. Kho chứa Redis còn: %s vé\n", stockStr)

	// ============================================
	// BƯỚC 2: ADMIN UPDATE ORDER SANG PAID
	// ============================================
	err = orderSvc.UpdateOrderStatusAdmin(ctx, order.ID, string(entity.OrderStatusPaid))
	if err != nil {
		t.Fatalf("Admin could not update to PAID: %v", err)
	}

	// Check DB
	paidOrder, _ := orderRepo.GetOrderByID(ctx, order.ID)
	if paidOrder.Status != entity.OrderStatusPaid {
		t.Fatalf("Expected status PAID, got %s", paidOrder.Status)
	}
	fmt.Println("✅ Admin duyệt đơn sang PAID thành công!")

	// ============================================
	// BƯỚC 3: ADMIN UPDATE ORDER TỪ PAID -> CANCELLED (Nên Fail do đã chốt)
	// ============================================
	err = orderSvc.UpdateOrderStatusAdmin(ctx, order.ID, string(entity.OrderStatusCancelled))
	if err == nil {
		// Nó không fail nghĩa là lọt qua kiểm tra
		t.Fatalf("Expected error when cancelling a PAID order, but got no error")
	}
	fmt.Printf("✅ Tính năng chặn Hủy đơn đã chốt hoạt động tốt: %v\n", err)

	// ============================================
	// BƯỚC 4: TẠO ĐƠN KHÁC (Mua 3 vé) RỒI CANCEL
	// ============================================
	items2 := []entity.OrderItem{
		{TicketTypeID: ticketID, Quantity: 3, UnitPrice: price},
	}
	order2, _ := orderSvc.PlaceOrder(ctx, userID, items2)
	time.Sleep(1 * time.Second)

	stock2Str, _ := rdb.Get(ctx, fmt.Sprintf("ticket:%s:stock", ticketID.String())).Result()
	fmt.Printf("✅ Mua tiếp 3 vé. Kho chứa Redis còn: %s vé\n", stock2Str) // Should be 8 - 3 = 5

	// Admin Hủy Đơn 2
	err = orderSvc.UpdateOrderStatusAdmin(ctx, order2.ID, string(entity.OrderStatusCancelled))
	if err != nil {
		t.Fatalf("Failed to cancel order 2: %v", err)
	}

	finalStockStr, _ := rdb.Get(ctx, fmt.Sprintf("ticket:%s:stock", ticketID.String())).Result()
	if finalStockStr != "8" {
		// 5 + 3 = 8
		t.Fatalf("Expected Redis stock to rollback to 8, but got %s", finalStockStr)
	}
	fmt.Printf("✅ Đơn hàng số 2 đã bị Hủy. Vé được trả về Redis thành công! Kho chứa hiện tại là: %s vé\n", finalStockStr)

	orderSvc.Shutdown()
}
