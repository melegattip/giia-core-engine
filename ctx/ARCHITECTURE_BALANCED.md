# GIIA - Arquitectura Balanceada (Recomendada)

## Microservicios con Buenas Prácticas - El Punto Óptimo

**Versión:** 1.0
**Fecha:** Diciembre 2025
**Enfoque:** Balance entre robustez, escalabilidad y pragmatismo
**Recomendación:** **Esta es la arquitectura recomendada para GIIA**

---

## Tabla de Contenidos

1. [Por Qué Esta Arquitectura](#1-por-qué-esta-arquitectura)
2. [Visión General](#2-visión-general)
3. [Microservicios Core (6 servicios)](#3-microservicios-core)
4. [Clean Architecture por Servicio](#4-clean-architecture-por-servicio)
5. [AI Agent Balanceado](#5-ai-agent-balanceado)
6. [Modelo de Datos](#6-modelo-de-datos)
7. [Comunicación entre Servicios](#7-comunicación-entre-servicios)
8. [Infraestructura](#8-infraestructura)
9. [Seguridad](#9-seguridad)
10. [Observabilidad](#10-observabilidad)
11. [Plan de Implementación](#11-plan-de-implementación)
12. [Comparación con Otras Arquitecturas](#12-comparación-con-otras-arquitecturas)

---

## 1. Por Qué Esta Arquitectura

### 1.1 El Problema con las Alternativas

**Arquitectura Enterprise (ARCHITECTURE.md)**:
- ✅ Escalable a millones de usuarios
- ❌ Over-engineering para GIIA (10+ servicios, 5 bases de datos)
- ❌ 12-18 meses de desarrollo
- ❌ Equipo de 10+ ingenieros requerido
- ❌ Costo operacional alto ($5K+/mes)

**Arquitectura Pragmática (ARCHITECTURE_PRAGMATIC.md)**:
- ✅ Rápida de implementar (4-6 meses)
- ✅ Bajo costo ($100/mes inicialmente)
- ❌ Monolito limita escalamiento de equipos
- ❌ Difícil separar responsabilidades a futuro
- ❌ AI Agent muy básico

### 1.2 La Solución: Arquitectura Balanceada

**Esta arquitectura es el punto óptimo**:
- ✅ **6 microservicios** (no 1, no 10+) - balance perfecto
- ✅ **Clean Architecture + DDD** en cada servicio
- ✅ **PostgreSQL + Redis** (suficiente para 90% de casos)
- ✅ **AI Agent robusto** pero sin complejidad innecesaria
- ✅ **Auth/IAM dedicado** (multi-tenancy + RBAC + JWT + 2FA)
- ✅ **Kubernetes** (estándar de industria, fácil contratar)
- ✅ **6-9 meses** de implementación
- ✅ **Equipo de 4-6 ingenieros**
- ✅ **Costo operacional razonable** ($500-1K/mes a escala)

### 1.3 ¿Para Quién Es Esta Arquitectura?

Esta arquitectura es ideal si:
- Quieres **lanzar un producto SaaS serio** (no un MVP rápido)
- Planeas **crecer a cientos/miles de clientes**
- Tienes un **equipo técnico competente** (o lo construirás)
- Valoras **buenas prácticas** pero no quieres **over-engineering**
- Quieres **atraer inversión** (VCs valoran arquitecturas sólidas)

---

## 2. Visión General

### 2.1 Diagrama de Arquitectura

```
┌────────────────────────────────────────────────────────────────┐
│                      FRONTEND (React/Next.js)                  │
│  • Dashboard DDMRP  • ABM  • AI Chat  • Mobile (React Native)  │
└────────────────────────────────────────────────────────────────┘
                             ↓ HTTPS/WSS
┌────────────────────────────────────────────────────────────────┐
│                     API GATEWAY (Go - Kong/Traefik)            │
│  • JWT Auth  • Rate Limiting  • Routing  • TLS                 │
└────────────────────────────────────────────────────────────────┘
                             ↓
┌────────────────────────────────────────────────────────────────┐
│                    MICROSERVICIOS CORE (Go)                    │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  1. DDMRP ENGINE SERVICE                             │     │
│  │  • Buffer calculation (zones, NFE)                   │     │
│  │  • CPD calculation                                   │     │
│  │  • Replenishment suggestions                         │     │
│  │  • FAP (Planned Adjustment Factors)                  │     │
│  │  Puertos: gRPC (interno) + REST (externo)            │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  2. CATALOG SERVICE                                  │     │
│  │  • Products, Suppliers, Nodes (ABM)                  │     │
│  │  • Buffer Profiles                                   │     │
│  │  • BOM (Bill of Materials)                           │     │
│  │  • Multi-tenancy management                          │     │
│  │  Puertos: gRPC + REST                                │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  3. EXECUTION SERVICE                                │     │
│  │  • Purchase Orders / Production Orders               │     │
│  │  • Inventory transactions                            │     │
│  │  • Synchronization alerts                            │     │
│  │  • ERP integrations (connectors)                     │     │
│  │  Puertos: gRPC + REST + Webhooks                     │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  4. ANALYTICS SERVICE                                │     │
│  │  • KPI Dashboard                                     │     │
│  │  • Variance analysis                                 │     │
│  │  • DDS&OP projections                                │     │
│  │  • Stockout predictions                              │     │
│  │  Puertos: REST + WebSocket (real-time)               │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  5. AI AGENT SERVICE                                 │     │
│  │  • Chat handler (WebSocket)                          │     │
│  │  • OpenAI integration (gpt-4o-mini)                  │     │
│  │  • Function calling → invoke other services via gRPC │     │
│  │  • Proactive analysis (cron triggers)                │     │
│  │  • Context enrichment (simple RAG)                   │     │
│  │  Puertos: WebSocket + gRPC (client)                  │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                                │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  6. AUTH/IAM SERVICE                                 │     │
│  │  • User management (CRUD)                            │     │
│  │  • Authentication (login, JWT, refresh tokens)       │     │
│  │  • Multi-tenancy (tenant management)                 │     │
│  │  • RBAC (roles, permissions)                         │     │
│  │  • 2FA/TOTP, password reset, email verification     │     │
│  │  • Session management, account locking               │     │
│  │  Puertos: gRPC + REST                                │     │
│  └──────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────┘
                             ↓
┌────────────────────────────────────────────────────────────────┐
│                      EVENT BUS (NATS Jetstream)                │
│  Topics: buffer.calculated, order.created, alert.generated     │
└────────────────────────────────────────────────────────────────┘
                             ↓
┌────────────────────────────────────────────────────────────────┐
│                       DATA LAYER                               │
│                                                                │
│  ┌──────────────────┐      ┌──────────────────┐              │
│  │  PostgreSQL 16   │      │  Redis 7         │              │
│  │  • Buffers       │      │  • Cache         │              │
│  │  • Products      │      │  • Sessions      │              │
│  │  │  • Orders        │      │  • AI context    │              │
│  │  • TimeSeries    │      │  • Rate limits   │              │
│  │    (con parti-   │      └──────────────────┘              │
│  │     tioning)     │                                         │
│  └──────────────────┘                                         │
└────────────────────────────────────────────────────────────────┘
                             ↓
┌────────────────────────────────────────────────────────────────┐
│                   OBSERVABILITY STACK                          │
│  • Prometheus (metrics)  • Grafana (dashboards)                │
│  • Loki (logs)  • Jaeger (traces - opcional inicialmente)     │
└────────────────────────────────────────────────────────────────┘
```

### 2.2 Decisiones de Diseño Clave

| Decisión | Justificación |
|----------|---------------|
| **6 microservicios** | Suficiente separación de responsabilidades sin fragmentación excesiva. Auth/IAM dedicado para security isolation |
| **gRPC interno** | Alta performance, type-safe, code generation |
| **REST externo** | Estándar web, fácil para frontend |
| **NATS (no Kafka)** | Más simple, suficiente para volumen esperado, built-in persistence (Jetstream) |
| **PostgreSQL único** | 1 DB simplifica operaciones, partitioning maneja escala |
| **No vector DB** | PostgreSQL con pgvector extension es suficiente (si se necesita) |
| **Kubernetes** | Estándar de industria, fácil contratar, cloud-agnostic |
| **Go para todos** | Stack unificado, fácil contratar, excelente performance |
| **Auth/IAM dedicado** | Basado en users-service existente, solo requiere agregar multi-tenancy + RBAC + gRPC + Events |

---

## 3. Microservicios Core

### 3.1 Service 1: DDMRP Engine Service

**Responsabilidad**: Core DDMRP logic - el corazón del sistema

**Bounded Context**: Buffer Management + Demand Planning

**Capabilities**:
- Calculate buffer zones (Red/Yellow/Green) usando buffer profiles
- Calculate CPD (Consumption Per Day) con historical data
- Compute Net Flow Equation (NFE) diario
- Apply Planned Adjustment Factors (FAP)
- Generate replenishment suggestions basado en buffer penetration
- Decoupled BOM explosion (stops at buffers)

**APIs**:
```go
// gRPC service
service DDMRPEngineService {
  // Buffer operations
  rpc CalculateBuffer(CalculateBufferRequest) returns (BufferResponse);
  rpc GetBufferStatus(GetBufferStatusRequest) returns (BufferStatusResponse);
  rpc BatchCalculateBuffers(BatchRequest) returns (stream BufferResponse);

  // CPD operations
  rpc CalculateCPD(CPDRequest) returns (CPDResponse);

  // Replenishment
  rpc GenerateReplenishmentSuggestions(SuggestionRequest) returns (stream Suggestion);

  // FAP
  rpc ApplyAdjustmentFactor(FAPRequest) returns (BufferResponse);
}

// REST API (público)
GET    /api/v1/buffers?zone=RED&product_id=X
GET    /api/v1/buffers/{id}
POST   /api/v1/buffers/{id}/recalculate
GET    /api/v1/replenishment-suggestions
POST   /api/v1/replenishment-suggestions/{id}/approve
```

**Database Tables**:
```sql
- buffers
- buffer_zones_history (partitioned by tenant)
- cpd_profiles
- replenishment_suggestions
- adjustment_factors
```

**Events Published**:
- `ddmrp.buffer.calculated`
- `ddmrp.buffer.penetrated` (Red/Yellow zone)
- `ddmrp.replenishment.suggested`

**Team Size**: 2 ingenieros (es el servicio más complejo)

---

### 3.2 Service 2: Catalog Service

**Responsabilidad**: Master data management

**Bounded Context**: Product Catalog + Configuration

**Capabilities**:
- CRUD Products (SKU, lead time, MOQ, cost, etc.)
- CRUD Suppliers
- CRUD Nodes/Warehouses
- CRUD Buffer Profiles (Lead Time × Variability matrix)
- BOM (Bill of Materials) management
- Multi-tenant management (organizations, users, roles)

**APIs**:
```go
service CatalogService {
  rpc CreateProduct(ProductRequest) returns (ProductResponse);
  rpc GetProduct(GetProductRequest) returns (ProductResponse);
  rpc UpdateProduct(UpdateProductRequest) returns (ProductResponse);
  rpc ListProducts(ListRequest) returns (stream ProductResponse);

  rpc CreateBufferProfile(ProfileRequest) returns (ProfileResponse);
  // ... similar para Suppliers, Nodes
}

// REST API
GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/{id}
PUT    /api/v1/products/{id}
DELETE /api/v1/products/{id}

GET    /api/v1/suppliers
GET    /api/v1/buffer-profiles
POST   /api/v1/buffer-profiles
```

**Database Tables**:
```sql
- products
- suppliers
- nodes
- buffer_profiles
- bill_of_materials
```

**Note**: Tenants y Users son gestionados por Auth/IAM Service

**Events Published**:
- `catalog.product.created`
- `catalog.product.updated`
- `catalog.profile.created`

**Team Size**: 1 ingeniero

---

### 3.3 Service 3: Execution Service

**Responsabilidad**: Supply chain execution + integrations

**Bounded Context**: Order Management + Integrations

**Capabilities**:
- Create/manage Purchase Orders (PO)
- Create/manage Production Orders (WO)
- Track on-order inventory
- Record inventory transactions (receipts, shipments)
- Synchronization alerts (late supply, early start)
- ERP integrations (connectors for SAP, Odoo, custom REST APIs)
- Webhook receivers

**APIs**:
```go
service ExecutionService {
  rpc CreatePurchaseOrder(PORequest) returns (POResponse);
  rpc UpdateOrderStatus(UpdateOrderRequest) returns (OrderResponse);
  rpc RecordInventoryTransaction(TransactionRequest) returns (TransactionResponse);
  rpc GetOnOrderInventory(OnOrderRequest) returns (OnOrderResponse);
  rpc GetSynchronizationAlerts(AlertQuery) returns (stream Alert);

  // ERP Integration
  rpc SyncFromERP(SyncRequest) returns (SyncResponse);
  rpc RegisterWebhook(WebhookRequest) returns (WebhookResponse);
}

// REST API
GET    /api/v1/orders?status=OPEN
POST   /api/v1/orders
GET    /api/v1/orders/{id}
PUT    /api/v1/orders/{id}/status

POST   /api/v1/inventory/transactions

GET    /api/v1/alerts/synchronization

POST   /api/v1/integrations/erp/sync
POST   /api/v1/webhooks
```

**Database Tables**:
```sql
- purchase_orders
- production_orders
- inventory_transactions
- synchronization_alerts
- erp_connections
- integration_jobs
```

**Events Published**:
- `execution.order.created`
- `execution.order.received`
- `execution.inventory.updated`
- `execution.alert.synchronization`

**Events Subscribed**:
- `ddmrp.replenishment.suggested` (auto-create orders if approved)

**Team Size**: 1-2 ingenieros

---

### 3.4 Service 4: Analytics Service

**Responsabilidad**: Reporting, KPIs, projections

**Bounded Context**: Business Intelligence

**Capabilities**:
- KPI dashboard (service level, inventory turns, ROCE)
- Variance analysis (signal integrity, model velocity)
- DDS&OP projections (space requirements, load capacity)
- Stockout projection alerts
- Supplier performance scoring
- Custom reports

**APIs**:
```go
service AnalyticsService {
  rpc GetKPIDashboard(DashboardRequest) returns (DashboardResponse);
  rpc GetVarianceReport(ReportRequest) returns (VarianceReport);
  rpc ProjectStockouts(ProjectionQuery) returns (stream StockoutAlert);
  rpc GetSupplierPerformance(SupplierID) returns (PerformanceReport);
  rpc GenerateReport(ReportRequest) returns (ReportResponse);
}

// REST API + WebSocket
GET    /api/v1/analytics/kpis
GET    /api/v1/analytics/variance
GET    /api/v1/analytics/projections/stockouts
GET    /api/v1/analytics/suppliers/{id}/performance

WS     /api/v1/analytics/stream (real-time KPI updates)
```

**Database Tables**:
```sql
- kpi_snapshots (time-series, partitioned)
- variance_reports
- stockout_projections
- supplier_performance_cache
```

**Events Subscribed**:
- Todos los eventos (para analytics)

**Team Size**: 1 ingeniero

---

### 3.5 Service 5: AI Agent Service

**Responsabilidad**: Intelligent assistant

**Bounded Context**: AI/ML Capabilities

**Capabilities**:
- Chat handler (WebSocket for real-time conversation)
- OpenAI integration (gpt-4o-mini for cost efficiency)
- Function calling (invoke other services as "tools")
- Proactive analysis (scheduled cron jobs)
- Context enrichment (simple RAG with PostgreSQL)
- User decision tracking (accept/reject suggestions)

**APIs**:
```go
service AIAgentService {
  rpc Chat(stream ChatMessage) returns (stream ChatResponse);
  rpc ExecuteProactiveAnalysis(AnalysisRequest) returns (AnalysisResponse);
  rpc GetConversationHistory(ConversationID) returns (HistoryResponse);
  rpc TrackDecision(DecisionRequest) returns (DecisionResponse);
}

// WebSocket API
WS     /api/v1/ai/chat
```

**Tools Available** (AI puede invocar via gRPC):
```go
// AI Agent llama a otros servicios
- ddmrp_engine.GetBufferStatus(product_id)
- ddmrp_engine.GetReplenishmentSuggestions()
- catalog.GetProduct(product_id)
- execution.GetOnOrderInventory(product_id)
- analytics.GetSupplierPerformance(supplier_id)
- analytics.ProjectStockouts(product_id)
```

**Database Tables**:
```sql
- ai_conversations
- ai_tool_invocations (audit log)
- ai_decisions (user accept/reject)
- ai_context_cache (RAG embeddings - optional)
```

**Events Published**:
- `ai.suggestion.generated`
- `ai.analysis.completed`

**Events Subscribed**:
- `ddmrp.buffer.penetrated` (proactive trigger)

**Team Size**: 1 ingeniero (con experiencia en LLMs)

---

### 3.6 Service 6: Auth/IAM Service

**Responsabilidad**: Authentication, Authorization, Multi-tenancy, User Management

**Bounded Context**: Identity & Access Management

**Capabilities**:
- User CRUD (registration, profile management, password reset)
- Authentication (JWT tokens: access + refresh)
- Multi-tenancy (tenant/organization management)
- RBAC (roles, permissions, authorization checks)
- 2FA/TOTP (two-factor authentication)
- Email verification, password reset flows
- Session management, account locking
- Audit logging (login attempts, user actions)

**APIs**:
```go
// gRPC service (for other services)
service AuthService {
  // Authentication
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (TokenResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);

  // User Management
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteResponse);
  rpc ListUsers(ListUsersRequest) returns (stream UserResponse);

  // Tenant Management
  rpc CreateTenant(CreateTenantRequest) returns (TenantResponse);
  rpc GetTenant(GetTenantRequest) returns (TenantResponse);
  rpc UpdateTenant(UpdateTenantRequest) returns (TenantResponse);

  // Authorization
  rpc CheckPermission(CheckPermissionRequest) returns (PermissionResponse);
  rpc AssignRole(AssignRoleRequest) returns (RoleResponse);
  rpc RevokeRole(RevokeRoleRequest) returns (RoleResponse);
  rpc GetUserRoles(GetUserRolesRequest) returns (RolesResponse);
  rpc GetUserPermissions(GetUserRequest) returns (PermissionsResponse);
}

// REST API (público)
// Authentication
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
POST   /api/v1/auth/verify-email
POST   /api/v1/auth/enable-2fa
POST   /api/v1/auth/verify-2fa

// User Management
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}
GET    /api/v1/users/me
PUT    /api/v1/users/me

// Tenant Management (admin only)
GET    /api/v1/tenants
POST   /api/v1/tenants
GET    /api/v1/tenants/{id}
PUT    /api/v1/tenants/{id}

// Roles & Permissions (admin only)
GET    /api/v1/roles
POST   /api/v1/roles
GET    /api/v1/roles/{id}
PUT    /api/v1/roles/{id}
POST   /api/v1/users/{id}/roles
DELETE /api/v1/users/{id}/roles/{roleId}
```

**Database Tables**:
```sql
- tenants
- users (with tenant_id)
- user_preferences
- user_notification_settings
- user_two_fa
- user_tokens (refresh tokens)
- login_attempts
- roles
- user_roles (many-to-many)
- audit_log
```

**Predefined Roles**:
- **Admin**: Full system access (`*:*`)
- **Planner**: Manage buffers, CPD, replenishment suggestions
  - `buffer:read`, `buffer:write`, `buffer:calculate`
  - `cpd:read`, `cpd:calculate`
  - `replenishment:read`, `replenishment:approve`
  - `product:read`, `order:read`
- **Viewer**: Read-only access to dashboards
  - `buffer:read`, `product:read`, `order:read`, `analytics:read`

**Events Published**:
- `auth.user.created`
- `auth.user.updated`
- `auth.user.logged_in`
- `auth.user.logged_out`
- `auth.tenant.created`
- `auth.tenant.updated`
- `auth.role.assigned`

**Security Features**:
- bcrypt password hashing (cost factor 12)
- JWT tokens (HS256, 15 min access, 7 day refresh)
- Rate limiting per tenant
- Account locking after failed attempts
- IP-based suspicious activity detection
- Password complexity validation
- Email verification required
- Audit trail for all auth operations

**Integration Notes**:
- Based on existing `users-service` implementation
- Well-architected with Clean Architecture + Repository pattern
- Already has JWT, 2FA, password management
- **Additions needed**: Multi-tenancy + RBAC + gRPC + Events
- **Implementation time**: 2-3 weeks

**Team Size**: 1 ingeniero (familiar with existing users-service codebase)

---

## 4. Clean Architecture por Servicio

### 4.1 Estructura de Directorios (Ejemplo: DDMRP Engine)

```
ddmrp-engine-service/
├── cmd/
│   ├── server/
│   │   └── main.go                # Entry point
│   └── worker/
│       └── main.go                # Background jobs (daily calc)
│
├── internal/
│   ├── domain/                    # LAYER 1: Domain (core business logic)
│   │   ├── buffer/
│   │   │   ├── buffer.go          # Aggregate root
│   │   │   ├── zones.go           # Value objects
│   │   │   ├── repository.go      # Port (interface)
│   │   │   ├── service.go         # Domain service (calculation logic)
│   │   │   └── events.go          # Domain events
│   │   ├── cpd/
│   │   │   ├── cpd.go
│   │   │   └── calculator.go
│   │   └── shared/
│   │       ├── types.go
│   │       └── errors.go
│   │
│   ├── application/               # LAYER 2: Use Cases
│   │   ├── calculate_buffer/
│   │   │   ├── usecase.go
│   │   │   ├── dto.go
│   │   │   └── usecase_test.go
│   │   ├── calculate_cpd/
│   │   ├── generate_suggestions/
│   │   └── apply_fap/
│   │
│   ├── adapter/                   # LAYER 3: Adapters
│   │   ├── grpc/
│   │   │   ├── server.go          # gRPC server
│   │   │   ├── handler.go         # RPC implementations
│   │   │   └── mapper.go          # Proto <-> Domain
│   │   ├── http/
│   │   │   └── handler.go         # REST API (public)
│   │   ├── repository/
│   │   │   └── postgres/
│   │   │       ├── buffer_repo.go
│   │   │       └── cpd_repo.go
│   │   └── events/
│   │       └── nats/
│   │           ├── publisher.go
│   │           └── subscriber.go
│   │
│   └── infrastructure/            # LAYER 4: Infrastructure
│       ├── config/
│       │   └── config.go
│       ├── database/
│       │   ├── postgres.go
│       │   └── migrations/
│       ├── nats/
│       │   └── client.go
│       └── logger/
│           └── logger.go
│
├── api/
│   └── proto/
│       └── ddmrp/v1/
│           └── ddmrp.proto        # gRPC service definition
│
├── migrations/
│   ├── 001_create_buffers.sql
│   └── 002_create_cpd_profiles.sql
│
├── scripts/
│   └── seed_data.sql
│
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 4.2 Dependency Injection (Wire)

```go
// internal/infrastructure/wire.go
//go:build wireinject
// +build wireinject

package infrastructure

import (
    "github.com/google/wire"
    "ddmrp-engine/internal/adapter/grpc"
    "ddmrp-engine/internal/adapter/repository/postgres"
    "ddmrp-engine/internal/application/calculate_buffer"
    "ddmrp-engine/internal/domain/buffer"
)

func InitializeApplication(cfg *Config) (*Application, error) {
    wire.Build(
        // Infrastructure
        NewPostgresDB,
        NewNATSClient,
        NewLogger,

        // Repositories
        postgres.NewBufferRepository,
        postgres.NewCPDRepository,
        wire.Bind(new(buffer.Repository), new(*postgres.BufferRepository)),

        // Use Cases
        calculate_buffer.NewUseCase,

        // Adapters
        grpc.NewServer,

        // Application
        NewApplication,
    )
    return nil, nil
}
```

### 4.3 Domain Example (Buffer Entity)

```go
// internal/domain/buffer/buffer.go
package buffer

import "time"

// Buffer is the aggregate root
type Buffer struct {
    ID              string
    TenantID        string
    ProductID       string
    NodeID          *string

    Zones           Zones
    NetFlow         NetFlowPosition
    Profile         *BufferProfile

    LastCalculatedAt time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}

type Zones struct {
    TopOfGreen  float64
    TopOfYellow float64
    TopOfRed    float64
    RedBase     float64
    RedSafety   float64
}

type NetFlowPosition struct {
    PhysicalInventory float64
    OnOrderInventory  float64
    QualifiedDemand   float64
}

// CalculateNetFlow - Business logic in domain
func (b *Buffer) CalculateNetFlow() float64 {
    return b.NetFlow.PhysicalInventory +
           b.NetFlow.OnOrderInventory -
           b.NetFlow.QualifiedDemand
}

// GetZone - Determine current zone
func (b *Buffer) GetZone() ZoneColor {
    nfe := b.CalculateNetFlow()

    if nfe <= b.Zones.TopOfRed {
        return ZoneRed
    }
    if nfe <= b.Zones.TopOfYellow {
        return ZoneYellow
    }
    return ZoneGreen
}

// ApplyFAP - Apply Planned Adjustment Factor
func (b *Buffer) ApplyFAP(multiplier float64) {
    b.Zones.TopOfGreen *= multiplier
    b.Zones.TopOfYellow *= multiplier
    b.Zones.TopOfRed *= multiplier
    b.Zones.RedBase *= multiplier
    b.Zones.RedSafety *= multiplier
}

// Repository interface (port)
type Repository interface {
    Save(ctx context.Context, buffer *Buffer) error
    FindByID(ctx context.Context, id string) (*Buffer, error)
    FindByProduct(ctx context.Context, productID string) (*Buffer, error)
    FindRedZoneBuffers(ctx context.Context, tenantID string) ([]*Buffer, error)
}
```

---

## 5. AI Agent Balanceado

### 5.1 Arquitectura del AI Agent

```
┌──────────────────────────────────────────────────────────┐
│  AI AGENT SERVICE                                        │
│                                                          │
│  ┌────────────────────────────────────────────────┐     │
│  │  CHAT HANDLER (WebSocket)                      │     │
│  │  • Manage connections                          │     │
│  │  • Session state (Redis)                       │     │
│  └────────────────────────────────────────────────┘     │
│                      ↓                                   │
│  ┌────────────────────────────────────────────────┐     │
│  │  CONVERSATION MANAGER                          │     │
│  │  • Multi-turn context                          │     │
│  │  • Conversation history (PostgreSQL)           │     │
│  │  • User intent classification                  │     │
│  └────────────────────────────────────────────────┘     │
│                      ↓                                   │
│  ┌────────────────────────────────────────────────┐     │
│  │  OPENAI CLIENT                                 │     │
│  │  • Model: gpt-4o-mini (cost-effective)         │     │
│  │  • Function calling enabled                    │     │
│  │  • Streaming responses                         │     │
│  │  • Rate limiting (per tenant)                  │     │
│  └────────────────────────────────────────────────┘     │
│                      ↓                                   │
│  ┌────────────────────────────────────────────────┐     │
│  │  TOOL EXECUTOR                                 │     │
│  │  • Invoke other services via gRPC              │     │
│  │  • get_buffer_status()                         │     │
│  │  • get_supplier_performance()                  │     │
│  │  • project_stockout()                          │     │
│  │  • explain_cpd_calculation()                   │     │
│  └────────────────────────────────────────────────┘     │
│                      ↓                                   │
│  ┌────────────────────────────────────────────────┐     │
│  │  CONTEXT ENRICHER (Simple RAG)                 │     │
│  │  • Recent buffer changes (Redis cache)         │     │
│  │  • User preferences                            │     │
│  │  • Similar past queries (PostgreSQL)           │     │
│  └────────────────────────────────────────────────┘     │
│                      ↓                                   │
│  ┌────────────────────────────────────────────────┐     │
│  │  PROACTIVE TRIGGER ENGINE                      │     │
│  │  • Cron job: Daily buffer analysis (6am)       │     │
│  │  • Event-driven: buffer.penetrated             │     │
│  │  • Generate insights & notifications           │     │
│  └────────────────────────────────────────────────┘     │
└──────────────────────────────────────────────────────────┘
```

### 5.2 AI Implementation (Simplified)

```go
// internal/application/chat/usecase.go
package chat

type ChatUseCase struct {
    openaiClient   *openai.Client
    ddmrpClient    ddmrppb.DDMRPEngineServiceClient  // gRPC
    catalogClient  catalogpb.CatalogServiceClient
    analyticsClient analyticspb.AnalyticsServiceClient
    conversationRepo ConversationRepository
}

func (uc *ChatUseCase) HandleMessage(ctx context.Context, userID, message string) (*ChatResponse, error) {
    // 1. Get conversation history
    history := uc.getHistory(ctx, userID)

    // 2. Add user message
    history = append(history, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: message,
    })

    // 3. Define tools
    tools := []openai.Tool{
        {
            Type: openai.ToolTypeFunction,
            Function: &openai.FunctionDefinition{
                Name:        "get_buffer_status",
                Description: "Get current status of a DDMRP buffer for a product",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "product_sku": {"type": "string"},
                    },
                    "required": []string{"product_sku"},
                },
            },
        },
        {
            Type: openai.ToolTypeFunction,
            Function: &openai.FunctionDefinition{
                Name:        "project_stockout",
                Description: "Project when a product will stock out based on current demand",
                Parameters: map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "product_sku": {"type": "string"},
                    },
                    "required": []string{"product_sku"},
                },
            },
        },
        // ... más tools
    }

    // 4. Call OpenAI
    resp, err := uc.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:    openai.GPT4oMini,
        Messages: history,
        Tools:    tools,
    })
    if err != nil {
        return nil, err
    }

    // 5. Execute tool calls if any
    if len(resp.Choices[0].Message.ToolCalls) > 0 {
        for _, toolCall := range resp.Choices[0].Message.ToolCalls {
            result := uc.executeTool(ctx, toolCall)

            history = append(history, openai.ChatCompletionMessage{
                Role:       openai.ChatMessageRoleTool,
                Content:    result,
                Name:       toolCall.Function.Name,
                ToolCallID: toolCall.ID,
            })
        }

        // 6. Call OpenAI again with tool results
        resp, err = uc.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
            Model:    openai.GPT4oMini,
            Messages: history,
            Tools:    tools,
        })
    }

    // 7. Save conversation
    uc.conversationRepo.Save(ctx, userID, history)

    return &ChatResponse{
        Content: resp.Choices[0].Message.Content,
    }, nil
}

func (uc *ChatUseCase) executeTool(ctx context.Context, toolCall openai.ToolCall) string {
    switch toolCall.Function.Name {
    case "get_buffer_status":
        var args struct {
            ProductSKU string `json:"product_sku"`
        }
        json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

        // Call DDMRP Engine Service via gRPC
        resp, err := uc.ddmrpClient.GetBufferStatus(ctx, &ddmrppb.GetBufferStatusRequest{
            ProductSku: args.ProductSKU,
        })
        if err != nil {
            return fmt.Sprintf("Error: %v", err)
        }

        return fmt.Sprintf(`Product %s is in %s zone.
Net Flow: %.2f, Top of Red: %.2f, Top of Yellow: %.2f, Top of Green: %.2f`,
            args.ProductSKU, resp.CurrentZone, resp.NetFlow,
            resp.Zones.TopOfRed, resp.Zones.TopOfYellow, resp.Zones.TopOfGreen)

    case "project_stockout":
        // ... similar
    }
    return "Unknown tool"
}
```

### 5.3 Proactive Analysis (Cron Job)

```go
// internal/application/proactive/daily_analysis.go
package proactive

type DailyAnalysisUseCase struct {
    ddmrpClient    ddmrppb.DDMRPEngineServiceClient
    openaiClient   *openai.Client
    notificationSvc NotificationService
}

func (uc *DailyAnalysisUseCase) Execute(ctx context.Context) error {
    // 1. Get all red zone buffers
    redBuffers, err := uc.ddmrpClient.GetRedZoneBuffers(ctx, &ddmrppb.GetRedZoneBuffersRequest{})
    if err != nil {
        return err
    }

    if len(redBuffers.Buffers) == 0 {
        return nil // All good
    }

    // 2. Ask AI to analyze
    prompt := fmt.Sprintf(`You are a DDMRP expert. Analyze these %d red zone buffers:

%s

Provide:
1. Top 3 most urgent products
2. Common patterns (supplier issues? demand spikes?)
3. Recommended actions

Be concise.`, len(redBuffers.Buffers), formatBuffers(redBuffers.Buffers))

    resp, err := uc.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4oMini,
        Messages: []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleUser, Content: prompt},
        },
    })
    if err != nil {
        return err
    }

    // 3. Send notification to all planners
    uc.notificationSvc.NotifyAllPlanners(ctx, Notification{
        Title:   "Daily Buffer Health Report",
        Content: resp.Choices[0].Message.Content,
    })

    return nil
}
```

---

## 6. Modelo de Datos

### 6.1 Base de Datos: PostgreSQL Único

**Por qué PostgreSQL único**:
- ✅ Simplifica operaciones (1 DB, no 5)
- ✅ Transacciones ACID nativas
- ✅ Partitioning para escala
- ✅ JSONB para flexibilidad
- ✅ TimescaleDB extension para time-series (opcional)
- ✅ pgvector para embeddings (si necesitas RAG avanzado después)

### 6.2 Schema Principal

```sql
-- =====================================================
-- AUTH/IAM SERVICE TABLES
-- =====================================================

