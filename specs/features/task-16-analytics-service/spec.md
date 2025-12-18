# Task 16: Analytics Service - Specification

**Task ID**: task-16-analytics-service
**Phase**: 2B - New Microservices
**Priority**: P3 (Low - Reporting Layer)
**Estimated Duration**: 2-3 weeks
**Dependencies**: All operational services (Tasks 12, 14, 15)

---

## Overview

Implement the Analytics Service for reporting, dashboards, KPI calculations, and historical trend analysis. This service aggregates data from all other services, provides business intelligence, and enables data-driven decision making for inventory management.

---

## User Scenarios

### US1: Inventory Performance Dashboard (P1)

**As an** operations manager
**I want to** view inventory performance metrics
**So that** I can monitor overall inventory health

**Acceptance Criteria**:
- Display key metrics: inventory turnover, stockout rate, excess inventory percentage
- Buffer performance: % in Green, Yellow, Red zones
- Service level achievement
- Inventory value by category
- Real-time and historical views
- Configurable time periods (day, week, month, quarter, year)

**Success Metrics**:
- Dashboard loads in <3s
- Data refreshes every 5 minutes
- 100% data accuracy vs source systems

---

### US2: DDMRP Performance Reports (P1)

**As a** DDMRP planner
**I want to** analyze buffer performance
**So that** I can optimize buffer settings

**Acceptance Criteria**:
- Buffer penetration history charts
- Stockout events and root causes
- Buffer adjustment effectiveness
- ADU trend analysis
- Lead time variance analysis
- Replenishment order history

**Success Metrics**:
- Reports generate in <10s
- Export to PDF, Excel, CSV
- Historical data retention (1+ years)

---

### US3: Supplier Performance Analytics (P2)

**As a** procurement manager
**I want to** evaluate supplier performance
**So that** I can make informed sourcing decisions

**Acceptance Criteria**:
- On-time delivery percentage
- Lead time accuracy
- Order fill rate
- Quality metrics (returns, rejections)
- Cost trends
- Supplier comparison

**Success Metrics**:
- Identify top/bottom 20% suppliers
- Track month-over-month trends

---

### US4: Demand Pattern Analysis (P2)

**As a** demand planner
**I want to** analyze historical demand patterns
**So that** I can improve forecasting

**Acceptance Criteria**:
- Demand volume trends
- Seasonality detection
- Demand variability (coefficient of variation)
- ABC/XYZ classification
- Pareto analysis (80/20 rule)

**Success Metrics**:
- Identify seasonal products
- Classify products by value/variability

---

### US5: Custom Reports and Exports (P2)

**As a** business analyst
**I want to** create custom reports
**So that** I can answer specific business questions

**Acceptance Criteria**:
- Query builder for custom data extraction
- Saved report templates
- Scheduled report generation
- Export to multiple formats (PDF, Excel, CSV, JSON)
- Email delivery of reports

**Success Metrics**:
- Support 50+ custom report templates
- Scheduled reports deliver on time

---

## Functional Requirements

### FR1: Data Aggregation
- Subscribe to events from all services (Auth, Catalog, DDMRP, Execution)
- Aggregate data into analytics database (separate from operational DBs)
- Calculate derived metrics and KPIs
- Time-series data storage
- Data retention policies

### FR2: KPI Calculations

**Inventory Performance KPIs**:
- **Inventory Turnover** (Requirements): (Sales Last 30 Days) / (Average Monthly Stock)
- **Stockout Rate**: Stockout Days / Total Days
- **Service Level**: Orders Fulfilled on Time / Total Orders
- **Excess Inventory %**: (Inventory - Top of Green) / Total Inventory
- **Buffer Performance Score**: Weighted average of buffer health

**Inventory Valuation KPIs** [NEW]:
- **Days in Inventory (Valorizado)**: Sum of (Current Date - Purchase Date) × Product Cost for all products
- **Immobilized Inventory %**: (Value of Products > X years) / (Total Stock Value)
- **Immobilized Product Count**: Count of products with purchase date > X years ago
- **Total Inventory Value**: Sum of (Quantity × Standard Cost) for all products

**Product Analysis KPIs**:
- **ABC Classification**: Products by value contribution (A: 80%, B: 15%, C: 5%)
- **XYZ Classification**: Products by demand variability (X: low, Y: medium, Z: high)
- **Slow Moving Items**: Products with rotation < threshold

### FR3: Dashboard API
- REST and gRPC APIs for dashboard data
- Real-time metrics endpoints
- Historical trend endpoints
- Aggregation by time period, category, supplier
- Pagination for large datasets

### FR4: Report Generation
- PDF generation with charts and tables
- Excel export with multiple sheets
- CSV export for raw data
- Templating engine for report layouts
- Scheduling engine for automated reports

### FR5: Query Engine
- SQL-like query interface for custom analytics
- Pre-aggregated views for common queries
- Materialized views for performance
- Query result caching

---

## Key Entities

### KPISnapshot
```go
type KPISnapshot struct {
    ID                   uuid.UUID
    OrganizationID       uuid.UUID
    SnapshotDate         time.Time
    InventoryTurnover    float64
    StockoutRate         float64
    ServiceLevel         float64
    ExcessInventoryPct   float64
    BufferScoreGreen     float64 // % products in green
    BufferScoreYellow    float64 // % products in yellow
    BufferScoreRed       float64 // % products in red
    TotalInventoryValue  float64
    CreatedAt            time.Time
}
```

### BufferPerformance
```go
type BufferPerformance struct {
    ID                  uuid.UUID
    ProductID           uuid.UUID
    OrganizationID      uuid.UUID
    Date                time.Time
    BufferPenetration   float64
    Zone                string // "green", "yellow", "red"
    StockoutOccurred    bool
    ReplenishmentOrdered bool
}
```

