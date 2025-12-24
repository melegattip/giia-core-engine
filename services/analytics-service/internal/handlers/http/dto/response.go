// Package dto provides Data Transfer Objects for HTTP handlers.
package dto

import (
	"time"

	"github.com/melegattip/giia-core-engine/services/analytics-service/internal/core/domain"
)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// PaginatedResponse represents a paginated response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response.
func NewPaginatedResponse(data interface{}, page, pageSize, total int) PaginatedResponse {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}
	return PaginatedResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

// DaysInInventoryKPIResponse represents the response for DII KPI.
type DaysInInventoryKPIResponse struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organization_id"`
	SnapshotDate      time.Time `json:"snapshot_date"`
	TotalValuedDays   float64   `json:"total_valued_days"`
	AverageValuedDays float64   `json:"average_valued_days"`
	TotalProducts     int       `json:"total_products"`
	CreatedAt         time.Time `json:"created_at"`
}

// ToDaysInInventoryKPIResponse converts domain to DTO.
func ToDaysInInventoryKPIResponse(kpi *domain.DaysInInventoryKPI) *DaysInInventoryKPIResponse {
	if kpi == nil {
		return nil
	}
	return &DaysInInventoryKPIResponse{
		ID:                kpi.ID.String(),
		OrganizationID:    kpi.OrganizationID.String(),
		SnapshotDate:      kpi.SnapshotDate,
		TotalValuedDays:   kpi.TotalValuedDays,
		AverageValuedDays: kpi.AverageValuedDays,
		TotalProducts:     kpi.TotalProducts,
		CreatedAt:         kpi.CreatedAt,
	}
}

// ImmobilizedInventoryKPIResponse represents the response for immobilized inventory KPI.
type ImmobilizedInventoryKPIResponse struct {
	ID                    string    `json:"id"`
	OrganizationID        string    `json:"organization_id"`
	SnapshotDate          time.Time `json:"snapshot_date"`
	ThresholdYears        int       `json:"threshold_years"`
	ImmobilizedCount      int       `json:"immobilized_count"`
	ImmobilizedValue      float64   `json:"immobilized_value"`
	TotalStockValue       float64   `json:"total_stock_value"`
	ImmobilizedPercentage float64   `json:"immobilized_percentage"`
	CreatedAt             time.Time `json:"created_at"`
}

// ToImmobilizedInventoryKPIResponse converts domain to DTO.
func ToImmobilizedInventoryKPIResponse(kpi *domain.ImmobilizedInventoryKPI) *ImmobilizedInventoryKPIResponse {
	if kpi == nil {
		return nil
	}
	return &ImmobilizedInventoryKPIResponse{
		ID:                    kpi.ID.String(),
		OrganizationID:        kpi.OrganizationID.String(),
		SnapshotDate:          kpi.SnapshotDate,
		ThresholdYears:        kpi.ThresholdYears,
		ImmobilizedCount:      kpi.ImmobilizedCount,
		ImmobilizedValue:      kpi.ImmobilizedValue,
		TotalStockValue:       kpi.TotalStockValue,
		ImmobilizedPercentage: kpi.ImmobilizedPercentage,
		CreatedAt:             kpi.CreatedAt,
	}
}

// RotatingProductResponse represents a rotating product in the response.
type RotatingProductResponse struct {
	ProductID     string  `json:"product_id"`
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Sales30Days   float64 `json:"sales_30_days"`
	AvgStockValue float64 `json:"avg_stock_value"`
	RotationRatio float64 `json:"rotation_ratio"`
}

// InventoryRotationKPIResponse represents the response for inventory rotation KPI.
type InventoryRotationKPIResponse struct {
	ID                   string                    `json:"id"`
	OrganizationID       string                    `json:"organization_id"`
	SnapshotDate         time.Time                 `json:"snapshot_date"`
	SalesLast30Days      float64                   `json:"sales_last_30_days"`
	AvgMonthlyStock      float64                   `json:"avg_monthly_stock"`
	RotationRatio        float64                   `json:"rotation_ratio"`
	TopRotatingProducts  []RotatingProductResponse `json:"top_rotating_products"`
	SlowRotatingProducts []RotatingProductResponse `json:"slow_rotating_products"`
	CreatedAt            time.Time                 `json:"created_at"`
}

// ToInventoryRotationKPIResponse converts domain to DTO.
func ToInventoryRotationKPIResponse(kpi *domain.InventoryRotationKPI) *InventoryRotationKPIResponse {
	if kpi == nil {
		return nil
	}

	topProducts := make([]RotatingProductResponse, len(kpi.TopRotatingProducts))
	for i, p := range kpi.TopRotatingProducts {
		topProducts[i] = RotatingProductResponse{
			ProductID:     p.ProductID.String(),
			SKU:           p.SKU,
			Name:          p.Name,
			Sales30Days:   p.Sales30Days,
			AvgStockValue: p.AvgStockValue,
			RotationRatio: p.RotationRatio,
		}
	}

	slowProducts := make([]RotatingProductResponse, len(kpi.SlowRotatingProducts))
	for i, p := range kpi.SlowRotatingProducts {
		slowProducts[i] = RotatingProductResponse{
			ProductID:     p.ProductID.String(),
			SKU:           p.SKU,
			Name:          p.Name,
			Sales30Days:   p.Sales30Days,
			AvgStockValue: p.AvgStockValue,
			RotationRatio: p.RotationRatio,
		}
	}

	return &InventoryRotationKPIResponse{
		ID:                   kpi.ID.String(),
		OrganizationID:       kpi.OrganizationID.String(),
		SnapshotDate:         kpi.SnapshotDate,
		SalesLast30Days:      kpi.SalesLast30Days,
		AvgMonthlyStock:      kpi.AvgMonthlyStock,
		RotationRatio:        kpi.RotationRatio,
		TopRotatingProducts:  topProducts,
		SlowRotatingProducts: slowProducts,
		CreatedAt:            kpi.CreatedAt,
	}
}

