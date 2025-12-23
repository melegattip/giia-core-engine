package repositories

import (
	"context"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BufferRepository struct {
	db *gorm.DB
}

func NewBufferRepository(db *gorm.DB) *BufferRepository {
	return &BufferRepository{db: db}
}

func (r *BufferRepository) Create(ctx context.Context, buffer *domain.Buffer) error {
	if err := buffer.Validate(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(buffer).Error
}

func (r *BufferRepository) Save(ctx context.Context, buffer *domain.Buffer) error {
	if err := buffer.Validate(); err != nil {
		return err
	}

	var existing domain.Buffer
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", buffer.ProductID, buffer.OrganizationID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(buffer).Error
	}
	if err != nil {
		return err
	}

	buffer.ID = existing.ID
	buffer.CreatedAt = existing.CreatedAt
	return r.db.WithContext(ctx).Save(buffer).Error
}

func (r *BufferRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Buffer, error) {
	var buffer domain.Buffer
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&buffer).Error
	if err != nil {
		return nil, err
	}
	return &buffer, nil
}

func (r *BufferRepository) GetByProduct(ctx context.Context, productID, organizationID uuid.UUID) (*domain.Buffer, error) {
	var buffer domain.Buffer
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND organization_id = ?", productID, organizationID).
		First(&buffer).Error
	if err != nil {
		return nil, err
	}
	return &buffer, nil
}

func (r *BufferRepository) List(ctx context.Context, organizationID uuid.UUID, limit, offset int) ([]domain.Buffer, error) {
	var buffers []domain.Buffer
	query := r.db.WithContext(ctx).Where("organization_id = ?", organizationID)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("updated_at DESC").Find(&buffers).Error
	return buffers, err
}

func (r *BufferRepository) ListByZone(ctx context.Context, organizationID uuid.UUID, zone domain.ZoneType) ([]domain.Buffer, error) {
	var buffers []domain.Buffer
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND zone = ?", organizationID, zone).
		Order("updated_at DESC").
		Find(&buffers).Error
	return buffers, err
}

func (r *BufferRepository) ListByAlertLevel(ctx context.Context, organizationID uuid.UUID, alertLevel domain.AlertLevel) ([]domain.Buffer, error) {
	var buffers []domain.Buffer
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND alert_level = ?", organizationID, alertLevel).
		Order("updated_at DESC").
		Find(&buffers).Error
	return buffers, err
}

func (r *BufferRepository) ListAll(ctx context.Context, organizationID uuid.UUID) ([]domain.Buffer, error) {
	var buffers []domain.Buffer
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Find(&buffers).Error
	return buffers, err
}

func (r *BufferRepository) UpdateNFP(ctx context.Context, bufferID uuid.UUID, onHand, onOrder, qualifiedDemand float64) error {
	return r.db.WithContext(ctx).Model(&domain.Buffer{}).
		Where("id = ?", bufferID).
		Updates(map[string]interface{}{
			"on_hand":          onHand,
			"on_order":         onOrder,
			"qualified_demand": qualifiedDemand,
		}).Error
}

func (r *BufferRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Buffer{}, "id = ?", id).Error
}
