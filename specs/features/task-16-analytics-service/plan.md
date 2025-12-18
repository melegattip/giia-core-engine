# Task 16: Analytics Service - Implementation Plan

**Task ID**: task-16-analytics-service
**Phase**: 2B - New Microservices
**Priority**: P3 (Low - Reporting Layer)
**Estimated Duration**: 2-3 weeks
**Dependencies**: All operational services (Tasks 12, 14, 15)

---

## 1. Technical Context

### Current State
- **Analytics Service**: Not yet implemented (new service)
- **Operational Services**: Catalog, DDMRP Engine, Execution complete
- **Data Sources**: Events from all services via NATS JetStream

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure)
- **Database**: PostgreSQL 16 (separate analytics DB for time-series data)
- **gRPC & REST**: Both APIs for flexibility
- **Event Streaming**: NATS JetStream (consumer for all service events)
- **Reporting**: PDF generation, Excel export, CSV export
- **Testing**: testify, httptest

### Key Design Decisions
- **Separate Analytics DB**: Isolate from operational databases
- **Event-Driven Data Aggregation**: Subscribe to all service events
- **Materialized Views**: Pre-calculate common aggregations
- **Time-Series Optimizations**: Partitioning by date
- **Daily KPI Snapshots**: Store daily calculations for trend analysis

---

## 2. Key Entities and Calculations

### New KPIs from Requirements

#### 1. Days in Inventory (Valorizado)
```go
type DaysInInventoryKPI struct {
    ID                uuid.UUID
    OrganizationID    uuid.UUID
    SnapshotDate      time.Time
    TotalValuedDays   float64   // Sum of (DaysInStock × UnitCost × Quantity)
    AverageValuedDays float64   // TotalValuedDays / TotalProducts
    TotalProducts     int
    CreatedAt         time.Time
}

// Calculation:
// For each product:
//   DaysInInventory = Current Date - LastPurchaseDate
//   ValuedDays = DaysInInventory × (Quantity × StandardCost)
// TotalValuedDays = Sum(ValuedDays for all products)
```

#### 2. Immobilized Inventory
```go
type ImmobilizedInventoryKPI struct {
    ID                     uuid.UUID
    OrganizationID         uuid.UUID
    SnapshotDate           time.Time
    ThresholdYears         int        // Configurable (e.g., 1, 2, 3 years)
    ImmobilizedCount       int        // Products with age > threshold
    ImmobilizedValue       float64    // Total value of immobilized products
    TotalStockValue        float64
    ImmobilizedPercentage  float64    // (ImmobilizedValue / TotalStockValue) × 100
    CreatedAt              time.Time
}

// Calculation:
// ImmobilizedProducts = Products WHERE (Current Date - PurchaseDate) > ThresholdYears
// ImmobilizedValue = Sum(Quantity × StandardCost) for immobilized products
// ImmobilizedPercentage = (ImmobilizedValue / TotalStockValue) × 100
```

#### 3. Inventory Rotation
```go
type InventoryRotationKPI struct {
    ID                  uuid.UUID
    OrganizationID      uuid.UUID
    SnapshotDate        time.Time
    SalesLast30Days     float64   // Total sales value (last 30 days)
    AvgMonthlyStock     float64   // Average stock value during period
    RotationRatio       float64   // SalesLast30Days / AvgMonthlyStock
    TopRotatingProducts []RotatingProduct
    SlowRotatingProducts []RotatingProduct
    CreatedAt           time.Time
}

// Calculation (per requirements):
// RotationRatio = (Sales Last 30 Days) / (Average Monthly Stock)
// Sales30Days = Sum(Sales Value) for last 30 days
// AvgMonthlyStock = Average(Daily Stock Value) over last 30 days
```

