package repositories

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DemandAdjustmentRepository struct {
	db *gorm.DB
}

func NewDemandAdjustmentRepository(db *gorm.DB) *DemandAdjustmentRepository {
	return &DemandAdjustmentRepository{db: db}
}

func (r *DemandAdjustmentRepository) Create(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	if err := adjustment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(adjustment).Error
}

func (r *DemandAdjustmentRepository) Update(ctx context.Context, adjustment *domain.DemandAdjustment) error {
	if err := adjustment.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(adjustment).Error
}

func (r *DemandAdjustmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DemandAdjustment, error) {
	var adjustment domain.DemandAdjustment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&adjustment).Error
	if err != nil {
		return nil, err
	}
	return &adjustment, nil
}

func (r *DemandAdjustmentRepository) GetActiveForDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) ([]domain.DemandAdjustment, error) {
	var adjustments []domain.DemandAdjustment

	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ? AND start_date <= ? AND end_date >= ?",
			productID, organizationID, dateOnly, dateOnly).
		Find(&adjustments).Error

	return adjustments, err
}

func (r *DemandAdjustmentRepository) ListByProduct(ctx context.Context, productID, organizationID uuid.UUID) ([]domain.DemandAdjustment, error) {
	var adjustments []domain.DemandAdjustment
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", productID, organizationID).
		Order("created_at DESC").
		Find(&adjustments).Error
	return adjustments, err
}

func (r *DemandAdjustmentRepository) ListByOrganization(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.DemandAdjustment, error) {
	var adjustments []domain.DemandAdjustment
	query := r.db.WithContext(ctx).Where("organization_id = ?", organizationID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&adjustments).Error
	return adjustments, err
}

func (r *DemandAdjustmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.DemandAdjustment{}, "id = ?", id).Error
}
