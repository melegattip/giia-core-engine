# Task 14: DDMRP Engine Service - Specification

**Task ID**: task-14-ddmrp-engine-service
**Phase**: 2B - New Microservices
**Priority**: P1 (High - Core Business Logic)
**Estimated Duration**: 3-4 weeks
**Dependencies**: Task 12 (Catalog at 100%)

---

## Overview

Implement the DDMRP (Demand Driven MRP) Engine Service, the core calculation engine for inventory buffer management. This service calculates Average Daily Usage (ADU), Decoupled Lead Time (DLT), Net Flow Equation, and maintains buffer zones (Red, Yellow, Green) for demand-driven inventory management.

---

## User Scenarios

### US1: ADU Calculation (P1)

**As a** DDMRP planner
**I want to** calculate Average Daily Usage for products
**So that** I can determine baseline demand for buffer sizing

**Acceptance Criteria**:
- Calculate ADU using configurable methods: Simple Average, Exponential Smoothing, Weighted Moving Average
- Support configurable time periods (30, 60, 90 days)
- Handle missing data and outliers
- Store ADU history for trending
- Multi-tenancy support

**Success Metrics**:
- <2s p95 for ADU calculation
- 100% calculation accuracy vs validated results

---

### US2: Buffer Calculation (P1)

**As a** DDMRP planner
**I want to** calculate buffer levels (Red, Yellow, Green zones)
**So that** I can maintain optimal inventory levels

**Acceptance Criteria**:
- Calculate Red Zone = DLT × ADU × Lead Time Factor
- Calculate Yellow Zone = ADU × Replenishment Cycle
- Calculate Green Zone = DLT × ADU × Variability Factor
- Support buffer profile customization per product
- Recalculate automatically on parameter changes

**Success Metrics**:
- <3s p95 for buffer calculation
- Buffer recommendations match manual calculations

---

### US3: Net Flow Equation (P1)

**As a** DDMRP planner
**I want to** calculate Net Flow Position for products
**So that** I know when to replenish inventory

**Acceptance Criteria**:
- NFP = On-Hand + On-Order - Qualified Demand
- Track on-hand inventory from Execution service
- Track on-order quantities from pending orders
- Identify qualified demand (confirmed orders)
- Calculate buffer penetration percentage
- Trigger replenishment alerts

**Success Metrics**:
- Real-time NFP updates (<5s)
- 100% accuracy of buffer penetration alerts

---

### US4: Demand and Buffer Adjustments (P1) [UPDATED]

**As a** DDMRP planner
**I want to** adjust demand (FAD) and buffer zones for planned events
**So that** buffers adapt to known future demand changes and special circumstances

**Acceptance Criteria**:

**FAD (Demand Adjustment Factor)**:
- Create demand adjustments with date ranges and multiplier factors
- Apply FAD to CPD calculation: CPD_Adjusted = CPD_Original × Factor
- Support adjustment types: seasonal, new product, discontinuation, promotion
- Multiple FAD adjustments can overlap (factors multiply)
- FAD applied during buffer recalculation for affected periods

**Buffer Zone Adjustments**:
- Manually adjust specific zones (red, yellow, green, or all) by factor
- Support planned events (promotional periods, seasonal prep)
- Apply adjustments during specified date ranges
- Maintain adjustment history for audit trail

**Daily Recalculation**:
- Recalculate all buffers daily with current CPD
- Apply active FAD adjustments to CPD
- Apply active buffer adjustments to zones
- Store daily snapshots in BufferHistory
- Track buffer trends (expanding/contracting)

**Success Metrics**:
- <10s p95 for buffer adjustment application
- Daily recalculation completes in <5 minutes for 10,000 products
- 100% accuracy of FAD application in calculations
- Historical buffer data available for trend analysis

---

### US5: gRPC API for Buffer Queries (P1)

**As a** client service (Execution, Analytics)
**I want to** query buffer information via gRPC
**So that** I can make inventory decisions

**Acceptance Criteria**:
- GetBuffer RPC returns buffer zones for product
- CalculateBuffers RPC triggers recalculation
- GetNetFlowPosition RPC returns current NFP
- ListBufferAlerts RPC returns products needing replenishment
- All RPCs enforce multi-tenancy

**Success Metrics**:
- <50ms p50 for GetBuffer
- <100ms p50 for List operations

---

## Functional Requirements

### FR1: ADU Calculation Methods
- **Simple Average**: Sum(demand) / N days
- **Exponential Smoothing**: Weighted average with alpha parameter
- **Weighted Moving Average**: Recent data weighted higher
- Configurable lookback period (30, 60, 90 days)
- Outlier detection and handling

### FR2: Decoupled Lead Time (DLT)
- Cumulative lead time for decoupled items
- Includes: supplier lead time, manufacturing time, inspection time
- Updated from Catalog service supplier data
- Historical DLT tracking