#### 4. Buffer Analytics
```go
type BufferAnalytics struct {
    ID                  uuid.UUID
    ProductID           uuid.UUID
    OrganizationID      uuid.UUID
    Date                time.Time
    CPD                 float64
    RedZone             float64
    RedBase             float64
    RedSafe             float64
    YellowZone          float64
    GreenZone           float64
    LTD                 int
    LeadTimeFactor      float64
    VariabilityFactor   float64
    MOQ                 int
    OrderFrequency      int
    OptimalOrderFreq    float64   // Green / CPD
    SafetyDays          float64   // Red / CPD
    AvgOpenOrders       float64   // Yellow / Green
    HasAdjustments      bool
    CreatedAt           time.Time
}

// Synchronized daily from DDMRP Engine BufferHistory
```

---

## 3. Implementation Steps

### Phase 1: Database Schema (Week 1 Days 1-2)

#### Migrations

**File**: `services/analytics-service/migrations/000001_create_kpi_snapshots.up.sql`

```sql
-- KPI Snapshots table
CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    inventory_turnover DECIMAL(10,2),
    stockout_rate DECIMAL(5,2),
    service_level DECIMAL(5,2),
    excess_inventory_pct DECIMAL(5,2),
    buffer_score_green DECIMAL(5,2),
    buffer_score_yellow DECIMAL(5,2),
    buffer_score_red DECIMAL(5,2),
    total_inventory_value DECIMAL(15,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_kpi_snapshot_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_kpi_snapshots_org ON kpi_snapshots(organization_id);
CREATE INDEX idx_kpi_snapshots_date ON kpi_snapshots(snapshot_date DESC);
```

**File**: `services/analytics-service/migrations/000002_create_days_in_inventory_kpi.up.sql`

```sql
-- Days in Inventory KPI table
CREATE TABLE IF NOT EXISTS days_in_inventory_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    total_valued_days DECIMAL(20,2) NOT NULL,
    average_valued_days DECIMAL(10,2) NOT NULL,
    total_products INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_days_in_inventory_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_days_in_inventory_org ON days_in_inventory_kpi(organization_id);
CREATE INDEX idx_days_in_inventory_date ON days_in_inventory_kpi(snapshot_date DESC);
```

**File**: `services/analytics-service/migrations/000003_create_immobilized_inventory_kpi.up.sql`

```sql
-- Immobilized Inventory KPI table
CREATE TABLE IF NOT EXISTS immobilized_inventory_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    threshold_years INTEGER NOT NULL,
    immobilized_count INTEGER NOT NULL,
    immobilized_value DECIMAL(15,2) NOT NULL,
    total_stock_value DECIMAL(15,2) NOT NULL,
    immobilized_percentage DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_immobilized_org_date UNIQUE (organization_id, snapshot_date, threshold_years)
);

CREATE INDEX idx_immobilized_org ON immobilized_inventory_kpi(organization_id);
CREATE INDEX idx_immobilized_date ON immobilized_inventory_kpi(snapshot_date DESC);
```

**File**: `services/analytics-service/migrations/000004_create_inventory_rotation_kpi.up.sql`

```sql
-- Inventory Rotation KPI table
CREATE TABLE IF NOT EXISTS inventory_rotation_kpi (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    sales_last_30_days DECIMAL(15,2) NOT NULL,
    avg_monthly_stock DECIMAL(15,2) NOT NULL,
    rotation_ratio DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_rotation_org_date UNIQUE (organization_id, snapshot_date)
);

CREATE INDEX idx_rotation_org ON inventory_rotation_kpi(organization_id);
CREATE INDEX idx_rotation_date ON inventory_rotation_kpi(snapshot_date DESC);

-- Top/Slow Rotating Products detail table
CREATE TABLE IF NOT EXISTS rotating_products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kpi_id UUID NOT NULL REFERENCES inventory_rotation_kpi(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(200) NOT NULL,
    sales_30_days DECIMAL(15,2) NOT NULL,
    avg_stock_value DECIMAL(15,2) NOT NULL,
    rotation_ratio DECIMAL(10,2) NOT NULL,
    category VARCHAR(20) NOT NULL, -- 'top' or 'slow'
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rotating_products_kpi ON rotating_products(kpi_id);
```

**File**: `services/analytics-service/migrations/000005_create_buffer_analytics.up.sql`

