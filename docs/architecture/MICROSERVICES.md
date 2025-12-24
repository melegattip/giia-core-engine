# GIIA Microservices Architecture

**Version**: 1.0  
**Last Updated**: 2025-12-23

---

## 📖 Overview

The GIIA platform consists of **6 microservices** designed around business capabilities. Each service is independently deployable, has its own database schema, and communicates via gRPC (synchronous) or NATS (asynchronous).

---

## 🧩 Service Catalog

### 1. Auth Service

**Purpose**: Authentication, authorization, and multi-tenancy management

**Responsibilities**:
- User registration, login, logout
- JWT token management (access + refresh)
- Role-Based Access Control (RBAC)
- Organization (tenant) management
- Permission validation

**Key Entities**:
- User
- Organization
- Role
- Permission
- RefreshToken
- ActivationToken
- PasswordResetToken

**API Exposure**:
- REST: `/api/v1/auth/*`
- gRPC: Port 9081

**Database Schema**: `auth`

**Status**: 🟢 95% Complete

---

### 2. Catalog Service

**Purpose**: Master data management for products and suppliers

**Responsibilities**:
- Product catalog management
- Supplier information
- Buffer profile definitions
- Product-supplier relationships

**Key Entities**:
- Product
- Supplier
- BufferProfile
- ProductSupplier (relationship)

**API Exposure**:
- REST: `/api/v1/products`, `/api/v1/suppliers`, `/api/v1/buffer-profiles`
- gRPC: Port 9082 (planned)

**Database Schema**: `catalog`

**Status**: 🟢 85% Complete

---

### 3. DDMRP Engine Service

**Purpose**: Core DDMRP calculations and buffer management

**Responsibilities**:
- Buffer zone calculations (Red, Yellow, Green)
- Average Daily Usage (ADU) calculation
- Net Flow Equation (NFP)
- Demand Adjustment Factor (FAD)
- Buffer status monitoring
- Daily buffer recalculation

**Key Entities**:
- Buffer
- ADUCalculation
- DemandAdjustment
- BufferAdjustment
- BufferHistory

**Implemented Use Cases**:
- `CalculateBuffer` - Buffer zone calculations with tests
- `GetBuffer` - Retrieve buffer details
- `ListBuffers` - List with filtering
- `RecalculateAllBuffers` - Batch recalculation
- ADU calculation use cases
- NFP update use cases
- Demand adjustment CRUD

**API Exposure**:
- gRPC: Port 50053 (in development)

**Database Schema**: `ddmrp`

**Status**: 🟢 65% Complete (Core functionality ready - Domain, Use Cases, Repositories, gRPC Handlers, Cron, Adapters implemented; Unit tests and service registration pending)

---

### 4. Execution Service

**Purpose**: Order and inventory transaction management

**Responsibilities**:
- Purchase order management
- Sales order management
- Inventory transactions
- Alert management
- Stock level tracking

**Key Entities** (with unit tests):
- PurchaseOrder / PurchaseOrderLine
- SalesOrder / SalesOrderLine
- InventoryTransaction
- InventoryBalance
- Alert

**Implemented Use Cases**:
- Purchase Order: `CreatePO`, `CancelPO`, `ReceivePO` (with tests)
- Sales Order: `CreateSO`, `CancelSO`, `ShipSO`, `CompleteSO`
- Inventory: Balance queries, Transaction recording
- Alerts: Alert generation and management

**API Exposure**:
- REST: `/api/v1/orders/*`, `/api/v1/inventory/*`
- gRPC: Port 50054

**Database Schema**: `execution`

**Status**: 🟢 75% Complete (Domain and Use Cases with full test coverage; Repositories, Handlers, and Infrastructure pending)

---

### 5. Analytics Service

**Purpose**: KPIs, reporting, and dashboards

**Responsibilities**:
- Dashboard KPI calculations
- Inventory rotation metrics
- Days in inventory analysis
- Immobilized inventory tracking
- Buffer analytics synchronization

**Key Entities**:
- KPISnapshot
- DaysInInventory
- ImmobilizedInventory
- InventoryRotation
- BufferAnalytics

**Implemented Use Cases** (92.5% test coverage):
- `CalculateDaysInInventory` - Valorized days calculation
- `CalculateImmobilizedInventory` - Aged stock analysis
- `CalculateInventoryRotation` - Turnover metrics
- `SyncBufferAnalytics` - DDMRP buffer data sync

**API Exposure**:
- REST: HTTP Port 8083
- gRPC: Port 50053

**Database Schema**: `analytics`

**Status**: 🟢 85% Complete (Core KPI use cases with 92.5% test coverage; Infrastructure adapters in progress)

---

### 6. AI Intelligence Hub

**Purpose**: AI-powered intelligence and notifications

