# AI Intelligence Hub gRPC API

**Package:** `ai.v1`  
**Port:** 9086  
**Proto File:** `services/ai-intelligence-hub/api/proto/ai/v1/ai.proto`

---

## Service Definition

```protobuf
service AIIntelligenceService {
  // Notification management
  rpc CreateNotification(CreateNotificationRequest) returns (NotificationResponse);
  rpc GetNotification(GetNotificationRequest) returns (NotificationResponse);
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc MarkNotificationRead(MarkReadRequest) returns (NotificationResponse);
  rpc DeleteNotification(DeleteNotificationRequest) returns (DeleteResponse);
  
  // User preferences
  rpc GetPreferences(GetPreferencesRequest) returns (PreferencesResponse);
  rpc UpdatePreferences(UpdatePreferencesRequest) returns (PreferencesResponse);
  
  // AI-powered features
  rpc GetRecommendations(GetRecommendationsRequest) returns (RecommendationsResponse);
  rpc AnalyzePattern(AnalyzePatternRequest) returns (PatternResponse);
}
```

---

## Message Types

### Notification

```protobuf
message Notification {
  string id = 1;
  string user_id = 2;
  string organization_id = 3;
  string type = 4;           // alert, recommendation, insight, action_required
  string priority = 5;       // low, medium, high, critical
  string title = 6;
  string message = 7;
  map<string, string> data = 8;
  bool read = 9;
  google.protobuf.Timestamp read_at = 10;
  google.protobuf.Timestamp created_at = 11;
  google.protobuf.Timestamp expires_at = 12;
}
```

### UserPreferences

```protobuf
message UserPreferences {
  string user_id = 1;
  string organization_id = 2;
  bool email_notifications = 3;
  bool push_notifications = 4;
  bool sms_notifications = 5;
  NotificationTypePrefs notification_types = 6;
  QuietHours quiet_hours = 7;
  string digest_frequency = 8;  // none, daily, weekly
  google.protobuf.Timestamp updated_at = 9;
}

message NotificationTypePrefs {
  bool alerts = 1;
  bool recommendations = 2;
  bool insights = 3;
  bool action_required = 4;
}

message QuietHours {
  bool enabled = 1;
  string start = 2;  // "22:00"
  string end = 3;    // "08:00"
}
```

### Recommendation

```protobuf
message Recommendation {
  string id = 1;
  string type = 2;           // replenishment, optimization, alert
  string title = 3;
  string description = 4;
  string priority = 5;
  double confidence = 6;     // 0.0 - 1.0
  string action_type = 7;    // create_po, adjust_buffer, etc.
  map<string, string> action_data = 8;
  google.protobuf.Timestamp created_at = 9;
}
```

---

## Key Methods

### CreateNotification

Creates a notification for a user (called by other services).

**Request:**
```protobuf
message CreateNotificationRequest {
  string user_id = 1;
  string organization_id = 2;
  string type = 3;
  string priority = 4;
  string title = 5;
  string message = 6;
  map<string, string> data = 7;
  google.protobuf.Timestamp expires_at = 8;
}
```

**Response:**
```protobuf
message NotificationResponse {
  Notification notification = 1;
}
```

**Example (Go):**
```go
// Called by DDMRP Engine when buffer goes red
resp, err := aiClient.CreateNotification(ctx, &aiv1.CreateNotificationRequest{
    UserId:         userID,
    OrganizationId: orgID,
    Type:           "alert",
    Priority:       "critical",
    Title:          "Buffer Critical: Product ABC",
    Message:        "Product ABC is in red zone. Immediate replenishment needed.",
    Data: map[string]string{
        "product_id": productID,
        "zone":       "red",
        "nfp":        fmt.Sprintf("%.2f", nfp),
    },
})
```

---

### ListNotifications

Lists notifications for a user with filtering.

**Request:**
```protobuf
message ListNotificationsRequest {
  string user_id = 1;
  string organization_id = 2;
  bool unread_only = 3;
  string type = 4;       // Optional filter
  string priority = 5;   // Optional filter
  int32 page = 6;
  int32 page_size = 7;
}
```

**Response:**
```protobuf
message ListNotificationsResponse {
  repeated Notification notifications = 1;
  int32 total = 2;
  int32 unread_count = 3;
}
```

---

### GetRecommendations

