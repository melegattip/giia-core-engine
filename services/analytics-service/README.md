# Analytics Service

**Version**: 1.0.0  
**Status**: 🟢 90% Complete - Core Functionality Ready  
**Phase**: 2B - New Microservices  
**Last Updated**: 2025-12-23

The Analytics Service provides KPI calculations, business intelligence metrics, and reporting capabilities for the GIIA platform.

## Overview

This service aggregates data from all operational services (Catalog, DDMRP Engine, Execution) and provides:
- Inventory performance KPIs
- DDMRP buffer analytics
- Days in inventory tracking
- Immobilized inventory analysis
- Inventory rotation metrics
- Dashboard APIs (gRPC and REST)
- Historical trend analysis

## Architecture

```
analytics-service/
├── api/proto/                    # Protocol Buffers definitions
├── cmd/server/                   # Main application entry point
├── internal/
│   ├── core/                    # Domain layer (business logic)
│   │   ├── domain/             # Entities with validation
│   │   ├── providers/          # Interface definitions
│   │   └── usecases/           # Application logic
│   └── infrastructure/          # Infrastructure layer
│       ├── adapters/           # External service adapters
│       ├── entrypoints/        # gRPC/HTTP handlers
│       └── persistence/        # Database repositories
├── Makefile                     # Build and development commands
└── go.mod                       # Go module dependencies
```

## Quick Start

### Prerequisites
- Go 1.24.0+
- PostgreSQL 16+
- Protocol Buffers compiler (protoc)
- NATS JetStream (for events)

### Installation

```bash
# Install dependencies
make deps

# Generate protobuf code
make proto

# Run tests
make test

# Run with coverage
make coverage
```

### Running Locally

```bash
# Run the service
make run

# Or build and run
make build
./bin/analytics-service
```

### Environment Variables

Create a `.env` file:

```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/analytics_db

# gRPC Server
GRPC_PORT=50053
HTTP_PORT=8083

# Service Clients
CATALOG_SERVICE_URL=localhost:50051
DDMRP_SERVICE_URL=localhost:50052
EXECUTION_SERVICE_URL=localhost:50054

# NATS
NATS_URL=nats://localhost:4222

# Logging
LOG_LEVEL=info
```

## KPIs Provided

### Days in Inventory (Valorizado)
Tracks the total and average valued days products remain in inventory.
- Formula: `ValuedDays = DaysInStock × (Quantity × UnitCost)`
- Endpoint: `GetDaysInInventoryKPI`

### Immobilized Inventory
Identifies inventory older than a configurable threshold.
- Formula: `ImmobilizedPercentage = (ImmobilizedValue / TotalStockValue) × 100`
- Endpoint: `GetImmobilizedInventoryKPI`

### Inventory Rotation
Measures how quickly inventory turns over.
- Formula: `RotationRatio = SalesLast30Days / AvgMonthlyStock`
- Endpoint: `GetInventoryRotationKPI`

### Buffer Analytics
Daily snapshots of DDMRP buffer configurations for trend analysis.
- Synchronized from DDMRP Engine service
- Endpoint: `GetBufferAnalytics`

## API Reference

### gRPC Service

```protobuf
service AnalyticsService {
  rpc GetKPISnapshot(GetKPISnapshotRequest) returns (GetKPISnapshotResponse);
  rpc GetDaysInInventoryKPI(GetDaysInInventoryKPIRequest) returns (GetDaysInInventoryKPIResponse);
  rpc GetImmobilizedInventoryKPI(GetImmobilizedInventoryKPIRequest) returns (GetImmobilizedInventoryKPIResponse);
  rpc GetInventoryRotationKPI(GetInventoryRotationKPIRequest) returns (GetInventoryRotationKPIResponse);
  rpc GetBufferAnalytics(GetBufferAnalyticsRequest) returns (GetBufferAnalyticsResponse);
}
```

## Database Migrations

Migrations are located in `internal/infrastructure/persistence/migrations/`.

```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Generate coverage report
make coverage

# View core package coverage
make coverage-core
```

**Test Coverage**: 92.5% (Domain: 91.6%, Use Cases: 94.4%)

## Development

### Adding a New KPI

1. Define domain entity in `internal/core/domain/`
2. Add repository methods to `internal/core/providers/kpi_repository.go`
3. Implement use case in `internal/core/usecases/kpi/`
4. Add tests (aim for 85%+ coverage)
5. Update proto definitions
6. Implement gRPC handler

### Code Quality

```bash
# Run linters
make lint

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Deployment

### Docker

```bash
# Build image
docker build -t analytics-service:latest .

# Run container
docker run -p 50053:50053 -p 8083:8083 analytics-service:latest
```

### Kubernetes

```bash
# Apply deployment
kubectl apply -f deployments/kubernetes/

# Check status
kubectl get pods -l app=analytics-service
```

## Monitoring

- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics`
- **gRPC Health**: Use `grpc.health.v1.Health` service

## Implementation Status

**Current**: 🟢 90% Complete

| Component | Status | Notes |
|-----------|--------|-------|
| Domain Entities | ✅ 100% | 5 KPI entities with validation |
| Domain Tests | ✅ 100% | 5 test files (91.6% coverage) |
| Use Cases | ✅ 100% | 4 KPI calculation use cases |
| Use Case Tests | ✅ 100% | 4 test files (94.4% coverage) |
| Providers/Interfaces | ✅ 100% | Repository contracts defined |
| Repository | ✅ 100% | Full GORM implementation (18KB) |
| Database Migrations | ✅ 100% | 10 migration files |
| Proto Definitions | ✅ 100% | gRPC service defined |
| Main Entry Point | ✅ 100% | cmd/server with DI |
| gRPC/HTTP Handlers | 🔨 50% | In progress |
| External Adapters | 🔨 50% | Catalog, DDMRP, Execution clients |
| Integration Tests | ⏸️ 0% | Not started |

**Overall Test Coverage**: 92.5% (Domain: 91.6%, Use Cases: 94.4%)

### 🔨 Pending Items

- Complete gRPC/HTTP handler implementations
- External service adapters (Catalog, DDMRP, Execution)
- Integration tests
- End-to-end testing
- API documentation

---

## Contributing

1. Follow Clean Architecture principles
2. Write comprehensive tests (85%+ coverage)
3. Use typed domain errors
4. Document all public APIs
5. Run linters before committing

## License

Proprietary - GIIA Platform

## Support

For issues or questions, contact the platform team.
