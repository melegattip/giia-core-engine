package providers

import (
	"context"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByEmailAndOrg(ctx context.Context, email string, orgID uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int) error
	UpdateLastLogin(ctx context.Context, userID int) error
	List(ctx context.Context, offset, limit int) ([]*domain.User, error)
}
