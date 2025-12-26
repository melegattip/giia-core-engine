# Analytics Service gRPC API

**Package:** `analytics.v1`  
**Port:** 9085  
**Proto File:** `services/analytics-service/api/proto/analytics/v1/analytics.proto`

---

## Service Definition

```protobuf
service AnalyticsService {
  rpc GetKPISnapshot(GetKPISnapshotRequest) returns (GetKPISnapshotResponse);
  rpc ListKPISnapshots(ListKPISnapshotsRequest) returns (ListKPISnapshotsResponse);

  rpc GetDaysInInventoryKPI(GetDaysInInventoryKPIRequest) returns (GetDaysInInventoryKPIResponse);
  rpc ListDaysInInventoryKPI(ListDaysInInventoryKPIRequest) returns (ListDaysInInventoryKPIResponse);

  rpc GetImmobilizedInventoryKPI(GetImmobilizedInventoryKPIRequest) returns (GetImmobilizedInventoryKPIResponse);
  rpc ListImmobilizedInventoryKPI(ListImmobilizedInventoryKPIRequest) returns (ListImmobilizedInventoryKPIResponse);

  rpc GetInventoryRotationKPI(GetInventoryRotationKPIRequest) returns (GetInventoryRotationKPIResponse);
  rpc ListInventoryRotationKPI(ListInventoryRotationKPIRequest) returns (ListInventoryRotationKPIResponse);

  rpc GetBufferAnalytics(GetBufferAnalyticsRequest) returns (GetBufferAnalyticsResponse);
  rpc ListBufferAnalytics(ListBufferAnalyticsRequest) returns (ListBufferAnalyticsResponse);
}
```

---

## Message Types

### KPISnapshot

```protobuf
message KPISnapshot {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  double inventory_turnover = 4;
  double stockout_rate = 5;
  double service_level = 6;
  double excess_inventory_pct = 7;
  double buffer_score_green = 8;
  double buffer_score_yellow = 9;
  double buffer_score_red = 10;
  double total_inventory_value = 11;
  google.protobuf.Timestamp created_at = 12;
}
```

### DaysInInventoryKPI

```protobuf
message DaysInInventoryKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  double total_valued_days = 4;
  double average_valued_days = 5;
  int32 total_products = 6;
  google.protobuf.Timestamp created_at = 7;
}
```

### ImmobilizedInventoryKPI

```protobuf
message ImmobilizedInventoryKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  int32 threshold_years = 4;
  int32 immobilized_count = 5;
  double immobilized_value = 6;
  double total_stock_value = 7;
  double immobilized_percentage = 8;
  google.protobuf.Timestamp created_at = 9;
}
```

### InventoryRotationKPI

```protobuf
message InventoryRotationKPI {
  string id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp snapshot_date = 3;
  double sales_last_30_days = 4;
  double avg_monthly_stock = 5;
  double rotation_ratio = 6;
  repeated RotatingProduct top_rotating_products = 7;
  repeated RotatingProduct slow_rotating_products = 8;
  google.protobuf.Timestamp created_at = 9;
}

message RotatingProduct {
  string product_id = 1;
  string sku = 2;
  string name = 3;
  double sales_30_days = 4;
  double avg_stock_value = 5;
  double rotation_ratio = 6;
}
```

### BufferAnalytics

```protobuf
message BufferAnalytics {
  string id = 1;
  string product_id = 2;
  string organization_id = 3;
  google.protobuf.Timestamp date = 4;
  double cpd = 5;
  double red_zone = 6;
  double red_base = 7;
  double red_safe = 8;
  double yellow_zone = 9;
  double green_zone = 10;
  int32 ltd = 11;
  double lead_time_factor = 12;
  double variability_factor = 13;
  int32 moq = 14;
  int32 order_frequency = 15;
  double optimal_order_freq = 16;
  double safety_days = 17;
  double avg_open_orders = 18;
  bool has_adjustments = 19;
  google.protobuf.Timestamp created_at = 20;
}
```

---

## Key Methods

### GetKPISnapshot

Retrieves the KPI snapshot for a specific date.

