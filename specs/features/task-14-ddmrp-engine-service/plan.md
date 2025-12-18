# Task 14: DDMRP Engine Service - Implementation Plan

**Task ID**: task-14-ddmrp-engine-service
**Phase**: 2B - New Microservices
**Priority**: P1 (High - Core Business Logic)
**Estimated Duration**: 3-4 weeks
**Dependencies**: Task 12 (Catalog at 100%)

---

## 1. Technical Context

### Current State
- **DDMRP Engine**: Not yet implemented (new service)
- **Catalog Service**: 100% complete with Product, Supplier, BufferProfile entities
- **Execution Service**: Will provide inventory transactions for NFP calculation

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure)
- **Database**: PostgreSQL 16 with GORM
- **gRPC**: Protocol Buffers v3
- **Event Streaming**: NATS JetStream
- **Scheduling**: Cron for daily buffer recalculation
- **Testing**: testify, httptest, gRPC testing framework

### Key Design Decisions
- **Daily Recalculation**: Scheduled cron job recalculates all buffers daily
- **FAD System**: Support multiple overlapping adjustments with factor multiplication
- **Buffer History**: Store daily snapshots for trend analysis and auditing
- **Event-Driven**: Publish buffer status changes for downstream services
- **Green Zone**: Use MAX(MOQ, FO × CPD, LTD × CPD × %LT) per requirements
- **Multi-tenancy**: organization_id filtering at all layers

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
└── services/ddmrp-engine-service/
    ├── api/
    │   └── proto/
    │       └── ddmrp/
    │           └── v1/
    │               ├── ddmrp.proto                        [NEW]
    │               ├── ddmrp.pb.go                        [GENERATED]
    │               └── ddmrp_grpc.pb.go                   [GENERATED]
    │
    ├── cmd/
    │   └── main.go                                         [NEW]
    │
    ├── internal/
    │   ├── core/
    │   │   ├── domain/
    │   │   │   ├── adu_calculation.go                     [NEW]
    │   │   │   ├── buffer.go                              [NEW]
    │   │   │   ├── demand_adjustment.go                   [NEW] FAD entity
    │   │   │   ├── buffer_adjustment.go                   [NEW] Zone adjustments
    │   │   │   ├── buffer_history.go                      [NEW] Daily snapshots
    │   │   │   ├── net_flow_position.go                   [NEW]
    │   │   │   └── errors.go                              [NEW]
    │   │   │
    │   │   ├── providers/
    │   │   │   ├── adu_repository.go                      [NEW]
    │   │   │   ├── buffer_repository.go                   [NEW]
    │   │   │   ├── demand_adjustment_repository.go        [NEW]
    │   │   │   ├── buffer_adjustment_repository.go        [NEW]
    │   │   │   ├── buffer_history_repository.go           [NEW]
    │   │   │   ├── catalog_service_client.go              [NEW]
    │   │   │   ├── execution_service_client.go            [NEW]
    │   │   │   └── event_publisher.go                     [NEW]
    │   │   │
    │   │   └── usecases/
    │   │       ├── adu/
    │   │       │   ├── calculate_adu.go                   [NEW]
    │   │       │   ├── get_adu.go                         [NEW]
    │   │       │   └── list_adu_history.go                [NEW]
    │   │       │
    │   │       ├── buffer/
    │   │       │   ├── calculate_buffer.go                [NEW]
    │   │       │   ├── recalculate_all_buffers.go         [NEW] Daily cron
    │   │       │   ├── get_buffer.go                      [NEW]
    │   │       │   ├── list_buffers.go                    [NEW]
    │   │       │   ├── get_buffer_status.go               [NEW]
    │   │       │   └── get_buffer_history.go              [NEW]
    │   │       │
    │   │       ├── demand_adjustment/
    │   │       │   ├── create_fad.go                      [NEW]
    │   │       │   ├── update_fad.go                      [NEW]
    │   │       │   ├── delete_fad.go                      [NEW]
    │   │       │   ├── list_fads.go                       [NEW]
    │   │       │   └── get_active_fads.go                 [NEW]
    │   │       │
    │   │       ├── buffer_adjustment/
    │   │       │   ├── create_adjustment.go               [NEW]
    │   │       │   ├── update_adjustment.go               [NEW]
    │   │       │   ├── delete_adjustment.go               [NEW]
    │   │       │   └── list_adjustments.go                [NEW]
    │   │       │
    │   │       └── nfp/
    │   │           ├── calculate_nfp.go                   [NEW]
    │   │           ├── update_nfp.go                      [NEW]
    │   │           └── check_replenishment_needed.go      [NEW]
    │   │
    │   └── infrastructure/
    │       ├── adapters/
    │       │   ├── catalog/
    │       │   │   ├── grpc_catalog_client.go             [NEW]
    │       │   │   └── catalog_client_mock.go             [NEW]
    │       │   │
    │       │   ├── execution/
    │       │   │   ├── grpc_execution_client.go           [NEW]
    │       │   │   └── execution_client_mock.go           [NEW]
    │       │   │
    │       │   └── events/
    │       │       ├── nats_publisher.go                  [NEW]
    │       │       └── publisher_mock.go                  [NEW]
    │       │
    │       ├── repositories/
    │       │   ├── adu_repository.go                      [NEW]
    │       │   ├── buffer_repository.go                   [NEW]
    │       │   ├── demand_adjustment_repository.go        [NEW]
    │       │   ├── buffer_adjustment_repository.go        [NEW]
    │       │   └── buffer_history_repository.go           [NEW]
    │       │
    │       ├── entrypoints/
    │       │   ├── grpc/
    │       │   │   ├── server.go                          [NEW]
    │       │   │   ├── adu_handler.go                     [NEW]
    │       │   │   ├── buffer_handler.go                  [NEW]
    │       │   │   ├── fad_handler.go                     [NEW]
    │       │   │   └── nfp_handler.go                     [NEW]
    │       │   │
    │       │   └── cron/
    │       │       └── daily_recalculation.go             [NEW]
    │       │
    │       └── database/
    │           └── migrations/
    │               ├── 000001_create_adu_calculations.up.sql    [NEW]
    │               ├── 000002_create_buffers.up.sql             [NEW]
    │               ├── 000003_create_demand_adjustments.up.sql  [NEW]
    │               ├── 000004_create_buffer_adjustments.up.sql  [NEW]
    │               └── 000005_create_buffer_history.up.sql      [NEW]
    │
    ├── config/
    │   └── config.yaml                                     [NEW]
    │
    ├── Dockerfile                                          [NEW]
    ├── Makefile                                            [NEW]
    ├── go.mod                                              [NEW]
    └── README.md                                           [NEW]