```sql
-- Buffer Analytics table (synchronized from DDMRP Engine)
CREATE TABLE IF NOT EXISTS buffer_analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    cpd DECIMAL(15,2) NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL,
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    ltd INTEGER NOT NULL,
    lead_time_factor DECIMAL(5,2) NOT NULL,
    variability_factor DECIMAL(5,2) NOT NULL,
    moq INTEGER,
    order_frequency INTEGER,
    optimal_order_freq DECIMAL(10,2),
    safety_days DECIMAL(10,2),
    avg_open_orders DECIMAL(10,2),
    has_adjustments BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_analytics_product_date UNIQUE (product_id, organization_id, snapshot_date)
);

CREATE INDEX idx_buffer_analytics_product ON buffer_analytics(product_id, organization_id);
CREATE INDEX idx_buffer_analytics_date ON buffer_analytics(snapshot_date DESC);
```

---

### Phase 2: Event Consumers (Week 1 Days 3-5)

#### Event Subscription

**File**: `services/analytics-service/internal/infrastructure/adapters/events/nats_consumer.go`

```go
package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"giia-core-engine/services/analytics-service/internal/core/usecases/aggregation"
)

type NATSConsumer struct {
	nc                  *nats.Conn
	aggregationUseCase  *aggregation.AggregateDataUseCase
}

func NewNATSConsumer(nc *nats.Conn, aggregationUseCase *aggregation.AggregateDataUseCase) *NATSConsumer {
	return &NATSConsumer{
		nc:                 nc,
		aggregationUseCase: aggregationUseCase,
	}
}

func (c *NATSConsumer) Start() error {
	// Subscribe to all relevant events
	subjects := []string{
		"catalog.product.created",
		"catalog.product.updated",
		"ddmrp.buffer.calculated",
		"ddmrp.buffer_history.created",
		"execution.po.received",
		"execution.so.created",
		"execution.inventory.updated",
	}

	for _, subject := range subjects {
		if _, err := c.nc.Subscribe(subject, c.handleEvent); err != nil {
			return err
		}
	}

	log.Println("Analytics event consumers started")
	return nil
}

func (c *NATSConsumer) handleEvent(msg *nats.Msg) {
	// Parse event and trigger aggregation
	var event map[string]interface{}
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Error parsing event: %v", err)
		return
	}

	// Process event based on subject
	ctx := context.Background()
	if err := c.aggregationUseCase.ProcessEvent(ctx, msg.Subject, event); err != nil {
		log.Printf("Error processing event %s: %v", msg.Subject, err)
	}
}
```

---

### Phase 3: KPI Calculation Use Cases (Week 2)

#### Days in Inventory Calculation

**File**: `services/analytics-service/internal/core/usecases/kpi/calculate_days_in_inventory.go`

```go
package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/analytics-service/internal/core/domain"
	"giia-core-engine/services/analytics-service/internal/core/providers"
)

type CalculateDaysInInventoryUseCase struct {
	kpiRepo       providers.KPIRepository
	catalogClient providers.CatalogServiceClient
}

func (uc *CalculateDaysInInventoryUseCase) Execute(ctx context.Context, orgID uuid.UUID, date time.Time) error {
	// 1. Get all products with inventory
	products, err := uc.catalogClient.ListProductsWithInventory(ctx, orgID)
	if err != nil {
		return err
	}

	totalValuedDays := 0.0
	totalProducts := 0

	// 2. Calculate valued days for each product
	for _, product := range products {
		if product.LastPurchaseDate == nil {
			continue
		}

		daysInInventory := int(date.Sub(*product.LastPurchaseDate).Hours() / 24)
		if daysInInventory < 0 {
			continue
		}

		totalValue := product.Quantity * product.StandardCost
		valuedDays := float64(daysInInventory) * totalValue

		totalValuedDays += valuedDays
		totalProducts++
	}

	// 3. Create KPI snapshot
	kpi := &domain.DaysInInventoryKPI{
		ID:                uuid.New(),
		OrganizationID:    orgID,
		SnapshotDate:      date,
		TotalValuedDays:   totalValuedDays,
		AverageValuedDays: totalValuedDays / float64(totalProducts),
		TotalProducts:     totalProducts,
		CreatedAt:         time.Now(),
	}

	// 4. Save KPI
	return uc.kpiRepo.SaveDaysInInventoryKPI(ctx, kpi)
}
```

