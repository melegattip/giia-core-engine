# DDMRP Engine Service gRPC API

**Package:** `ddmrp.v1`  
**Port:** 9083  
**Proto File:** `services/ddmrp-engine-service/api/proto/ddmrp/v1/ddmrp.proto`

---

## Service Definition

```protobuf
service DDMRPService {
  // Buffer calculations
  rpc CalculateBuffer(CalculateBufferRequest) returns (CalculateBufferResponse);
  rpc GetBuffer(GetBufferRequest) returns (GetBufferResponse);
  rpc ListBuffers(ListBuffersRequest) returns (ListBuffersResponse);

  // Forward-looking Adjustments (FAD)
  rpc CreateFAD(CreateFADRequest) returns (CreateFADResponse);
  rpc UpdateFAD(UpdateFADRequest) returns (UpdateFADResponse);
  rpc DeleteFAD(DeleteFADRequest) returns (DeleteFADResponse);
  rpc ListFADs(ListFADsRequest) returns (ListFADsResponse);

  // Net Flow Position updates
  rpc UpdateNFP(UpdateNFPRequest) returns (UpdateNFPResponse);
  rpc CheckReplenishment(CheckReplenishmentRequest) returns (CheckReplenishmentResponse);
}
```

---

## Message Types

### Buffer

```protobuf
message Buffer {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  string buffer_profile_id = 4;
  
  // Consumption data
  double cpd = 5;                 // Consumption Per Day (ADU)
  int32 ltd = 6;                  // Lead Time Days
  
  // Zone calculations
  double red_base = 7;            // Red Base = CPD × LTD × LTF
  double red_safe = 8;            // Red Safe = Red Base × VF
  double red_zone = 9;            // Red Zone = Red Base + Red Safe
  double yellow_zone = 10;        // Yellow Zone = CPD × LTD
  double green_zone = 11;         // Green Zone = max(CPD × LTD × LTF, MOQ, cycle)
  
  // Zone thresholds
  double top_of_red = 12;         // Top of Red
  double top_of_yellow = 13;      // Top of Yellow
  double top_of_green = 14;       // Top of Green (total buffer size)
  
  // Current state
  double on_hand = 15;            // Current on-hand inventory
  double on_order = 16;           // Pending purchase orders
  double qualified_demand = 17;   // Confirmed sales orders
  double net_flow_position = 18;  // NFP = On Hand + On Order - Qualified Demand
  double buffer_penetration = 19; // NFP as % of Top of Green
  
  // Status
  string zone = 20;               // red, yellow, green
  string alert_level = 21;        // ok, warning, critical
  
  google.protobuf.Timestamp last_recalculated_at = 22;
  google.protobuf.Timestamp created_at = 23;
  google.protobuf.Timestamp updated_at = 24;
}
```

### DemandAdjustment (FAD)

```protobuf
message DemandAdjustment {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  google.protobuf.Timestamp start_date = 4;
  google.protobuf.Timestamp end_date = 5;
  string adjustment_type = 6;     // spike, reduction, promotion
  double factor = 7;              // Multiplier (e.g., 1.5 = 50% increase)
  string reason = 8;              // Description
  google.protobuf.Timestamp created_at = 9;
  string created_by = 10;
}
```

---

## Key Methods

### CalculateBuffer

Recalculates buffer zones for a product.

**Request:**
```protobuf
message CalculateBufferRequest {
  string product_id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message CalculateBufferResponse {
  Buffer buffer = 1;
}
```

**Example (Go):**
```go
resp, err := ddmrpClient.CalculateBuffer(ctx, &ddmrpv1.CalculateBufferRequest{
    ProductId:      productID,
    OrganizationId: orgID,
})
if err != nil {
    return err
}
buffer := resp.Buffer
log.Printf("Buffer zones: Red=%.2f, Yellow=%.2f, Green=%.2f",
    buffer.RedZone, buffer.YellowZone, buffer.GreenZone)
```

---

### GetBuffer

Retrieves current buffer status for a product.

**Request:**
```protobuf
message GetBufferRequest {
  string product_id = 1;
  string organization_id = 2;
}
```

**Response:**
```protobuf
message GetBufferResponse {
  Buffer buffer = 1;
}
```

---

### ListBuffers

Lists buffers with optional filtering by zone or alert level.