### FR3: Buffer Zone Calculations

**Red Zone (Safety Stock)** - Composed of Base + Safe:
- **Red Base**: DLT × CPD × %LT
- **Red Safe**: Red Base × %CV
- **Total Red**: Red Base + Red Safe

**Yellow Zone (Demand Coverage)**:
- Formula: CPD × DLT

**Green Zone (Order Frequency)** - Takes the MAXIMUM of three values:
1. MOQ (Minimum Order Quantity from supplier)
2. FO × CPD (Order Frequency × Average Daily Consumption)
3. DLT × CPD × %LT

Where:
- **CPD**: Consumo Promedio Diario (Average Daily Consumption) - rounded up
- **DLT**: Lead Time Desacoplado (Decoupled Lead Time) in days
- **%LT**: Lead Time Factor from Buffer Profile Matrix (0.2 to 0.7)
- **%CV**: Variability Coefficient from Buffer Profile Matrix (0.25 to 1.0)
- **FO**: Order Frequency in days
- **MOQ**: Minimum Order Quantity (per supplier-product)

**Total Buffer**: Red + Yellow + Green
**Buffer Thresholds**:
- Top of Red = Red Zone
- Top of Yellow = Red + Yellow
- Top of Green = Red + Yellow + Green

### FR4: Net Flow Equation
- **On-Hand**: Current physical inventory
- **On-Order**: Open purchase/work orders
- **Qualified Demand**: Confirmed customer orders + forecasted demand
- **NFP**: On-Hand + On-Order - Qualified Demand
- **Buffer Penetration**: NFP position within zones

### FR5: Buffer Status & Alerts
- **Green Status**: NFP in Green zone (healthy)
- **Yellow Status**: NFP in Yellow zone (monitor)
- **Red Status**: NFP in Red zone (replenish now)
- **Stockout Risk**: NFP below Red zone (critical)
- Alert generation and notification

---

## Key Entities

### Buffer
```go
type Buffer struct {
    ID              uuid.UUID
    ProductID       uuid.UUID
    OrganizationID  uuid.UUID
    ADU             float64      // Average Daily Usage
    DLT             int          // Decoupled Lead Time (days)
    RedZoneQty      float64      // Safety stock quantity
    YellowZoneQty   float64      // Reorder quantity
    GreenZoneQty    float64      // Strategic excess
    TopOfRed        float64      // Red zone threshold
    TopOfYellow     float64      // Yellow zone threshold
    TopOfGreen      float64      // Green zone threshold (max inventory)
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

### BufferStatus
```go
type BufferStatus struct {
    ID               uuid.UUID
    BufferID         uuid.UUID
    ProductID        uuid.UUID
    OnHand           float64
    OnOrder          float64
    QualifiedDemand  float64
    NetFlowPosition  float64
    BufferPenetration float64 // Percentage (0-100)
    Zone             ZoneType // "green", "yellow", "red", "below_red"
    AlertLevel       AlertLevel // "none", "monitor", "replenish", "critical"
    Timestamp        time.Time
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
    AlertNone      AlertLevel = "none"
    AlertMonitor   AlertLevel = "monitor"
    AlertReplenish AlertLevel = "replenish"
    AlertCritical  AlertLevel = "critical"
)
```

### ADUHistory
```go
type ADUHistory struct {
    ID             uuid.UUID
    ProductID      uuid.UUID
    OrganizationID uuid.UUID
    ADU            float64
    Method         ADUMethod // "simple", "exponential", "weighted"
    Period         int       // Days
    CalculatedAt   time.Time
}
```

### DemandAdjustment (FAD - Factor de Ajuste de Demanda)
```go
type DemandAdjustment struct {
    ID              uuid.UUID
    ProductID       uuid.UUID
    OrganizationID  uuid.UUID
    StartDate       time.Time
    EndDate         time.Time
    AdjustmentType  DemandAdjustmentType
    Factor          float64        // Multiplier for CPD (e.g., 1.5 = 50% increase, 0.0 = discontinue)
    Reason          string
    CreatedAt       time.Time
    CreatedBy       uuid.UUID
}

type DemandAdjustmentType string

const (
    DemandAdjustmentFAD         DemandAdjustmentType = "fad"         // General FAD adjustment
    DemandAdjustmentSeasonal    DemandAdjustmentType = "seasonal"    // Seasonal variation
    DemandAdjustmentNewProduct  DemandAdjustmentType = "new_product" // New product launch
    DemandAdjustmentDiscontinue DemandAdjustmentType = "discontinue" // Product discontinuation (CPD → 0)
    DemandAdjustmentPromotion   DemandAdjustmentType = "promotion"   // Promotional period
)

