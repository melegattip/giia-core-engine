package repositories

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BufferAdjustmentRepository struct {
	db *gorm.DB
}

func NewBufferAdjustmentRepository(db *gorm.DB) *BufferAdjustmentRepository {
	return &BufferAdjustmentRepository{db: db}
}

func (r *BufferAdjustmentRepository) Create(ctx context.Context, adjustment *domain.BufferAdjustment) error {
	if err := adjustment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(adjustment).Error
}

func (r *BufferAdjustmentRepository) Update(ctx context.Context, adjustment *domain.BufferAdjustment) error {
	if err := adjustment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(adjustment).Error
}

func (r *BufferAdjustmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferAdjustment, error) {
	var adjustment domain.BufferAdjustment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&adjustment).Error
	if err != nil {
		return nil, err
	}
	return &adjustment, nil
}

func (r *BufferAdjustmentRepository) GetActiveForDate(ctx context.Context, bufferID uuid.UUID, date time.Time) ([]domain.BufferAdjustment, error) {
	var adjustments []domain.BufferAdjustment

	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("buffer_id = ? AND start_date <= ? AND end_date >= ?",
			bufferID, dateOnly, dateOnly).
		Find(&adjustments).Error

	return adjustments, err
}

func (r *BufferAdjustmentRepository) ListByBuffer(ctx context.Context, bufferID uuid.UUID) ([]domain.BufferAdjustment, error) {
	var adjustments []domain.BufferAdjustment
	err := r.db.WithContext(ctx).
		Where("buffer_id = ?", bufferID).
		Order("created_at DESC").
		Find(&adjustments).Error
	return adjustments, err
}

func (r *BufferAdjustmentRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.BufferAdjustment, error) {
	var adjustments []domain.BufferAdjustment
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", productID, organizationID).
		Order("created_at DESC").
		Find(&adjustments).Error
	return adjustments, err
}

func (r *BufferAdjustmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.BufferAdjustment{}, "id = ?", id).Error
}
