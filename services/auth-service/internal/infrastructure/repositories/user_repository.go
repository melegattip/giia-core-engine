package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) providers.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	orgID := getOrgIDFromContext(ctx)

	var user domain.User
	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	err := query.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	orgID := getOrgIDFromContext(ctx)

	var user domain.User
	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	err := query.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmailAndOrg(ctx context.Context, email string, orgID uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Where("email = ? AND organization_id = ?", email, orgID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	orgID := getOrgIDFromContext(ctx)

	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	return query.Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	orgID := getOrgIDFromContext(ctx)

	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	return query.Delete(&domain.User{}, "id = ?", id).Error
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	orgID := getOrgIDFromContext(ctx)

	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	return query.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", gorm.Expr("NOW()")).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, error) {
	orgID := getOrgIDFromContext(ctx)

	var users []*domain.User
	query := r.db.WithContext(ctx)

	if orgID != uuid.Nil {
		query = query.Scopes(TenantScope(orgID))
	}

	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func getOrgIDFromContext(ctx context.Context) uuid.UUID {
	if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok {
		return orgID
	}
	return uuid.Nil
}
