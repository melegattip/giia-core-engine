package providers

import (
	"context"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
)

type OrganizationRepository interface {
	Create(ctx context.Context, org *domain.Organization) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	Update(ctx context.Context, org *domain.Organization) error
	List(ctx context.Context, offset, limit int) ([]*domain.Organization, error)
}
