# Task 15: Execution Service - Implementation Plan

**Task ID**: task-15-execution-service
**Phase**: 2B - New Microservices
**Priority**: P2 (Medium)
**Estimated Duration**: 2-3 weeks
**Dependencies**: Task 12 (Catalog at 100%), Task 14 (DDMRP Engine)

---

## 1. Technical Context

### Current State
- **Execution Service**: Not yet implemented (new service)
- **Catalog Service**: 100% complete (products, suppliers)
- **DDMRP Engine**: Complete with buffer calculations and NFP

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure)
- **Database**: PostgreSQL 16 with GORM
- **gRPC**: Protocol Buffers v3
- **Event Streaming**: NATS JetStream
- **Testing**: testify, httptest, gRPC testing framework

### Key Design Decisions
- **ACID Transactions**: Inventory balance updates use database transactions
- **Event-Driven**: Publish events for inventory changes → update DDMRP NFP
- **Alert System**: Generate alerts for PO delays, buffer status, stockouts
- **Delivery Notes (Remitos)**: Track order fulfillment and qualified demand
- **Multi-tenancy**: organization_id filtering at all layers

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
└── services/execution-service/
    ├── api/
    │   └── proto/
    │       └── execution/
    │           └── v1/
    │               ├── execution.proto                    [NEW]
    │               ├── execution.pb.go                    [GENERATED]
    │               └── execution_grpc.pb.go               [GENERATED]
    │
    ├── cmd/
    │   └── main.go                                         [NEW]
    │
    ├── internal/
    │   ├── core/
    │   │   ├── domain/
    │   │   │   ├── purchase_order.go                      [NEW]
    │   │   │   ├── sales_order.go                         [NEW]
    │   │   │   ├── inventory_transaction.go               [NEW]
    │   │   │   ├── inventory_balance.go                   [NEW]
    │   │   │   ├── alert.go                               [NEW]
    │   │   │   ├── replenishment_recommendation.go        [NEW]
    │   │   │   └── errors.go                              [NEW]
    │   │   │
    │   │   ├── providers/
    │   │   │   ├── purchase_order_repository.go           [NEW]
    │   │   │   ├── sales_order_repository.go              [NEW]
    │   │   │   ├── inventory_transaction_repository.go    [NEW]
    │   │   │   ├── inventory_balance_repository.go        [NEW]
    │   │   │   ├── alert_repository.go                    [NEW]
    │   │   │   ├── ddmrp_service_client.go                [NEW]
    │   │   │   ├── catalog_service_client.go              [NEW]
    │   │   │   └── event_publisher.go                     [NEW]
    │   │   │
    │   │   └── usecases/
    │   │       ├── purchase_order/
    │   │       │   ├── create_po.go                       [NEW]
    │   │       │   ├── receive_po.go                      [NEW]
    │   │       │   ├── cancel_po.go                       [NEW]
    │   │       │   ├── list_pos.go                        [NEW]
    │   │       │   └── check_po_delays.go                 [NEW]
    │   │       │
    │   │       ├── sales_order/
    │   │       │   ├── create_so.go                       [NEW]
    │   │       │   ├── issue_delivery_note.go             [NEW]
    │   │       │   ├── cancel_so.go                       [NEW]
    │   │       │   └── list_sos.go                        [NEW]
    │   │       │
    │   │       ├── inventory/
    │   │       │   ├── record_transaction.go              [NEW]
    │   │       │   ├── get_balance.go                     [NEW]
    │   │       │   ├── transfer_stock.go                  [NEW]
    │   │       │   └── adjust_stock.go                    [NEW]
    │   │       │
    │   │       ├── alert/
    │   │       │   ├── generate_alerts.go                 [NEW]
    │   │       │   ├── acknowledge_alert.go               [NEW]
    │   │       │   └── list_alerts.go                     [NEW]
    │   │       │
    │   │       └── replenishment/
    │   │           ├── get_recommendations.go             [NEW]
    │   │           └── create_po_from_recommendation.go   [NEW]
    │   │
    │   └── infrastructure/
    │       ├── adapters/
    │       │   ├── ddmrp/
    │       │   │   ├── grpc_ddmrp_client.go               [NEW]
    │       │   │   └── ddmrp_client_mock.go               [NEW]
    │       │   │
    │       │   ├── catalog/
    │       │   │   ├── grpc_catalog_client.go             [NEW]
    │       │   │   └── catalog_client_mock.go             [NEW]
    │       │   │
    │       │   └── events/
    │       │       ├── nats_publisher.go                  [NEW]
    │       │       └── publisher_mock.go                  [NEW]
    │       │
    │       ├── repositories/
    │       │   ├── purchase_order_repository.go           [NEW]
    │       │   ├── sales_order_repository.go              [NEW]
    │       │   ├── inventory_transaction_repository.go    [NEW]
    │       │   ├── inventory_balance_repository.go        [NEW]
    │       │   └── alert_repository.go                    [NEW]
    │       │
    │       ├── entrypoints/
    │       │   ├── grpc/
    │       │   │   ├── server.go                          [NEW]
    │       │   │   ├── po_handler.go                      [NEW]
    │       │   │   ├── so_handler.go                      [NEW]
    │       │   │   ├── inventory_handler.go               [NEW]
    │       │   │   └── alert_handler.go                   [NEW]
    │       │   │
    │       │   └── cron/
    │       │       └── alert_checker.go                   [NEW]
    │       │
    │       └── database/
    │           └── migrations/
    │               ├── 000001_create_purchase_orders.up.sql      [NEW]
    │               ├── 000002_create_sales_orders.up.sql         [NEW]
    │               ├── 000003_create_inventory_transactions.up.sql [NEW]
    │               ├── 000004_create_inventory_balances.up.sql   [NEW]
    │               └── 000005_create_alerts.up.sql               [NEW]
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

