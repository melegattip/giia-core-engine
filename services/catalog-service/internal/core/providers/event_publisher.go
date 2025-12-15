package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
)

type EventPublisher interface {
	PublishProductCreated(ctx context.Context, product *domain.Product) error
	PublishProductUpdated(ctx context.Context, product *domain.Product) error
	PublishProductDeleted(ctx context.Context, product *domain.Product) error
	PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error
	PublishSupplierUpdated(ctx context.Context, supplier *domain.Supplier) error
	PublishSupplierDeleted(ctx context.Context, supplier *domain.Supplier) error
	PublishBufferProfileAssigned(ctx context.Context, product *domain.Product, profile *domain.BufferProfile) error
}
