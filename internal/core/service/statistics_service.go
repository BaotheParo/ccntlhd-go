package service

import (
	"context"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
)

type statisticsService struct {
	repo port.StatisticsRepositoryPort
}

func NewStatisticsService(repo port.StatisticsRepositoryPort) port.StatisticsServicePort {
	return &statisticsService{repo: repo}
}

func (s *statisticsService) GetEventStatistics(ctx context.Context) (entity.EventStatistics, error) {
	return s.repo.GetEventStatistics(ctx)
}

// GetDashboardStats tính toán các con số tổng quát bằng cách dùng vòng lặp Go cơ bản
// Cách làm này giúp sinh viên dễ hiểu, dễ giải thích hơn so với dùng Query SQL phức tạp.
func (s *statisticsService) GetDashboardStats(ctx context.Context) (entity.DashboardStatsResponse, error) {
	// B1: Lấy toàn bộ đơn hàng đã thanh toán từ Repository
	orders, err := s.repo.GetAllPaidOrders(ctx)
	if err != nil {
		return entity.DashboardStatsResponse{}, err
	}

	// B2: Khởi tạo các biến tích lũy
	var totalOrders int
	var totalRevenue float64
	var totalTickets int

	// B3: Dùng vòng lặp for cơ bản để duyệt qua danh sách đơn hàng
	for _, order := range orders {
		// Tăng số lượng đơn hàng
		totalOrders++

		// Cộng dồn doanh thu của đơn hàng này
		// Chuyển decimal sang float64 để tính toán đơn giản
		revenue, _ := order.TotalAmount.Float64()
		totalRevenue += revenue

		// B4: Duyệt tiếp các vé (Items) trong đơn hàng này để cộng tổng số vé
		for _, item := range order.Items {
			totalTickets += item.Quantity
		}
	}

	// B5: Trả về kết quả cuối cùng
	return entity.DashboardStatsResponse{
		TotalOrders:      totalOrders,
		TotalRevenue:     totalRevenue,
		TotalTicketsSold: totalTickets,
	}, nil
}