**File**: `services/execution-service/migrations/000001_create_purchase_orders.up.sql`

```sql
-- Purchase Orders table
CREATE TABLE IF NOT EXISTS purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    po_number VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    order_date TIMESTAMP NOT NULL,
    expected_arrival_date TIMESTAMP NOT NULL,
    actual_arrival_date TIMESTAMP,
    delay_days INTEGER DEFAULT 0,
    is_delayed BOOLEAN NOT NULL DEFAULT FALSE,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_po_number_org UNIQUE (organization_id, po_number),
    CONSTRAINT chk_po_status CHECK (status IN ('draft', 'pending', 'confirmed', 'partial', 'received', 'closed', 'cancelled'))
);

CREATE INDEX idx_po_org ON purchase_orders(organization_id);
CREATE INDEX idx_po_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_expected_date ON purchase_orders(expected_arrival_date);

-- Purchase Order Line Items table
CREATE TABLE IF NOT EXISTS po_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity DECIMAL(15,2) NOT NULL,
    received_qty DECIMAL(15,2) NOT NULL DEFAULT 0,
    unit_cost DECIMAL(15,2) NOT NULL,
    line_total DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_po_line_po ON po_line_items(purchase_order_id);
CREATE INDEX idx_po_line_product ON po_line_items(product_id);
```

**File**: `services/execution-service/migrations/000002_create_sales_orders.up.sql`

```sql
-- Sales Orders table
CREATE TABLE IF NOT EXISTS sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    so_number VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_date TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    ship_date TIMESTAMP,
    delivery_note_issued BOOLEAN NOT NULL DEFAULT FALSE,
    delivery_note_number VARCHAR(50),
    delivery_note_date TIMESTAMP,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_so_number_org UNIQUE (organization_id, so_number),
    CONSTRAINT chk_so_status CHECK (status IN ('pending', 'confirmed', 'picking', 'packed', 'shipped', 'delivered'))
);

CREATE INDEX idx_so_org ON sales_orders(organization_id);
CREATE INDEX idx_so_customer ON sales_orders(customer_id);
CREATE INDEX idx_so_status ON sales_orders(status);
CREATE INDEX idx_so_delivery_note ON sales_orders(delivery_note_issued);

-- Sales Order Line Items table
CREATE TABLE IF NOT EXISTS so_line_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sales_order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity DECIMAL(15,2) NOT NULL,
    unit_price DECIMAL(15,2) NOT NULL,
    line_total DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_so_line_so ON so_line_items(sales_order_id);
CREATE INDEX idx_so_line_product ON so_line_items(product_id);
```