```

---

## 3. Implementation Steps

### Phase 1: Foundation (Week 1 - Days 1-3)

#### T001: Database Migrations

**File**: `services/ddmrp-engine-service/migrations/000001_create_adu_calculations.up.sql`

```sql
-- ADU Calculations table
CREATE TABLE IF NOT EXISTS adu_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    calculation_date DATE NOT NULL,
    adu_value DECIMAL(15,2) NOT NULL,
    method VARCHAR(20) NOT NULL, -- 'average', 'exponential', 'weighted'
    period_days INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_adu_product_date UNIQUE (product_id, organization_id, calculation_date)
);

CREATE INDEX idx_adu_product ON adu_calculations(product_id, organization_id);
CREATE INDEX idx_adu_calc_date ON adu_calculations(calculation_date DESC);
CREATE INDEX idx_adu_org ON adu_calculations(organization_id);
```

**File**: `services/ddmrp-engine-service/migrations/000002_create_buffers.up.sql`

```sql
-- Buffers table
CREATE TABLE IF NOT EXISTS buffers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    buffer_profile_id UUID NOT NULL,
    cpd DECIMAL(15,2) NOT NULL, -- Current/Adjusted CPD (ceiling)
    ltd INTEGER NOT NULL, -- Lead Time Decoupled (days)
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL, -- red_base + red_safe
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    top_of_red DECIMAL(15,2) NOT NULL,
    top_of_yellow DECIMAL(15,2) NOT NULL,
    top_of_green DECIMAL(15,2) NOT NULL,
    on_hand DECIMAL(15,2) NOT NULL DEFAULT 0,
    on_order DECIMAL(15,2) NOT NULL DEFAULT 0,
    qualified_demand DECIMAL(15,2) NOT NULL DEFAULT 0,
    net_flow_position DECIMAL(15,2) NOT NULL DEFAULT 0,
    buffer_penetration DECIMAL(5,2) NOT NULL DEFAULT 0, -- Percentage
    zone VARCHAR(20) NOT NULL DEFAULT 'green', -- 'green', 'yellow', 'red', 'below_red'
    alert_level VARCHAR(20) NOT NULL DEFAULT 'normal', -- 'normal', 'monitor', 'replenish', 'critical'
    last_recalculated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_product UNIQUE (product_id, organization_id),
    CONSTRAINT chk_zone CHECK (zone IN ('green', 'yellow', 'red', 'below_red')),
    CONSTRAINT chk_alert_level CHECK (alert_level IN ('normal', 'monitor', 'replenish', 'critical'))
);

