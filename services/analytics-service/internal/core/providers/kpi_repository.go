package providers

import (
	"context"
	"time"

	"github.com/giia/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/google/uuid"
)

type KPIRepository interface {
	SaveKPISnapshot(ctx context.Context, snapshot *domain.KPISnapshot) error
	GetKPISnapshot(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.KPISnapshot, error)
	ListKPISnapshots(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.KPISnapshot, error)

	SaveDaysInInventoryKPI(ctx context.Context, kpi *domain.DaysInInventoryKPI) error
	GetDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.DaysInInventoryKPI, error)
	ListDaysInInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.DaysInInventoryKPI, error)

	SaveImmobilizedInventoryKPI(ctx context.Context, kpi *domain.ImmobilizedInventoryKPI) error
	GetImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, date time.Time, thresholdYears int) (*domain.ImmobilizedInventoryKPI, error)
	ListImmobilizedInventoryKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time, thresholdYears int) ([]*domain.ImmobilizedInventoryKPI, error)

	SaveInventoryRotationKPI(ctx context.Context, kpi *domain.InventoryRotationKPI) error
	GetInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, date time.Time) (*domain.InventoryRotationKPI, error)
	ListInventoryRotationKPI(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.InventoryRotationKPI, error)

	SaveBufferAnalytics(ctx context.Context, analytics *domain.BufferAnalytics) error
	GetBufferAnalyticsByProduct(ctx context.Context, organizationID, productID uuid.UUID, date time.Time) (*domain.BufferAnalytics, error)
	ListBufferAnalytics(ctx context.Context, organizationID uuid.UUID, startDate, endDate time.Time) ([]*domain.BufferAnalytics, error)
}