**File**: `services/execution-service/migrations/000003_create_inventory_transactions.up.sql`

```sql
-- Inventory Transactions table
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    quantity DECIMAL(15,2) NOT NULL, -- Positive for receipt, negative for issue
    unit_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    reference_type VARCHAR(50),
    reference_id UUID,
    reason TEXT,
    transaction_date TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_txn_type CHECK (transaction_type IN ('receipt', 'issue', 'transfer', 'adjustment'))
);

CREATE INDEX idx_inv_txn_org ON inventory_transactions(organization_id);
CREATE INDEX idx_inv_txn_product ON inventory_transactions(product_id);
CREATE INDEX idx_inv_txn_location ON inventory_transactions(location_id);
CREATE INDEX idx_inv_txn_date ON inventory_transactions(transaction_date DESC);
CREATE INDEX idx_inv_txn_ref ON inventory_transactions(reference_type, reference_id);
```

**File**: `services/execution-service/migrations/000004_create_inventory_balances.up.sql`

```sql
-- Inventory Balances table
CREATE TABLE IF NOT EXISTS inventory_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    product_id UUID NOT NULL,
    location_id UUID NOT NULL,
    on_hand DECIMAL(15,2) NOT NULL DEFAULT 0,
    reserved DECIMAL(15,2) NOT NULL DEFAULT 0,
    available DECIMAL(15,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_balance_product_location UNIQUE (organization_id, product_id, location_id)
);

CREATE INDEX idx_inv_balance_org ON inventory_balances(organization_id);
CREATE INDEX idx_inv_balance_product ON inventory_balances(product_id);
CREATE INDEX idx_inv_balance_location ON inventory_balances(location_id);
```

**File**: `services/execution-service/migrations/000005_create_alerts.up.sql`

```sql
-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID NOT NULL,
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    acknowledged_at TIMESTAMP,
    acknowledged_by UUID,
    resolved_at TIMESTAMP,
    resolved_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_alert_type CHECK (alert_type IN (
        'po_delayed', 'po_late_warning',
        'buffer_red', 'buffer_below_red', 'buffer_stockout',
        'stock_deviation', 'obsolescence_risk', 'excess_inventory',
        'supplier_delay_pattern'
    )),
    CONSTRAINT chk_alert_severity CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical'))
);

CREATE INDEX idx_alerts_org ON alerts(organization_id);
CREATE INDEX idx_alerts_type ON alerts(alert_type);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_resource ON alerts(resource_type, resource_id);
CREATE INDEX idx_alerts_acknowledged ON alerts(acknowledged_at) WHERE acknowledged_at IS NULL;
CREATE INDEX idx_alerts_resolved ON alerts(resolved_at) WHERE resolved_at IS NULL;
```

---

### Phase 2: Core Domain Entities (Week 1 Days 4-5)

#### T002: Domain Entities

