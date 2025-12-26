# GIIA Integration Test Suite

Comprehensive end-to-end integration tests for the GIIA platform, testing all 6 microservices working together.

## 🎯 Overview

This test suite verifies:
- **Authentication flows** across all services
- **Purchase order lifecycle** (create → receive → inventory update → DDMRP update)
- **Sales order lifecycle** (create → ship → inventory decrease)
- **Multi-tenancy isolation** between organizations
- **Analytics aggregation** across services
- **NATS event publishing** and consumption

## 📂 Structure

```
tests/integration/
├── docker-compose.yml          # All services + infrastructure
├── setup.go                    # Test environment setup
├── teardown.go                 # Cleanup utilities
├── test_data_factory.go        # Factory for test data
├── clients.go                  # Client re-exports
├── clients/
│   ├── auth_client.go          # Auth service client
│   ├── catalog_client.go       # Catalog service client
│   ├── execution_client.go     # Execution service client
│   ├── ddmrp_client.go         # DDMRP Engine client
│   ├── analytics_client.go     # Analytics service client
│   └── ai_hub_client.go        # AI Hub client
├── purchase_order_flow_test.go # PO lifecycle tests
├── sales_order_flow_test.go    # SO lifecycle tests
├── analytics_aggregation_test.go # Analytics tests
├── auth_across_services_test.go # Cross-service auth tests
├── multi_tenancy_test.go       # Multi-tenancy isolation tests
├── nats_events_test.go         # NATS event tests
├── run-tests.sh                # Linux/Mac test runner
└── run-tests.bat               # Windows test runner
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.24+
- Access to Docker Hub for pulling images

### Running Tests

**Linux/Mac:**
```bash
cd tests/integration
chmod +x run-tests.sh
./run-tests.sh
```

**Windows:**
```batch
cd tests\integration
run-tests.bat
```

### Options

| Option | Description |
|--------|-------------|
| `-v, --verbose` | Run tests with verbose output |
| `-run PATTERN` | Run only tests matching pattern |
| `--timeout DURATION` | Set test timeout (default: 10m) |
| `--skip-setup` | Skip docker-compose up |
| `--skip-teardown` | Skip docker-compose down |

### Examples

```bash
# Run all tests verbosely
./run-tests.sh -v

# Run only auth tests
./run-tests.sh -run "Auth"

# Run only multi-tenancy tests
./run-tests.sh -run "MultiTenancy"

# Keep services running after tests
./run-tests.sh --skip-teardown
```

## 🔧 Test Scenarios

### Purchase Order Flow (5 tests)
- `TestPurchaseOrderFlow_CreateToReceive` - Complete PO lifecycle
- `TestPurchaseOrderFlow_CreateAndCancel` - PO cancellation
- `TestPurchaseOrderFlow_PartialReceive` - Partial goods receiving

### Sales Order Flow (4 tests)
- `TestSalesOrderFlow_CreateToShip` - Complete SO lifecycle
- `TestSalesOrderFlow_CreateAndCancel` - SO cancellation
- `TestSalesOrderFlow_InsufficientInventory` - Inventory validation
- `TestSalesOrderFlow_MultipleItems` - Multi-item orders

### Authentication Across Services (6 tests)
- `TestAuth_JWTWorksAcrossServices` - Token works on all services
- `TestAuth_ExpiredJWTRejected` - Expired tokens rejected
- `TestAuth_InvalidJWTRejected` - Invalid tokens rejected
- `TestAuth_TokenRefresh` - Token refresh flow
- `TestAuth_Logout` - Session invalidation
- `TestAuth_OrganizationIsolation` - Org-scoped access

### Multi-Tenancy Isolation (4 tests)
- `TestMultiTenancy_Isolation` - Data isolation between orgs
- `TestMultiTenancy_ConcurrentOperations` - Safe concurrent access
- `TestMultiTenancy_DataLeakPrevention` - No cross-org data leaks
- `TestMultiTenancy_OrganizationScopedInventory` - Inventory isolation

### Analytics Aggregation (5 tests)
- `TestAnalyticsAggregation_BufferAnalytics` - Buffer analytics
- `TestAnalyticsAggregation_Snapshot` - Snapshot retrieval
- `TestAnalyticsAggregation_SyncBufferData` - Data sync
- `TestAnalyticsAggregation_AfterTransactions` - Post-transaction updates
- `TestAnalyticsAggregation_CrossServiceData` - Cross-service aggregation

### NATS Events (6 tests)
- `TestNATSEvents_ProductCreated` - Product creation events
- `TestNATSEvents_PurchaseOrderCreated` - PO creation events
- `TestNATSEvents_GoodsReceived` - Goods receipt events
- `TestNATSEvents_DDMRPBufferUpdate` - DDMRP buffer events
- `TestNATSEvents_EventOrdering` - Event sequence verification
- `TestNATSEvents_JetStreamDurability` - JetStream verification

## 🏗️ Architecture

### Service Ports (Test Environment)

| Service | HTTP Port | gRPC Port |
|---------|-----------|-----------|
| Auth | 8183 | 9191 |
| Catalog | 8182 | - |
| Execution | 8184 | 9192 |
| DDMRP | 8185 | 9193 |
| Analytics | 8186 | 9194 |
| AI Hub | 8187 | 9195 |

### Infrastructure Ports

| Service | Port |
|---------|------|
| PostgreSQL | 5433 |
| Redis | 6380 |
| NATS | 4223 |
| NATS Monitoring | 8223 |

## 📊 Success Criteria

- [x] 20+ integration test scenarios
- [x] All critical flows tested E2E
- [ ] Tests run in CI/CD <10 minutes
- [ ] 100% pass before deploy
- [x] Test environment fully automated
- [x] Multi-tenancy isolation verified

## 🔍 Debugging

### View Service Logs
```bash
docker-compose -f docker-compose.yml logs auth-service
docker-compose -f docker-compose.yml logs catalog-service
```

### Connect to Test Database
```bash
docker exec -it giia-test-postgres psql -U giia_test -d giia_test
```

### Check NATS
```bash
curl http://localhost:8223/varz
```

## 🧪 Writing New Tests

1. Use the `DefaultTestEnvironment()` for setup
2. Use `clients` package for service interactions
3. Use `generateTestEmail()` and `generateTestSKU()` helpers
4. Always clean up with `defer env.Teardown()`
5. Skip in short mode with `testing.Short()`

Example:
```go
func TestMyFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    env := DefaultTestEnvironment()
    defer env.Teardown()

    authClient := clients.NewAuthClient(env.AuthService.HTTPURL)
    // ... test logic
}
```

## 📝 License

Part of the GIIA Core Engine project.
