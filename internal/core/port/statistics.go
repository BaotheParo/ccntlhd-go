package port

import (
	"context"

	"github.com/yourname/ticketing-system/internal/core/entity"
)

type StatisticsRepositoryPort interface {
	GetEventStatistics(ctx context.Context) (entity.EventStatistics, error)
}

type StatisticsServicePort interface {
	GetEventStatistics(ctx context.Context) (entity.EventStatistics, error)
}
