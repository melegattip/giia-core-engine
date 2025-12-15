package repositories

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) providers.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) tenantScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
			return db.Where("organization_id = ?", orgID)
		}
		return db
	}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
		supplier.OrganizationID = orgID
	}

	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		if errors.IsDuplicateKeyError(err) {
			return errors.NewConflict("supplier with this code already exists for this organization")
		}
		return errors.NewInternalServerError("failed to create supplier")
	}

	return nil
}

func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", id).
		First(&supplier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("supplier not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve supplier")
	}

	return &supplier, nil
}

func (r *supplierRepository) GetByCode(ctx context.Context, code string) (*domain.Supplier, error) {
	var supplier domain.Supplier
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("code = ?", code).
		First(&supplier).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("supplier not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve supplier")
	}

	return &supplier, nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", supplier.ID).
		Updates(supplier)

	if result.Error != nil {
		if errors.IsDuplicateKeyError(result.Error) {
			return errors.NewConflict("supplier with this code already exists for this organization")
		}
		return errors.NewInternalServerError("failed to update supplier")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("supplier not found")
	}

	return nil
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.Supplier{}).
		Where("id = ?", id).
		Update("status", domain.SupplierStatusInactive)

	if result.Error != nil {
		return errors.NewInternalServerError("failed to delete supplier")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("supplier not found")
	}

	return nil
}

func (r *supplierRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Supplier, int64, error) {
	var suppliers []*domain.Supplier
	var total int64

	query := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.Supplier{})

	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count suppliers")
	}

	offset := (page - 1) * pageSize
	err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&suppliers).Error

	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to list suppliers")
	}

	return suppliers, total, nil
}