#### Immobilized Inventory Calculation

**File**: `services/analytics-service/internal/core/usecases/kpi/calculate_immobilized_inventory.go`

```go
package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/analytics-service/internal/core/domain"
	"giia-core-engine/services/analytics-service/internal/core/providers"
)

type CalculateImmobilizedInventoryUseCase struct {
	kpiRepo       providers.KPIRepository
	catalogClient providers.CatalogServiceClient
}

func (uc *CalculateImmobilizedInventoryUseCase) Execute(ctx context.Context, orgID uuid.UUID, date time.Time, thresholdYears int) error {
	// 1. Get all products
	products, err := uc.catalogClient.ListProductsWithInventory(ctx, orgID)
	if err != nil {
		return err
	}

	thresholdDate := date.AddDate(-thresholdYears, 0, 0)
	immobilizedCount := 0
	immobilizedValue := 0.0
	totalStockValue := 0.0

	// 2. Identify immobilized products
	for _, product := range products {
		productValue := product.Quantity * product.StandardCost
		totalStockValue += productValue

		if product.LastPurchaseDate != nil && product.LastPurchaseDate.Before(thresholdDate) {
			immobilizedCount++
			immobilizedValue += productValue
		}
	}

	// 3. Calculate percentage
	immobilizedPercentage := 0.0
	if totalStockValue > 0 {
		immobilizedPercentage = (immobilizedValue / totalStockValue) * 100
	}

	// 4. Create KPI snapshot
	kpi := &domain.ImmobilizedInventoryKPI{
		ID:                    uuid.New(),
		OrganizationID:        orgID,
		SnapshotDate:          date,
		ThresholdYears:        thresholdYears,
		ImmobilizedCount:      immobilizedCount,
		ImmobilizedValue:      immobilizedValue,
		TotalStockValue:       totalStockValue,
		ImmobilizedPercentage: immobilizedPercentage,
		CreatedAt:             time.Now(),
	}

	// 5. Save KPI
	return uc.kpiRepo.SaveImmobilizedInventoryKPI(ctx, kpi)
}
```

#### Inventory Rotation Calculation

**File**: `services/analytics-service/internal/core/usecases/kpi/calculate_inventory_rotation.go`

```go
package kpi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/analytics-service/internal/core/domain"
	"giia-core-engine/services/analytics-service/internal/core/providers"
)

type CalculateInventoryRotationUseCase struct {
	kpiRepo         providers.KPIRepository
	executionClient providers.ExecutionServiceClient
	catalogClient   providers.CatalogServiceClient
}

func (uc *CalculateInventoryRotationUseCase) Execute(ctx context.Context, orgID uuid.UUID, date time.Time) error {
	// 1. Get sales data for last 30 days
	startDate := date.AddDate(0, 0, -30)
	salesData, err := uc.executionClient.GetSalesData(ctx, orgID, startDate, date)
	if err != nil {
		return err
	}

	salesLast30Days := salesData.TotalValue

	// 2. Get inventory data for last 30 days
	inventoryData, err := uc.executionClient.GetInventorySnapshots(ctx, orgID, startDate, date)
	if err != nil {
		return err
	}

	// Calculate average monthly stock
	totalStockValue := 0.0
	for _, snapshot := range inventoryData {
		totalStockValue += snapshot.TotalValue
	}
	avgMonthlyStock := totalStockValue / float64(len(inventoryData))

	// 3. Calculate rotation ratio
	rotationRatio := 0.0
	if avgMonthlyStock > 0 {
		rotationRatio = salesLast30Days / avgMonthlyStock
	}

	// 4. Get top and slow rotating products
	topProducts := uc.getTopRotatingProducts(ctx, orgID, startDate, date, 10)
	slowProducts := uc.getSlowRotatingProducts(ctx, orgID, startDate, date, 10)

	// 5. Create KPI snapshot
	kpi := &domain.InventoryRotationKPI{
		ID:                   uuid.New(),
		OrganizationID:       orgID,
		SnapshotDate:         date,
		SalesLast30Days:      salesLast30Days,
		AvgMonthlyStock:      avgMonthlyStock,
		RotationRatio:        rotationRatio,
		TopRotatingProducts:  topProducts,
		SlowRotatingProducts: slowProducts,
		CreatedAt:            time.Now(),
	}

	// 6. Save KPI
	return uc.kpiRepo.SaveInventoryRotationKPI(ctx, kpi)
}
```

