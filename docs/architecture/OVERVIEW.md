# GIIA Architecture Overview

**Version**: 1.0  
**Last Updated**: 2025-12-23  
**Architecture Style**: Monorepo Microservices with Clean Architecture

---

## 📖 Table of Contents

- [Executive Summary](#executive-summary)
- [System Architecture](#system-architecture)
- [Microservices Overview](#microservices-overview)
- [Communication Patterns](#communication-patterns)
- [Data Architecture](#data-architecture)
- [Infrastructure](#infrastructure)
- [Security Architecture](#security-architecture)

---

## Executive Summary

GIIA (Gestión Inteligente de Inventario con IA) is a **SaaS platform** implementing **DDMRP (Demand Driven Material Requirements Planning)** with AI-powered assistance. The platform follows a **monorepo microservices architecture** with 6 independent services sharing common infrastructure packages.

### Key Architectural Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Language** | Go 1.23 | Performance, concurrency, strong typing |
| **Architecture** | Clean Architecture | Testability, maintainability, independence |
| **Communication** | gRPC + NATS | Low latency sync + async event-driven |
| **Multi-tenancy** | Row-level isolation | organization_id in all entities |
| **Deployment** | Kubernetes | Scalability, resilience, portability |

---

## System Architecture

### High-Level Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                           EXTERNAL CLIENTS                                │
│                   (Web App, Mobile App, ERP Systems)                     │
└─────────────────────────────────┬────────────────────────────────────────┘
                                  │ HTTPS/WSS
                                  ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                        API GATEWAY / INGRESS                              │
│                    (NGINX Ingress Controller)                             │
└─────────────────────────────────┬────────────────────────────────────────┘
                                  │
         ┌────────────────────────┼────────────────────────┐
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Auth Service   │    │ Catalog Service │    │ DDMRP Engine    │
│  (Multi-tenant) │    │  (Master Data)  │    │ (Core Logic)    │
│   REST + gRPC   │    │   REST + gRPC   │    │     gRPC        │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                      │                      │
         └──────────────────────┼──────────────────────┘
                                │ gRPC (inter-service)
         ┌──────────────────────┼──────────────────────┐
         │                      │                      │
         ▼                      ▼                      ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Execution Svc   │    │ Analytics Svc   │    │  AI Intel Hub   │
│ (Orders/Inv)    │    │ (KPIs/Reports)  │    │ (AI Assistant)  │
│   REST + gRPC   │    │   REST + gRPC   │    │  REST + WSS     │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                      │                      │
         └──────────────────────┼──────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                         MESSAGE BUS (NATS JETSTREAM)                      │
│            AUTH_EVENTS | CATALOG_EVENTS | DDMRP_EVENTS | ...             │
└──────────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                         DATA LAYER                                        │
│  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐              │
│  │  PostgreSQL 16 │  │   Redis 7      │  │  NATS Jetstream│              │
│  │  (Persistence) │  │  (Cache/Session)│  │  (Event Store) │              │
│  └────────────────┘  └────────────────┘  └────────────────┘              │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Microservices Overview

### Service Catalog

| Service | Responsibility | Status | Port (HTTP/gRPC) |
|---------|----------------|--------|------------------|
| **Auth Service** | Authentication, authorization, multi-tenancy, RBAC | 95% | 8081/9081 |
| **Catalog Service** | Products, suppliers, buffer profiles | 85% | 8082/9082 |
| **DDMRP Engine** | Buffer calculations, ADU, Net Flow Equation | 0% | 8083/9083 |
| **Execution Service** | Orders, inventory transactions, replenishment | 0% | 8084/9084 |
| **Analytics Service** | KPIs, reports, dashboards | 0% | 8085/9085 |
| **AI Intelligence Hub** | AI assistant, notifications, insights | 40% | 8086/9086 |

### Clean Architecture Layers

Each microservice follows **Clean Architecture** with clear layer separation:

```
service/
├── cmd/api/                    # Entry point
│   └── main.go
├── internal/
│   ├── core/                   # 🧠 BUSINESS LOGIC (Framework-independent)
│   │   ├── domain/            # Entities, value objects
│   │   ├── usecases/          # Use cases (business logic)
│   │   └── providers/         # Interface contracts
│   │
│   └── infrastructure/         # 🔌 EXTERNAL ADAPTERS
│       ├── adapters/          # External service implementations
│       ├── repositories/      # Data access (GORM)
│       └── entrypoints/       # HTTP/gRPC handlers
│
├── api/proto/                  # gRPC definitions
└── migrations/                 # Database migrations
```

**Dependency Rule**: Inner layers never depend on outer layers. Core knows nothing about infrastructure.

---

## Communication Patterns

### Synchronous (gRPC)

Used for **request-response** patterns where immediate result is needed:

- Auth token validation
- Permission checks
- Data queries between services
- User information retrieval

**Example: Auth Service gRPC**

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  rpc BatchCheckPermissions(BatchCheckPermissionsRequest) returns (BatchCheckPermissionsResponse);
  rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
```

### Asynchronous (NATS Jetstream)

Used for **event-driven** patterns where decoupling is important:

- Audit logging
- Notification triggers
- Analytics data collection
- Cross-service state synchronization

**Event Streams:**

| Stream | Purpose | Events |
|--------|---------|--------|
| `AUTH_EVENTS` | Authentication events | user.logged_in, role.assigned |
| `CATALOG_EVENTS` | Catalog changes | product.created, supplier.updated |
| `DDMRP_EVENTS` | Buffer calculations | buffer.calculated, zone.breached |
| `EXECUTION_EVENTS` | Order/Inventory | order.created, inventory.adjusted |
| `ANALYTICS_EVENTS` | KPI updates | kpi.calculated, report.generated |
| `AI_AGENT_EVENTS` | AI activities | insight.generated, notification.sent |
| `DLQ_EVENTS` | Dead letter queue | Failed events for retry |

---

## Data Architecture

### Database Strategy

- **PostgreSQL 16**: Primary data store for all services
- **Multi-schema approach**: Each service owns its schema in shared database
- **Row-level tenant isolation**: `organization_id` in all tenant-scoped tables

**Schemas:**
- `auth` - Users, organizations, roles, permissions
- `catalog` - Products, suppliers, buffer profiles
- `ddmrp` - Buffers, calculations, history
- `execution` - Orders, inventory, transactions
- `analytics` - KPIs, reports, snapshots
- `ai_hub` - Notifications, recommendations

### Redis Usage

| Purpose | Database | TTL |
|---------|----------|-----|
| Session/Token blacklist | 0 | 15 min |
| Permission cache | 0 | 5 min |
| Rate limiting | 1 | Varies |
| Buffer calculations cache | 2 | 1 hour |

---

## Infrastructure

### Container Orchestration

- **Kubernetes** for production and staging
- **Docker Compose** for local development
- **Helm Charts** for service deployment

### Observability Stack

| Tool | Purpose |
|------|---------|
| **Prometheus** | Metrics collection |
| **Grafana** | Dashboards and visualization |
| **Loki** | Log aggregation |
| **Jaeger** | Distributed tracing (planned) |

---

## Security Architecture

### Authentication

- **JWT-based** with access (15 min) and refresh (7 days) tokens
- Access tokens include: `user_id`, `organization_id`, `roles`, `permissions`
- Refresh tokens stored hashed in database

### Authorization (RBAC)

- **Role-based Access Control** with permission inheritance
- Permissions cached in Redis for performance
- Checked via gRPC for all protected operations

### Multi-Tenancy

- **Row-level isolation** via `organization_id`
- Automatic query filtering using GORM scopes
- Organization context injected from JWT claims

### Security Controls

- bcrypt password hashing (cost 12)
- Rate limiting (Redis-based)
- Input validation at all layers
- SQL injection prevention (parameterized queries)
- Secrets via environment variables or K8s secrets

---

## Shared Packages

Located in `/pkg/`:

| Package | Purpose | Status |
|---------|---------|--------|
| `config` | Viper configuration management | 90% |
| `logger` | Zerolog structured logging | 95% |
| `database` | GORM connection pool | 90% |
| `errors` | Typed error system | 100% |
| `events` | NATS publisher/subscriber | 85% |
| `middleware` | Common HTTP/gRPC middleware | 80% |
| `monitoring` | Prometheus metrics | 70% |

---

## Related Documentation

- [Microservices Deep Dive](./MICROSERVICES.md)
- [Clean Architecture Patterns](./CLEAN_ARCHITECTURE.md)
- [Data Model](./DATA_MODEL.md)
- [DDMRP Methodology](./DDMRP_METHODOLOGY.md)

---

**Architecture maintained by the GIIA Team** 🏗️