-- Multi-tenancy: todas las tablas con tenant_id
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    subscription_plan VARCHAR(50) DEFAULT 'FREE',
    subscription_status VARCHAR(20) DEFAULT 'ACTIVE',
    max_users INT DEFAULT 5,
    settings JSONB,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(subscription_status);

-- Users table (Auth/IAM Service)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    avatar TEXT,
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    email_verification_token VARCHAR(255),
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMPTZ,
    totp_secret VARCHAR(255),
    totp_enabled BOOLEAN DEFAULT false,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_tenant_email ON users(tenant_id, email);

-- Roles (RBAC)
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    permissions JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);
CREATE INDEX idx_roles_system ON roles(is_system);

-- User Roles (many-to-many)
CREATE TABLE user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT now(),
    assigned_by INTEGER REFERENCES users(id),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID REFERENCES tenants(id),
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(255),
    ip_address INET,
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_audit_log_tenant ON audit_log(tenant_id, created_at DESC);
CREATE INDEX idx_audit_log_user ON audit_log(user_id, created_at DESC);
CREATE INDEX idx_audit_log_action ON audit_log(action);

-- User Preferences
CREATE TABLE user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    theme VARCHAR(20) DEFAULT 'light',
    preferences JSONB,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id)
);

-- User Tokens (refresh tokens)
CREATE TABLE user_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    token_type VARCHAR(20) DEFAULT 'refresh',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_user_tokens_user ON user_tokens(user_id);
CREATE INDEX idx_user_tokens_token ON user_tokens(token);