---

### Phase 4: Daily KPI Cron Job (Week 2 Day 4)

**File**: `services/analytics-service/internal/infrastructure/entrypoints/cron/daily_kpi_calculation.go`

```go
package cron

import (
	"context"
	"log"
	"time"

	"giia-core-engine/services/analytics-service/internal/core/usecases/kpi"
	"github.com/robfig/cron/v3"
)

type DailyKPICalculation struct {
	daysInInventoryUC  *kpi.CalculateDaysInInventoryUseCase
	immobilizedUC      *kpi.CalculateImmobilizedInventoryUseCase
	rotationUC         *kpi.CalculateInventoryRotationUseCase
	cronScheduler      *cron.Cron
}

func (kc *DailyKPICalculation) Start() {
	// Run daily at 3 AM
	kc.cronScheduler.AddFunc("0 3 * * *", func() {
		log.Println("Starting daily KPI calculation...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		today := time.Now()

		// Calculate all KPIs
		// ... (execute all KPI calculations)

		log.Println("Daily KPI calculation completed")
	})

	kc.cronScheduler.Start()
}
```

---

### Phase 5: gRPC & REST APIs (Week 3)

#### Protocol Buffers

**File**: `services/analytics-service/api/proto/analytics/v1/analytics.proto`

```protobuf
syntax = "proto3";

package analytics.v1;

option go_package = "giia-core-engine/services/analytics-service/api/proto/analytics/v1;analyticsv1";

import "google/protobuf/timestamp.proto";

service AnalyticsService {
  rpc GetKPISnapshot(GetKPISnapshotRequest) returns (GetKPISnapshotResponse);
  rpc GetDaysInInventoryKPI(GetDaysInInventoryKPIRequest) returns (GetDaysInInventoryKPIResponse);
  rpc GetImmobilizedInventoryKPI(GetImmobilizedInventoryKPIRequest) returns (GetImmobilizedInventoryKPIResponse);
  rpc GetInventoryRotationKPI(GetInventoryRotationKPIRequest) returns (GetInventoryRotationKPIResponse);
  rpc GetBufferAnalytics(GetBufferAnalyticsRequest) returns (GetBufferAnalyticsResponse);
}

message DaysInInventoryKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  double total_valued_days = 4;
  double average_valued_days = 5;
  int32 total_products = 6;
}

message ImmobilizedInventoryKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  int32 threshold_years = 4;
  int32 immobilized_count = 5;
  double immobilized_value = 6;
  double total_stock_value = 7;
  double immobilized_percentage = 8;
}

message InventoryRotationKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  double sales_last_30_days = 4;
  double avg_monthly_stock = 5;
  double rotation_ratio = 6;
}

// ... (other messages)
```

---

## 4. Success Criteria

### Mandatory
- ✅ Event consumers for all services
- ✅ Days in Inventory (Valorizado) KPI calculation
- ✅ Immobilized Inventory KPI calculation
- ✅ Inventory Rotation KPI calculation
- ✅ Buffer Analytics synchronization from DDMRP Engine
- ✅ Daily KPI cron job
- ✅ gRPC and REST APIs
- ✅ Export to PDF, Excel, CSV
- ✅ 80%+ test coverage
- ✅ Multi-tenancy support

---

## 5. Dependencies

- **All Services**: For event data (Catalog, DDMRP, Execution)
- **Shared packages**: pkg/events, pkg/database, pkg/logger

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Event processing lag | NATS JetStream guarantees, consumer groups |
| KPI calculation performance | Batch processing, database indexing, materialized views |
| Data consistency | Eventually consistent model acceptable (5min lag) |

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
