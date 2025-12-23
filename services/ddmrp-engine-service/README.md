# DDMRP Engine Service

**Version**: 1.0.0
**Status**: Foundation Complete - In Development
**Phase**: 2B - New Microservices

---

## Overview

The DDMRP (Demand Driven Material Requirements Planning) Engine Service is the core calculation engine for inventory buffer management in the GIIA platform. It calculates Average Daily Usage (ADU), Decoupled Lead Time (DLT), Net Flow Equation, and maintains buffer zones (Red, Yellow, Green) for demand-driven inventory management.

---

## Features

### ✅ Implemented
- **Buffer Zone Calculation**: Red, Yellow, Green zones with DDMRP formulas
- **FAD System**: Factor de Ajuste de Demanda (Demand Adjustment Factor) with multi-factor support
- **Net Flow Position**: NFP calculation and zone determination
- **Buffer History**: Daily snapshots for trend analysis
- **Multi-Tenancy**: Organization-level isolation
- **Domain Models**: Complete entities with business logic
- **Database Migrations**: PostgreSQL schema with indexes and constraints

### 🔨 In Progress
- Repository implementations (GORM)
- ADU calculation use cases
- NFP update use cases
- gRPC API implementation
- Daily recalculation cron job
- Catalog service integration
- Event publishing (NATS)

### 📋 Planned
- Unit tests (85%+ coverage goal)
- Integration tests
- Configuration management
- Docker containerization
- API documentation

---

## Architecture

### Clean Architecture Layers

```
ddmrp-engine-service/
├── internal/core/              # Domain & Business Logic
│   ├── domain/                # Entities with business rules
│   ├── providers/             # Interface contracts
│   └── usecases/              # Application business logic
│
├── internal/infrastructure/    # External Adapters
│   ├── adapters/              # gRPC clients, Events
│   ├── repositories/          # Data access (GORM)
│   ├── entrypoints/           # gRPC handlers, Cron
│   └── database/              # Migrations
│
├── api/proto/                 # Protocol Buffers
├── cmd/server/                # Application entry point
└── config/                    # Configuration files
```

### Dependency Flow

```
External → Infrastructure → Use Cases → Domain
         (Adapters)       (Application) (Business Logic)
```

---

## DDMRP Calculations

### Buffer Zones

```
Red Zone (Safety Stock):
  Red Base = DLT × CPD × %LT
  Red Safe = Red Base × %CV
  Red Zone = Red Base + Red Safe

Yellow Zone (Demand Coverage):
  Yellow Zone = CPD × DLT

Green Zone (Order Frequency):
  Green Zone = MAX(MOQ, FO × CPD, DLT × CPD × %LT)

Total Buffer:
  Top of Red = Red Zone
  Top of Yellow = Red + Yellow
  Top of Green = Red + Yellow + Green
```

### Net Flow Position

```
NFP = On-Hand + On-Order - Qualified Demand

Buffer Penetration = (NFP / Top of Green) × 100

Zone Assignment:
  - NFP >= Top of Yellow  → Green (Normal)
  - NFP >= Top of Red     → Yellow (Monitor)
  - NFP > 0               → Red (Replenish Now)
  - NFP <= 0              → Below Red (Critical)
```

### FAD (Demand Adjustment)

```
CPD_Adjusted = CPD_Base × Factor₁ × Factor₂ × ... × Factorₙ
CPD_Final = CEILING(CPD_Adjusted)

Examples:
  - Factor = 1.5 → 50% demand increase (promotion)
  - Factor = 0.8 → 20% demand decrease
  - Factor = 0.0 → Product discontinuation
```

---

## Database Schema

### Tables

1. **adu_calculations** - Average Daily Usage records
2. **buffers** - Buffer zones and status for products
3. **demand_adjustments** - FAD adjustments (seasonal, promotions, etc.)
4. **buffer_adjustments** - Manual buffer zone adjustments
5. **buffer_history** - Daily buffer snapshots for trend analysis

### Key Indexes

- Multi-tenant: `organization_id` on all tables
- Performance: Compound indexes on frequently queried columns
- Lookups: `product_id + organization_id` unique constraints
- Time-series: Date indexes for historical queries

---

## API (Planned)

