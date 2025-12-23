# Task 15: Execution Service - Specification

**Task ID**: task-15-execution-service
**Phase**: 2B - New Microservices
**Priority**: P2 (Medium)
**Estimated Duration**: 2-3 weeks
**Dependencies**: Task 12 (Catalog at 100%), Task 14 (DDMRP Engine)

---

## Overview

Implement the Execution Service for order management, inventory transactions, and stock movements. This service handles purchase orders, work orders, sales orders, inventory adjustments, and integrates with DDMRP Engine for replenishment decisions and with Catalog service for product data.

---

## User Scenarios

### US1: Purchase Order Management (P1)

**As an** inventory manager
**I want to** create and manage purchase orders
**So that** I can replenish inventory from suppliers

**Acceptance Criteria**:
- Create purchase orders with supplier, products, quantities, expected dates
- Track order status: pending, confirmed, shipped, received, cancelled
- Receive inventory and automatically update stock levels
- Partial receipts supported
- Cancel or modify pending orders
- Integration with Catalog service for product/supplier data
- Multi-tenancy and audit logging

**Success Metrics**:
- <3s p95 for PO creation
- <2s p95 for inventory receipt
- 100% inventory accuracy after receipt

---

### US2: Inventory Transactions (P1)

**As an** warehouse operator
**I want to** record inventory movements
**So that** stock levels remain accurate

**Acceptance Criteria**:
- Stock receipts: purchases, production completions, returns
- Stock issues: sales, production consumption, adjustments
- Stock transfers between locations
- Cycle count adjustments
- Transaction history with timestamps and reasons
- Real-time balance calculation (on-hand quantity)
- Notify DDMRP Engine of inventory changes via events

**Success Metrics**:
- <1s p95 for transaction recording
- 99.9% inventory accuracy
- Real-time updates to DDMRP Engine

---

### US3: Replenishment Recommendations (P1)

**As an** inventory planner
**I want to** receive replenishment recommendations
**So that** I know which products to order and in what quantities

**Acceptance Criteria**:
- Query DDMRP Engine for products in Red/Yellow zones
- Calculate recommended order quantities based on buffer gaps
- Consider minimum order quantities (MOQ) from supplier
- Consider order lot sizes and multiples
- Generate purchase order drafts
- Support manual approval or automatic execution
- Optimize orders across multiple products from same supplier

**Success Metrics**:
- <5s p95 for replenishment calculation
- 30% reduction in stockouts
- 20% reduction in excess inventory

---

### US4: Sales Order Integration (P2)

**As a** sales manager
**I want to** create sales orders
**So that** I can fulfill customer demand

**Acceptance Criteria**:
- Create sales orders with customer, products, quantities, due dates
- Reserve inventory for confirmed orders
- Update qualified demand in DDMRP Engine
- Track fulfillment status: pending, picking, packed, shipped, delivered
- Back-order management for out-of-stock items
- Cancel or modify orders

**Success Metrics**:
- <3s p95 for sales order creation
- 95%+ on-time fulfillment rate
- Accurate demand signal to DDMRP Engine

---

### US5: Stock Location Management (P2)

**As a** warehouse manager
**I want to** manage multiple stock locations
**So that** I can track inventory across warehouses, zones, bins

**Acceptance Criteria**:
- Create and manage locations (warehouse, zone, aisle, bin)
- Track inventory by location
- Transfer stock between locations
- Location-specific replenishment
- Multi-location visibility

**Success Metrics**:
- Support 100+ locations per organization
- <2s p95 for stock transfers

---

## Functional Requirements

### FR1: Purchase Order Lifecycle
- Create PO with line items (product, quantity, unit cost, expected date)
- Confirm PO (sent to supplier)
- Receive PO (full or partial)
- Close PO when complete
- Cancel PO if needed
- PO history and versioning

### FR2: Inventory Transaction Types
- **Receipt**: Purchase receipt, production completion, customer return
- **Issue**: Sale, production consumption, damage/loss
- **Transfer**: Move between locations
- **Adjustment**: Cycle count, system correction
- **Reservation**: Allocate for sales orders
- Transaction locking to prevent double-counting

### FR3: Inventory Balance Calculation
- On-Hand = Previous Balance + Receipts - Issues ± Adjustments
- Available = On-Hand - Reserved
- On-Order = Sum of open purchase orders
- Allocated = Sum of sales order reservations

### FR4: Replenishment Logic
- Query DDMRP buffers with alert status
- Calculate order quantity: Top of Green - Net Flow Position
- Apply MOQ, lot size, rounding rules
- Aggregate orders by supplier
- Generate PO recommendations or auto-create POs

### FR5: Integration Points
- **DDMRP Engine**: Query buffer status, update NFP on inventory changes
- **Catalog Service**: Get product details, supplier info, pricing
- **Auth Service**: Validate tokens, check permissions
- **Events**: Publish order.created, inventory.updated, replenishment.needed

