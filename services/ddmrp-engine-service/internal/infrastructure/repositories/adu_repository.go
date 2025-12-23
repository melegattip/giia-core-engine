package repositories

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ADURepository struct {
	db *gorm.DB
}

func NewADURepository(db *gorm.DB) *ADURepository {
	return &ADURepository{db: db}
}

func (r *ADURepository) Create(ctx context.Context, adu *domain.ADUCalculation) error {
	if err := adu.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(adu).Error
}

func (r *ADURepository) GetLatest(ctx context.Context, productID, organizationID uuid.UUID) (*domain.ADUCalculation, error) {
	var adu domain.ADUCalculation
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", productID, organizationID).
		Order("calculation_date DESC").
		First(&adu).Error

	if err != nil {
		return nil, err
	}
	return &adu, nil
}

func (r *ADURepository) GetByDate(ctx context.Context, productID, organizationID uuid.UUID, date time.Time) (*domain.ADUCalculation, error) {
	var adu domain.ADUCalculation
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ? AND calculation_date = ?",
			productID, organizationID, dateOnly).
		First(&adu).Error

	if err != nil {
		return nil, err
	}
	return &adu, nil
}

func (r *ADURepository) ListHistory(ctx context.Context, productID, organizationID uuid.UUID, limit int) ([]domain.ADUCalculation, error) {
	var adus []domain.ADUCalculation
	query := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", productID, organizationID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("calculation_date DESC").Find(&adus).Error
	return adus, err
}

func (r *ADURepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ADUCalculation{}, "id = ?", id).Error
}
