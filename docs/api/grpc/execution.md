# Execution Service gRPC API

**Package:** `execution.v1`  
**Port:** 9084  
**Proto File:** `services/execution-service/api/proto/execution/v1/execution.proto`

---

## Service Definition

```protobuf
service ExecutionService {
  rpc GetInventoryBalance(GetBalanceRequest) returns (BalanceResponse);
  rpc GetPendingPurchaseOrders(GetPendingPORequest) returns (PendingPOResponse);
  rpc GetPendingOnOrder(GetPendingOnOrderRequest) returns (OnOrderResponse);
  rpc GetQualifiedDemand(GetQualifiedDemandRequest) returns (DemandResponse);
  rpc RecordTransaction(RecordTransactionRequest) returns (TransactionResponse);
}
```

---

## Message Types

### InventoryBalance

```protobuf
message InventoryBalance {
  string id = 1;
  string organization_id = 2;
  string product_id = 3;
  string location_id = 4;
  double on_hand = 5;
  double reserved = 6;
  double available = 7;
  google.protobuf.Timestamp updated_at = 8;
}
```

### PendingPurchaseOrder

```protobuf
message PendingPurchaseOrder {
  string id = 1;
  string po_number = 2;
  string status = 3;
  google.protobuf.Timestamp expected_arrival_date = 4;
  double pending_quantity = 5;
}
```

### InventoryTransaction

```protobuf
message InventoryTransaction {
  string id = 1;
  string organization_id = 2;
  string product_id = 3;
  string location_id = 4;
  string type = 5;           // receipt, issue, transfer, adjustment
  double quantity = 6;
  double unit_cost = 7;
  string reference_type = 8;
  string reference_id = 9;
  string reason = 10;
  google.protobuf.Timestamp transaction_date = 11;
  string created_by = 12;
  google.protobuf.Timestamp created_at = 13;
}
```

---

## Key Methods

### GetInventoryBalance

Retrieves current inventory balance for a product.

**Request:**
```protobuf
message GetBalanceRequest {
  string organization_id = 1;
  string product_id = 2;
  string location_id = 3;  // Optional - if empty, returns all locations
}
```

**Response:**
```protobuf
message BalanceResponse {
  repeated InventoryBalance balances = 1;
  double total_on_hand = 2;
  double total_available = 3;
}
```

**Example (Go):**
```go
resp, err := execClient.GetInventoryBalance(ctx, &executionv1.GetBalanceRequest{
    OrganizationId: orgID,
    ProductId:      productID,
})
if err != nil {
    return err
}
onHand := resp.TotalOnHand
available := resp.TotalAvailable
```

---

### GetPendingOnOrder

Retrieves total quantity on order for DDMRP calculations.

**Request:**
```protobuf
message GetPendingOnOrderRequest {
  string organization_id = 1;
  string product_id = 2;
}
```

**Response:**
```protobuf
message OnOrderResponse {
  double on_order = 1;
  int32 pending_po_count = 2;
}
```

**Example (Go):**
```go
resp, err := execClient.GetPendingOnOrder(ctx, &executionv1.GetPendingOnOrderRequest{
    OrganizationId: orgID,
    ProductId:      productID,
})
// Use for NFP calculation
onOrder := resp.OnOrder
```

---

### GetQualifiedDemand

Retrieves qualified demand (confirmed sales orders) for NFP.

**Request:**
```protobuf
message GetQualifiedDemandRequest {
  string organization_id = 1;
  string product_id = 2;
}
```

**Response:**
```protobuf
message DemandResponse {
  double qualified_demand = 1;
  int32 confirmed_so_count = 2;
}
```

---

### RecordTransaction

Records an inventory transaction (receipt, issue, adjustment).

**Request:**
```protobuf
message RecordTransactionRequest {
  string organization_id = 1;
  string product_id = 2;
  string location_id = 3;
  string type = 4;           // receipt, issue, transfer, adjustment
  double quantity = 5;
  double unit_cost = 6;
  string reference_type = 7; // purchase_order, sales_order, manual
  string reference_id = 8;
  string reason = 9;
  string created_by = 10;
}
```

**Response:**
```protobuf
message TransactionResponse {
  InventoryTransaction transaction = 1;
  InventoryBalance updated_balance = 2;
}
```

**Example (Go):**
```go
// Record receipt from purchase order
resp, err := execClient.RecordTransaction(ctx, &executionv1.RecordTransactionRequest{
    OrganizationId: orgID,
    ProductId:      productID,
    LocationId:     warehouseID,
    Type:           "receipt",
    Quantity:       100,
    UnitCost:       25.50,
    ReferenceType:  "purchase_order",
    ReferenceId:    poID,
    Reason:         "PO-2024-001 received",
    CreatedBy:      userID,
})
newBalance := resp.UpdatedBalance.OnHand
```

---

## Usage from DDMRP Engine

The Execution Service provides the inventory data needed for NFP calculations:

```go
func (s *DDMRPService) GetCurrentNFPData(ctx context.Context, productID string) (*NFPData, error) {
    // Get on-hand inventory
    balanceResp, err := s.execClient.GetInventoryBalance(ctx, &executionv1.GetBalanceRequest{
        OrganizationId: s.orgID,
        ProductId:      productID,
    })
    if err != nil {
        return nil, err
    }
    
    // Get on-order quantity (pending POs)
    orderResp, err := s.execClient.GetPendingOnOrder(ctx, &executionv1.GetPendingOnOrderRequest{
        OrganizationId: s.orgID,
        ProductId:      productID,
    })
    if err != nil {
        return nil, err
    }
    
    // Get qualified demand (confirmed SOs)
    demandResp, err := s.execClient.GetQualifiedDemand(ctx, &executionv1.GetQualifiedDemandRequest{
        OrganizationId: s.orgID,
        ProductId:      productID,
    })
    if err != nil {
        return nil, err
    }
    
    return &NFPData{
        OnHand:          balanceResp.TotalOnHand,
        OnOrder:         orderResp.OnOrder,
        QualifiedDemand: demandResp.QualifiedDemand,
        NFP:             balanceResp.TotalOnHand + orderResp.OnOrder - demandResp.QualifiedDemand,
    }, nil
}
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `NOT_FOUND` (5) | Product or location not found |
| `INVALID_ARGUMENT` (3) | Invalid transaction type or quantity |
| `FAILED_PRECONDITION` (9) | Insufficient inventory for issue |
| `INTERNAL` (13) | Database error |

---

## Connection Example

```go
import (
    executionv1 "github.com/melegattip/giia-core-engine/services/execution-service/api/proto/execution/v1"
    "google.golang.org/grpc"
)

func NewExecutionClient() (executionv1.ExecutionServiceClient, error) {
    conn, err := grpc.Dial("execution-service:9084", grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return executionv1.NewExecutionServiceClient(conn), nil
}
```

---

**Related Documentation:**
- [Execution Service OpenAPI](/services/execution-service/docs/openapi.yaml)
- [gRPC Contracts Overview](/docs/api/GRPC_CONTRACTS.md)