**File**: `services/execution-service/internal/core/domain/purchase_order.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseOrder struct {
	ID                  uuid.UUID
	OrganizationID      uuid.UUID
	PONumber            string
	SupplierID          uuid.UUID
	Status              POStatus
	OrderDate           time.Time
	ExpectedArrivalDate time.Time
	ActualArrivalDate   *time.Time
	DelayDays           int
	IsDelayed           bool
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
	POStatusPartial   POStatus = "partial"
	POStatusReceived  POStatus = "received"
	POStatusClosed    POStatus = "closed"
	POStatusCancelled POStatus = "cancelled"
)

func (s POStatus) IsValid() bool {
	switch s {
	case POStatusDraft, POStatusPending, POStatusConfirmed, POStatusPartial,
		POStatusReceived, POStatusClosed, POStatusCancelled:
		return true
	}
	return false
}

func NewPurchaseOrder(
	orgID, supplierID, createdBy uuid.UUID,
	poNumber string,
	orderDate, expectedArrivalDate time.Time,
	lineItems []POLineItem,
) (*PurchaseOrder, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if supplierID == uuid.Nil {
		return nil, NewValidationError("supplier_id is required")
	}
	if poNumber == "" {
		return nil, NewValidationError("po_number is required")
	}
	if len(lineItems) == 0 {
		return nil, NewValidationError("at least one line item is required")
	}

	totalAmount := 0.0
	for _, item := range lineItems {
		totalAmount += item.LineTotal
	}

	return &PurchaseOrder{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		PONumber:            poNumber,
		SupplierID:          supplierID,
		Status:              POStatusDraft,
		OrderDate:           orderDate,
		ExpectedArrivalDate: expectedArrivalDate,
		TotalAmount:         totalAmount,
		LineItems:           lineItems,
		CreatedBy:           createdBy,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}

func (po *PurchaseOrder) CheckDelay() {
	if po.ActualArrivalDate != nil {
		po.DelayDays = int(po.ActualArrivalDate.Sub(po.ExpectedArrivalDate).Hours() / 24)
		po.IsDelayed = po.DelayDays > 0
	} else if time.Now().After(po.ExpectedArrivalDate) &&
		po.Status != POStatusReceived &&
		po.Status != POStatusClosed &&
		po.Status != POStatusCancelled {
		po.IsDelayed = true
		po.DelayDays = int(time.Since(po.ExpectedArrivalDate).Hours() / 24)
	}
}
```

**File**: `services/execution-service/internal/core/domain/sales_order.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type SalesOrder struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	SONumber           string
	CustomerID         uuid.UUID
	Status             SOStatus
	OrderDate          time.Time
	DueDate            time.Time
	ShipDate           *time.Time
	DeliveryNoteIssued bool
	DeliveryNoteNumber string
	DeliveryNoteDate   *time.Time
	TotalAmount        float64
	LineItems          []SOLineItem
	CreatedAt          time.Time
	UpdatedAt          time.Time
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

func (s SOStatus) IsValid() bool {
	switch s {
	case SOStatusPending, SOStatusConfirmed, SOStatusPicking,
		SOStatusPacked, SOStatusShipped, SOStatusDelivered:
		return true
	}
	return false
}

// IsQualifiedDemand returns true if this SO contributes to qualified demand
// Qualified Demand = Confirmed orders WITHOUT delivery note
func (so *SalesOrder) IsQualifiedDemand() bool {
	return so.Status == SOStatusConfirmed && !so.DeliveryNoteIssued
}

func (so *SalesOrder) IssueDeliveryNote(noteNumber string) error {
	if so.DeliveryNoteIssued {
		return NewValidationError("delivery note already issued")
	}
	if noteNumber == "" {
		return NewValidationError("delivery note number is required")
	}

	now := time.Now()
	so.DeliveryNoteIssued = true
	so.DeliveryNoteNumber = noteNumber
	so.DeliveryNoteDate = &now
	so.UpdatedAt = now

	return nil
}
```