-- Login Attempts
CREATE TABLE login_attempts (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    ip_address INET,
    success BOOLEAN DEFAULT false,
    failed_reason VARCHAR(100),
    attempted_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_login_attempts_email ON login_attempts(email, attempted_at DESC);
CREATE INDEX idx_login_attempts_ip ON login_attempts(ip_address, attempted_at DESC);

-- =====================================================
-- CATALOG SERVICE TABLES
-- =====================================================
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255),
    is_buffered BOOLEAN DEFAULT true,
    lead_time_days INT,
    moq DECIMAL,
    unit_cost DECIMAL,
    supplier_id UUID,
    buffer_profile_id UUID,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, sku)
);

CREATE INDEX idx_products_tenant_sku ON products(tenant_id, sku);
CREATE INDEX idx_products_buffered ON products(tenant_id, is_buffered) WHERE is_buffered = true;

CREATE TABLE buffer_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(100),
    lead_time_category VARCHAR(20),      -- SHORT, MEDIUM, LONG
    variability_category VARCHAR(20),    -- LOW, MEDIUM, HIGH
    red_base_percentage DECIMAL,
    red_safety_percentage DECIMAL,
    yellow_percentage DECIMAL,
    green_percentage DECIMAL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- DDMRP ENGINE SERVICE
