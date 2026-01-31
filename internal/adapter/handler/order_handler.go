package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/yourname/ticketing-system/internal/core/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type CreateOrderRequest struct {
	Items []struct {
		TicketTypeID string `json:"ticket_type_id"`
		Quantity     int    `json:"quantity"`
	} `json:"items"`
}

func (h *OrderHandler) PlaceOrder(c *fiber.Ctx) error {
	// parse va validate
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(req.Items) == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Items cannot be empty"})
	}

	// Lay UserID tu AuthMiddleware
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	// Du lieu tu client
	var serviceItems []service.RequestItem
	for _, item := range req.Items {
		ticketID, err := uuid.Parse(item.TicketTypeID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid ticket_type_id: %s", item.TicketTypeID),
			})
		}
		if item.Quantity <= 0 {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Quantity must be greater than 0 for ticket: %s", item.TicketTypeID),
			})
		}

		serviceItems = append(serviceItems, service.RequestItem{
			TicketTypeID: ticketID,
			Quantity:     item.Quantity,
		})
	}

	// goi Service
	order, err := h.svc.PlaceOrder(c.Context(), userID, serviceItems)
	if err != nil {
		// Goi Service bi loi
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Tra ve response
	return c.Status(http.StatusCreated).JSON(order)
}