**File**: `services/execution-service/internal/core/domain/alert.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	AlertType      AlertType
	Severity       AlertSeverity
	ResourceType   string
	ResourceID     uuid.UUID
	Title          string
	Message        string
	Data           map[string]interface{}
	AcknowledgedAt *time.Time
	AcknowledgedBy *uuid.UUID
	ResolvedAt     *time.Time
	ResolvedBy     *uuid.UUID
	CreatedAt      time.Time
}

type AlertType string

const (
	// Purchase Order Alerts
	AlertTypePODelayed     AlertType = "po_delayed"
	AlertTypePOLateWarning AlertType = "po_late_warning"

	// Buffer Alerts
	AlertTypeBufferRed      AlertType = "buffer_red"
	AlertTypeBufferBelowRed AlertType = "buffer_below_red"
	AlertTypeBufferStockout AlertType = "buffer_stockout"

	// Inventory Alerts
	AlertTypeStockDeviation   AlertType = "stock_deviation"
	AlertTypeObsolescenceRisk AlertType = "obsolescence_risk"
	AlertTypeExcessInventory  AlertType = "excess_inventory"

	// Supplier Alerts
	AlertTypeSupplierDelayPattern AlertType = "supplier_delay_pattern"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

func NewPODelayedAlert(po *PurchaseOrder) *Alert {
	return &Alert{
		ID:             uuid.New(),
		OrganizationID: po.OrganizationID,
		AlertType:      AlertTypePODelayed,
		Severity:       AlertSeverityHigh,
		ResourceType:   "purchase_order",
		ResourceID:     po.ID,
		Title:          "Purchase Order Delayed",
		Message:        "PO " + po.PONumber + " is delayed by " + string(po.DelayDays) + " days",
		Data: map[string]interface{}{
			"po_number":             po.PONumber,
			"supplier_id":           po.SupplierID.String(),
			"expected_arrival_date": po.ExpectedArrivalDate,
			"delay_days":            po.DelayDays,
		},
		CreatedAt: time.Now(),
	}
}
```

---

### Phase 3: Use Cases (Week 2)

#### T003: Purchase Order Use Cases

**File**: `services/execution-service/internal/core/usecases/purchase_order/create_po.go`

```go
package purchase_order

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/execution-service/internal/core/domain"
	"giia-core-engine/services/execution-service/internal/core/providers"
)

type CreatePOUseCase struct {
	poRepo         providers.PurchaseOrderRepository
	catalogClient  providers.CatalogServiceClient
	eventPublisher providers.EventPublisher
}

func NewCreatePOUseCase(
	poRepo providers.PurchaseOrderRepository,
	catalogClient providers.CatalogServiceClient,
	publisher providers.EventPublisher,
) *CreatePOUseCase {
	return &CreatePOUseCase{
		poRepo:         poRepo,
		catalogClient:  catalogClient,
		eventPublisher: publisher,
	}
}

type CreatePOInput struct {
	OrganizationID      uuid.UUID
	PONumber            string
	SupplierID          uuid.UUID
	OrderDate           time.Time
	ExpectedArrivalDate time.Time
	LineItems           []domain.POLineItem
	CreatedBy           uuid.UUID
}

func (uc *CreatePOUseCase) Execute(ctx context.Context, input CreatePOInput) (*domain.PurchaseOrder, error) {
	// 1. Validate supplier exists
	_, err := uc.catalogClient.GetSupplier(ctx, input.SupplierID)
	if err != nil {
		return nil, domain.NewValidationError("invalid supplier_id")
	}

	// 2. Validate all products exist
	for _, item := range input.LineItems {
		_, err := uc.catalogClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return nil, domain.NewValidationError("invalid product_id: " + item.ProductID.String())
		}
	}

	// 3. Create purchase order
	po, err := domain.NewPurchaseOrder(
		input.OrganizationID,
		input.SupplierID,
		input.CreatedBy,
		input.PONumber,
		input.OrderDate,
		input.ExpectedArrivalDate,
		input.LineItems,
	)
	if err != nil {
		return nil, err
	}

	// 4. Save to repository
	if err := uc.poRepo.Create(ctx, po); err != nil {
		return nil, err
	}

	// 5. Publish event
	uc.eventPublisher.PublishPOCreated(ctx, po)

	return po, nil
}
```

**File**: `services/execution-service/internal/core/usecases/purchase_order/receive_po.go`