**Request:**
```protobuf
message GetKPISnapshotRequest {
  string organization_id = 1;
  google.protobuf.Timestamp snapshot_date = 2;
}
```

**Response:**
```protobuf
message GetKPISnapshotResponse {
  KPISnapshot kpi_snapshot = 1;
}
```

**Example (Go):**
```go
resp, err := analyticsClient.GetKPISnapshot(ctx, &analyticsv1.GetKPISnapshotRequest{
    OrganizationId: orgID,
    SnapshotDate:   timestamppb.Now(),
})
if err != nil {
    return err
}
kpi := resp.KpiSnapshot
log.Printf("Service Level: %.2f%%, Turnover: %.2f", 
    kpi.ServiceLevel*100, kpi.InventoryTurnover)
```

---

### ListKPISnapshots

Lists KPI snapshots for a date range (trend analysis).

**Request:**
```protobuf
message ListKPISnapshotsRequest {
  string organization_id = 1;
  google.protobuf.Timestamp start_date = 2;
  google.protobuf.Timestamp end_date = 3;
}
```

**Response:**
```protobuf
message ListKPISnapshotsResponse {
  repeated KPISnapshot kpi_snapshots = 1;
}
```

---

### GetBufferAnalytics

Retrieves buffer performance analytics for a product.

**Request:**
```protobuf
message GetBufferAnalyticsRequest {
  string product_id = 1;
  string organization_id = 2;
  google.protobuf.Timestamp date = 3;
}
```

**Response:**
```protobuf
message GetBufferAnalyticsResponse {
  BufferAnalytics analytics = 1;
}
```

---

### ListBufferAnalytics

Lists buffer analytics for all products in a date range.

**Request:**
```protobuf
message ListBufferAnalyticsRequest {
  string organization_id = 1;
  google.protobuf.Timestamp start_date = 2;
  google.protobuf.Timestamp end_date = 3;
}
```

**Response:**
```protobuf
message ListBufferAnalyticsResponse {
  repeated BufferAnalytics analytics_list = 1;
}
```

---

## Usage from AI Intelligence Hub

The Analytics Service provides data for AI recommendations:

```go
func (s *AIService) GetInsights(ctx context.Context, orgID string) ([]*Insight, error) {
    // Get current KPI snapshot
    kpiResp, err := s.analyticsClient.GetKPISnapshot(ctx, &analyticsv1.GetKPISnapshotRequest{
        OrganizationId: orgID,
        SnapshotDate:   timestamppb.Now(),
    })
    if err != nil {
        return nil, err
    }
    
    insights := []*Insight{}
    
    // Analyze stockout rate
    if kpiResp.KpiSnapshot.StockoutRate > 0.05 {
        insights = append(insights, &Insight{
            Type:     "warning",
            Title:    "High Stockout Rate",
            Message:  fmt.Sprintf("Stockout rate is %.1f%%, above 5%% threshold", 
                kpiResp.KpiSnapshot.StockoutRate*100),
            Priority: "high",
        })
    }
    
    // Analyze buffer distribution
    redPct := kpiResp.KpiSnapshot.BufferScoreRed
    if redPct > 0.1 {
        insights = append(insights, &Insight{
            Type:     "critical",
            Title:    "Many Products in Red Zone",
            Message:  fmt.Sprintf("%.1f%% of products are in red zone", redPct*100),
            Priority: "critical",
        })
    }
    
    return insights, nil
}
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `NOT_FOUND` (5) | Snapshot not found for date |
| `INVALID_ARGUMENT` (3) | Invalid date range |
| `INTERNAL` (13) | Calculation error |

---

## Connection Example

```go
import (
    analyticsv1 "github.com/giia/giia-core-engine/services/analytics-service/api/proto/analytics/v1"
    "google.golang.org/grpc"
)

func NewAnalyticsClient() (analyticsv1.AnalyticsServiceClient, error) {
    conn, err := grpc.Dial("analytics-service:9085", grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return analyticsv1.NewAnalyticsServiceClient(conn), nil
}
```

---

**Related Documentation:**
- [Analytics Service OpenAPI](/services/analytics-service/docs/openapi.yaml)
- [gRPC Contracts Overview](/docs/api/GRPC_CONTRACTS.md)