CREATE INDEX idx_buffers_product ON buffers(product_id, organization_id);
CREATE INDEX idx_buffers_org ON buffers(organization_id);
CREATE INDEX idx_buffers_zone ON buffers(zone);
CREATE INDEX idx_buffers_alert ON buffers(alert_level);
```

**File**: `services/ddmrp-engine-service/migrations/000003_create_demand_adjustments.up.sql`

```sql
-- Demand Adjustments (FAD) table
CREATE TABLE IF NOT EXISTS demand_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    adjustment_type VARCHAR(30) NOT NULL,
    factor DECIMAL(5,2) NOT NULL, -- Multiplier (e.g., 1.5 = 50% increase, 0.0 = discontinue)
    reason TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    CONSTRAINT chk_adjustment_type CHECK (adjustment_type IN ('fad', 'seasonal', 'new_product', 'discontinue', 'promotion')),
    CONSTRAINT chk_factor CHECK (factor >= 0),
    CONSTRAINT chk_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_demand_adj_product ON demand_adjustments(product_id, organization_id);
CREATE INDEX idx_demand_adj_dates ON demand_adjustments(start_date, end_date);
CREATE INDEX idx_demand_adj_org ON demand_adjustments(organization_id);
```

**File**: `services/ddmrp-engine-service/migrations/000004_create_buffer_adjustments.up.sql`

```sql
-- Buffer Adjustments table
CREATE TABLE IF NOT EXISTS buffer_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buffer_id UUID NOT NULL REFERENCES buffers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    adjustment_type VARCHAR(30) NOT NULL,
    target_zone VARCHAR(20) NOT NULL, -- 'red', 'yellow', 'green', 'all'
    factor DECIMAL(5,2) NOT NULL, -- Multiplier (e.g., 1.2 = 20% increase)
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    CONSTRAINT chk_buffer_adj_type CHECK (adjustment_type IN ('zone_factor', 'planned_event', 'spike_management', 'seasonal_prepare')),
    CONSTRAINT chk_buffer_target_zone CHECK (target_zone IN ('red', 'yellow', 'green', 'all')),
    CONSTRAINT chk_buffer_factor CHECK (factor > 0),
    CONSTRAINT chk_buffer_dates CHECK (end_date >= start_date)
);

CREATE INDEX idx_buffer_adj_buffer ON buffer_adjustments(buffer_id);
CREATE INDEX idx_buffer_adj_product ON buffer_adjustments(product_id, organization_id);
CREATE INDEX idx_buffer_adj_dates ON buffer_adjustments(start_date, end_date);
```

**File**: `services/ddmrp-engine-service/migrations/000005_create_buffer_history.up.sql`

```sql
-- Buffer History table (daily snapshots)
CREATE TABLE IF NOT EXISTS buffer_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buffer_id UUID NOT NULL REFERENCES buffers(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    snapshot_date DATE NOT NULL,
    cpd DECIMAL(15,2) NOT NULL,
    dlt INTEGER NOT NULL,
    red_zone DECIMAL(15,2) NOT NULL,
    red_base DECIMAL(15,2) NOT NULL,
    red_safe DECIMAL(15,2) NOT NULL,
    yellow_zone DECIMAL(15,2) NOT NULL,
    green_zone DECIMAL(15,2) NOT NULL,
    lead_time_factor DECIMAL(5,2) NOT NULL,
    variability_factor DECIMAL(5,2) NOT NULL,
    moq INTEGER,
    order_frequency INTEGER,
    has_adjustments BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_buffer_history_date UNIQUE (buffer_id, snapshot_date)
);

CREATE INDEX idx_buffer_history_buffer ON buffer_history(buffer_id);
CREATE INDEX idx_buffer_history_product ON buffer_history(product_id, organization_id);
CREATE INDEX idx_buffer_history_date ON buffer_history(snapshot_date DESC);
```

#### T002: Define Domain Entities

**File**: `services/ddmrp-engine-service/internal/core/domain/demand_adjustment.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type DemandAdjustment struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	AdjustmentType DemandAdjustmentType
	Factor         float64 // Multiplier for CPD (e.g., 1.5 = 50% increase, 0.0 = discontinue)
	Reason         string
	CreatedAt      time.Time
	CreatedBy      uuid.UUID
}

type DemandAdjustmentType string

const (
	DemandAdjustmentFAD         DemandAdjustmentType = "fad"
	DemandAdjustmentSeasonal    DemandAdjustmentType = "seasonal"
	DemandAdjustmentNewProduct  DemandAdjustmentType = "new_product"
	DemandAdjustmentDiscontinue DemandAdjustmentType = "discontinue"
	DemandAdjustmentPromotion   DemandAdjustmentType = "promotion"
)