**Responsibilities**:
- Event monitoring and processing (NATS JetStream)
- AI-powered analysis (Claude integration)
- RAG-based knowledge retrieval (DDMRP methodology)
- Intelligent notifications with recommendations
- Impact assessment and priority scoring

**Key Entities**:
- AINotification (with ImpactAssessment)
- Recommendation (action, effort, impact)
- UserNotificationPreferences

**Implemented Features**:
- ✅ NATS JetStream event subscription
- ✅ Buffer event processing
- ✅ Claude AI client (with mock fallback)
- ✅ RAG knowledge retrieval (7,400+ words)
- ✅ Stockout risk analysis use case
- ✅ Notification persistence (PostgreSQL)
- ✅ Event handlers (buffer, execution, user)

**API Exposure**:
- REST: HTTP endpoints (planned)
- WebSocket: Real-time notifications (planned)
- gRPC: Port 50086 (planned)

**Database Schema**: `intelligence_hub`

**Status**: 🟢 80% Complete - MVP Operational (Event processing, AI analysis, notifications fully working; HTTP/gRPC API and advanced features pending)

---

## 🔗 Service Dependencies

```
                    ┌──────────────────┐
                    │   Auth Service   │
                    │ (Authentication) │
                    └────────┬─────────┘
                             │ gRPC: ValidateToken, CheckPermission
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│ Catalog Service │ │ Execution Svc   │ │ Analytics Svc   │
│ (Master Data)   │ │ (Transactions)  │ │ (Reporting)     │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         │ gRPC: GetProduct  │ Events: order.*  │ Events: kpi.*
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                    DDMRP Engine Service                      │
│                  (Buffer Calculations)                       │
└──────────────────────────┬──────────────────────────────────┘
                           │ Events: buffer.*
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                   AI Intelligence Hub                        │
│           (Notifications, Recommendations)                   │
└─────────────────────────────────────────────────────────────┘
```

---

## 📡 Communication Patterns

### Synchronous (gRPC)

Used when immediate response is required:

| From | To | Method | Purpose |
|------|----|--------|---------|
| Any Service | Auth | ValidateToken | Verify JWT token |
| Any Service | Auth | CheckPermission | Verify user permission |
| Execution | Catalog | GetProduct | Get product details |
| DDMRP | Catalog | GetBufferProfile | Get buffer profile |

### Asynchronous (NATS Events)

Used for decoupled, event-driven communication:

| Event | Published By | Consumed By | Purpose |
|-------|-------------|-------------|---------|
| user.logged_in | Auth | Analytics, AI Hub | Audit, insights |
| product.created | Catalog | DDMRP, AI Hub | Buffer creation |
| buffer.breached | DDMRP | Execution, AI Hub | Alerts, orders |
| order.created | Execution | Analytics, AI Hub | KPI updates |
| kpi.calculated | Analytics | AI Hub | AI analysis |

---

## 🗄️ Data Ownership

Each service owns its data exclusively:

| Service | Schema | Tables |
|---------|--------|--------|
| Auth | `auth` | users, organizations, roles, permissions, tokens |
| Catalog | `catalog` | products, suppliers, buffer_profiles |
| DDMRP | `ddmrp` | buffers, calculations, adjustments |
| Execution | `execution` | orders, inventory, transactions |
| Analytics | `analytics` | snapshots, reports, kpis |
| AI Hub | `ai_hub` | notifications, recommendations |

**Rule**: Services NEVER directly access another service's database. Use gRPC or events.

---

## 🛡️ Multi-Tenancy

All tenant-scoped data includes `organization_id`:

- Enforced via GORM scopes in repositories
- JWT tokens include organization context
- Middleware injects tenant context into request

---

## 🚀 Deployment

### Local Development

```bash
# Start infrastructure
docker-compose up -d

# Run specific service
cd services/auth-service
go run cmd/api/main.go
```

### Kubernetes

Each service has a Helm chart in `/k8s/services/`:

```bash
# Deploy all services
make k8s-deploy-services
```

---

## 📊 Service Health

| Service | Health Endpoint | Readiness | Liveness |
|---------|----------------|-----------|----------|
| Auth | `/health` | `/ready` | `/live` |
| Catalog | `/health` | `/ready` | `/live` |
| DDMRP | `/health` | `/ready` | `/live` |
| Execution | `/health` | `/ready` | `/live` |
| Analytics | `/health` | `/ready` | `/live` |
| AI Hub | `/health` | `/ready` | `/live` |

---

## 📚 Related Documentation

- [Architecture Overview](./OVERVIEW.md)
- [Clean Architecture Patterns](./CLEAN_ARCHITECTURE.md)
- [Data Model](./DATA_MODEL.md)
- [API Reference](/docs/api/PUBLIC_RFC.md)

---

**Maintained by the GIIA Team** 🧩