Gets AI-powered recommendations for an organization.

**Request:**
```protobuf
message GetRecommendationsRequest {
  string organization_id = 1;
  string category = 2;    // Optional: replenishment, optimization, all
  int32 limit = 3;
}
```

**Response:**
```protobuf
message RecommendationsResponse {
  repeated Recommendation recommendations = 1;
}
```

**Example (Go):**
```go
resp, err := aiClient.GetRecommendations(ctx, &aiv1.GetRecommendationsRequest{
    OrganizationId: orgID,
    Category:       "replenishment",
    Limit:          10,
})
for _, rec := range resp.Recommendations {
    if rec.Confidence > 0.8 {
        // High confidence recommendation
        log.Printf("Recommendation: %s (%.0f%% confidence)", 
            rec.Title, rec.Confidence*100)
    }
}
```

---

### AnalyzePattern

Analyzes patterns in data for insights.

**Request:**
```protobuf
message AnalyzePatternRequest {
  string organization_id = 1;
  string pattern_type = 2;    // demand, stockout, seasonal
  string product_id = 3;      // Optional
  google.protobuf.Timestamp start_date = 4;
  google.protobuf.Timestamp end_date = 5;
}
```

**Response:**
```protobuf
message PatternResponse {
  string pattern_type = 1;
  repeated PatternInsight insights = 2;
  double confidence = 3;
}

message PatternInsight {
  string description = 1;
  map<string, double> metrics = 2;
  string recommendation = 3;
}
```

---

## Usage from Other Services

### DDMRP Engine → AI Hub (Buffer Alerts)

```go
func (s *DDMRPService) NotifyBufferAlert(ctx context.Context, buffer *Buffer) error {
    if buffer.Zone != "red" {
        return nil
    }
    
    // Get users to notify (admins and inventory managers)
    users := s.getUsersToNotify(buffer.OrganizationId)
    
    for _, userID := range users {
        _, err := s.aiClient.CreateNotification(ctx, &aiv1.CreateNotificationRequest{
            UserId:         userID,
            OrganizationId: buffer.OrganizationId,
            Type:           "alert",
            Priority:       "critical",
            Title:          fmt.Sprintf("Red Zone Alert: %s", buffer.ProductName),
            Message:        fmt.Sprintf("NFP at %.1f%% of buffer. Consider creating a purchase order.",
                buffer.BufferPenetration * 100),
            Data: map[string]string{
                "product_id": buffer.ProductId,
                "nfp":        fmt.Sprintf("%.2f", buffer.NetFlowPosition),
            },
        })
        if err != nil {
            log.Printf("Failed to create notification: %v", err)
        }
    }
    return nil
}
```

### Analytics → AI Hub (Daily Insights)

```go
func (s *AnalyticsService) GenerateDailyInsights(ctx context.Context, orgID string) error {
    kpi := s.getKPISnapshot(ctx, orgID)
    
    if kpi.StockoutRate > 0.05 {
        s.aiClient.CreateNotification(ctx, &aiv1.CreateNotificationRequest{
            Type:     "insight",
            Priority: "medium",
            Title:    "Weekly Insight: Stockout Rate",
            Message:  fmt.Sprintf("Stockout rate is %.1f%%, above target.", kpi.StockoutRate*100),
        })
    }
    
    return nil
}
```

---

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| `NOT_FOUND` (5) | Notification or user not found |
| `PERMISSION_DENIED` (7) | Cannot access other user's notifications |
| `INVALID_ARGUMENT` (3) | Invalid notification type or priority |
| `RESOURCE_EXHAUSTED` (8) | Too many notifications |
| `INTERNAL` (13) | AI processing error |

---

## Connection Example

```go
import (
    aiv1 "github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/api/proto/ai/v1"
    "google.golang.org/grpc"
)

func NewAIClient() (aiv1.AIIntelligenceServiceClient, error) {
    conn, err := grpc.Dial("ai-intelligence-hub:9086", grpc.WithInsecure())
    if err != nil {
        return nil, err
    }
    return aiv1.NewAIIntelligenceServiceClient(conn), nil
}
```

---

**Related Documentation:**
- [AI Hub OpenAPI](/services/ai-intelligence-hub/docs/openapi.yaml)
- [gRPC Contracts Overview](/docs/api/GRPC_CONTRACTS.md)
