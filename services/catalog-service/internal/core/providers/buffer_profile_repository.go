package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
)

type BufferProfileRepository interface {
	Create(ctx context.Context, profile *domain.BufferProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.BufferProfile, error)
	GetByName(ctx context.Context, name string) (*domain.BufferProfile, error)
	Update(ctx context.Context, profile *domain.BufferProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.BufferProfile, int64, error)
}