CREATE TABLE buffers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    product_id UUID NOT NULL REFERENCES products(id),
    node_id UUID,

    -- Zones
    top_of_green DECIMAL NOT NULL,
    top_of_yellow DECIMAL NOT NULL,
    top_of_red DECIMAL NOT NULL,
    red_base DECIMAL NOT NULL,
    red_safety DECIMAL NOT NULL,

    -- Current state
    physical_inventory DECIMAL DEFAULT 0,
    on_order_inventory DECIMAL DEFAULT 0,
    qualified_demand DECIMAL DEFAULT 0,

    -- Computed column (virtual)
    net_flow DECIMAL GENERATED ALWAYS AS (
        physical_inventory + on_order_inventory - qualified_demand
    ) STORED,

    profile_id UUID REFERENCES buffer_profiles(id),
    last_calculated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE(tenant_id, product_id, COALESCE(node_id, '00000000-0000-0000-0000-000000000000'))
);

CREATE INDEX idx_buffers_tenant ON buffers(tenant_id);
CREATE INDEX idx_buffers_product ON buffers(product_id);
CREATE INDEX idx_buffers_zone ON buffers(tenant_id, (
    CASE
        WHEN net_flow <= top_of_red THEN 'RED'
        WHEN net_flow <= top_of_yellow THEN 'YELLOW'
        ELSE 'GREEN'
    END
)) WHERE net_flow <= top_of_yellow;

