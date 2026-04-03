package repository

import (
	"context"

	"github.com/yourname/ticketing-system/internal/core/entity"
	"github.com/yourname/ticketing-system/internal/core/port"
	"gorm.io/gorm"
)

type statisticsRepository struct {
	db *gorm.DB
}

func NewStatisticRepository(db *gorm.DB) port.StatisticsRepositoryPort {
	return &statisticsRepository{db: db}
}

func (r *statisticsRepository) GetEventStatistics(ctx context.Context) (entity.EventStatistics, error) {
	var stats entity.EventStatistics

	rows, err := r.db.WithContext(ctx).
		Model(&entity.Event{}).
		Select("status, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("status").
		Rows()
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var row struct {
			Status string
			Count  int64
		}

		if err := r.db.ScanRows(rows, &row); err != nil {
			return stats, err
		}

		switch row.Status {
		case string(entity.EventStatusDraft):
			stats.Draft = row.Count
		case string(entity.EventStatusPublished):
			stats.Published = row.Count
		case string(entity.EventStatusCancelled):
			stats.Cancelled = row.Count
		case string(entity.EventStatusEnded):
			stats.Ended = row.Count
		}
	}

	stats.Total = stats.Draft + stats.Published + stats.Cancelled + stats.Ended

	return stats, nil
}

// GetAllPaidOrders lấy toàn bộ đơn hàng đã thanh toán (PAID) kèm theo chi tiết vé (Items)
func (r *statisticsRepository) GetAllPaidOrders(ctx context.Context) ([]entity.Order, error) {
	var orders []entity.Order

	// Chỉ lấy đơn hàng có trạng thái PAID
	// Preload("Items") để lấy danh sách vé mua bên trong từng đơn hàng
	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", "PAID").
		Find(&orders).Error

	return orders, err
}
