# Agent Prompt: Task 25 - AI Intelligence Hub API Layer

## 🤖 Agent Identity
Expert Go API Engineer for real-time notification systems with REST, WebSocket, and gRPC.

---

## 📋 Mission
Build API layer for AI Intelligence Hub: REST for notification management, WebSocket for real-time push, and user preferences API.

---

## 📂 Files to Create

### Handlers (internal/handlers/)
- `http/notification_handler.go` + `_test.go`
- `http/preferences_handler.go` + `_test.go`
- `http/router.go`
- `websocket/hub.go` + `_test.go`
- `websocket/client.go`
- `grpc/notification_service.go` + `_test.go`

---

## 🔧 REST Endpoints

### Notifications
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/notifications` | List (paginated, filtered) |
| GET | `/api/v1/notifications/{id}` | Get details + recommendations |
| PATCH | `/api/v1/notifications/{id}` | Update status (read/acted_upon) |
| DELETE | `/api/v1/notifications/{id}` | Delete notification |
| GET | `/api/v1/notifications/unread-count` | Quick count |

### Preferences
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/notifications/preferences` | Get user preferences |
| PUT | `/api/v1/notifications/preferences` | Update preferences |

---

## 🔧 WebSocket

```go
// Connection endpoint: /ws/notifications
// - Authenticate on connection via JWT
// - Subscribe by user ID + org ID
// - Push notifications in real-time
// - Handle reconnection with missed notifications
```

---

## 🔧 Key Requirements

### Filtering & Pagination
Support filtering by priority, status, type, date range.

### WebSocket Reconnection
Store last_seen_at per client, on reconnect send missed notifications.

### Preferences Schema
```go
type NotificationPreferences struct {
    EmailEnabled     bool
    EmailMinPriority string
    PushEnabled      bool
    QuietHoursStart  string
    QuietHoursEnd    string
    DigestEnabled    bool
    DigestTime       string
}
```

---

## ✅ Success Criteria
- [ ] 8+ REST endpoints
- [ ] WebSocket push <500ms from event
- [ ] Preferences respected for delivery
- [ ] 1000+ concurrent WebSocket connections
- [ ] Notification list <100ms p95
- [ ] 85%+ test coverage

---

## 🚀 Commands
```bash
cd services/ai-intelligence-hub
go test ./internal/handlers/... -cover
go build -o bin/ai-hub ./cmd/api
wscat -c ws://localhost:8086/ws/notifications -H "Authorization: Bearer $TOKEN"
```
