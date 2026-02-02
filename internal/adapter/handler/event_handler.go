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
