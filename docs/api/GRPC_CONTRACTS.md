# GIIA gRPC Service Contracts

**Version**: 1.0  
**Last Updated**: 2025-12-23

---

## 📖 Overview

GIIA uses gRPC for **internal service-to-service communication**. All external client communication uses REST APIs.

### Benefits of gRPC

- **Performance**: Binary protocol, faster than JSON
- **Type Safety**: Strong typing via Protocol Buffers
- **Code Generation**: Auto-generated clients in any language
- **Bi-directional Streaming**: For real-time features

---

## 📁 Proto File Locations

```
api/proto/
├── auth/v1/
│   └── auth.proto          # Auth service definitions
├── catalog/v1/
│   └── catalog.proto       # Catalog service definitions
├── ddmrp/v1/
│   └── ddmrp.proto         # DDMRP engine definitions
├── execution/v1/
│   └── execution.proto     # Execution service definitions
├── analytics/v1/
│   └── analytics.proto     # Analytics service definitions
└── ai/v1/
    └── ai.proto            # AI service definitions
```

---

## 🔐 Auth Service (Port 9081)

**Proto File**: `api/proto/auth/v1/auth.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.auth.v1;

service AuthService {
  // Validate JWT access token
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  
  // Check single permission
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  
  // Check multiple permissions at once
  rpc BatchCheckPermissions(BatchCheckPermissionsRequest) returns (BatchCheckPermissionsResponse);
  
  // Get user details
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

### Message Types

```protobuf
message ValidateTokenRequest {
  string token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  User user = 2;
  string error_message = 3;
}

message CheckPermissionRequest {
  string user_id = 1;
  string organization_id = 2;
  string resource = 3;        // e.g., "products", "orders"
  string action = 4;          // e.g., "create", "read", "update", "delete"
}

message CheckPermissionResponse {
  bool allowed = 1;
  string reason = 2;          // Explanation if denied
}

message BatchCheckPermissionsRequest {
  string user_id = 1;
  string organization_id = 2;
  repeated PermissionCheck permissions = 3;
}

message PermissionCheck {
  string resource = 1;
  string action = 2;
}

message BatchCheckPermissionsResponse {
  repeated PermissionResult results = 1;
}

message PermissionResult {
  string resource = 1;
  string action = 2;
  bool allowed = 3;
}

message GetUserRequest {
  string user_id = 1;
}

message GetUserResponse {
  User user = 1;
  string error_message = 2;
}

message User {
  string id = 1;
  string email = 2;
  string first_name = 3;
  string last_name = 4;
  string organization_id = 5;
  string status = 6;
  repeated string roles = 7;
  repeated string permissions = 8;
}
```

### Usage Example (Go Client)

```go
import (
    authpb "github.com/giia/giia-core-engine/api/proto/auth/v1"
    "google.golang.org/grpc"
)

// Create client
conn, err := grpc.Dial("auth-service:9081", grpc.WithInsecure())
if err != nil {
    return err
}
defer conn.Close()

client := authpb.NewAuthServiceClient(conn)

// Validate token
resp, err := client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
    Token: accessToken,
})
if err != nil {
    return err
}

if !resp.Valid {
    return errors.New(resp.ErrorMessage)
}

user := resp.User
```

---

## 📦 Catalog Service (Port 9082)

**Proto File**: `api/proto/catalog/v1/catalog.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.catalog.v1;

service CatalogService {
  // Product operations
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc GetProductWithBuffer(GetProductRequest) returns (ProductWithBufferResponse);
  
  // Supplier operations
  rpc GetSupplier(GetSupplierRequest) returns (GetSupplierResponse);
  rpc GetProductSuppliers(GetProductSuppliersRequest) returns (GetProductSuppliersResponse);
  
  // Buffer profile operations
  rpc GetBufferProfile(GetBufferProfileRequest) returns (GetBufferProfileResponse);
}
```

### Message Types

```protobuf
message Product {
  string id = 1;
  string organization_id = 2;
  string sku = 3;
  string name = 4;
  string description = 5;
  string category = 6;
  string unit_of_measure = 7;
  string status = 8;
  string buffer_profile_id = 9;
  string created_at = 10;
  string updated_at = 11;
}

message Supplier {
  string id = 1;
  string organization_id = 2;
  string code = 3;
  string name = 4;
  int32 lead_time_days = 5;
  string status = 6;
}

message BufferProfile {
  string id = 1;
  string organization_id = 2;
  string name = 3;
  string adu_method = 4;            // "average", "exponential", "weighted"
  string lead_time_category = 5;    // "long", "medium", "short"
  string variability_category = 6;  // "high", "medium", "low"
  double lead_time_factor = 7;
  double variability_factor = 8;
  string status = 9;
}
```

---

## 📊 DDMRP Engine Service (Port 9083)

**Proto File**: `api/proto/ddmrp/v1/ddmrp.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.ddmrp.v1;

