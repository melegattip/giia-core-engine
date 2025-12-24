# Agent Prompt: Task 28 - Cross-Service Integration Testing

## 🤖 Agent Identity
Expert QA Engineer for end-to-end integration testing of microservices with Go, docker-compose, and CI/CD.

---

## 📋 Mission
Build comprehensive integration test suite for GIIA platform: E2E flows for orders, inventory, DDMRP, analytics, and auth across all 6 services.

---

## 📂 Files to Create

### Test Infrastructure
- `tests/integration/docker-compose.yml` - All services + infra
- `tests/integration/setup.go` - Test environment setup
- `tests/integration/teardown.go` - Cleanup
- `tests/integration/test_data_factory.go` - Factory for test data

### Service Clients
- `tests/integration/clients/auth_client.go`
- `tests/integration/clients/catalog_client.go`
- `tests/integration/clients/execution_client.go`
- `tests/integration/clients/ddmrp_client.go`
- `tests/integration/clients/analytics_client.go`
- `tests/integration/clients/ai_hub_client.go`

### E2E Test Suites
- `tests/integration/purchase_order_flow_test.go`
- `tests/integration/sales_order_flow_test.go`
- `tests/integration/analytics_aggregation_test.go`
- `tests/integration/auth_across_services_test.go`
- `tests/integration/multi_tenancy_test.go`
- `tests/integration/nats_events_test.go`

---

## 🔧 E2E Test Patterns

### Purchase Order Flow
```go
func TestPurchaseOrderFlow_CreateToReceive(t *testing.T) {
    // 1. Create user and authenticate
    // 2. Create product in catalog
    // 3. Create purchase order
    // 4. Wait for NATS event
    // 5. Verify DDMRP updated on-order
    // 6. Receive goods
    // 7. Verify inventory increased
    // 8. Verify buffer NFP updated
}
```

### Multi-Tenancy Isolation
```go
func TestMultiTenancy_Isolation(t *testing.T) {
    // Create two organizations
    // Create data in org A
    // Try to access from org B - should get 404
}
```

### Auth Across Services
```go
func TestAuth_JWTWorksAcrossServices(t *testing.T) {
    // Token works on all services
}

func TestAuth_ExpiredJWTRejected(t *testing.T) {
    // Expired token rejected by all services
}
```

---

## 🔧 Test Data Factory

```go
type TestDataFactory struct {
    authClient    *AuthClient
    catalogClient *CatalogClient
    executionClient *ExecutionClient
}

func (f *TestDataFactory) CreateCompleteSetup(ctx context.Context) *TestSetup {
    // Create org, users, products, buffers, orders
}
```

---

## ✅ Success Criteria
- [ ] 20+ integration test scenarios
- [ ] All critical flows tested E2E
- [ ] Tests run in CI/CD <10 minutes
- [ ] 100% pass before deploy
- [ ] Test environment fully automated
- [ ] Multi-tenancy isolation verified

---

## 🚀 Commands
```bash
cd tests/integration
docker-compose up -d
./wait-for-services.sh
go test -v -timeout 10m ./...
docker-compose down -v
```