**Request:**
```protobuf
message ListBuffersRequest {
  string organization_id = 1;
  string zone = 2;           // Optional: red, yellow, green
  string alert_level = 3;    // Optional: ok, warning, critical
  int32 limit = 4;
  int32 offset = 5;
}
```

**Response:**
```protobuf
message ListBuffersResponse {
  repeated Buffer buffers = 1;
}
```

---

### UpdateNFP

Updates the Net Flow Position when inventory changes.

**Request:**
```protobuf
message UpdateNFPRequest {
  string product_id = 1;
  string organization_id = 2;
  double on_hand = 3;
  double on_order = 4;
  double qualified_demand = 5;
}
```

**Response:**
```protobuf
message UpdateNFPResponse {
  Buffer buffer = 1;  // Updated buffer with new NFP
}
```

**Example (Go):**
```go
// Called after inventory transaction
resp, err := ddmrpClient.UpdateNFP(ctx, &ddmrpv1.UpdateNFPRequest{
    ProductId:       productID,
    OrganizationId:  orgID,
    OnHand:          newOnHandQty,
    OnOrder:         pendingPOQty,
    QualifiedDemand: confirmedSOQty,
})
// Check if replenishment is needed
if resp.Buffer.Zone == "red" {
    // Trigger replenishment alert
}
```

---

### CheckReplenishment

Checks all products and returns those needing replenishment.

**Request:**
```protobuf
message CheckReplenishmentRequest {
  string organization_id = 1;
}
```

**Response:**
```protobuf
message CheckReplenishmentResponse {
  repeated Buffer buffers = 1;  // Buffers in red or yellow zone
}
```

---

### CreateFAD

Creates a Forward-looking Adjustment for demand spikes.

**Request:**
```protobuf
message CreateFADRequest {
  string product_id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp start_date = 3;
  google.protobuf.Timestamp end_date = 4;
  string adjustment_type = 5;
  double factor = 6;
  string reason = 7;
  string created_by = 8;
}
```

**Response:**
```protobuf
message CreateFADResponse {
  DemandAdjustment demand_adjustment = 1;
}
```

**Example (Go):**
```go
// Create promotion-related demand spike
resp, err := ddmrpClient.CreateFAD(ctx, &ddmrpv1.CreateFADRequest{
    ProductId:       productID,
    OrganizationId:  orgID,
    StartDate:       timestamppb.New(promoStart),
    EndDate:         timestamppb.New(promoEnd),
    AdjustmentType:  "promotion",
    Factor:          1.75,  // 75% increase
    Reason:          "Holiday promotion campaign",
    CreatedBy:       userID,
})
```

---

## DDMRP Calculation Flow

```
1. Get Product from Catalog Service
       ↓
2. Get Buffer Profile (LTF, VF, ADU method)
       ↓
3. Get Lead Time from Product-Supplier
       ↓
4. Calculate CPD (Average Daily Usage)
       ↓
5. Apply any active FADs (demand adjustments)
       ↓
6. Calculate Zones:
   - Red Base = CPD × LTD × LTF
   - Red Safe = Red Base × VF
   - Red Zone = Red Base + Red Safe
   - Yellow Zone = CPD × LTD
   - Green Zone = max(CPD × LTD × LTF, MOQ, cycle)
       ↓
7. Get current inventory from Execution Service
       ↓
8. Calculate NFP = On Hand + On Order - Qualified Demand
       ↓
9. Determine Zone and Alert Level
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `NOT_FOUND` (5) | Product or buffer not found |
| `INVALID_ARGUMENT` (3) | Invalid parameters |
| `FAILED_PRECONDITION` (9) | Missing buffer profile |
| `INTERNAL` (13) | Calculation error |

---

## Connection Example

```go
import (
    ddmrpv1 "github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/api/proto/ddmrp/v1"
    "google.golang.org/grpc"
)

func NewDDMRPClient() (ddmrpv1.DDMRPServiceClient, error) {
    conn, err := grpc.Dial("ddmrp-engine-service:9083", grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return ddmrpv1.NewDDMRPServiceClient(conn), nil
}
```

---

**Related Documentation:**
- [DDMRP Concepts](/docs/specifications/DDMRP_CONCEPTS.md)
- [gRPC Contracts Overview](/docs/api/GRPC_CONTRACTS.md)
