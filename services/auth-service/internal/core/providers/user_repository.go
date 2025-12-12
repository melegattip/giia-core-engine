package providers

import (
	"context"

	"github.com/google/uuid"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailAndOrg(ctx context.Context, email string, orgID uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