func (t DemandAdjustmentType) IsValid() bool {
	switch t {
	case DemandAdjustmentFAD, DemandAdjustmentSeasonal, DemandAdjustmentNewProduct,
		DemandAdjustmentDiscontinue, DemandAdjustmentPromotion:
		return true
	}
	return false
}

func NewDemandAdjustment(
	productID, orgID, createdBy uuid.UUID,
	startDate, endDate time.Time,
	adjustmentType DemandAdjustmentType,
	factor float64,
	reason string,
) (*DemandAdjustment, error) {
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if !adjustmentType.IsValid() {
		return nil, NewValidationError("invalid adjustment type")
	}
	if factor < 0 {
		return nil, NewValidationError("factor must be non-negative")
	}
	if endDate.Before(startDate) {
		return nil, NewValidationError("end_date must be >= start_date")
	}
	if reason == "" {
		return nil, NewValidationError("reason is required")
	}

	return &DemandAdjustment{
		ID:             uuid.New(),
		ProductID:      productID,
		OrganizationID: orgID,
		StartDate:      startDate,
		EndDate:        endDate,
		AdjustmentType: adjustmentType,
		Factor:         factor,
		Reason:         reason,
		CreatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}, nil
}

// IsActive returns true if the adjustment is active for the given date
func (da *DemandAdjustment) IsActive(date time.Time) bool {
	return !date.Before(da.StartDate) && !date.After(da.EndDate)
}
```

**File**: `services/ddmrp-engine-service/internal/core/domain/buffer.go`

```go
package domain

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type Buffer struct {
	ID                 uuid.UUID
	ProductID          uuid.UUID
	OrganizationID     uuid.UUID
	BufferProfileID    uuid.UUID
	CPD                float64 // Current/Adjusted CPD (ceiling)
	LTD                int     // Lead Time Decoupled (days)
	RedBase            float64
	RedSafe            float64
	RedZone            float64 // red_base + red_safe
	YellowZone         float64
	GreenZone          float64
	TopOfRed           float64
	TopOfYellow        float64
	TopOfGreen         float64
	OnHand             float64
	OnOrder            float64
	QualifiedDemand    float64
	NetFlowPosition    float64
	BufferPenetration  float64
	Zone               ZoneType
	AlertLevel         AlertLevel
	LastRecalculatedAt time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ZoneType string

const (
	ZoneGreen    ZoneType = "green"
	ZoneYellow   ZoneType = "yellow"
	ZoneRed      ZoneType = "red"
	ZoneBelowRed ZoneType = "below_red"
)

type AlertLevel string

const (
	AlertNormal   AlertLevel = "normal"
	AlertMonitor  AlertLevel = "monitor"
	AlertReplenish AlertLevel = "replenish"
	AlertCritical AlertLevel = "critical"
)

// CalculateBufferZones calculates Red, Yellow, Green zones
// Green Zone = MAX(MOQ, FO × CPD, LTD × CPD × %LT)
func CalculateBufferZones(
	cpd float64,
	ltd int,
	leadTimeFactor float64,
	variabilityFactor float64,
	moq int,
	orderFrequency int,
) (redBase, redSafe, redZone, yellowZone, greenZone float64) {
	// Red Zone = Red Base + Red Safe
	// Red Base = DLT × CPD × %LT
	redBase = float64(ltd) * cpd * leadTimeFactor

	// Red Safe = Red Base × %CV
	redSafe = redBase * variabilityFactor

	// Total Red Zone
	redZone = redBase + redSafe

	// Yellow Zone = CPD × LTD
	yellowZone = cpd * float64(ltd)

	// Green Zone = MAX(MOQ, FO × CPD, LTD × CPD × %LT)
	option1 := float64(moq)
	option2 := float64(orderFrequency) * cpd
	option3 := float64(ltd) * cpd * leadTimeFactor

	greenZone = math.Max(option1, math.Max(option2, option3))

	return
}

// CalculateNFP calculates Net Flow Position
func (b *Buffer) CalculateNFP() {
	b.NetFlowPosition = b.OnHand + b.OnOrder - b.QualifiedDemand
}

// DetermineZone determines which zone the NFP is in
func (b *Buffer) DetermineZone() {
	b.CalculateNFP()

	topOfRed := b.RedZone
	topOfYellow := topOfRed + b.YellowZone
	topOfGreen := topOfYellow + b.GreenZone

	b.TopOfRed = topOfRed
	b.TopOfYellow = topOfYellow
	b.TopOfGreen = topOfGreen

	switch {
	case b.NetFlowPosition >= topOfYellow:
		b.Zone = ZoneGreen
		b.AlertLevel = AlertNormal
	case b.NetFlowPosition >= topOfRed:
		b.Zone = ZoneYellow
		b.AlertLevel = AlertMonitor
	case b.NetFlowPosition > 0:
		b.Zone = ZoneRed
		b.AlertLevel = AlertReplenish
	default:
		b.Zone = ZoneBelowRed
		b.AlertLevel = AlertCritical
	}

	// Calculate buffer penetration percentage
	if topOfGreen > 0 {
		b.BufferPenetration = (b.NetFlowPosition / topOfGreen) * 100
	}
}

// ApplyAdjustedCPD applies FAD adjustments to CPD
func ApplyAdjustedCPD(baseCPD float64, activeFADs []DemandAdjustment) float64 {
	adjustedCPD := baseCPD

	for _, fad := range activeFADs {
		adjustedCPD *= fad.Factor
	}

	// CPD must be ceiling (round up)
	return math.Ceil(adjustedCPD)
}
```

**File**: `services/ddmrp-engine-service/internal/core/domain/buffer_adjustment.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferAdjustment struct {
	ID             uuid.UUID
	BufferID       uuid.UUID
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	AdjustmentType BufferAdjustmentType
	TargetZone     ZoneType // Which zone to adjust
	Factor         float64  // Multiplier (e.g., 1.2 = 20% increase)
	StartDate      time.Time
	EndDate        time.Time
	Reason         string
	CreatedAt      time.Time
	CreatedBy      uuid.UUID
}

type BufferAdjustmentType string

const (
	BufferAdjustmentZoneFactor     BufferAdjustmentType = "zone_factor"
	BufferAdjustmentPlannedEvent   BufferAdjustmentType = "planned_event"
	BufferAdjustmentSpikeManagement BufferAdjustmentType = "spike_management"
	BufferAdjustmentSeasonalPrepare BufferAdjustmentType = "seasonal_prepare"
)

func (t BufferAdjustmentType) IsValid() bool {
	switch t {
	case BufferAdjustmentZoneFactor, BufferAdjustmentPlannedEvent,
		BufferAdjustmentSpikeManagement, BufferAdjustmentSeasonalPrepare:
		return true
	}
	return false
}

// IsActive returns true if the adjustment is active for the given date
func (ba *BufferAdjustment) IsActive(date time.Time) bool {
	return !date.Before(ba.StartDate) && !date.After(ba.EndDate)
}
```

**File**: `services/ddmrp-engine-service/internal/core/domain/buffer_history.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferHistory struct {
	ID                 uuid.UUID
	BufferID           uuid.UUID
	ProductID          uuid.UUID
	OrganizationID     uuid.UUID
	SnapshotDate       time.Time
	CPD                float64
	DLT                int
	RedZone            float64
	RedBase            float64
	RedSafe            float64
	YellowZone         float64
	GreenZone          float64
	LeadTimeFactor     float64
	VariabilityFactor  float64
	MOQ                *int
	OrderFrequency     *int
	HasAdjustments     bool
	CreatedAt          time.Time
}

func NewBufferHistory(buffer *Buffer, leadTimeFactor, variabilityFactor float64, moq, orderFrequency *int, hasAdjustments bool) *BufferHistory {
	return &BufferHistory{
		ID:                uuid.New(),
		BufferID:          buffer.ID,
		ProductID:         buffer.ProductID,
		OrganizationID:    buffer.OrganizationID,
		SnapshotDate:      time.Now(),
		CPD:               buffer.CPD,
		DLT:               buffer.LTD,
		RedZone:           buffer.RedZone,
		RedBase:           buffer.RedBase,
		RedSafe:           buffer.RedSafe,
		YellowZone:        buffer.YellowZone,
		GreenZone:         buffer.GreenZone,
		LeadTimeFactor:    leadTimeFactor,
		VariabilityFactor: variabilityFactor,
		MOQ:               moq,
		OrderFrequency:    orderFrequency,
		HasAdjustments:    hasAdjustments,
		CreatedAt:         time.Now(),
	}
}
```

---

### Phase 2: Core Use Cases (Week 1 Days 4-5, Week 2)

#### T003: FAD (Demand Adjustment) Use Cases

**File**: `services/ddmrp-engine-service/internal/core/usecases/demand_adjustment/create_fad.go`

```go
package demand_adjustment

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
)