-- Time-series: Buffer history (partitioned by tenant for scale)
CREATE TABLE buffer_history (
    id UUID DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    buffer_id UUID NOT NULL,
    net_flow DECIMAL,
    zone VARCHAR(10),
    snapshot_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (tenant_id, snapshot_at, id)
) PARTITION BY LIST (tenant_id);

-- CPD profiles
CREATE TABLE cpd_profiles (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    product_id UUID NOT NULL,
    node_id UUID,
    cpd DECIMAL NOT NULL,
    calculation_window INT,               -- Days
    historical_avg DECIMAL,
    forecast_avg DECIMAL,
    blend_ratio DECIMAL,                  -- 0.0-1.0
    calculated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, product_id, COALESCE(node_id, '00000000-0000-0000-0000-000000000000'))
);

-- Replenishment suggestions
CREATE TABLE replenishment_suggestions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    buffer_id UUID NOT NULL,
    product_id UUID NOT NULL,
    suggested_quantity DECIMAL,
    priority INT,                         -- 1 (urgent) - 5 (low)
    status VARCHAR(20),                   -- PENDING, APPROVED, REJECTED
    created_at TIMESTAMPTZ DEFAULT now(),
    approved_at TIMESTAMPTZ,
    approved_by UUID
);

-- EXECUTION SERVICE
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    product_id UUID NOT NULL,
    supplier_id UUID,
    quantity DECIMAL,
    order_date TIMESTAMPTZ,
    promised_date TIMESTAMPTZ,
    actual_receipt_date TIMESTAMPTZ,
    status VARCHAR(20),                   -- OPEN, IN_TRANSIT, RECEIVED, DELAYED
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_orders_status ON purchase_orders(tenant_id, status)
WHERE status IN ('OPEN', 'IN_TRANSIT');

CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    product_id UUID NOT NULL,
    node_id UUID,
    transaction_type VARCHAR(20),         -- RECEIPT, SHIPMENT, ADJUSTMENT
    quantity DECIMAL,
    reference_order_id UUID,
    timestamp TIMESTAMPTZ DEFAULT now()
);

-- AI SERVICE
CREATE TABLE ai_conversations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    messages JSONB NOT NULL,              -- Array de {role, content, timestamp}
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_ai_conversations_user ON ai_conversations(user_id, updated_at DESC);

CREATE TABLE ai_decisions (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    suggestion_id UUID,
    decision VARCHAR(20),                 -- ACCEPTED, REJECTED, MODIFIED
    timestamp TIMESTAMPTZ DEFAULT now()
);
```

### 6.3 Partitioning Strategy

**Para escalar a millones de registros**:

```sql
-- Ejemplo: Crear particiones por tenant para buffer_history
CREATE TABLE buffer_history_tenant_abc PARTITION OF buffer_history
    FOR VALUES IN ('abc-tenant-uuid');

-- Automatizar creación de particiones (trigger o script)
```

---

## 7. Comunicación entre Servicios

### 7.1 Protocols

**gRPC (interno)**:
- Entre servicios (service-to-service)
- Alta performance, type-safe
- Streaming support

**REST (externo)**:
- Frontend → Backend
- Webhooks
- ERP integrations

**WebSocket**:
- AI Chat
- Real-time dashboard updates

**NATS (eventos)**:
- Async communication
- Event-driven patterns

### 7.2 Service Communication Matrix

| From ↓ To → | Auth/IAM | DDMRP Engine | Catalog | Execution | Analytics | AI Agent |
|-------------|----------|--------------|---------|-----------|-----------|----------|
| **Auth/IAM** | - | - | - | - | - | - |
| **DDMRP Engine** | gRPC | - | gRPC | Events | Events | - |
| **Catalog** | gRPC | - | - | - | - | - |
| **Execution** | gRPC | gRPC | gRPC | - | Events | - |
| **Analytics** | gRPC | gRPC | gRPC | gRPC | - | - |
| **AI Agent** | gRPC | gRPC | gRPC | gRPC | gRPC | - |
| **API Gateway** | gRPC | REST | REST | REST | REST | WebSocket |

**Principles**:
- **Auth/IAM** only provides services (never calls other services)
- **All services** call Auth/IAM via gRPC for token validation and permission checks
- **API Gateway** validates tokens with Auth/IAM before routing requests
- **AI Agent** only consumes (read-only via gRPC)
- **Analytics** only consumes (subscribes to all events)
- **Core services** communicate via events for decoupling

### 7.3 Event Schema (NATS)

```go
// Event envelope
type DomainEvent struct {
    EventID     string                 `json:"event_id"`
    EventType   string                 `json:"event_type"`
    TenantID    string                 `json:"tenant_id"`
    AggregateID string                 `json:"aggregate_id"`
    Payload     map[string]interface{} `json:"payload"`
    OccurredAt  time.Time              `json:"occurred_at"`
}

