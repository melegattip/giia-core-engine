# Agent Prompt: Task 24 - Analytics Service API Completion

## 🤖 Agent Identity
Expert Go API Engineer for analytics and KPI microservices with REST, gRPC, and multi-service data aggregation.

---

## 📋 Mission
Complete Analytics Service API layer: REST/gRPC handlers for KPIs, and service adapters for cross-service data fetching.

---

## 📂 Files to Create

### Handlers (internal/handlers/)
- `http/kpi_handler.go` + `_test.go`
- `http/router.go`
- `grpc/analytics_service.go` + `_test.go`
- `grpc/server.go`

### Service Adapters (internal/adapters/)
- `catalog/client.go` - Product data fetching
- `ddmrp/client.go` - Buffer data fetching
- `execution/client.go` - Transaction data fetching

---

## 🔧 REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/analytics/days-in-inventory` | DII KPI with breakdown |
| GET | `/api/v1/analytics/immobilized-inventory` | Aged stock analysis |
| GET | `/api/v1/analytics/inventory-rotation` | Turnover metrics |
| GET | `/api/v1/analytics/buffer-analytics` | DDMRP buffer stats |
| GET | `/api/v1/analytics/snapshot` | Consolidated KPI view |

---

## 🔧 gRPC Service

```protobuf
service AnalyticsService {
  rpc GetDaysInInventory(DIIRequest) returns (DIIResponse);
  rpc GetImmobilizedInventory(ImmobilizedRequest) returns (ImmobilizedResponse);
  rpc GetInventoryRotation(RotationRequest) returns (RotationResponse);
  rpc GetBufferAnalytics(BufferAnalyticsRequest) returns (BufferAnalyticsResponse);
  rpc SyncBufferData(SyncRequest) returns (SyncResponse);
}
```

---

## 🔧 Key Requirements

### Response Caching
Cache KPI results for performance (5 min TTL).

### Service Adapters with Retry
All adapters must implement exponential backoff retry.

### Graceful Degradation
When adapters fail, return partial data with warnings.

---

## ✅ Success Criteria
- [ ] 5+ REST endpoints with caching
- [ ] gRPC service matching proto
- [ ] 3 service adapters with retry logic
- [ ] KPI queries <500ms p95
- [ ] Graceful degradation on adapter failures
- [ ] 85%+ test coverage

---

## 🚀 Commands
```bash
cd services/analytics-service
go test ./internal/handlers/... -cover
go build -o bin/analytics-service ./cmd/api
curl http://localhost:8083/api/v1/analytics/snapshot
```