---

## Key Entities

### PurchaseOrder
```go
type PurchaseOrder struct {
    ID                  uuid.UUID
    OrganizationID      uuid.UUID
    PONumber            string      // Unique PO number
    SupplierID          uuid.UUID
    Status              POStatus    // "draft", "pending", "confirmed", "received", "closed", "cancelled"
    OrderDate           time.Time
    ExpectedArrivalDate time.Time   // Expected delivery date [UPDATED]
    ActualArrivalDate   *time.Time  // Actual delivery date [NEW]
    DelayDays           int         // [NEW] Calculated: ActualArrivalDate - ExpectedArrivalDate (negative if early)
    IsDelayed           bool        // [NEW] True if current date > ExpectedArrivalDate and not received
    TotalAmount         float64
    LineItems           []POLineItem
    CreatedBy           uuid.UUID
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

type POLineItem struct {
    ID              uuid.UUID
    PurchaseOrderID uuid.UUID
    ProductID       uuid.UUID
    Quantity        float64
    ReceivedQty     float64
    UnitCost        float64
    LineTotal       float64
}

type POStatus string

const (
    POStatusDraft     POStatus = "draft"
    POStatusPending   POStatus = "pending"
    POStatusConfirmed POStatus = "confirmed"
    POStatusPartial   POStatus = "partial"  // [NEW] Partially received
    POStatusReceived  POStatus = "received"
    POStatusClosed    POStatus = "closed"
    POStatusCancelled POStatus = "cancelled"
)
```

### InventoryTransaction
```go
type InventoryTransaction struct {
    ID              uuid.UUID
    OrganizationID  uuid.UUID
    ProductID       uuid.UUID
    LocationID      uuid.UUID
    Type            TransactionType
    Quantity        float64     // Positive for receipt, negative for issue
    UnitCost        float64
    ReferenceType   string      // "purchase_order", "sales_order", "adjustment"
    ReferenceID     uuid.UUID
    Reason          string
    TransactionDate time.Time
    CreatedBy       uuid.UUID
    CreatedAt       time.Time
}

type TransactionType string

const (
    TransactionReceipt     TransactionType = "receipt"
    TransactionIssue       TransactionType = "issue"
    TransactionTransfer    TransactionType = "transfer"
    TransactionAdjustment  TransactionType = "adjustment"
)
```

### InventoryBalance
```go
type InventoryBalance struct {
    ID             uuid.UUID
    OrganizationID uuid.UUID
    ProductID      uuid.UUID
    LocationID     uuid.UUID
    OnHand         float64
    Reserved       float64
    Available      float64
    UpdatedAt      time.Time
}
```

### SalesOrder
```go
type SalesOrder struct {
    ID                  uuid.UUID
    OrganizationID      uuid.UUID
    SONumber            string
    CustomerID          uuid.UUID
    Status              SOStatus // "pending", "confirmed", "picking", "packed", "shipped", "delivered"
    OrderDate           time.Time
    DueDate             time.Time
    ShipDate            *time.Time
    DeliveryNoteIssued  bool       // [NEW] Has delivery note (remito) been issued?
    DeliveryNoteNumber  string     // [NEW] Remito number
    DeliveryNoteDate    *time.Time // [NEW] Date of delivery note
    TotalAmount         float64
    LineItems           []SOLineItem
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

type SOLineItem struct {
    ID           uuid.UUID
    SalesOrderID uuid.UUID
    ProductID    uuid.UUID
    Quantity     float64
    UnitPrice    float64
    LineTotal    float64
}

type SOStatus string

const (
    SOStatusPending   SOStatus = "pending"
    SOStatusConfirmed SOStatus = "confirmed"
    SOStatusPicking   SOStatus = "picking"
    SOStatusPacked    SOStatus = "packed"
    SOStatusShipped   SOStatus = "shipped"
    SOStatusDelivered SOStatus = "delivered"
)

// Qualified Demand calculation:
// - Confirmed sales orders WITHOUT delivery note = firm demand
// - Orders WITH delivery note = already fulfilled, not counted as demand
```

### ReplenishmentRecommendation
```go
type ReplenishmentRecommendation struct {
    ID               uuid.UUID
    OrganizationID   uuid.UUID
    ProductID        uuid.UUID
    SupplierID       uuid.UUID
    RecommendedQty   float64
    BufferStatus     string // "red", "yellow"
    NetFlowPosition  float64
    TargetLevel      float64 // Top of Green
    Priority         int     // 1 (critical) to 5 (low)
    GeneratedAt      time.Time
}
```