```go
package purchase_order

import (
	"context"
	"time"

	"github.com/google/uuid"
	"giia-core-engine/services/execution-service/internal/core/domain"
	"giia-core-engine/services/execution-service/internal/core/providers"
)

type ReceivePOUseCase struct {
	poRepo               providers.PurchaseOrderRepository
	inventoryTxnRepo     providers.InventoryTransactionRepository
	inventoryBalanceRepo providers.InventoryBalanceRepository
	ddmrpClient          providers.DDMRPServiceClient
	eventPublisher       providers.EventPublisher
}

type ReceivePOInput struct {
	POID           uuid.UUID
	OrganizationID uuid.UUID
	Receipts       []struct {
		LineItemID  uuid.UUID
		ReceivedQty float64
	}
	ReceivedBy uuid.UUID
}

func (uc *ReceivePOUseCase) Execute(ctx context.Context, input ReceivePOInput) (*domain.PurchaseOrder, error) {
	// 1. Get PO
	po, err := uc.poRepo.GetByID(ctx, input.POID, input.OrganizationID)
	if err != nil {
		return nil, err
	}

	// 2. Record inventory transactions (ACID transaction)
	for _, receipt := range input.Receipts {
		// Find line item
		var lineItem *domain.POLineItem
		for i := range po.LineItems {
			if po.LineItems[i].ID == receipt.LineItemID {
				lineItem = &po.LineItems[i]
				break
			}
		}

		if lineItem == nil {
			return nil, domain.NewValidationError("invalid line_item_id")
		}

		// Create inventory transaction
		txn := &domain.InventoryTransaction{
			ID:              uuid.New(),
			OrganizationID:  input.OrganizationID,
			ProductID:       lineItem.ProductID,
			Type:            domain.TransactionReceipt,
			Quantity:        receipt.ReceivedQty,
			UnitCost:        lineItem.UnitCost,
			ReferenceType:   "purchase_order",
			ReferenceID:     po.ID,
			TransactionDate: time.Now(),
			CreatedBy:       input.ReceivedBy,
		}

		if err := uc.inventoryTxnRepo.Create(ctx, txn); err != nil {
			return nil, err
		}

		// Update inventory balance
		if err := uc.inventoryBalanceRepo.UpdateBalance(ctx, lineItem.ProductID, input.OrganizationID, receipt.ReceivedQty); err != nil {
			return nil, err
		}

		// Update line item received quantity
		lineItem.ReceivedQty += receipt.ReceivedQty
	}

	// 3. Update PO status
	allReceived := true
	for _, item := range po.LineItems {
		if item.ReceivedQty < item.Quantity {
			allReceived = false
			break
		}
	}

	if allReceived {
		po.Status = domain.POStatusReceived
		po.ActualArrivalDate = new(time.Time)
		*po.ActualArrivalDate = time.Now()
	} else {
		po.Status = domain.POStatusPartial
	}

	po.CheckDelay()
	po.UpdatedAt = time.Now()

	if err := uc.poRepo.Update(ctx, po); err != nil {
		return nil, err
	}

	// 4. Update DDMRP NFP (on-hand increased)
	for _, receipt := range input.Receipts {
		uc.ddmrpClient.UpdateNFP(ctx, receipt.LineItemID, input.OrganizationID)
	}

	// 5. Publish event
	uc.eventPublisher.PublishPOReceived(ctx, po)

	return po, nil
}
```

---

### Phase 4: Alert System (Week 2 Day 4-5)

#### T004: Alert Generation Cron

**File**: `services/execution-service/internal/infrastructure/entrypoints/cron/alert_checker.go`

```go
package cron

import (
	"context"
	"log"
	"time"

	"giia-core-engine/services/execution-service/internal/core/usecases/alert"
	"github.com/robfig/cron/v3"
)

type AlertChecker struct {
	generateAlertsUseCase *alert.GenerateAlertsUseCase
	cronScheduler         *cron.Cron
}

func NewAlertChecker(generateAlertsUseCase *alert.GenerateAlertsUseCase) *AlertChecker {
	return &AlertChecker{
		generateAlertsUseCase: generateAlertsUseCase,
		cronScheduler:         cron.New(),
	}
}

func (ac *AlertChecker) Start() {
	// Run every hour
	ac.cronScheduler.AddFunc("0 * * * *", func() {
		log.Println("Checking for alerts...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := ac.generateAlertsUseCase.Execute(ctx); err != nil {
			log.Printf("Error generating alerts: %v", err)
		} else {
			log.Println("Alert check completed successfully")
		}
	})

	ac.cronScheduler.Start()
}

func (ac *AlertChecker) Stop() {
	ac.cronScheduler.Stop()
}
```