// Adjusted CPD Calculation
// CPD_Adjusted = CPD_Original × Factor (for dates within StartDate and EndDate)
```

### BufferAdjustment
```go
type BufferAdjustment struct {
    ID              uuid.UUID
    BufferID        uuid.UUID
    ProductID       uuid.UUID
    OrganizationID  uuid.UUID
    AdjustmentType  BufferAdjustmentType
    TargetZone      ZoneType // Which zone to adjust ("red", "yellow", "green", "all")
    Factor          float64  // Multiplier (e.g., 1.2 = 20% increase)
    StartDate       time.Time
    EndDate         time.Time
    Reason          string
    CreatedAt       time.Time
    CreatedBy       uuid.UUID
}

type BufferAdjustmentType string

const (
    BufferAdjustmentZoneFactor     BufferAdjustmentType = "zone_factor"      // Manual zone adjustment
    BufferAdjustmentPlannedEvent   BufferAdjustmentType = "planned_event"    // Known future event
    BufferAdjustmentSpikeManagement BufferAdjustmentType = "spike_management" // Demand spike handling
    BufferAdjustmentSeasonalPrepare BufferAdjustmentType = "seasonal_prepare" // Pre-season buildup
)

// Buffer zones are multiplied by Factor during the adjustment period
// Example: Red Zone = (DLT × CPD × %LT) × Factor
```

### BufferHistory
```go
type BufferHistory struct {
    ID              uuid.UUID
    BufferID        uuid.UUID
    ProductID       uuid.UUID
    OrganizationID  uuid.UUID
    Date            time.Time
    CPD             float64      // Actual or adjusted CPD used
    DLT             int
    RedZone         float64
    RedBase         float64
    RedSafe         float64
    YellowZone      float64
    GreenZone       float64
    LeadTimeFactor  float64      // %LT used
    VariabilityFactor float64    // %CV used
    MOQ             int          // MOQ considered
    OrderFrequency  int          // FO used
    HasAdjustments  bool         // Whether FAD or buffer adjustments were applied
    CreatedAt       time.Time
}

// Enables daily buffer recalculation tracking and trend analysis
```

---

## Non-Functional Requirements

### Performance
- ADU calculation: <2s p95
- Buffer calculation: <3s p95
- Net Flow Position query: <500ms p95
- Batch buffer updates: <10s for 1000 products

### Accuracy
- Calculation precision: 0.01 units
- No rounding errors in cumulative calculations
- Match validated reference calculations 100%

### Reliability
- Graceful handling of missing data
- Fallback to last known values
- Retry failed calculations (3 attempts)
- Event-driven updates from Catalog/Execution services

### Scalability
- Support 10,000+ products per organization
- Support 100+ concurrent calculations
- Horizontal scaling via stateless design

---

## Success Criteria

### Mandatory (Must Have)
- ✅ ADU calculation with 3 methods (Simple, Exponential, Weighted)
- ✅ Buffer zone calculation (Red, Yellow, Green)
- ✅ Net Flow Position calculation
- ✅ Buffer status and alerts
- ✅ gRPC API for all operations
- ✅ Integration with Catalog service for product/supplier data
- ✅ Integration with Execution service for inventory levels
- ✅ Event-driven updates via NATS
- ✅ Multi-tenancy support
- ✅ 85%+ test coverage

### Optional (Nice to Have)
- ⚪ Machine learning for demand forecasting
- ⚪ Seasonal adjustment factors
- ⚪ Automatic buffer tuning
- ⚪ Buffer simulation and what-if analysis

---

## Out of Scope

- ❌ Order generation - Execution service responsibility
- ❌ Supplier selection - Catalog service responsibility
- ❌ Demand forecasting - AI Agent service (future)
- ❌ Cost optimization - Analytics service (future)

---

## Dependencies

- **Task 12**: Catalog service at 100% (product, supplier, buffer profile data)
- **Task 15**: Execution service (for inventory levels, orders) - Can mock initially
- **Shared Packages**: pkg/events, pkg/database, pkg/logger, pkg/errors
- **Infrastructure**: PostgreSQL, NATS Jetstream

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Calculation complexity and errors | Critical | Medium | Extensive testing, validation with domain experts |
| Performance at scale (10k+ products) | High | Medium | Batch processing, caching, database optimization |
| Data synchronization across services | High | Medium | Event-driven updates, eventual consistency |
| Missing or invalid input data | Medium | High | Fallback values, data validation, error handling |
| Domain knowledge gaps | High | Low | Consult DDMRP experts, reference implementations |

---

## References

- **DDMRP Book**: "Demand Driven Material Requirements Planning (DDMRP)" by Carol Ptak and Chad Smith
- **DDMRP Standard**: Demand Driven Institute specifications
- [Task 12 Spec](../task-12-catalog-service-integration/spec.md) - Catalog service integration
- [Task 15 Spec](../task-15-execution-service/spec.md) - Execution service integration

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)