### Alert [NEW]
```go
type Alert struct {
    ID              uuid.UUID
    OrganizationID  uuid.UUID
    AlertType       AlertType
    Severity        AlertSeverity
    ResourceType    string      // "purchase_order", "sales_order", "buffer", "inventory", "supplier"
    ResourceID      uuid.UUID
    Title           string      // Short alert title
    Message         string      // Detailed alert message
    Data            map[string]interface{} // Additional context data
    AcknowledgedAt  *time.Time
    AcknowledgedBy  *uuid.UUID
    ResolvedAt      *time.Time
    ResolvedBy      *uuid.UUID
    CreatedAt       time.Time
}

type AlertType string

const (
    // Purchase Order Alerts (Requirements: "Debe avisar al usuario si se detecta un desvío en la llegada de la mercancía")
    AlertTypePODelayed          AlertType = "po_delayed"           // PO past expected arrival date
    AlertTypePOLateWarning      AlertType = "po_late_warning"      // PO approaching expected date

    // Buffer Alerts
    AlertTypeBufferRed          AlertType = "buffer_red"           // NFP in red zone
    AlertTypeBufferBelowRed     AlertType = "buffer_below_red"     // NFP below red zone (critical stockout risk)
    AlertTypeBufferStockout     AlertType = "buffer_stockout"      // Actual stockout occurred

    // Inventory Alerts
    AlertTypeStockDeviation     AlertType = "stock_deviation"      // Unexpected stock variance
    AlertTypeObsolescenceRisk   AlertType = "obsolescence_risk"    // Product aging beyond threshold
    AlertTypeExcessInventory    AlertType = "excess_inventory"     // Inventory well above top of green

    // Supplier Alerts
    AlertTypeSupplierDelayPattern AlertType = "supplier_delay_pattern" // Supplier consistently late
)

type AlertSeverity string

const (
    AlertSeverityInfo     AlertSeverity = "info"      // Informational
    AlertSeverityLow      AlertSeverity = "low"       // Low priority
    AlertSeverityMedium   AlertSeverity = "medium"    // Medium priority
    AlertSeverityHigh     AlertSeverity = "high"      // High priority - action needed
    AlertSeverityCritical AlertSeverity = "critical"  // Critical - urgent action required
)

// Alert generation examples:
// 1. PO Delayed: Current Date > ExpectedArrivalDate && Status NOT IN (received, closed, cancelled)
// 2. PO Late Warning: (ExpectedArrivalDate - Current Date) <= 3 days && Status = confirmed
// 3. Buffer Red: Zone = "red" && AlertLevel = "replenish"
// 4. Stockout: Zone = "below_red" && OnHand = 0
```

---

## Non-Functional Requirements

### Performance
- PO/SO creation: <3s p95
- Inventory transaction: <1s p95
- Balance query: <500ms p95
- Replenishment calculation: <5s p95

### Accuracy
- 99.9% inventory accuracy
- Zero double-counting of transactions
- Transactional consistency for balance updates

### Reliability
- ACID transactions for inventory changes
- Event publishing retry logic
- Graceful degradation if DDMRP Engine unavailable

### Scalability
- Support 50,000+ products
- Support 10,000+ transactions/day
- Support 1,000+ concurrent orders

---

## Success Criteria

### Mandatory (Must Have)
- ✅ Purchase order CRUD operations
- ✅ Sales order CRUD operations
- ✅ Inventory transaction recording (receipts, issues, adjustments)
- ✅ Real-time inventory balance calculation
- ✅ Replenishment recommendations from DDMRP Engine
- ✅ gRPC API for all operations
- ✅ Integration with Catalog service (products, suppliers)
- ✅ Integration with DDMRP Engine (buffer status, NFP updates)
- ✅ Event publishing for all state changes
- ✅ Multi-tenancy support
- ✅ 85%+ test coverage

### Optional (Nice to Have)
- ⚪ Automatic PO generation from replenishment
- ⚪ Multi-warehouse routing
- ⚪ Barcode scanning integration
- ⚪ Shipping integration (FedEx, UPS APIs)

---

## Out of Scope

- ❌ Supplier portal - Future task
- ❌ Customer portal - Future task
- ❌ Advanced warehouse management (pick/pack optimization) - Future WMS module
- ❌ Transportation management - Future TMS module

---

## Dependencies

- **Task 12**: Catalog service at 100%
- **Task 14**: DDMRP Engine service (for buffer queries)
- **Shared Packages**: pkg/events, pkg/database, pkg/logger, pkg/errors
- **Infrastructure**: PostgreSQL, NATS Jetstream

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Inventory balance discrepancies | Critical | Medium | Transactional updates, audit logs, reconciliation |
| Performance with high transaction volume | High | Medium | Database optimization, caching, batch processing |
| Integration failures with DDMRP Engine | High | Medium | Retry logic, fallback to last known values |
| Concurrent transaction conflicts | Medium | High | Optimistic locking, database transactions |

---

## References

- [Task 12 Spec](../task-12-catalog-service-integration/spec.md) - Catalog integration
- [Task 14 Spec](../task-14-ddmrp-engine-service/spec.md) - DDMRP Engine integration
- **Inventory Management Best Practices**: APICS CPIM standards

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)