### gRPC Endpoints

#### Buffer Operations
- `CalculateBuffer` - Calculate/recalculate buffer zones
- `GetBuffer` - Retrieve buffer details
- `ListBuffers` - List buffers with filters
- `GetBufferStatus` - Get current status and alerts

#### FAD Operations
- `CreateFAD` - Create demand adjustment
- `UpdateFAD` - Update existing FAD
- `DeleteFAD` - Remove FAD
- `ListFADs` - List adjustments for product

#### Buffer Adjustment Operations
- `CreateBufferAdjustment` - Manual zone adjustment
- `ListBufferAdjustments` - List adjustments

#### NFP Operations
- `UpdateNFP` - Update inventory levels
- `CheckReplenishmentNeeded` - Get replenishment alerts

---

## Dependencies

### Internal Packages
- `pkg/config` - Configuration management
- `pkg/database` - Database connection with GORM
- `pkg/errors` - Typed error handling
- `pkg/events` - NATS event publishing
- `pkg/logger` - Structured logging

### External Services
- **Catalog Service** (gRPC) - Product, Supplier, BufferProfile data
- **Execution Service** (gRPC) - Inventory levels, orders
- **PostgreSQL 16** - Primary database
- **NATS JetStream** - Event streaming

### Third-Party Libraries
- `github.com/google/uuid` - UUID generation
- `github.com/robfig/cron/v3` - Scheduled jobs
- `google.golang.org/grpc` - gRPC framework
- `gorm.io/gorm` - ORM
- `github.com/stretchr/testify` - Testing

---

## Development

### Prerequisites

```bash
# Go 1.23.4 or later
go version

# PostgreSQL 16
psql --version

# Protocol Buffer Compiler
protoc --version
```

### Setup

```bash
# Navigate to service directory
cd services/ddmrp-engine-service

# Install dependencies
go mod download
go mod tidy

# Run migrations
make migrate-up
```

### Running Tests

```bash
# Unit tests
go test ./... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Specific package
go test ./internal/core/usecases/buffer -v
```

### Running the Service

```bash
# Development mode
make run

# Build binary
make build

# Run binary
./bin/ddmrp-engine-service
```

---

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=giia_ddmrp
DB_USER=postgres
DB_PASSWORD=postgres

# gRPC
GRPC_PORT=50053
GRPC_CATALOG_URL=localhost:50051

# NATS
NATS_URL=nats://localhost:4222

# Cron
DAILY_RECALC_ENABLED=true
DAILY_RECALC_CRON=0 2 * * *  # 2 AM daily
```

---

## Events Published

### Buffer Events

```
ddmrp.buffer.calculated       - Buffer zones recalculated
ddmrp.buffer.status_changed   - Zone changed (Green→Yellow, etc.)
ddmrp.buffer.alert_triggered  - Replenishment alert
```

### FAD Events

```
ddmrp.fad.created             - Demand adjustment created
ddmrp.fad.updated             - Demand adjustment updated
ddmrp.fad.deleted             - Demand adjustment deleted
```

---

## Performance Targets

- **ADU Calculation**: <2s p95
- **Buffer Calculation**: <3s p95
- **NFP Query**: <500ms p95
- **Batch Updates**: <10s for 1000 products
- **Daily Recalculation**: <5 minutes for 10,000 products

---

## Implementation Status

See [TASK_14_IMPLEMENTATION_STATUS.md](./TASK_14_IMPLEMENTATION_STATUS.md) for detailed progress.

**Current**: 30% Complete (Foundation Ready)
**Next**: Repository layer, remaining use cases, gRPC API

---

## Contributing

### Code Standards

- Follow Clean Architecture principles
- No external dependencies in `core/domain`
- Use typed errors from `pkg/errors`
- GORM tags on domain entities
- Validate all inputs at use case boundary
- snake_case for directories, camelCase for import aliases
- Multi-tenancy: Always filter by `organization_id`
- Test coverage: 85%+ goal

### Testing Standards

- Given-When-Then structure
- Specific mock expectations (avoid `mock.Anything`)
- One assertion per test case
- Test file naming: `*_test.go`

---

## License

Copyright © 2025 GIIA Platform. All rights reserved.
