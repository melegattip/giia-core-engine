# Agent Prompt: Task 23 - Execution Service API Layer

## 🤖 Agent Identity
Expert Go API Engineer for REST and gRPC services with Clean Architecture, JWT authentication, and NATS events.

---

## 📋 Mission
Build the API layer for Execution Service: REST handlers for orders/inventory, gRPC service for inter-service communication, and service adapters.

---

## 📂 Files to Create

### REST Handlers (internal/handlers/http/)
- `purchase_order_handler.go` + `_test.go`
- `sales_order_handler.go` + `_test.go`
- `inventory_handler.go` + `_test.go`
- `router.go` - Chi router setup with middleware

### gRPC Handlers (internal/handlers/grpc/)
- `execution_service.go` + `_test.go`
- `server.go`

### Service Adapters (internal/adapters/)
- `catalog/client.go` - Catalog service gRPC client
- `ddmrp/client.go` - DDMRP Engine gRPC client

---

## 🔧 REST Endpoints

### Purchase Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/purchase-orders` | Create PO |
| GET | `/api/v1/purchase-orders` | List POs (paginated) |
| GET | `/api/v1/purchase-orders/{id}` | Get PO details |
| POST | `/api/v1/purchase-orders/{id}/receive` | Receive goods |
| POST | `/api/v1/purchase-orders/{id}/cancel` | Cancel PO |

### Sales Orders
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/sales-orders` | Create SO |
| GET | `/api/v1/sales-orders` | List SOs |
| GET | `/api/v1/sales-orders/{id}` | Get SO details |
| POST | `/api/v1/sales-orders/{id}/ship` | Ship SO |
| POST | `/api/v1/sales-orders/{id}/cancel` | Cancel SO |

### Inventory
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/inventory/balances` | Get balances |
| GET | `/api/v1/inventory/transactions` | Get transactions |

---

## 🔧 gRPC Service

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

## 🔧 Key Requirements

### JWT Authentication Middleware
```go
func AuthMiddleware(authClient auth.AuthServiceClient) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        // Extract JWT from Authorization header
        // Validate via auth-service gRPC
        // Add user/org context
    }
}
```

### NATS Event Publishing
All mutations MUST publish events:
- `execution.purchase_order.created`
- `execution.purchase_order.received`
- `execution.sales_order.created`
- `execution.inventory.updated`

---

## ✅ Success Criteria
- [ ] 12+ REST endpoints implemented
- [ ] 5+ gRPC methods for inter-service
- [ ] 2 service adapters (Catalog, DDMRP)
- [ ] JWT auth on all endpoints
- [ ] NATS events published
- [ ] Response time <200ms p95
- [ ] 80%+ test coverage

---

## 🚀 Commands
```bash
cd services/execution-service
go test ./internal/handlers/... -cover
go build -o bin/execution-service ./cmd/api
curl -H "Authorization: Bearer $TOKEN" http://localhost:8084/api/v1/purchase-orders
```