### SupplierMetrics
```go
type SupplierMetrics struct {
    ID                  uuid.UUID
    SupplierID          uuid.UUID
    OrganizationID      uuid.UUID
    Period              string // "2025-Q4"
    OnTimeDeliveryPct   float64
    OrderFillRate       float64
    AverageLeadTime     float64
    LeadTimeVariance    float64
    TotalOrderValue     float64
}
```

### DaysInInventoryKPI [NEW]
```go
type DaysInInventoryKPI struct {
    ID                  uuid.UUID
    OrganizationID      uuid.UUID
    SnapshotDate        time.Time
    TotalValuedDays     float64   // Sum of (DaysInStock × UnitCost × Quantity)
    AverageValuedDays   float64   // TotalValuedDays / TotalProducts
    TotalProducts       int
    CreatedAt           time.Time
}

// Per-product detail
type ProductInventoryAge struct {
    ProductID           uuid.UUID
    SKU                 string
    Name                string
    Quantity            int
    PurchaseDate        time.Time
    DaysInInventory     int        // Current Date - Purchase Date
    UnitCost            float64
    TotalValue          float64    // Quantity × UnitCost
    ValuedDays          float64    // DaysInInventory × TotalValue
}

// Calculation:
// DaysInInventory = Current Date - LastPurchaseDate (from Product table)
// ValuedDays = DaysInInventory × (Quantity × StandardCost)
// TotalValuedDays = Sum(ValuedDays for all products)
```

### ImmobilizedInventoryKPI [NEW]
```go
type ImmobilizedInventoryKPI struct {
    ID                     uuid.UUID
    OrganizationID         uuid.UUID
    SnapshotDate           time.Time
    ThresholdYears         int        // Configurable (e.g., 1, 2, 3 years)
    ImmobilizedCount       int        // Products with age > threshold
    ImmobilizedValue       float64    // Total value of immobilized products
    TotalStockValue        float64    // Total value of all inventory
    ImmobilizedPercentage  float64    // (ImmobilizedValue / TotalStockValue) × 100
    CreatedAt              time.Time
}

// Immobilized products detail
type ImmobilizedProduct struct {
    ProductID       uuid.UUID
    SKU             string
    Name            string
    Category        string
    Quantity        int
    PurchaseDate    time.Time
    YearsInStock    float64   // (Current Date - Purchase Date) / 365
    UnitCost        float64
    TotalValue      float64   // Quantity × UnitCost
    LastSaleDate    *time.Time // For additional context
}

// Calculation:
// ImmobilizedProducts = Products WHERE (Current Date - PurchaseDate) > ThresholdYears
// ImmobilizedValue = Sum(Quantity × StandardCost) for immobilized products
// ImmobilizedPercentage = (ImmobilizedValue / TotalStockValue) × 100
```

### InventoryRotationKPI [NEW]
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

type RotatingProduct struct {
    ProductID       uuid.UUID
    SKU             string
    Name            string
    Sales30Days     float64
    AvgStockValue   float64
    RotationRatio   float64
}

// Calculation (per requirements):
// RotationRatio = (Sales Last 30 Days) / (Average Monthly Stock)
// Sales30Days = Sum(Sales Value) for last 30 days
// AvgMonthlyStock = Average(Daily Stock Value) over last 30 days
```

### BufferAnalytics [NEW]
```go
type BufferAnalytics struct {
    ID                  uuid.UUID
    ProductID           uuid.UUID
    OrganizationID      uuid.UUID
    Date                time.Time
    CPD                 float64   // Actual or adjusted CPD used
    RedZone             float64
    RedBase             float64
    RedSafe             float64
    YellowZone          float64
    GreenZone           float64
    LTD                 int
    LeadTimeFactor      float64   // %LT used
    VariabilityFactor   float64   // %CV used
    MOQ                 int       // MOQ considered
    OrderFrequency      int       // FO used
    OptimalOrderFreq    float64   // Green / CPD
    SafetyDays          float64   // Red / CPD
    AvgOpenOrders       float64   // Yellow / Green
    HasAdjustments      bool      // FAD or buffer adjustments applied
    CreatedAt           time.Time
}

// Enables daily buffer recalculation tracking and trend analysis
// Synchronized from DDMRP Engine BufferHistory
```

---

## Non-Functional Requirements

### Performance
- Dashboard load: <3s p95
- Report generation: <10s p95
- Query execution: <5s p95
- Data refresh: Every 5 minutes

### Data Volume
- Store 2+ years of historical data
- Support 1M+ data points per metric
- Efficient aggregation and rollup

### Reliability
- Event processing: At-least-once delivery
- Data consistency: Eventually consistent (5-minute lag acceptable)
- Backup and disaster recovery

---

## Success Criteria

### Mandatory (Must Have)
- ✅ Inventory performance dashboard
- ✅ DDMRP performance reports
- ✅ REST/gRPC API for metrics
- ✅ Event subscription from all services
- ✅ Data aggregation and KPI calculation
- ✅ Export to PDF, Excel, CSV
- ✅ Multi-tenancy support
- ✅ 80%+ test coverage

### Optional (Nice to Have)
- ⚪ Machine learning-based anomaly detection
- ⚪ Predictive analytics
- ⚪ Interactive data exploration UI
- ⚪ Real-time streaming dashboards

---

## Dependencies

- **All Services**: Auth, Catalog, DDMRP Engine, Execution (for event data)
- **Shared Packages**: pkg/events, pkg/database, pkg/logger
- **Infrastructure**: PostgreSQL (analytics DB), NATS Jetstream

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning