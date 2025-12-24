// Package grpc provides gRPC handlers for the Analytics Service.
package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/handlers/http/cache"
)

// AnalyticsService implements the gRPC AnalyticsService interface.
type AnalyticsService struct {
	kpiRepo            providers.KPIRepository
	diiUseCase         *kpi.CalculateDaysInInventoryUseCase
	immobilizedUseCase *kpi.CalculateImmobilizedInventoryUseCase
	rotationUseCase    *kpi.CalculateInventoryRotationUseCase
	syncBufferUseCase  *kpi.SyncBufferAnalyticsUseCase
	cache              *cache.Cache
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(
	kpiRepo providers.KPIRepository,
	diiUseCase *kpi.CalculateDaysInInventoryUseCase,
	immobilizedUseCase *kpi.CalculateImmobilizedInventoryUseCase,
	rotationUseCase *kpi.CalculateInventoryRotationUseCase,
	syncBufferUseCase *kpi.SyncBufferAnalyticsUseCase,
) *AnalyticsService {
	return &AnalyticsService{
		kpiRepo:            kpiRepo,
		diiUseCase:         diiUseCase,
		immobilizedUseCase: immobilizedUseCase,
		rotationUseCase:    rotationUseCase,
		syncBufferUseCase:  syncBufferUseCase,
		cache:              cache.New(5 * time.Minute),
	}
}

// KPISnapshotProto represents the proto message for KPISnapshot.
type KPISnapshotProto struct {
	ID                  string
	OrganizationID      string
	SnapshotDate        *timestamppb.Timestamp
	InventoryTurnover   float64
	StockoutRate        float64
	ServiceLevel        float64
	ExcessInventoryPct  float64
	BufferScoreGreen    float64
	BufferScoreYellow   float64
	BufferScoreRed      float64
	TotalInventoryValue float64
	CreatedAt           *timestamppb.Timestamp
}

// DaysInInventoryKPIProto represents the proto message for DaysInInventoryKPI.
type DaysInInventoryKPIProto struct {
	ID                string
	OrganizationID    string
	SnapshotDate      *timestamppb.Timestamp
	TotalValuedDays   float64
	AverageValuedDays float64
	TotalProducts     int32
	CreatedAt         *timestamppb.Timestamp
}

// ImmobilizedInventoryKPIProto represents the proto message for ImmobilizedInventoryKPI.
type ImmobilizedInventoryKPIProto struct {
	ID                    string
	OrganizationID        string
	SnapshotDate          *timestamppb.Timestamp
	ThresholdYears        int32
	ImmobilizedCount      int32
	ImmobilizedValue      float64
	TotalStockValue       float64
	ImmobilizedPercentage float64
	CreatedAt             *timestamppb.Timestamp
}

// InventoryRotationKPIProto represents the proto message for InventoryRotationKPI.
type InventoryRotationKPIProto struct {
	ID                   string
	OrganizationID       string
	SnapshotDate         *timestamppb.Timestamp
	SalesLast30Days      float64
	AvgMonthlyStock      float64
	RotationRatio        float64
	TopRotatingProducts  []*RotatingProductProto
	SlowRotatingProducts []*RotatingProductProto
	CreatedAt            *timestamppb.Timestamp
}

// RotatingProductProto represents the proto message for RotatingProduct.
type RotatingProductProto struct {
	ProductID     string
	SKU           string
	Name          string
	Sales30Days   float64
	AvgStockValue float64
	RotationRatio float64
}

// BufferAnalyticsProto represents the proto message for BufferAnalytics.
type BufferAnalyticsProto struct {
	ID                string
	ProductID         string
	OrganizationID    string
	Date              *timestamppb.Timestamp
	CPD               float64
	RedZone           float64
	RedBase           float64
	RedSafe           float64
	YellowZone        float64
	GreenZone         float64
	LTD               int32
	LeadTimeFactor    float64
	VariabilityFactor float64
	MOQ               int32
	OrderFrequency    int32
	OptimalOrderFreq  float64
	SafetyDays        float64
	AvgOpenOrders     float64
	HasAdjustments    bool
	CreatedAt         *timestamppb.Timestamp
}

// GetKPISnapshot retrieves a KPI snapshot for an organization.
func (s *AnalyticsService) GetKPISnapshot(ctx context.Context, orgID uuid.UUID, date time.Time) (*KPISnapshotProto, error) {
	cacheKey := fmt.Sprintf("grpc:snapshot:%s:%s", orgID.String(), date.Format("2006-01-02"))

	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*KPISnapshotProto), nil
	}

	snapshot, err := s.kpiRepo.GetKPISnapshot(ctx, orgID, date)
	if err != nil {
		return nil, err
	}
	if snapshot == nil {
		return nil, fmt.Errorf("snapshot not found")
	}

	result := toKPISnapshotProto(snapshot)
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListKPISnapshots retrieves KPI snapshots for a date range.
func (s *AnalyticsService) ListKPISnapshots(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]*KPISnapshotProto, error) {
	snapshots, err := s.kpiRepo.ListKPISnapshots(ctx, orgID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make([]*KPISnapshotProto, len(snapshots))
	for i, snapshot := range snapshots {
		result[i] = toKPISnapshotProto(snapshot)
	}
	return result, nil
}

// GetDaysInInventoryKPI retrieves DII KPI for an organization.
func (s *AnalyticsService) GetDaysInInventoryKPI(ctx context.Context, orgID uuid.UUID, date time.Time) (*DaysInInventoryKPIProto, error) {
	cacheKey := fmt.Sprintf("grpc:dii:%s:%s", orgID.String(), date.Format("2006-01-02"))

	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*DaysInInventoryKPIProto), nil
	}

	kpiData, err := s.kpiRepo.GetDaysInInventoryKPI(ctx, orgID, date)
	if err != nil {
		return nil, err
	}
	if kpiData == nil {
		// Calculate new KPI if not found and use case is available
		if s.diiUseCase != nil {
			input := &kpi.CalculateDaysInInventoryInput{
				OrganizationID: orgID,
				SnapshotDate:   date,
			}
			kpiData, err = s.diiUseCase.Execute(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate DII KPI: %w", err)
			}
		} else {
			return nil, fmt.Errorf("DII KPI not found")
		}
	}

	result := toDaysInInventoryKPIProto(kpiData)
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListDaysInInventoryKPI retrieves DII KPIs for a date range.
func (s *AnalyticsService) ListDaysInInventoryKPI(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]*DaysInInventoryKPIProto, error) {
	kpis, err := s.kpiRepo.ListDaysInInventoryKPI(ctx, orgID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make([]*DaysInInventoryKPIProto, len(kpis))
	for i, k := range kpis {
		result[i] = toDaysInInventoryKPIProto(k)
	}
	return result, nil
}

// GetImmobilizedInventoryKPI retrieves immobilized inventory KPI.
func (s *AnalyticsService) GetImmobilizedInventoryKPI(ctx context.Context, orgID uuid.UUID, date time.Time, thresholdYears int) (*ImmobilizedInventoryKPIProto, error) {
	cacheKey := fmt.Sprintf("grpc:immobilized:%s:%s:%d", orgID.String(), date.Format("2006-01-02"), thresholdYears)

	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*ImmobilizedInventoryKPIProto), nil
	}

	kpiData, err := s.kpiRepo.GetImmobilizedInventoryKPI(ctx, orgID, date, thresholdYears)
	if err != nil {
		return nil, err
	}
	if kpiData == nil {
		if s.immobilizedUseCase != nil {
			input := &kpi.CalculateImmobilizedInventoryInput{
				OrganizationID: orgID,
				SnapshotDate:   date,
				ThresholdYears: thresholdYears,
			}
			kpiData, err = s.immobilizedUseCase.Execute(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate immobilized inventory KPI: %w", err)
			}
		} else {
			return nil, fmt.Errorf("immobilized inventory KPI not found")
		}
	}

	result := toImmobilizedInventoryKPIProto(kpiData)
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListImmobilizedInventoryKPI retrieves immobilized inventory KPIs for a date range.
func (s *AnalyticsService) ListImmobilizedInventoryKPI(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time, thresholdYears int) ([]*ImmobilizedInventoryKPIProto, error) {
	kpis, err := s.kpiRepo.ListImmobilizedInventoryKPI(ctx, orgID, startDate, endDate, thresholdYears)
	if err != nil {
		return nil, err
	}

	result := make([]*ImmobilizedInventoryKPIProto, len(kpis))
	for i, k := range kpis {
		result[i] = toImmobilizedInventoryKPIProto(k)
	}
	return result, nil
}

