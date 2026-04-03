package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type EventHandler struct {
	svc port.EventServicePort
}

func NewEventHandler(svc port.EventServicePort) *EventHandler {
	return &EventHandler{svc: svc}
}

// CreateEvent godoc
// @Summary Create a new event
// @Tags Event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body entity.CreateEventRequest true "Event Payload"
// @Success 201 {object} entity.Event
// @Router /api/v1/events [post]
func (h *EventHandler) CreateEvent(c *fiber.Ctx) error {
	var req entity.CreateEventRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dữ liệu không hợp lệ",
		})
	}

	event, err := h.svc.CreateEvent(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(event)
}

// CreateEventWithTickets godoc
// @Summary Create a new event with ticket types
// @Tags Event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body entity.CreateEventWithTicketsRequest true "Event with Tickets Payload"
// @Success 201 {object} map[string]interface{}
// @Router /api/v1/events/with-tickets [post]
func (h *EventHandler) CreateEventWithTickets(c *fiber.Ctx) error {
	var req entity.CreateEventWithTicketsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dữ liệu không hợp lệ",
		})
	}

	if req.Event.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Dữ liệu không hợp lệ",
		})
	}

	event, err := h.svc.CreateEventWithTickets(c.Context(), req.Event, req.TicketTypes)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "event created successfully",
		"data":    event,
	})
}

// GetEvent godoc
// @Summary Get event by ID
// @Tags Event
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} entity.Event
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse UUID
	eventID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID không hợp lệ",
		})
	}

	event, err := h.svc.GetEvent(c.Context(), eventID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Sự kiện không tìm thấy",
		})
	}

	return c.JSON(event)
}

// GetEventBySlug godoc
// @Summary Get event by slug
// @Tags Event
// @Accept json
// @Produce json
// @Param slug path string true "Event Slug"
// @Success 200 {object} entity.Event
// @Router /api/v1/events/slug/{slug} [get]
func (h *EventHandler) GetEventBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")

	event, err := h.svc.GetEventBySlug(c.Context(), slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Sự kiện không tìm thấy",
		})
	}

	return c.JSON(event)
}

// ListEvents godoc
// @Summary List all events
// @Tags Event
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	offset := c.QueryInt("offset", 0)

	events, err := h.svc.ListEvents(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data":   events,
		"limit":  limit,
		"offset": offset,
	})
}

// ListEventsAdvanced godoc
// @Summary Search events with filters
// @Tags Event
// @Accept json
// @Produce json
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Param search query string false "Search query"
// @Param status query string false "Status filter"
// @Param from_time query string false "From time"
// @Param to_time query string false "To time"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/events/search [get]
func (h *EventHandler) ListEventsAdvanced(c *fiber.Ctx) error {
	var req entity.ListEventRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query không hợp lệ",
		})
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	events, total, err := h.svc.ListEventsAdvanced(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON((fiber.Map{
			"error": err.Error(),
		}))
	}
	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return c.JSON(fiber.Map{
		"data": events,
		"pagination": fiber.Map{
			"page":        req.Page,
			"limit":       req.Limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// UpdateEvent godoc
// @Summary Update an existing event
// @Tags Event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param payload body entity.UpdateEventRequest true "Update Payload"
// @Success 200 {object} entity.Event
// @Router /api/v1/events/{id} [put]
func (h *EventHandler) UpdateEvent(c *fiber.Ctx) error {
	id := c.Params("id")

	eventID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID không hợp lệ"})
	}

	var req entity.UpdateEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Dữ liệu không hợp lệ"})
	}

	event, err := h.svc.UpdateEvent(c.Context(), eventID, req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(event)
}

// DeleteEvent godoc
// @Summary Delete an event
// @Tags Event
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/events/{id} [delete]
func (h *EventHandler) DeleteEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID không hợp lệ"})
	}

	if err := h.svc.DeleteEvent(c.Context(), eventID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Xóa thành công"})
}

// CreateTicketType godoc
// @Summary Add a new ticket type to an event
// @Tags Event (Admin)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Param payload body entity.CreateTicketTypeRequest true "Ticket Type Payload"
// @Success 201 {object} entity.TicketType
// @Router /api/v1/admin/events/{id}/ticket-types [post]
func (h *EventHandler) CreateTicketType(c *fiber.Ctx) error {
	id := c.Params("id")
	eventID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID sự kiện không hợp lệ"})
	}

	var req entity.CreateTicketTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dữ liệu vé gửi lên không hợp lệ"})
	}

	ticketType, err := h.svc.CreateTicketType(c.Context(), eventID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(ticketType)
}

// UpdateTicketType godoc
// @Summary Update an existing ticket type (Price, Quantity)
// @Tags Event (Admin)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ticket_id path string true "Ticket Type ID"
// @Param payload body entity.UpdateTicketTypeRequest true "Update Ticket Type Payload"
// @Success 200 {object} entity.TicketType
// @Router /api/v1/admin/ticket-types/{ticket_id} [put]
func (h *EventHandler) UpdateTicketType(c *fiber.Ctx) error {
	ticketIDStr := c.Params("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID hạng vé không hợp lệ"})
	}

	var req entity.UpdateTicketTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dữ liệu cập nhật vé không hợp lệ"})
	}

	ticketType, err := h.svc.UpdateTicketType(c.Context(), ticketID, req)
	if err != nil {
		// Báo lỗi 400 rõ ràng khi sinh viên test thử giảm số lượng hoặc cập nhật lỗi
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(ticketType)
}

