package port

import (
	"context"

	"github.com/yourname/ticketing-system/internal/core/entity"
)

type StatisticsRepositoryPort interface {
	GetEventStatistics(ctx context.Context) (entity.EventStatistics, error)
	GetAllPaidOrders(ctx context.Context) ([]entity.Order, error)
}

type StatisticsServicePort interface {
	GetEventStatistics(ctx context.Context) (entity.EventStatistics, error)
	GetDashboardStats(ctx context.Context) (entity.DashboardStatsResponse, error)
}
