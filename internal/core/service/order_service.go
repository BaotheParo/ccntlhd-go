package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

// OrderPayload chứa thông tin cần thiết để xử lý giao dịch mua vé ở background
type OrderPayload struct {
	OrderID uuid.UUID
	UserID  uuid.UUID
	Items   []entity.OrderItem
}

type OrderService struct {
	repo       port.OrderRepositoryPort
	eventRepo  port.EventRepositoryPort
	cacheRepo  port.TicketCacheRepository
	orderQueue chan OrderPayload
	wg         *sync.WaitGroup
}

// NewOrderService tạo service mới, khởi tạo worker pool
func NewOrderService(repo port.OrderRepositoryPort, eventRepo port.EventRepositoryPort, cacheRepo port.TicketCacheRepository, queueSize int, numWorkers int) *OrderService {
	s := &OrderService{
		repo:       repo,
		eventRepo:  eventRepo,
		cacheRepo:  cacheRepo,
		orderQueue: make(chan OrderPayload, queueSize),
		wg:         &sync.WaitGroup{},
	}
	s.StartWorkers(numWorkers)
	return s
}

// StartWorkers khởi tạo các worker goroutines để xử lý order queue
func (s *OrderService) StartWorkers(numWorkers int) {
	for i := 1; i <= numWorkers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
	log.Printf("🚀 Khởi chạy %d workers xử lý đơn hàng", numWorkers)
}

func (s *OrderService) worker(id int) {
	defer s.wg.Done()
	log.Printf("Worker %d bắt đầu chạy", id)
	for payload := range s.orderQueue {
		s.processOrderInDB(payload)
	}
	log.Printf("Worker %d đã dừng", id)
}

func (s *OrderService) processOrderInDB(payload OrderPayload) {
	ctx := context.Background()

	// Tính tổng tiền (tái tạo lại cho Entity Order)
	totalAmount := decimal.Zero
	for _, item := range payload.Items {
		itemTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)
	}

	order := &entity.Order{
		ID:          payload.OrderID,
		UserID:      payload.UserID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       payload.Items,
	}

	// Gọi hàm lưu Order có Transaction
	err := s.repo.CreateOrderWithTransaction(ctx, order, payload.Items)
	if err != nil {
		log.Printf("lưu đơn hàng %s vào DB: %v", payload.OrderID, err)
		// Hoàn lại vé vào Redis
		for _, item := range payload.Items {
			rollbackErr := s.cacheRepo.RollbackStock(ctx, item.TicketTypeID, item.Quantity)
			if rollbackErr != nil {
				log.Printf(" LỖI NGHIÊM TRỌNG: Lỗi rollback Redis cho vé %s: %v", item.TicketTypeID, rollbackErr)
			}
		}
		// Cập nhật trạng thái đơn hàng thành CANCELLED hoặc FAILED nếu hỗ trợ
		rErr := s.repo.UpdateOrderStatus(ctx, payload.OrderID, entity.OrderStatusCancelled)
		if rErr != nil {
			log.Printf("Cảnh báo: Không thể cập nhật trạng thái lỗi cho đơn hàng %s", payload.OrderID)
		}
	} else {
		log.Printf("Đơn hàng %s tạo thành công", payload.OrderID)
	}
}

// PlaceOrder tiếp nhận đơn hàng, trừ khóa Redis và đẩy vào queue (HTTP Response trả về ngay)
func (s *OrderService) PlaceOrder(ctx context.Context, userID uuid.UUID, items []entity.OrderItem) (*entity.Order, error) {
	// Validate input
	if len(items) == 0 {
		return nil, errors.New("đơn hàng phải có ít nhất một item")
	}

	totalAmount := decimal.Zero
	for _, item := range items {
		if item.TicketTypeID == uuid.Nil {
			return nil, errors.New("ticket_type_id không được để trống")
		}
		if item.Quantity <= 0 {
			return nil, errors.New("số lượng vé phải lớn hơn 0")
		}
		if item.UnitPrice.IsNegative() || item.UnitPrice.IsZero() {
			return nil, errors.New("giá vé phải lớn hơn 0")
		}

		itemTotal := item.UnitPrice.Mul(decimal.NewFromInt(int64(item.Quantity)))
		totalAmount = totalAmount.Add(itemTotal)
	}

	// 1. Phase 1 (Redis): Trừ tồn kho trên Redis trước (Atomic)
	for _, item := range items {
		_, err := s.cacheRepo.DeductStock(ctx, item.TicketTypeID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("không thể giữ vé (loại %s): %w", item.TicketTypeID, err)
		}
	}

	// 2. Prepare Payload
	orderID := uuid.New()
	for i := range items {
		items[i].ID = uuid.New()
		items[i].OrderID = orderID
	}

	payload := OrderPayload{
		OrderID: orderID,
		UserID:  userID,
		Items:   items,
	}

	// 3. Đẩy vào Job Queue
	select {
	case s.orderQueue <- payload:
		// Queue nhận job thành công
	default:
		// Hàng đợi đầy -> Quá tải hệ thống -> Rollback Redis
		for _, item := range items {
			_ = s.cacheRepo.RollbackStock(ctx, item.TicketTypeID, item.Quantity)
		}
		return nil, errors.New("hệ thống đang quá tải, vui lòng thử lại sau")
	}

	// 4. Trả về kết quả ngay lập tức cho client
	order := &entity.Order{
		ID:          orderID,
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      entity.OrderStatusPending, // Báo cho client là đơn đang được xử lý
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       items,
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	return s.repo.GetOrderByID(ctx, id)
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID uuid.UUID, limit int, offset int) ([]entity.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetOrdersByUserID(ctx, userID, limit, offset)
}

func (s *OrderService) CancelOrder(ctx context.Context, id uuid.UUID) error {
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return err
	}

	if order.Status != entity.OrderStatusPending {
		return errors.New("chỉ có thể hủy đơn hàng ở trạng thái PENDING")
	}

	return s.repo.UpdateOrderStatus(ctx, id, entity.OrderStatusCancelled)
}

// Shutdown đóng queue và đợi tất cả worker hoàn thành công việc đang dở dang
func (s *OrderService) Shutdown() error {
	log.Println("Đang dừng xử lý đơn hàng (Đợi các worker hoàn thành)...")
	close(s.orderQueue)
	s.wg.Wait()
	log.Println("Tất cả worker đã dừng hoàn toàn")
	return nil
}