// GetInventoryRotationKPI retrieves inventory rotation KPI.
func (s *AnalyticsService) GetInventoryRotationKPI(ctx context.Context, orgID uuid.UUID, date time.Time) (*InventoryRotationKPIProto, error) {
	cacheKey := fmt.Sprintf("grpc:rotation:%s:%s", orgID.String(), date.Format("2006-01-02"))

	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*InventoryRotationKPIProto), nil
	}

	kpiData, err := s.kpiRepo.GetInventoryRotationKPI(ctx, orgID, date)
	if err != nil {
		return nil, err
	}
	if kpiData == nil {
		if s.rotationUseCase != nil {
			input := &kpi.CalculateInventoryRotationInput{
				OrganizationID: orgID,
				SnapshotDate:   date,
			}
			kpiData, err = s.rotationUseCase.Execute(ctx, input)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate rotation KPI: %w", err)
			}
		} else {
			return nil, fmt.Errorf("rotation KPI not found")
		}
	}

	result := toInventoryRotationKPIProto(kpiData)
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListInventoryRotationKPI retrieves inventory rotation KPIs for a date range.
func (s *AnalyticsService) ListInventoryRotationKPI(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]*InventoryRotationKPIProto, error) {
	kpis, err := s.kpiRepo.ListInventoryRotationKPI(ctx, orgID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make([]*InventoryRotationKPIProto, len(kpis))
	for i, k := range kpis {
		result[i] = toInventoryRotationKPIProto(k)
	}
	return result, nil
}

// GetBufferAnalytics retrieves buffer analytics for a product.
func (s *AnalyticsService) GetBufferAnalytics(ctx context.Context, orgID, productID uuid.UUID, date time.Time) (*BufferAnalyticsProto, error) {
	cacheKey := fmt.Sprintf("grpc:buffer:%s:%s:%s", orgID.String(), productID.String(), date.Format("2006-01-02"))

	if cached, found := s.cache.Get(cacheKey); found {
		return cached.(*BufferAnalyticsProto), nil
	}

	analytics, err := s.kpiRepo.GetBufferAnalyticsByProduct(ctx, orgID, productID, date)
	if err != nil {
		return nil, err
	}
	if analytics == nil {
		return nil, fmt.Errorf("buffer analytics not found")
	}

	result := toBufferAnalyticsProto(analytics)
	s.cache.Set(cacheKey, result)
	return result, nil
}

// ListBufferAnalytics retrieves buffer analytics for a date range.
func (s *AnalyticsService) ListBufferAnalytics(ctx context.Context, orgID uuid.UUID, startDate, endDate time.Time) ([]*BufferAnalyticsProto, error) {
	analytics, err := s.kpiRepo.ListBufferAnalytics(ctx, orgID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make([]*BufferAnalyticsProto, len(analytics))
	for i, a := range analytics {
		result[i] = toBufferAnalyticsProto(a)
	}
	return result, nil
}

// SyncBufferData syncs buffer data from DDMRP service.
func (s *AnalyticsService) SyncBufferData(ctx context.Context, orgID uuid.UUID, date time.Time) (int, error) {
	if s.syncBufferUseCase == nil {
		return 0, fmt.Errorf("sync use case not configured")
	}

	input := &kpi.SyncBufferAnalyticsInput{
		OrganizationID: orgID,
		Date:           date,
	}

	syncedCount, err := s.syncBufferUseCase.Execute(ctx, input)
	if err != nil {
		return 0, err
	}

	// Clear relevant cache entries
	s.cache.Clear()

	return syncedCount, nil
}

// Conversion functions

func toKPISnapshotProto(snapshot *domain.KPISnapshot) *KPISnapshotProto {
	return &KPISnapshotProto{
		ID:                  snapshot.ID.String(),
		OrganizationID:      snapshot.OrganizationID.String(),
		SnapshotDate:        timestamppb.New(snapshot.SnapshotDate),
		InventoryTurnover:   snapshot.InventoryTurnover,
		StockoutRate:        snapshot.StockoutRate,
		ServiceLevel:        snapshot.ServiceLevel,
		ExcessInventoryPct:  snapshot.ExcessInventoryPct,
		BufferScoreGreen:    snapshot.BufferScoreGreen,
		BufferScoreYellow:   snapshot.BufferScoreYellow,
		BufferScoreRed:      snapshot.BufferScoreRed,
		TotalInventoryValue: snapshot.TotalInventoryValue,
		CreatedAt:           timestamppb.New(snapshot.CreatedAt),
	}
}

func toDaysInInventoryKPIProto(kpiData *domain.DaysInInventoryKPI) *DaysInInventoryKPIProto {
	return &DaysInInventoryKPIProto{
		ID:                kpiData.ID.String(),
		OrganizationID:    kpiData.OrganizationID.String(),
		SnapshotDate:      timestamppb.New(kpiData.SnapshotDate),
		TotalValuedDays:   kpiData.TotalValuedDays,
		AverageValuedDays: kpiData.AverageValuedDays,
		TotalProducts:     int32(kpiData.TotalProducts),
		CreatedAt:         timestamppb.New(kpiData.CreatedAt),
	}
}

func toImmobilizedInventoryKPIProto(kpiData *domain.ImmobilizedInventoryKPI) *ImmobilizedInventoryKPIProto {
	return &ImmobilizedInventoryKPIProto{
		ID:                    kpiData.ID.String(),
		OrganizationID:        kpiData.OrganizationID.String(),
		SnapshotDate:          timestamppb.New(kpiData.SnapshotDate),
		ThresholdYears:        int32(kpiData.ThresholdYears),
		ImmobilizedCount:      int32(kpiData.ImmobilizedCount),
		ImmobilizedValue:      kpiData.ImmobilizedValue,
		TotalStockValue:       kpiData.TotalStockValue,
		ImmobilizedPercentage: kpiData.ImmobilizedPercentage,
		CreatedAt:             timestamppb.New(kpiData.CreatedAt),
	}
}

func toInventoryRotationKPIProto(kpiData *domain.InventoryRotationKPI) *InventoryRotationKPIProto {
	topProducts := make([]*RotatingProductProto, len(kpiData.TopRotatingProducts))
	for i, p := range kpiData.TopRotatingProducts {
		topProducts[i] = &RotatingProductProto{
			ProductID:     p.ProductID.String(),
			SKU:           p.SKU,
			Name:          p.Name,
			Sales30Days:   p.Sales30Days,
			AvgStockValue: p.AvgStockValue,
			RotationRatio: p.RotationRatio,
		}
	}

	slowProducts := make([]*RotatingProductProto, len(kpiData.SlowRotatingProducts))
	for i, p := range kpiData.SlowRotatingProducts {
		slowProducts[i] = &RotatingProductProto{
			ProductID:     p.ProductID.String(),
			SKU:           p.SKU,
			Name:          p.Name,
			Sales30Days:   p.Sales30Days,
			AvgStockValue: p.AvgStockValue,
			RotationRatio: p.RotationRatio,
		}
	}

	return &InventoryRotationKPIProto{
		ID:                   kpiData.ID.String(),
		OrganizationID:       kpiData.OrganizationID.String(),
		SnapshotDate:         timestamppb.New(kpiData.SnapshotDate),
		SalesLast30Days:      kpiData.SalesLast30Days,
		AvgMonthlyStock:      kpiData.AvgMonthlyStock,
		RotationRatio:        kpiData.RotationRatio,
		TopRotatingProducts:  topProducts,
		SlowRotatingProducts: slowProducts,
		CreatedAt:            timestamppb.New(kpiData.CreatedAt),
	}
}

func toBufferAnalyticsProto(analytics *domain.BufferAnalytics) *BufferAnalyticsProto {
	return &BufferAnalyticsProto{
		ID:                analytics.ID.String(),
		ProductID:         analytics.ProductID.String(),
		OrganizationID:    analytics.OrganizationID.String(),
		Date:              timestamppb.New(analytics.Date),
		CPD:               analytics.CPD,
		RedZone:           analytics.RedZone,
		RedBase:           analytics.RedBase,
		RedSafe:           analytics.RedSafe,
		YellowZone:        analytics.YellowZone,
		GreenZone:         analytics.GreenZone,
		LTD:               int32(analytics.LTD),
		LeadTimeFactor:    analytics.LeadTimeFactor,
		VariabilityFactor: analytics.VariabilityFactor,
		MOQ:               int32(analytics.MOQ),
		OrderFrequency:    int32(analytics.OrderFrequency),
		OptimalOrderFreq:  analytics.OptimalOrderFreq,
		SafetyDays:        analytics.SafetyDays,
		AvgOpenOrders:     analytics.AvgOpenOrders,
		HasAdjustments:    analytics.HasAdjustments,
		CreatedAt:         timestamppb.New(analytics.CreatedAt),
	}
}