// Topics
const (
    TopicBufferCalculated       = "ddmrp.buffer.calculated"
    TopicBufferPenetrated       = "ddmrp.buffer.penetrated"
    TopicReplenishmentSuggested = "ddmrp.replenishment.suggested"
    TopicOrderCreated           = "execution.order.created"
    TopicInventoryUpdated       = "execution.inventory.updated"
    TopicAISuggestionGenerated  = "ai.suggestion.generated"
)
```

---

## 8. Infraestructura

### 8.1 Deployment (Kubernetes)

**Cluster Setup**:
- **Cloud Provider**: GCP (GKE), AWS (EKS), or Azure (AKS)
- **Nodes**: 3-6 nodes inicialmente (e2-standard-4 en GCP)
- **Namespaces**: `giia-prod`, `giia-staging`, `giia-dev`

**Services Deployed**:
```yaml
# Namespace: giia-prod
deployments/
  ├── api-gateway           (Kong/Traefik, 2 replicas)
  ├── auth-service          (2 replicas, critical for all requests)
  ├── ddmrp-engine-service  (3 replicas, HPA enabled)
  ├── catalog-service       (2 replicas)
  ├── execution-service     (2 replicas)
  ├── analytics-service     (2 replicas)
  └── ai-agent-service      (2 replicas)

statefulsets/
  ├── postgres              (1 master, 2 read replicas)
  ├── redis                 (Redis cluster con sentinel)
  └── nats                  (NATS Jetstream cluster)
```

**Example Deployment** (ddmrp-engine):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ddmrp-engine
  namespace: giia-prod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ddmrp-engine
  template:
    metadata:
      labels:
        app: ddmrp-engine
    spec:
      containers:
      - name: ddmrp-engine
        image: ghcr.io/giia/ddmrp-engine:v1.2.3
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: giia-secrets
              key: database-url
        - name: NATS_URL
          value: "nats://nats:4222"
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: ddmrp-engine
spec:
  selector:
    app: ddmrp-engine
  ports:
  - name: grpc
    port: 9090
  type: ClusterIP
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: ddmrp-engine-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: ddmrp-engine
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 8.2 CI/CD (GitHub Actions)

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run tests
        run: go test -v -race ./...

      - name: Build Docker images
        run: |
          docker build -t giia/ddmrp-engine:${{ github.sha }} ./services/ddmrp-engine
          # ... otros servicios

      - name: Push to registry
        run: |
          echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
          docker push giia/ddmrp-engine:${{ github.sha }}

      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/ddmrp-engine ddmrp-engine=giia/ddmrp-engine:${{ github.sha }} -n giia-prod
```

### 8.3 Costos Estimados

| Fase | Clientes | Infra Mensual | Desglose |
|------|----------|---------------|----------|
| **Launch (0-6m)** | 10-50 | $500 | GKE: $300, DB: $100, Redis: $50, OpenAI: $50 |
| **Growth (6-18m)** | 100-500 | $1,200 | GKE: $700, DB: $300, Redis: $100, OpenAI: $100 |
| **Scale (18m+)** | 1000+ | $3,000+ | GKE: $1,500, DB: $800, Redis: $200, OpenAI: $500 |

**Revenue Target** (Growth phase):
- 300 clientes × $100/mes = $30K MRR
- **Gross Margin**: 96% 💰

---

## 9. Seguridad

### 9.1 Authentication

- **JWT tokens** (15 min access, 7 day refresh)
- **OAuth2** support (Google, Microsoft SSO)
- **MFA** (TOTP) para admin users

### 9.2 Authorization

- **RBAC** (Role-Based Access Control)
- Roles: Admin, Planner, Viewer
- **Multi-tenancy isolation** (tenant_id en todas las queries)

### 9.3 OWASP Top 10 Compliance

✅ SQL Injection: Parameterized queries
✅ XSS: Content Security Policy headers
✅ CSRF: SameSite cookies
✅ Secrets: Kubernetes Secrets / Vault
✅ TLS: 1.3 only
✅ Rate Limiting: Redis-based (per tenant)

---

## 10. Observability

### 10.1 Stack