type CreateFADUseCase struct {
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	eventPublisher       providers.EventPublisher
}

func NewCreateFADUseCase(
	repo providers.DemandAdjustmentRepository,
	publisher providers.EventPublisher,
) *CreateFADUseCase {
	return &CreateFADUseCase{
		demandAdjustmentRepo: repo,
		eventPublisher:       publisher,
	}
}

type CreateFADInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	StartDate      time.Time
	EndDate        time.Time
	AdjustmentType domain.DemandAdjustmentType
	Factor         float64
	Reason         string
	CreatedBy      uuid.UUID
}

func (uc *CreateFADUseCase) Execute(ctx context.Context, input CreateFADInput) (*domain.DemandAdjustment, error) {
	fad, err := domain.NewDemandAdjustment(
		input.ProductID,
		input.OrganizationID,
		input.CreatedBy,
		input.StartDate,
		input.EndDate,
		input.AdjustmentType,
		input.Factor,
		input.Reason,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.demandAdjustmentRepo.Create(ctx, fad); err != nil {
		return nil, err
	}

	// Publish event for buffer recalculation trigger
	uc.eventPublisher.PublishFADCreated(ctx, fad)

	return fad, nil
}
```

#### T004: Buffer Calculation Use Case

**File**: `services/ddmrp-engine-service/internal/core/usecases/buffer/calculate_buffer.go`

```go
package buffer

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
)

