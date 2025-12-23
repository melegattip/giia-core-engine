package repositories

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BufferHistoryRepository struct {
	db *gorm.DB
}

func NewBufferHistoryRepository(db *gorm.DB) *BufferHistoryRepository {
	return &BufferHistoryRepository{db: db}
}

func (r *BufferHistoryRepository) Create(ctx context.Context, history *domain.BufferHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *BufferHistoryRepository) GetByBufferAndDate(ctx context.Context, bufferID uuid.UUID, date time.Time) (*domain.BufferHistory, error) {
	var history domain.BufferHistory
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("buffer_id = ? AND snapshot_date = ?", bufferID, dateOnly).
		First(&history).Error

	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *BufferHistoryRepository) ListByBuffer(ctx context.Context, bufferID uuid.UUID, limit int) ([]domain.BufferHistory, error) {
	var histories []domain.BufferHistory
	query := r.db.WithContext(ctx).Where("buffer_id = ?", bufferID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("snapshot_date DESC").Find(&histories).Error
	return histories, err
}

func (r *BufferHistoryRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID, startDate, endDate time.Time) ([]domain.BufferHistory, error) {
	var histories []domain.BufferHistory

	startOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ? AND snapshot_date >= ? AND snapshot_date <= ?",
			productID, organizationID, startOnly, endOnly).
		Order("snapshot_date DESC").
		Find(&histories).Error

	return histories, err
}

func (r *BufferHistoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.BufferHistory{}, "id = ?", id).Error
}
