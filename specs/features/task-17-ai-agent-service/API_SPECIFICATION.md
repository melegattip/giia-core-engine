# AI Intelligence Hub - API Specification

**Version:** 1.0
**Last Updated:** 2025-12-23
**Status:** Draft

---

## Overview

The AI Intelligence Hub exposes both gRPC and REST APIs for:
1. **Notification Management** - Query, update, dismiss notifications
2. **User Preferences** - Configure notification channels and settings
3. **Analytics** - Retrieve notification effectiveness metrics
4. **Admin Operations** - Knowledge base management, system health

---

## Table of Contents

1. [Authentication](#authentication)
2. [gRPC API](#grpc-api)
3. [REST API](#rest-api)
4. [WebSocket API](#websocket-api)
5. [Error Handling](#error-handling)
6. [Rate Limiting](#rate-limiting)

---

## Authentication

All API requests must include authentication credentials.

### gRPC Authentication

```protobuf
// Include in metadata
authorization: Bearer <jwt_token>
x-organization-id: <uuid>
```

### REST Authentication

```http
GET /api/v1/notifications HTTP/1.1
Host: intelligence-hub.giia.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000
```

### Authentication Flow

1. User authenticates via Auth Service
2. Receives JWT token with claims:
   ```json
   {
     "user_id": "uuid",
     "organization_id": "uuid",
     "roles": ["admin", "planner"],
     "exp": 1735603200
   }
   ```
3. Include JWT in all Intelligence Hub requests
4. Token validated and user context extracted

---

## gRPC API

### Proto Definition

```protobuf
// api/proto/intelligence/v1/intelligence.proto

syntax = "proto3";

package intelligence.v1;

option go_package = "github.com/giia/giia-core-engine/services/ai-intelligence-hub/api/proto/intelligence/v1;intelligencev1";

import "google/protobuf/timestamp.proto";

service IntelligenceService {
  // Notification Management
  rpc ListNotifications(ListNotificationsRequest) returns (ListNotificationsResponse);
  rpc GetNotification(GetNotificationRequest) returns (GetNotificationResponse);
  rpc MarkNotificationAsRead(MarkNotificationAsReadRequest) returns (MarkNotificationAsReadResponse);
  rpc MarkNotificationAsActedUpon(MarkNotificationAsActedUponRequest) returns (MarkNotificationAsActedUponResponse);
  rpc DismissNotification(DismissNotificationRequest) returns (DismissNotificationResponse);

  // User Preferences
  rpc GetUserPreferences(GetUserPreferencesRequest) returns (GetUserPreferencesResponse);
  rpc UpdateUserPreferences(UpdateUserPreferencesRequest) returns (UpdateUserPreferencesResponse);

  // Analytics
  rpc GetNotificationAnalytics(GetNotificationAnalyticsRequest) returns (GetNotificationAnalyticsResponse);

  // Admin Operations
  rpc ReprocessEvent(ReprocessEventRequest) returns (ReprocessEventResponse);
  rpc GetSystemHealth(GetSystemHealthRequest) returns (GetSystemHealthResponse);
}

// ==================== Messages ====================

// Notification
message Notification {
  string id = 1;
  string organization_id = 2;
  string user_id = 3;

  NotificationType type = 4;
  NotificationPriority priority = 5;

  string title = 6;
  string summary = 7;
  string full_analysis = 8;
  string reasoning = 9;

  ImpactAssessment impact = 10;
  repeated Recommendation recommendations = 11;

  repeated string source_events = 12;
  map<string, EntityList> related_entities = 13;

  NotificationStatus status = 14;

  google.protobuf.Timestamp created_at = 15;
  google.protobuf.Timestamp read_at = 16;
  google.protobuf.Timestamp acted_at = 17;
  google.protobuf.Timestamp dismissed_at = 18;
}

enum NotificationType {
  NOTIFICATION_TYPE_UNSPECIFIED = 0;
  NOTIFICATION_TYPE_ALERT = 1;
  NOTIFICATION_TYPE_WARNING = 2;
  NOTIFICATION_TYPE_INFO = 3;
  NOTIFICATION_TYPE_SUGGESTION = 4;
  NOTIFICATION_TYPE_INSIGHT = 5;
  NOTIFICATION_TYPE_DIGEST = 6;
}

enum NotificationPriority {
  NOTIFICATION_PRIORITY_UNSPECIFIED = 0;
  NOTIFICATION_PRIORITY_LOW = 1;
  NOTIFICATION_PRIORITY_MEDIUM = 2;
  NOTIFICATION_PRIORITY_HIGH = 3;
  NOTIFICATION_PRIORITY_CRITICAL = 4;
}

enum NotificationStatus {
  NOTIFICATION_STATUS_UNSPECIFIED = 0;
  NOTIFICATION_STATUS_UNREAD = 1;
  NOTIFICATION_STATUS_READ = 2;
  NOTIFICATION_STATUS_ACTED_UPON = 3;
  NOTIFICATION_STATUS_DISMISSED = 4;
}

message ImpactAssessment {
  string risk_level = 1;  // low, medium, high, critical
  double revenue_impact = 2;
  double cost_impact = 3;
  int64 time_to_impact_hours = 4;
  int32 affected_orders = 5;
  int32 affected_products = 6;
}

message Recommendation {
  string action = 1;
  string reasoning = 2;
  string expected_outcome = 3;
  string effort = 4;  // low, medium, high
  string impact = 5;  // low, medium, high
  string action_url = 6;
  int32 priority_order = 7;
}

message EntityList {
  repeated string ids = 1;
}

// List Notifications
message ListNotificationsRequest {
  string organization_id = 1;
  string user_id = 2;

  // Filters
  repeated NotificationType types = 3;
  repeated NotificationPriority priorities = 4;
  repeated NotificationStatus statuses = 5;

  // Pagination
  int32 page_size = 6;  // Default: 20, Max: 100
  string page_token = 7;

  // Sorting
  string order_by = 8;  // created_at, priority, status
  bool descending = 9;
}

message ListNotificationsResponse {
  repeated Notification notifications = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

// Get Notification
message GetNotificationRequest {
  string notification_id = 1;
  string organization_id = 2;
}

message GetNotificationResponse {
  Notification notification = 1;
}

// Mark As Read
message MarkNotificationAsReadRequest {
  string notification_id = 1;
  string organization_id = 2;
}

message MarkNotificationAsReadResponse {
  Notification notification = 1;
}

// Mark As Acted Upon
message MarkNotificationAsActedUponRequest {
  string notification_id = 1;
  string organization_id = 2;
  string action_taken = 3;  // Optional: describe what action was taken
}

message MarkNotificationAsActedUponResponse {
  Notification notification = 1;
}

// Dismiss Notification
message DismissNotificationRequest {
  string notification_id = 1;
  string organization_id = 2;
  string dismissal_reason = 3;  // Optional: why dismissed
}

message DismissNotificationResponse {
  Notification notification = 1;
}

// User Preferences
message UserPreferences {
  string user_id = 1;
  string organization_id = 2;

  // Channel preferences
  bool enable_in_app = 3;
  bool enable_email = 4;
  bool enable_sms = 5;
  bool enable_slack = 6;
  string slack_webhook_url = 7;

  // Priority thresholds
  NotificationPriority in_app_min_priority = 8;
  NotificationPriority email_min_priority = 9;
  NotificationPriority sms_min_priority = 10;

  // Timing
  string digest_time = 11;  // "06:00"
  string quiet_hours_start = 12;  // "22:00"
  string quiet_hours_end = 13;  // "07:00"
  string timezone = 14;  // "America/New_York"

  // Frequency limits
  int32 max_alerts_per_hour = 15;
  int32 max_emails_per_day = 16;

  // Content preferences
  string detail_level = 17;  // brief, detailed, comprehensive
  bool include_charts = 18;
  bool include_historical = 19;

  google.protobuf.Timestamp updated_at = 20;
}

message GetUserPreferencesRequest {
  string user_id = 1;
  string organization_id = 2;
}

message GetUserPreferencesResponse {
  UserPreferences preferences = 1;
}

message UpdateUserPreferencesRequest {
  UserPreferences preferences = 1;
}

message UpdateUserPreferencesResponse {
  UserPreferences preferences = 1;
}

// Analytics
message GetNotificationAnalyticsRequest {
  string organization_id = 1;
  google.protobuf.Timestamp start_date = 2;
  google.protobuf.Timestamp end_date = 3;
}

message GetNotificationAnalyticsResponse {
  int32 total_notifications = 1;
  int32 critical_alerts = 2;
  int32 acted_upon = 3;
  int32 dismissed = 4;
  double action_rate = 5;  // acted_upon / total

  map<string, int32> by_type = 6;
  map<string, int32> by_priority = 7;

  double avg_time_to_read_minutes = 8;
  double avg_time_to_action_minutes = 9;
}

// Admin Operations
message ReprocessEventRequest {
  string event_id = 1;
  string organization_id = 2;
}

message ReprocessEventResponse {
  bool success = 1;
  string message = 2;
}

message GetSystemHealthRequest {}

message GetSystemHealthResponse {
  bool healthy = 1;
  string version = 2;

  ComponentHealth database = 3;
  ComponentHealth nats = 4;
  ComponentHealth redis = 5;
  ComponentHealth chromadb = 6;
  ComponentHealth claude_api = 7;
}

message ComponentHealth {
  bool healthy = 1;
  string status = 2;
  int64 latency_ms = 3;
  string last_error = 4;
}
```

### gRPC Usage Examples

#### Go Client Example

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"

    intelligencev1 "github.com/giia/giia-core-engine/services/ai-intelligence-hub/api/proto/intelligence/v1"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("intelligence-hub.giia.com:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := intelligencev1.NewIntelligenceServiceClient(conn)

    // Add authentication metadata
    ctx := metadata.AppendToOutgoingContext(
        context.Background(),
        "authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "x-organization-id", "550e8400-e29b-41d4-a716-446655440000",
    )

    // List notifications
    resp, err := client.ListNotifications(ctx, &intelligencev1.ListNotificationsRequest{
        OrganizationId: "550e8400-e29b-41d4-a716-446655440000",
        UserId:         "660e8400-e29b-41d4-a716-446655440000",
        Priorities: []intelligencev1.NotificationPriority{
            intelligencev1.NotificationPriority_NOTIFICATION_PRIORITY_CRITICAL,
            intelligencev1.NotificationPriority_NOTIFICATION_PRIORITY_HIGH,
        },
        Statuses: []intelligencev1.NotificationStatus{
            intelligencev1.NotificationStatus_NOTIFICATION_STATUS_UNREAD,
        },
        PageSize: 20,
        OrderBy: "created_at",
        Descending: true,
    })

    if err != nil {
        log.Fatalf("Failed to list notifications: %v", err)
    }

    log.Printf("Found %d notifications", resp.TotalCount)
    for _, notif := range resp.Notifications {
        log.Printf("- [%s] %s: %s", notif.Priority, notif.Type, notif.Title)
    }
}
```

---

## REST API

### Base URL

```
Production: https://intelligence-hub.giia.com/api/v1
Staging: https://staging-intelligence-hub.giia.com/api/v1
```

### Endpoints

#### 1. List Notifications

```http
GET /api/v1/notifications
```

**Query Parameters:**
- `organization_id` (required): Organization UUID
- `user_id` (required): User UUID
- `type` (optional): Filter by type (can repeat: `type=alert&type=warning`)
- `priority` (optional): Filter by priority (low, medium, high, critical)
- `status` (optional): Filter by status (unread, read, acted_upon, dismissed)
- `page_size` (optional): Number of results (default: 20, max: 100)
- `page_token` (optional): Pagination token
- `order_by` (optional): Sort field (created_at, priority, status)
- `descending` (optional): Sort direction (true/false)

**Example Request:**
```http
GET /api/v1/notifications?organization_id=550e8400-e29b-41d4-a716-446655440000&user_id=660e8400-e29b-41d4-a716-446655440000&status=unread&priority=critical&priority=high&page_size=20&order_by=created_at&descending=true HTTP/1.1
Host: intelligence-hub.giia.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Example Response:**
```json
{
  "notifications": [
    {
      "id": "770e8400-e29b-41d4-a716-446655440000",
      "organization_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "660e8400-e29b-41d4-a716-446655440000",
      "type": "alert",
      "priority": "critical",
      "title": "Imminent Stockout: Widget-A",
      "summary": "Critical: Widget-A will stockout in 1.5 days. Immediate action required to prevent $15,000 revenue loss.",
      "full_analysis": "Detailed analysis...",
      "reasoning": "Based on DDMRP buffer methodology...",
      "impact": {
        "risk_level": "critical",
        "revenue_impact": 15000.00,
        "cost_impact": 0,
        "time_to_impact_hours": 36,
        "affected_orders": 5,
        "affected_products": 1
      },
      "recommendations": [
        {
          "action": "Place emergency order with Supplier B (2-day lead time)",
          "reasoning": "Primary supplier has 7-day lead time which exceeds stockout timeline",
          "expected_outcome": "Stockout prevented, $15K revenue protected, $200 expedite cost",
          "effort": "low",
          "impact": "high",
          "action_url": "/orders/create?supplier=supplier-b&product=widget-a&quantity=70",
          "priority_order": 1
        },
        {
          "action": "Increase buffer by 20% to prevent future occurrences",
          "reasoning": "Buffer penetration pattern indicates systematic under-buffering",
          "expected_outcome": "Reduced stockout frequency, improved service level",
          "effort": "medium",
          "impact": "high",
          "action_url": "/buffers/adjust?product=widget-a&factor=1.2",
          "priority_order": 2
        }
      ],
      "source_events": ["event-id-1", "event-id-2"],
      "related_entities": {
        "product_ids": ["product-uuid-1"],
        "supplier_ids": ["supplier-uuid-1", "supplier-uuid-2"]
      },
      "status": "unread",
      "created_at": "2025-12-23T10:30:00Z",
      "read_at": null,
      "acted_at": null,
      "dismissed_at": null
    }
  ],
  "next_page_token": "eyJvZmZzZXQiOjIwfQ==",
  "total_count": 45
}
```

#### 2. Get Notification by ID

```http
GET /api/v1/notifications/{notification_id}
```

**Path Parameters:**
- `notification_id` (required): Notification UUID

**Query Parameters:**
- `organization_id` (required): Organization UUID

**Example Request:**
```http
GET /api/v1/notifications/770e8400-e29b-41d4-a716-446655440000?organization_id=550e8400-e29b-41d4-a716-446655440000 HTTP/1.1
Host: intelligence-hub.giia.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Example Response:**
```json
{
  "notification": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    ...
  }
}
```

#### 3. Mark Notification as Read

```http
POST /api/v1/notifications/{notification_id}/read
```

**Request Body:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "notification": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    ...
    "status": "read",
    "read_at": "2025-12-23T11:15:00Z"
  }
}
```

#### 4. Mark Notification as Acted Upon

```http
POST /api/v1/notifications/{notification_id}/acted
```

**Request Body:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "action_taken": "Placed emergency order with Supplier B for 70 units"
}
```

**Response:**
```json
{
  "notification": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    ...
    "status": "acted_upon",
    "acted_at": "2025-12-23T11:20:00Z"
  }
}
```

#### 5. Dismiss Notification

```http
POST /api/v1/notifications/{notification_id}/dismiss
```

**Request Body:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "dismissal_reason": "False alarm - pending order already covers requirement"
}
```

#### 6. Get User Preferences

```http
GET /api/v1/users/{user_id}/preferences
```

**Query Parameters:**
- `organization_id` (required): Organization UUID

**Example Response:**
```json
{
  "preferences": {
    "user_id": "660e8400-e29b-41d4-a716-446655440000",
    "organization_id": "550e8400-e29b-41d4-a716-446655440000",
    "enable_in_app": true,
    "enable_email": true,
    "enable_sms": true,
    "enable_slack": false,
    "slack_webhook_url": null,
    "in_app_min_priority": "low",
    "email_min_priority": "medium",
    "sms_min_priority": "critical",
    "digest_time": "06:00",
    "quiet_hours_start": "22:00",
    "quiet_hours_end": "07:00",
    "timezone": "America/New_York",
    "max_alerts_per_hour": 10,
    "max_emails_per_day": 50,
    "detail_level": "detailed",
    "include_charts": true,
    "include_historical": true,
    "updated_at": "2025-12-23T00:00:00Z"
  }
}
```

#### 7. Update User Preferences

```http
PUT /api/v1/users/{user_id}/preferences
```

**Request Body:**
```json
{
  "organization_id": "550e8400-e29b-41d4-a716-446655440000",
  "enable_email": false,
  "email_min_priority": "high",
  "quiet_hours_start": "23:00",
  "quiet_hours_end": "08:00"
}
```

---

## WebSocket API

Real-time notification delivery via WebSocket for in-app notifications.

### Connection

```javascript
const ws = new WebSocket('wss://intelligence-hub.giia.com/ws');

// Send authentication
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
    organization_id: '550e8400-e29b-41d4-a716-446655440000',
    user_id: '660e8400-e29b-41d4-a716-446655440000'
  }));
};

// Receive notifications
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  switch (message.type) {
    case 'auth_success':
      console.log('Authenticated successfully');
      break;

    case 'notification':
      handleNotification(message.notification);
      break;

    case 'ping':
      ws.send(JSON.stringify({ type: 'pong' }));
      break;
  }
};
```

### Message Types

#### Server → Client: New Notification

```json
{
  "type": "notification",
  "notification": {
    "id": "770e8400-e29b-41d4-a716-446655440000",
    "priority": "critical",
    "title": "Imminent Stockout: Widget-A",
    ...
  }
}
```

#### Client → Server: Acknowledge Notification

```json
{
  "type": "ack",
  "notification_id": "770e8400-e29b-41d4-a716-446655440000"
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "NOTIFICATION_NOT_FOUND",
    "message": "Notification with ID 770e8400-e29b-41d4-a716-446655440000 not found",
    "details": {
      "notification_id": "770e8400-e29b-41d4-a716-446655440000"
    }
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `UNAUTHENTICATED` | 401 | Missing or invalid authentication token |
| `PERMISSION_DENIED` | 403 | User lacks permission for requested resource |
| `NOTIFICATION_NOT_FOUND` | 404 | Notification does not exist |
| `INVALID_ARGUMENT` | 400 | Invalid request parameters |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `SERVICE_UNAVAILABLE` | 503 | Service temporarily unavailable |

---

## Rate Limiting

### Limits

| Endpoint | Limit |
|----------|-------|
| List Notifications | 100 requests/minute |
| Get Notification | 200 requests/minute |
| Update Preferences | 20 requests/minute |
| WebSocket Connection | 10 connections/user |

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1735603200
```

### Rate Limit Exceeded Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again in 45 seconds.",
    "details": {
      "retry_after_seconds": 45
    }
  }
}
```

---

## Versioning

API versioning follows semantic versioning in the URL path:
- `/api/v1/...` - Current stable version
- `/api/v2/...` - Future version (backward incompatible changes)

Breaking changes will always result in a new major version.

---

## SDK Examples

### TypeScript/JavaScript

```typescript
import { IntelligenceHubClient } from '@giia/intelligence-hub-sdk';

const client = new IntelligenceHubClient({
  baseURL: 'https://intelligence-hub.giia.com',
  apiKey: 'your-api-key',
  organizationId: '550e8400-e29b-41d4-a716-446655440000'
});

// List unread critical notifications
const notifications = await client.notifications.list({
  userId: '660e8400-e29b-41d4-a716-446655440000',
  status: ['unread'],
  priority: ['critical', 'high'],
  pageSize: 20
});

// Mark as acted upon
await client.notifications.markAsActedUpon(
  '770e8400-e29b-41d4-a716-446655440000',
  { actionTaken: 'Placed emergency order' }
);
```

### Python

```python
from giia_intelligence_hub import IntelligenceHubClient

client = IntelligenceHubClient(
    base_url='https://intelligence-hub.giia.com',
    api_key='your-api-key',
    organization_id='550e8400-e29b-41d4-a716-446655440000'
)

# List notifications
notifications = client.notifications.list(
    user_id='660e8400-e29b-41d4-a716-446655440000',
    status=['unread'],
    priority=['critical', 'high'],
    page_size=20
)

# Mark as read
client.notifications.mark_as_read('770e8400-e29b-41d4-a716-446655440000')
```

---

**End of API Specification**