type CalculateBufferUseCase struct {
	bufferRepo           providers.BufferRepository
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	bufferAdjustmentRepo providers.BufferAdjustmentRepository
	catalogClient        providers.CatalogServiceClient
	aduRepo              providers.ADURepository
}

func NewCalculateBufferUseCase(
	bufferRepo providers.BufferRepository,
	demandAdjustmentRepo providers.DemandAdjustmentRepository,
	bufferAdjustmentRepo providers.BufferAdjustmentRepository,
	catalogClient providers.CatalogServiceClient,
	aduRepo providers.ADURepository,
) *CalculateBufferUseCase {
	return &CalculateBufferUseCase{
		bufferRepo:           bufferRepo,
		demandAdjustmentRepo: demandAdjustmentRepo,
		bufferAdjustmentRepo: bufferAdjustmentRepo,
		catalogClient:        catalogClient,
		aduRepo:              aduRepo,
	}
}

type CalculateBufferInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
}

func (uc *CalculateBufferUseCase) Execute(ctx context.Context, input CalculateBufferInput) (*domain.Buffer, error) {
	// 1. Get product and buffer profile from Catalog
	product, err := uc.catalogClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}

	if product.BufferProfileID == nil {
		return nil, domain.NewValidationError("product has no buffer profile assigned")
	}

	bufferProfile, err := uc.catalogClient.GetBufferProfile(ctx, *product.BufferProfileID)
	if err != nil {
		return nil, err
	}

	// 2. Get latest ADU
	adu, err := uc.aduRepo.GetLatest(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	baseCPD := math.Ceil(adu.ADUValue) // CPD is ceiling of ADU

	// 3. Get active FADs for today
	today := time.Now()
	activeFADs, err := uc.demandAdjustmentRepo.GetActiveForDate(ctx, input.ProductID, input.OrganizationID, today)
	if err != nil {
		return nil, err
	}

	// 4. Apply FAD adjustments to CPD
	adjustedCPD := domain.ApplyAdjustedCPD(baseCPD, activeFADs)

	// 5. Get supplier info for MOQ
	supplier, err := uc.catalogClient.GetPrimarySupplier(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}

	moq := supplier.MOQ
	orderFrequency := bufferProfile.OrderFrequency // From buffer profile

	// 6. Calculate buffer zones
	redBase, redSafe, redZone, yellowZone, greenZone := domain.CalculateBufferZones(
		adjustedCPD,
		product.LeadTime,
		bufferProfile.LeadTimeFactor,
		bufferProfile.VariabilityFactor,
		moq,
		orderFrequency,
	)

	// 7. Get or create buffer
	buffer, err := uc.bufferRepo.GetByProduct(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		// Create new buffer
		buffer = &domain.Buffer{
			ID:              uuid.New(),
			ProductID:       input.ProductID,
			OrganizationID:  input.OrganizationID,
			BufferProfileID: *product.BufferProfileID,
			CreatedAt:       time.Now(),
		}
	}

	// 8. Update buffer zones
	buffer.CPD = adjustedCPD
	buffer.LTD = product.LeadTime
	buffer.RedBase = redBase
	buffer.RedSafe = redSafe
	buffer.RedZone = redZone
	buffer.YellowZone = yellowZone
	buffer.GreenZone = greenZone
	buffer.LastRecalculatedAt = time.Now()
	buffer.UpdatedAt = time.Now()

	// 9. Apply buffer zone adjustments if any
	activeBufferAdjs, err := uc.bufferAdjustmentRepo.GetActiveForDate(ctx, buffer.ID, today)
	if err == nil && len(activeBufferAdjs) > 0 {
		for _, adj := range activeBufferAdjs {
			switch adj.TargetZone {
			case "red":
				buffer.RedZone *= adj.Factor
			case "yellow":
				buffer.YellowZone *= adj.Factor
			case "green":
				buffer.GreenZone *= adj.Factor
			case "all":
				buffer.RedZone *= adj.Factor
				buffer.YellowZone *= adj.Factor
				buffer.GreenZone *= adj.Factor
			}
		}
	}

	// 10. Calculate NFP and determine zone
	buffer.DetermineZone()

	// 11. Save buffer
	if err := uc.bufferRepo.Save(ctx, buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}
```

#### T005: Daily Recalculation Cron Job

**File**: `services/ddmrp-engine-service/internal/infrastructure/entrypoints/cron/daily_recalculation.go`

```go
package cron

import (
	"context"
	"log"
	"time"

	"giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
	"github.com/robfig/cron/v3"
)

type DailyRecalculation struct {
	recalculateUseCase *buffer.RecalculateAllBuffersUseCase
	cronScheduler      *cron.Cron
}

func NewDailyRecalculation(recalculateUseCase *buffer.RecalculateAllBuffersUseCase) *DailyRecalculation {
	return &DailyRecalculation{
		recalculateUseCase: recalculateUseCase,
		cronScheduler:      cron.New(),
	}
}

func (dr *DailyRecalculation) Start() {
	// Run daily at 2 AM
	dr.cronScheduler.AddFunc("0 2 * * *", func() {
		log.Println("Starting daily buffer recalculation...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := dr.recalculateUseCase.Execute(ctx); err != nil {
			log.Printf("Error in daily recalculation: %v", err)
		} else {
			log.Println("Daily buffer recalculation completed successfully")
		}
	})

	dr.cronScheduler.Start()
}

func (dr *DailyRecalculation) Stop() {
	dr.cronScheduler.Stop()
}
```

---

### Phase 3: gRPC API (Week 3)

#### T006: Protocol Buffers Definition

**File**: `services/ddmrp-engine-service/api/proto/ddmrp/v1/ddmrp.proto`

```protobuf
syntax = "proto3";

package ddmrp.v1;

option go_package = "giia-core-engine/services/ddmrp-engine-service/api/proto/ddmrp/v1;ddmrpv1";

import "google/protobuf/timestamp.proto";

// DDMRP Engine Service
service DDMRPService {
  // Buffer operations
  rpc CalculateBuffer(CalculateBufferRequest) returns (CalculateBufferResponse);
  rpc GetBuffer(GetBufferRequest) returns (GetBufferResponse);
  rpc ListBuffers(ListBuffersRequest) returns (ListBuffersResponse);
  rpc GetBufferStatus(GetBufferStatusRequest) returns (GetBufferStatusResponse);

  // FAD (Demand Adjustment) operations
  rpc CreateFAD(CreateFADRequest) returns (CreateFADResponse);
  rpc UpdateFAD(UpdateFADRequest) returns (UpdateFADResponse);
  rpc DeleteFAD(DeleteFADRequest) returns (DeleteFADResponse);
  rpc ListFADs(ListFADsRequest) returns (ListFADsResponse);

  // Buffer Adjustment operations
  rpc CreateBufferAdjustment(CreateBufferAdjustmentRequest) returns (CreateBufferAdjustmentResponse);
  rpc ListBufferAdjustments(ListBufferAdjustmentsRequest) returns (ListBufferAdjustmentsResponse);

  // NFP operations
  rpc UpdateNFP(UpdateNFPRequest) returns (UpdateNFPResponse);
  rpc CheckReplenishmentNeeded(CheckReplenishmentRequest) returns (CheckReplenishmentResponse);
}

message Buffer {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  double cpd = 4;
  int32 ltd = 5;
  double red_base = 6;
  double red_safe = 7;
  double red_zone = 8;
  double yellow_zone = 9;
  double green_zone = 10;
  double top_of_red = 11;
  double top_of_yellow = 12;
  double top_of_green = 13;
  double on_hand = 14;
  double on_order = 15;
  double qualified_demand = 16;
  double net_flow_position = 17;
  double buffer_penetration = 18;
  string zone = 19;
  string alert_level = 20;
  google.protobuf.Timestamp last_recalculated_at = 21;
}

message DemandAdjustment {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  google.protobuf.Timestamp start_date = 4;
  google.protobuf.Timestamp end_date = 5;
  string adjustment_type = 6;
  double factor = 7;
  string reason = 8;
  google.protobuf.Timestamp created_at = 9;
}

message CalculateBufferRequest {
  string product_id = 1;
  string organization_id = 2;
}

message CalculateBufferResponse {
  Buffer buffer = 1;
}

message CreateFADRequest {
  string product_id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp start_date = 3;
  google.protobuf.Timestamp end_date = 4;
  string adjustment_type = 5;
  double factor = 6;
  string reason = 7;
}

message CreateFADResponse {
  DemandAdjustment demand_adjustment = 1;
}

// ... (other message definitions)
```

---

### Phase 4: Testing (Week 3-4)

#### T007: Unit Tests

**File**: `services/ddmrp-engine-service/internal/core/usecases/buffer/calculate_buffer_test.go`

```go
package buffer_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"giia-core-engine/services/ddmrp-engine-service/internal/core/usecases/buffer"
)

