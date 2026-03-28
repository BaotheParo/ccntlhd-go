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