---

### Phase 5: gRPC API & Testing (Week 3)

#### T005: Protocol Buffers

**File**: `services/execution-service/api/proto/execution/v1/execution.proto`

```protobuf
syntax = "proto3";

package execution.v1;

option go_package = "giia-core-engine/services/execution-service/api/proto/execution/v1;executionv1";

import "google/protobuf/timestamp.proto";

service ExecutionService {
  // Purchase Order operations
  rpc CreatePO(CreatePORequest) returns (CreatePOResponse);
  rpc ReceivePO(ReceivePORequest) returns (ReceivePOResponse);
  rpc ListPOs(ListPOsRequest) returns (ListPOsResponse);

  // Sales Order operations
  rpc CreateSO(CreateSORequest) returns (CreateSOResponse);
  rpc IssueDeliveryNote(IssueDeliveryNoteRequest) returns (IssueDeliveryNoteResponse);
  rpc ListSOs(ListSOsRequest) returns (ListSOsResponse);

  // Inventory operations
  rpc RecordTransaction(RecordTransactionRequest) returns (RecordTransactionResponse);
  rpc GetBalance(GetBalanceRequest) returns (GetBalanceResponse);

  // Alert operations
  rpc ListAlerts(ListAlertsRequest) returns (ListAlertsResponse);
  rpc AcknowledgeAlert(AcknowledgeAlertRequest) returns (AcknowledgeAlertResponse);
}

message PurchaseOrder {
  string id = 1;
  string organization_id = 2;
  string po_number = 3;
  string supplier_id = 4;
  string status = 5;
  google.protobuf.Timestamp order_date = 6;
  google.protobuf.Timestamp expected_arrival_date = 7;
  google.protobuf.Timestamp actual_arrival_date = 8;
  int32 delay_days = 9;
  bool is_delayed = 10;
  double total_amount = 11;
  repeated POLineItem line_items = 12;
}

message POLineItem {
  string id = 1;
  string product_id = 2;
  double quantity = 3;
  double received_qty = 4;
  double unit_cost = 5;
  double line_total = 6;
}

message SalesOrder {
  string id = 1;
  string organization_id = 2;
  string so_number = 3;
  string customer_id = 4;
  string status = 5;
  google.protobuf.Timestamp order_date = 6;
  google.protobuf.Timestamp due_date = 7;
  bool delivery_note_issued = 8;
  string delivery_note_number = 9;
  google.protobuf.Timestamp delivery_note_date = 10;
  repeated SOLineItem line_items = 11;
}

message Alert {
  string id = 1;
  string organization_id = 2;
  string alert_type = 3;
  string severity = 4;
  string resource_type = 5;
  string resource_id = 6;
  string title = 7;
  string message = 8;
  google.protobuf.Timestamp created_at = 9;
}

// ... (request/response messages)
```

---

## 4. Success Criteria

### Mandatory
- ✅ Purchase Order CRUD with delay tracking
- ✅ Sales Order CRUD with delivery notes (remitos)
- ✅ Inventory transactions with ACID guarantees
- ✅ Inventory balance calculation
- ✅ Alert system for PO delays and buffer status
- ✅ Integration with DDMRP Engine for NFP updates
- ✅ Integration with Catalog service
- ✅ gRPC API for all operations
- ✅ Event publishing for state changes
- ✅ 85%+ test coverage
- ✅ Multi-tenancy support

---

## 5. Dependencies

- **Task 12**: Catalog service (products, suppliers)
- **Task 14**: DDMRP Engine (NFP updates)
- **Shared packages**: pkg/events, pkg/database, pkg/logger, pkg/errors

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Inventory balance discrepancies | ACID transactions, audit logging |
| PO delay detection accuracy | Automated cron checks, manual override |
| NFP update failures | Retry logic, event replay |
| Concurrent transaction conflicts | Optimistic locking, database transactions |

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