// BufferAnalyticsResponse represents the response for buffer analytics.
type BufferAnalyticsResponse struct {
	ID                string    `json:"id"`
	ProductID         string    `json:"product_id"`
	OrganizationID    string    `json:"organization_id"`
	Date              time.Time `json:"date"`
	CPD               float64   `json:"cpd"`
	RedZone           float64   `json:"red_zone"`
	RedBase           float64   `json:"red_base"`
	RedSafe           float64   `json:"red_safe"`
	YellowZone        float64   `json:"yellow_zone"`
	GreenZone         float64   `json:"green_zone"`
	LTD               int       `json:"ltd"`
	LeadTimeFactor    float64   `json:"lead_time_factor"`
	VariabilityFactor float64   `json:"variability_factor"`
	MOQ               int       `json:"moq"`
	OrderFrequency    int       `json:"order_frequency"`
	OptimalOrderFreq  float64   `json:"optimal_order_freq"`
	SafetyDays        float64   `json:"safety_days"`
	AvgOpenOrders     float64   `json:"avg_open_orders"`
	HasAdjustments    bool      `json:"has_adjustments"`
	CreatedAt         time.Time `json:"created_at"`
}

// ToBufferAnalyticsResponse converts domain to DTO.
func ToBufferAnalyticsResponse(ba *domain.BufferAnalytics) *BufferAnalyticsResponse {
	if ba == nil {
		return nil
	}
	return &BufferAnalyticsResponse{
		ID:                ba.ID.String(),
		ProductID:         ba.ProductID.String(),
		OrganizationID:    ba.OrganizationID.String(),
		Date:              ba.Date,
		CPD:               ba.CPD,
		RedZone:           ba.RedZone,
		RedBase:           ba.RedBase,
		RedSafe:           ba.RedSafe,
		YellowZone:        ba.YellowZone,
		GreenZone:         ba.GreenZone,
		LTD:               ba.LTD,
		LeadTimeFactor:    ba.LeadTimeFactor,
		VariabilityFactor: ba.VariabilityFactor,
		MOQ:               ba.MOQ,
		OrderFrequency:    ba.OrderFrequency,
		OptimalOrderFreq:  ba.OptimalOrderFreq,
		SafetyDays:        ba.SafetyDays,
		AvgOpenOrders:     ba.AvgOpenOrders,
		HasAdjustments:    ba.HasAdjustments,
		CreatedAt:         ba.CreatedAt,
	}
}

// ToBufferAnalyticsListResponse converts a slice of domain to DTOs.
func ToBufferAnalyticsListResponse(analytics []*domain.BufferAnalytics) []*BufferAnalyticsResponse {
	result := make([]*BufferAnalyticsResponse, len(analytics))
	for i, a := range analytics {
		result[i] = ToBufferAnalyticsResponse(a)
	}
	return result
}

// KPISnapshotResponse represents the consolidated KPI snapshot response.
type KPISnapshotResponse struct {
	ID                  string    `json:"id"`
	OrganizationID      string    `json:"organization_id"`
	SnapshotDate        time.Time `json:"snapshot_date"`
	InventoryTurnover   float64   `json:"inventory_turnover"`
	StockoutRate        float64   `json:"stockout_rate"`
	ServiceLevel        float64   `json:"service_level"`
	ExcessInventoryPct  float64   `json:"excess_inventory_pct"`
	BufferScoreGreen    float64   `json:"buffer_score_green"`
	BufferScoreYellow   float64   `json:"buffer_score_yellow"`
	BufferScoreRed      float64   `json:"buffer_score_red"`
	TotalInventoryValue float64   `json:"total_inventory_value"`
	CreatedAt           time.Time `json:"created_at"`
}

// ToKPISnapshotResponse converts domain to DTO.
func ToKPISnapshotResponse(snapshot *domain.KPISnapshot) *KPISnapshotResponse {
	if snapshot == nil {
		return nil
	}
	return &KPISnapshotResponse{
		ID:                  snapshot.ID.String(),
		OrganizationID:      snapshot.OrganizationID.String(),
		SnapshotDate:        snapshot.SnapshotDate,
		InventoryTurnover:   snapshot.InventoryTurnover,
		StockoutRate:        snapshot.StockoutRate,
		ServiceLevel:        snapshot.ServiceLevel,
		ExcessInventoryPct:  snapshot.ExcessInventoryPct,
		BufferScoreGreen:    snapshot.BufferScoreGreen,
		BufferScoreYellow:   snapshot.BufferScoreYellow,
		BufferScoreRed:      snapshot.BufferScoreRed,
		TotalInventoryValue: snapshot.TotalInventoryValue,
		CreatedAt:           snapshot.CreatedAt,
	}
}

// SyncBufferResponse represents the response for buffer sync.
type SyncBufferResponse struct {
	SyncedCount int    `json:"synced_count"`
	Message     string `json:"message"`
}

// AnalyticsResponseWithWarnings wraps a response with potential warnings.
type AnalyticsResponseWithWarnings struct {
	Data     interface{} `json:"data"`
	Warnings []string    `json:"warnings,omitempty"`
	Partial  bool        `json:"partial,omitempty"`
}