service DDMRPService {
  // Buffer calculations
  rpc CalculateBuffer(CalculateBufferRequest) returns (CalculateBufferResponse);
  rpc GetBufferStatus(GetBufferStatusRequest) returns (BufferStatusResponse);
  rpc RecalculateAllBuffers(RecalculateRequest) returns (RecalculateResponse);
  
  // ADU calculations
  rpc CalculateADU(CalculateADURequest) returns (CalculateADUResponse);
  
  // Net Flow Equation
  rpc CalculateNetFlowPosition(NetFlowRequest) returns (NetFlowResponse);
}
```

### Message Types

```protobuf
message Buffer {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  double red_zone = 4;
  double yellow_zone = 5;
  double green_zone = 6;
  double top_of_green = 7;
  double adu = 8;
  int32 lead_time_days = 9;
  string status = 10;
  string last_calculated = 11;
}

message BufferStatus {
  string buffer_id = 1;
  double current_stock = 2;
  double net_flow_position = 3;
  string zone = 4;              // "red", "yellow", "green"
  double penetration_percent = 5;
  bool replenishment_needed = 6;
  double suggested_order_qty = 7;
}

message CalculateBufferRequest {
  string product_id = 1;
  string organization_id = 2;
  string buffer_profile_id = 3;
  int32 lead_time_days = 4;
  double adu = 5;
  int32 min_order_qty = 6;
  int32 order_frequency = 7;
}

message CalculateBufferResponse {
  Buffer buffer = 1;
  string error_message = 2;
}
```

---

## 📋 Execution Service (Port 9084)

**Proto File**: `api/proto/execution/v1/execution.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.execution.v1;

service ExecutionService {
  // Orders
  rpc CreatePurchaseOrder(CreatePurchaseOrderRequest) returns (PurchaseOrderResponse);
  rpc GetPurchaseOrder(GetPurchaseOrderRequest) returns (PurchaseOrderResponse);
  rpc ListPurchaseOrders(ListPurchaseOrdersRequest) returns (ListPurchaseOrdersResponse);
  
  // Inventory
  rpc GetStockLevel(GetStockLevelRequest) returns (StockLevelResponse);
  rpc RecordTransaction(RecordTransactionRequest) returns (TransactionResponse);
  
  // Replenishment
  rpc GetReplenishmentRecommendations(ReplenishmentRequest) returns (ReplenishmentResponse);
}
```

---

## 📈 Analytics Service (Port 9085)

**Proto File**: `api/proto/analytics/v1/analytics.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.analytics.v1;

service AnalyticsService {
  // Dashboard KPIs
  rpc GetDashboardKPIs(DashboardKPIRequest) returns (DashboardKPIResponse);
  
  // Inventory metrics
  rpc GetInventoryRotation(InventoryRotationRequest) returns (InventoryRotationResponse);
  rpc GetDaysInInventory(DaysInInventoryRequest) returns (DaysInInventoryResponse);
  rpc GetImmobilizedInventory(ImmobilizedInventoryRequest) returns (ImmobilizedInventoryResponse);
  
  // Buffer analytics
  rpc GetBufferPerformance(BufferPerformanceRequest) returns (BufferPerformanceResponse);
}
```

---

## 🤖 AI Intelligence Hub (Port 9086)

**Proto File**: `api/proto/ai/v1/ai.proto`

### Service Definition

```protobuf
syntax = "proto3";

package giia.ai.v1;

service AIService {
  // Notifications
  rpc GetNotifications(GetNotificationsRequest) returns (GetNotificationsResponse);
  rpc MarkNotificationRead(MarkReadRequest) returns (MarkReadResponse);
  
  // Recommendations
  rpc GetRecommendations(GetRecommendationsRequest) returns (GetRecommendationsResponse);
  
  // Forecasting
  rpc GetDemandForecast(DemandForecastRequest) returns (DemandForecastResponse);
}
```

---

## 🔧 Code Generation

### Generate Go Code

```bash
# Using protoc
protoc --go_out=. --go-grpc_out=. api/proto/auth/v1/auth.proto

# Using buf (recommended)
buf generate

# Using Make
make proto
```

### Generated Files

For each `.proto` file:
- `*.pb.go` - Message types
- `*_grpc.pb.go` - gRPC service code

---

## 🔒 Security

### Authentication

All gRPC calls between services use:
1. **mTLS** (in production)
2. **Service mesh** (Istio) for traffic encryption
3. **JWT validation** via Auth service

### Error Handling

gRPC status codes:

| Status | gRPC Code | HTTP Equivalent |
|--------|-----------|-----------------|
| OK | 0 | 200 |
| Invalid Argument | 3 | 400 |
| Unauthenticated | 16 | 401 |
| Permission Denied | 7 | 403 |
| Not Found | 5 | 404 |
| Internal | 13 | 500 |
| Unavailable | 14 | 503 |

---

## 📚 Related Documentation

- [Public API RFC](./PUBLIC_RFC.md)
- [Auth Service API](./AUTH_SERVICE_API.md)
- [Architecture Overview](/docs/architecture/OVERVIEW.md)

---

**gRPC contracts maintained by the GIIA Team** 📡
