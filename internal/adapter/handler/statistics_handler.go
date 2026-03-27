package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type StatisticsHandler struct {
	service port.StatisticsServicePort
}

func NewStatisticsHandler(service port.StatisticsServicePort) *StatisticsHandler {
	return &StatisticsHandler{service: service}
}

func (h *StatisticsHandler) GetEventStatistics(c *fiber.Ctx) error {
	data, err := h.service.GetEventStatistics(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Lỗi khi lấy thống kê",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}
