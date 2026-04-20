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
	cacheRepo port.TicketCacheRepository
}

func NewEventService(eventRepo port.EventRepositoryPort, cacheRepo port.TicketCacheRepository) port.EventServicePort {
	return &eventService{
		eventRepo: eventRepo,
		cacheRepo: cacheRepo,
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

func (s *eventService) ListEventsAdvanced(ctx context.Context, req entity.ListEventRequest) ([]entity.Event, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	offset := (req.Page - 1) * req.Limit //offset: bỏ bao nhiêu event trước khi lấy

	var fromtime *time.Time
	var totime *time.Time
	if req.FromTime != "" {
		if t, err := time.Parse(time.RFC3339, req.FromTime); err == nil {
			fromtime = &t
		}
	}
	if req.ToTime != "" {
		if t, err := time.Parse(time.RFC3339, req.ToTime); err == nil {
			totime = &t
		}
	}

	var status entity.EventStatus
	switch req.Status {
	case string(entity.EventStatusDraft),
		string(entity.EventStatusPublished),
		string(entity.EventStatusCancelled),
		string(entity.EventStatusEnded):
		status = entity.EventStatus(req.Status)
	default:
		status = ""
	}

	filter := entity.EventFilter{
		Limit:    req.Limit,
		Offset:   offset,
		Search:   req.Search,
		Status:   status,
		FromTime: fromtime,
		ToTime:   totime,
	}

	return s.eventRepo.ListEventsAdvanced(ctx, filter)
}

func (s *eventService) UpdateEvent(ctx context.Context, id uuid.UUID, req entity.UpdateEventRequest) (*entity.Event, error) {
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return nil, errors.New("event không tồn tại")
	}

	if req.Name != "" {
		event.Name = req.Name
	}
	if req.Location != "" {
		event.Location = req.Location
	}
	if !req.StartTime.IsZero() {
		event.StartTime = req.StartTime
	}
	if !req.EndTime.IsZero() {
		event.EndTime = req.EndTime
	}
	if req.Status != "" {
		event.Status = entity.EventStatus(req.Status)
	}
	event.UpdatedAt = time.Now()

	if err := s.eventRepo.UpdateEvent(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *eventService) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return errors.New("Không tìm thấy id này")
	}
	return s.eventRepo.DeleteEvent(ctx, id)
}

func (s *eventService) CreateTicketType(ctx context.Context, eventID uuid.UUID, req entity.CreateTicketTypeRequest) (*entity.TicketType, error) {
	if err := validateCreateTicketTypeRequest(req); err != nil {
		return nil, err
	}

	// 1. Kiểm tra sự kiện có tồn tại không
	_, err := s.eventRepo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, errors.New("sự kiện không tồn tại")
	}

	// 2. Tạo đối tượng TicketType mới
	ticketType := &entity.TicketType{
		ID:                uuid.New(),
		EventID:           eventID,
		Name:              req.Name,
		Price:             req.Price,
		InitialQuantity:   req.InitialQuantity,
		RemainingQuantity: req.InitialQuantity,
	}

	// 3. Lưu vào Database (PostgreSQL)
	if err := s.eventRepo.CreateTicketType(ctx, ticketType); err != nil {
		return nil, errors.New("lỗi khi lưu hạng vé vào cơ sở dữ liệu")
	}

	// 4. Đồng bộ số lượng vé lên Redis để phục vụ mua vé nhanh (Flash Sale)
	if s.cacheRepo != nil {
		err = s.cacheRepo.SetStock(ctx, ticketType.ID, ticketType.InitialQuantity)
		if err != nil {
			// Lưu ý: Lỗi Redis ở đây không làm hỏng dữ liệu PostgreSQL, nhưng nên cảnh báo
			return ticketType, errors.New("tạo vé thành công nhưng lỗi đẩy lên Redis: " + err.Error())
		}
	}

	return ticketType, nil
}

func (s *eventService) UpdateTicketType(ctx context.Context, ticketID uuid.UUID, req entity.UpdateTicketTypeRequest) (*entity.TicketType, error) {
	// B1: Lấy thông tin hạng vé cũ từ Database ra
	ticketType, err := s.eventRepo.GetTicketTypeByID(ctx, ticketID)
	if err != nil {
		return nil, errors.New("hạng vé không tồn tại")
	}

	// B2: Cập nhật Giá vé nếu có truyền lên (lớn hơn 0)
	if req.Price.GreaterThan(req.Price.Sub(req.Price)) { // Cách an toàn kiểm tra lớn hơn 0
		ticketType.Price = req.Price
	}

	// Xử lý Cập nhật Số lượng vé
	if req.Quantity > 0 {
		oldQuantity := ticketType.InitialQuantity

		// Luật nghiệp vụ: Không được phép giảm số lượng vé đã phát hành
		if req.Quantity < oldQuantity {
			return nil, errors.New("chỉ được phép tăng số lượng vé, không được giảm")
		}

		// Nếu số lượng mới lớn hơn số lượng cũ -> Xử lý tăng thêm
		if req.Quantity > oldQuantity {
			// B3: Tính phần số lượng vé chênh lệch (tăng thêm)
			diff := req.Quantity - oldQuantity

			// B4: Cập nhật lại số liệu vào biến để chuẩn bị lưu DB
			ticketType.InitialQuantity = req.Quantity
			ticketType.RemainingQuantity = ticketType.RemainingQuantity + diff

			// Cập nhật Database
			if err := s.eventRepo.UpdateTicketType(ctx, ticketType); err != nil {
				return nil, errors.New("lỗi khi cập nhật hạng vé vào cơ sở dữ liệu")
			}

			// B5: Cập nhật Redis: Gọi SetStock hoặc hàm tăng. Ở đây để an toàn ta sẽ tính số lượng cộng thêm
			// hoặc gọi SetStock đè lại số RemainingQuantity mới nhất.
			// Tuy nhỏ, nhưng nếu ta dùng cách set mới thì an toàn hơn cho sinh viên dễ hiểu:
			if s.cacheRepo != nil {
				// Mẹo an toàn: Ta đẩy số tồn kho hiện tại lên lại Redis (Đây là cách xử lý đơn giản nhất)
				err = s.cacheRepo.SetStock(ctx, ticketType.ID, ticketType.RemainingQuantity)
				if err != nil {
					return ticketType, errors.New("cập nhật vé thành công nhưng lỗi đồng bộ Redis: " + err.Error())
				}
			}
		} else {
			// Chỉ đổi giá, không đổi số lượng -> Lưu lại bình thường
			if err := s.eventRepo.UpdateTicketType(ctx, ticketType); err != nil {
				return nil, errors.New("lỗi khi cập nhật giá vé")
			}
		}
	} else {
		// Chỉ đổi giá, không đổi số lượng (Quantity = 0)
		if err := s.eventRepo.UpdateTicketType(ctx, ticketType); err != nil {
			return nil, errors.New("lỗi khi cập nhật giá vé")
		}
	}

	return ticketType, nil
}

func (s *eventService) GetAllTicketTypeByIDEvent(ctx context.Context, id uuid.UUID) ([]entity.TicketType, error) {
	return s.eventRepo.GetAllTicketTypeByIDEvent(ctx, id)
}