- **Metrics**: Prometheus + Grafana
- **Logs**: Loki (o ELK si presupuesto permite)
- **Traces**: Jaeger (opcional inicialmente, agregar cuando escales)

### 10.2 Key Metrics

```go
// Prometheus metrics
var (
    bufferCalculationsTotal = prometheus.NewCounterVec(...)
    aiRequestDuration = prometheus.NewHistogram(...)
    redBuffersGauge = prometheus.NewGaugeVec(...)
    orderCreationTotal = prometheus.NewCounterVec(...)
)
```

### 10.3 Dashboards

1. **Golden Signals**: Latency, Traffic, Errors, Saturation
2. **DDMRP Metrics**: Red buffers count, replenishment approval rate
3. **AI Metrics**: Request latency, cost tracking, acceptance rate
4. **Business KPIs**: Service level, inventory turns, ROCE

---

## 11. Plan de Implementación

### Fase 1: Fundación (Mes 1-3)

**Objetivo**: Core platform + 3 servicios iniciales

- [ ] Setup monorepo + CI/CD
- [ ] Deploy Kubernetes cluster (dev + staging)
- [ ] PostgreSQL + Redis setup
- [ ] NATS setup
- [ ] **Servicio 1**: Auth/IAM Service (basado en users-service existente)
  - [ ] Migrar users-service al monorepo
  - [ ] Agregar multi-tenancy (tenants table, tenant_id en users)
  - [ ] Agregar RBAC (roles, permissions, user_roles)
  - [ ] Implementar gRPC server (ValidateToken, CheckPermission)
  - [ ] Agregar NATS event publisher
  - [ ] Tests E2E multi-tenant + RBAC
- [ ] **Servicio 2**: Catalog Service (ABM completo)
- [ ] **Servicio 3**: DDMRP Engine Service (buffer calculation básico)
- [ ] Frontend básico (React dashboard con login)

**Team**: 3-4 ingenieros full-time

**Entregable**: Demo funcional con autenticación multi-tenant + productos + buffers básicos

---

### Fase 2: DDMRP Completo (Mes 4-6)

**Objetivo**: Sistema DDMRP funcional end-to-end

- [ ] **Servicio 4**: Execution Service (órdenes, inventory)
- [ ] CPD calculation (en DDMRP Engine)
- [ ] Replenishment suggestions
- [ ] Synchronization alerts
- [ ] Dashboard visual (buffer zones, NFE charts)
- [ ] API documentation (OpenAPI)
- [ ] Integración con Auth/IAM (permisos por operación)

**Team**: 4 ingenieros

**Entregable**: Sistema usable por primeros clientes beta

---

### Fase 3: AI Agent (Mes 7-9)

**Objetivo**: Agregar inteligencia AI

- [ ] **Servicio 5**: AI Agent Service
- [ ] OpenAI integration
- [ ] Function calling (6-8 tools)
- [ ] WebSocket chat
- [ ] Proactive analysis (cron jobs)
- [ ] Notification system
- [ ] Integración con Auth/IAM (validación de permisos antes de tool execution)

**Team**: 5 ingenieros (1 especialista AI)

**Entregable**: AI conversacional + análisis proactivo

---

### Fase 4: Analytics & Scale (Mes 10-12)

**Objetivo**: Preparar para primeros 100 clientes

- [ ] **Servicio 6**: Analytics Service
- [ ] KPI dashboards
- [ ] Variance analysis
- [ ] DDS&OP projections
- [ ] ERP integrations (Odoo, SAP connectors)
- [ ] Performance optimization
- [ ] Load testing (1000 concurrent users)
- [ ] Security audit (Auth/IAM penetration testing)
- [ ] Production deployment

**Team**: 6 ingenieros

**Entregable**: Producto listo para lanzamiento comercial

---

## 12. Comparación con Otras Arquitecturas

| Aspecto | Monolito Pragmático | **Balanced (RECOMENDADA)** | Enterprise Completa |
|---------|---------------------|----------------------------|---------------------|
| **Servicios** | 1 monolito | 6 microservicios (incluye Auth/IAM dedicado) | 10+ microservicios |
| **Bases de Datos** | PostgreSQL + Redis | PostgreSQL + Redis | PostgreSQL + TimescaleDB + Redis + Qdrant + S3 |
| **AI Complexity** | OpenAI directo básico | OpenAI + RAG simple | AI Gateway + múltiples LLMs + ML services |
| **Auth/IAM** | Embebido en monolito | ✅ Servicio dedicado (multi-tenancy + RBAC) | Servicio dedicado + OAuth2 + SSO |
| **Deploy Target** | Railway/Fly.io | Kubernetes (GKE/EKS) | Kubernetes + Service Mesh |
| **Time to Market** | 4-6 meses | 6-9 meses | 12-18 meses |
| **Team Size** | 2-3 ingenieros | 4-6 ingenieros | 8-12 ingenieros |
| **Costo Operacional (inicial)** | $100/mes | $500/mes | $2K+/mes |
| **Escala máxima** | 100-500 clientes | 1K-10K clientes | 100K+ clientes |
| **Complejidad Operacional** | Baja | Media | Alta |
| **Separación de Equipos** | No | ✅ Sí (por servicio) | ✅ Sí (múltiples equipos) |
| **Inversión atractiva** | ⚠️ Cuestionable | ✅ **Sí** | ✅ Sí |
| **Hiring** | Go fullstack | Go engineers (común) | Go + DevOps + AI specialists |

### Recomendación Final

**Usa Balanced si**:
- Buscas inversión de VCs (quieren ver arquitectura seria)
- Planeas equipo de 5+ ingenieros
- Objetivo: 500+ clientes en 18 meses
- Valoras Clean Architecture + DDD

**Usa Monolito Pragmático si**:
- Bootstrap/autofinanciado
- Equipo pequeño (2-3 personas)
- Necesitas validar mercado rápido

**Usa Enterprise Completa si**:
- Ya tienes clientes enterprise grandes
- Equipo de 10+ ingenieros
- Compliance/regulatorio estricto

---

## Conclusión

Esta **Arquitectura Balanceada** representa el **punto óptimo** para GIIA:

✅ **Microservicios** pero no over-engineered
✅ **Clean Architecture** y **DDD** en cada servicio
✅ **AI robusto** pero sin complejidad innecesaria
✅ **Kubernetes** (estándar de industria)
✅ **Escalable** a miles de clientes
✅ **Implementable** en 6-9 meses con equipo competente
✅ **Atractiva para inversión**

**Esta es la arquitectura que recomiendo implementar.**

---

**Siguiente Paso**: Revisar los 3 documentos y elegir cuál implementar basándose en:
1. Recursos disponibles (equipo, presupuesto)
2. Timeline deseado
3. Ambición de escala
4. Estrategia de fundraising