func TestCalculateBuffer_Success_WithFAD(t *testing.T) {
	// Given
	mockBufferRepo := new(MockBufferRepository)
	mockDemandAdjRepo := new(MockDemandAdjustmentRepository)
	mockBufferAdjRepo := new(MockBufferAdjustmentRepository)
	mockCatalogClient := new(MockCatalogClient)
	mockADURepo := new(MockADURepository)

	useCase := buffer.NewCalculateBufferUseCase(
		mockBufferRepo,
		mockDemandAdjRepo,
		mockBufferAdjRepo,
		mockCatalogClient,
		mockADURepo,
	)

	givenProductID := uuid.New()
	givenOrgID := uuid.New()
	givenBufferProfileID := uuid.New()
	givenBaseCPD := 10.0
	givenFADFactor := 1.5

	mockCatalogClient.On("GetProduct", mock.Anything, givenProductID).
		Return(&domain.Product{
			ID:              givenProductID,
			BufferProfileID: &givenBufferProfileID,
			LeadTime:        30,
		}, nil)

	mockCatalogClient.On("GetBufferProfile", mock.Anything, givenBufferProfileID).
		Return(&domain.BufferProfile{
			LeadTimeFactor:    0.5,
			VariabilityFactor: 0.5,
			OrderFrequency:    7,
		}, nil)

	mockADURepo.On("GetLatest", mock.Anything, givenProductID, givenOrgID).
		Return(&domain.ADUCalculation{
			ADUValue: givenBaseCPD,
		}, nil)

	mockDemandAdjRepo.On("GetActiveForDate", mock.Anything, givenProductID, givenOrgID, mock.Anything).
		Return([]domain.DemandAdjustment{
			{Factor: givenFADFactor},
		}, nil)

	mockCatalogClient.On("GetPrimarySupplier", mock.Anything, givenProductID).
		Return(&domain.Supplier{
			MOQ: 100,
		}, nil)

	mockBufferRepo.On("GetByProduct", mock.Anything, givenProductID, givenOrgID).
		Return(nil, domain.NewNotFoundError("buffer not found"))

	mockBufferRepo.On("Save", mock.Anything, mock.AnythingOfType("*domain.Buffer")).
		Return(nil)

	mockBufferAdjRepo.On("GetActiveForDate", mock.Anything, mock.Anything, mock.Anything).
		Return([]domain.BufferAdjustment{}, nil)

	// When
	result, err := useCase.Execute(context.Background(), buffer.CalculateBufferInput{
		ProductID:      givenProductID,
		OrganizationID: givenOrgID,
	})

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 15.0, result.CPD) // baseCPD * FADFactor = 10 * 1.5 = 15
	mockBufferRepo.AssertExpectations(t)
	mockDemandAdjRepo.AssertExpectations(t)
	mockCatalogClient.AssertExpectations(t)
}
```

---

## 4. Success Criteria

### Mandatory
- ✅ All domain entities created (Buffer, DemandAdjustment, BufferAdjustment, BufferHistory)
- ✅ FAD system functional with multi-factor application
- ✅ Buffer calculation with MAX(MOQ, FO × CPD, LTD × CPD × %LT) for Green Zone
- ✅ Daily recalculation cron job working
- ✅ Buffer history snapshots saved daily
- ✅ NFP calculation accurate
- ✅ gRPC API for all operations
- ✅ Integration with Catalog service
- ✅ Event publishing for buffer status changes
- ✅ 85%+ test coverage
- ✅ Multi-tenancy support

---

## 5. Dependencies

- **Task 12**: Catalog service must be 100% complete
- **Shared packages**: pkg/events, pkg/database, pkg/logger, pkg/errors
- **External**: NATS JetStream, PostgreSQL

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Complex buffer calculations | Extensive unit tests with known values |
| FAD factor multiplication errors | Detailed test scenarios with multiple overlapping FADs |
| Daily cron failures | Error logging, retry logic, monitoring alerts |
| Integration with Catalog | Mock clients for testing, contract tests |

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
