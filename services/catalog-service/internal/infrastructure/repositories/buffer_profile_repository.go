package repositories

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bufferProfileRepository struct {
	db *gorm.DB
}

func NewBufferProfileRepository(db *gorm.DB) providers.BufferProfileRepository {
	return &bufferProfileRepository{db: db}
}

func (r *bufferProfileRepository) tenantScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
			return db.Where("organization_id = ?", orgID)
		}
		return db
	}
}

func (r *bufferProfileRepository) Create(ctx context.Context, profile *domain.BufferProfile) error {
	if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
		profile.OrganizationID = orgID
	}

	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		if errors.IsDuplicateKeyError(err) {
			return errors.NewConflict("buffer profile with this name already exists for this organization")
		}
		return errors.NewInternalServerError("failed to create buffer profile")
	}

	return nil
}

func (r *bufferProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferProfile, error) {
	var profile domain.BufferProfile
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", id).
		First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("buffer profile not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve buffer profile")
	}

	return &profile, nil
}

func (r *bufferProfileRepository) GetByName(ctx context.Context, name string) (*domain.BufferProfile, error) {
	var profile domain.BufferProfile
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("name = ?", name).
		First(&profile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("buffer profile not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve buffer profile")
	}

	return &profile, nil
}

func (r *bufferProfileRepository) Update(ctx context.Context, profile *domain.BufferProfile) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", profile.ID).
		Updates(profile)

	if result.Error != nil {
		if errors.IsDuplicateKeyError(result.Error) {
			return errors.NewConflict("buffer profile with this name already exists for this organization")
		}
		return errors.NewInternalServerError("failed to update buffer profile")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("buffer profile not found")
	}

	return nil
}

func (r *bufferProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", id).
		Delete(&domain.BufferProfile{})

	if result.Error != nil {
		return errors.NewInternalServerError("failed to delete buffer profile")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("buffer profile not found")
	}

	return nil
}

func (r *bufferProfileRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.BufferProfile, int64, error) {
	var profiles []*domain.BufferProfile
	var total int64

	query := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.BufferProfile{})

	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count buffer profiles")
	}

	offset := (page - 1) * pageSize
	err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&profiles).Error

	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to list buffer profiles")
	}

	return profiles, total, nil
}
