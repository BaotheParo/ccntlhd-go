package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type OrderHandler struct {
	orderService port.OrderServicePort
}

func NewOrderHandler(orderService port.OrderServicePort) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

type CreateOrderRequest struct {
	Items []struct {
		TicketTypeID string `json:"ticket_type_id"`
		Quantity     int    `json:"quantity"`
		UnitPrice    string `json:"unit_price"`
	} `json:"items"`
}

// PlaceOrder godoc
// @Summary Place a new order
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body CreateOrderRequest true "Order Payload"
// @Success 202 {object} entity.Order
// @Router /api/v1/orders [post]
func (h *OrderHandler) PlaceOrder(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user_id from JWT middleware
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "user_id not found in context",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id",
		})
	}

	// Parse request body
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Convert to entity.OrderItem
	var items []entity.OrderItem
	for _, item := range req.Items {
		ticketTypeID, err := uuid.Parse(item.TicketTypeID)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid ticket_type_id",
			})
		}

		unitPrice, err := decimal.NewFromString(item.UnitPrice)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid unit_price",
			})
		}

		items = append(items, entity.OrderItem{
			TicketTypeID: ticketTypeID,
			Quantity:     item.Quantity,
			UnitPrice:    unitPrice,
		})
	}

	// Call service
	order, err := h.orderService.PlaceOrder(ctx, userID, items)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusAccepted).JSON(order)
}

// GetOrder godoc
// @Summary Get order by ID
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} entity.Order
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order id",
		})
	}

	order, err := h.orderService.GetOrder(ctx, id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "order not found",
		})
	}

	return c.Status(http.StatusOK).JSON(order)
}

// GetUserOrders godoc
// @Summary Get user's orders
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} entity.Order
// @Router /api/v1/orders [get]
// @Router /api/v1/users/me/orders [get]
func (h *OrderHandler) GetUserOrders(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get user_id from JWT middleware
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "user_id not found in context",
		})
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user_id",
		})
	}

	// Get query params
	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	orders, err := h.orderService.GetUserOrders(ctx, userID, limit, offset)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(orders)
}

// CancelOrder godoc
// @Summary Cancel an order
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} map[string]string
// @Router /api/v1/orders/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid order id",
		})
	}

	err = h.orderService.CancelOrder(ctx, id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "order cancelled successfully",
	})
}

// ListAdminOrders godoc
// @Summary Get all orders for admin
// @Tags Order (Admin)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param status query string false "Order Status"
// @Param event_id query string false "Event ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/orders [get]
func (h *OrderHandler) ListAdminOrders(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	status := c.Query("status", "")
	eventID := c.Query("event_id", "")

	orders, total, err := h.orderService.ListAdminOrders(ctx, page, limit, status, eventID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  orders,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// UpdateOrderStatus godoc
// @Summary Update Order Status (Admin)
// @Tags Order (Admin)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param payload body entity.UpdateOrderStatusRequest true "Order Status Payload"
// @Success 200 {object} map[string]string
// @Router /api/v1/admin/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	orderIDStr := c.Params("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID đơn hàng không hợp lệ"})
	}

	var req entity.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Dữ liệu trạng thái không hợp lệ"})
	}

	err = h.orderService.UpdateOrderStatusAdmin(ctx, orderID, req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Cập nhật trạng thái đơn hàng thành công",
	})
}

