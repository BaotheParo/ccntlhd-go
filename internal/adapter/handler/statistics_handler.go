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

// GetEventStatistics godoc
// @Summary Get event statistics (admin only)
// @Tags Statistics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]entity.EventStatistics
// @Router /api/v1/admin/statistics/events [get]
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

// GetDashboardStats godoc
// @Summary Xem các con số thống kê Dashboard chính (Sử dụng vòng lặp Go cơ bản)
// @Tags Admin Statistics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} entity.DashboardStatsResponse
// @Router /api/v1/admin/statistics/dashboard [get]
func (h *StatisticsHandler) GetDashboardStats(c *fiber.Ctx) error {
	// Gọi xuống tầng Service để tính toán (Service sẽ dùng vòng lặp for cơ bản)
	data, err := h.service.GetDashboardStats(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": "Không thể lấy số liệu thống kê Dashboard",
			"error":   err.Error(),
		})
	}

	return c.JSON(data)
}
