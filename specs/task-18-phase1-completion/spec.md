# Task 18: Phase 1 Completion - Final Polish

**Task ID**: task-18-phase1-completion
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Duration**: 1 week
**Dependencies**: Tasks 11, 12, 13 (in progress/complete)

---

## Overview

This task consolidates all remaining items to bring Phase 1 from 93% to 100% completion. It focuses on polish, configuration, and production-readiness items across Auth service (RBAC seed data), gRPC (reflection and examples), NATS (additional events and DLQ), and overall system quality.

---

## User Scenarios

### US1: RBAC Seed Data and Performance (P1)

**As a** system administrator
**I want to** predefined roles and permissions in the system
**So that** I can quickly assign users to standard roles without manual setup

**Acceptance Criteria**:
- Seed data script creates 4 standard roles: Admin, Manager, Analyst, Viewer
- Each role has appropriate permissions predefined
- Seed script is idempotent (can run multiple times safely)
- Seed data runs automatically on first deployment
- Performance testing shows <10ms p95 for permission checks
- Audit logging captures all permission check attempts

**Success Metrics**:
- 100% role coverage for common use cases
- <10ms p95 permission check latency
- Complete audit trail for security compliance

---

### US2: gRPC Reflection and Client Examples (P1)

**As a** service developer
**I want to** gRPC reflection and client examples
**So that** I can debug services and integrate with them easily

**Acceptance Criteria**:
- gRPC reflection enabled on all services (auth, catalog)
- Can use grpcurl or similar tools to explore APIs
- Client example code for calling Auth service from Catalog service
- Client example code for token validation
- Client example code for permission checking
- Integration tests for gRPC endpoints
- Proto files for Catalog service gRPC API

**Success Metrics**:
- 100% gRPC service discoverability via reflection
- Working client examples for all major RPCs
- All gRPC integration tests pass

---

### US3: NATS Event System Polish (P1)

**As a** system architect
**I want to** complete event publishing and DLQ handling
**So that** services communicate reliably via events

**Acceptance Criteria**:
- Additional auth events published: user.registered, user.password_changed, user.deactivated
- Additional catalog events: supplier.created, supplier.updated, buffer_profile.created
- Dead letter queue (DLQ) consumer processes failed events
- DLQ monitoring and alerting configured
- Event retry logic with exponential backoff
- Event schema validation
- Event replay capability for debugging

**Success Metrics**:
- 100% domain event coverage for auth and catalog
- <1% events sent to DLQ
- Zero lost events (99.99% delivery guarantee)

---

### US4: Production Readiness Checklist (P1)

**As a** DevOps engineer
**I want to** production-ready configuration
**So that** services can be deployed to staging/production safely

**Acceptance Criteria**:
- Health check endpoints return detailed status
- Readiness probes for Kubernetes
- Graceful shutdown handling (SIGTERM)
- Configuration validation on startup
- Environment-specific configs (dev, staging, prod)
- Secrets management documented
- Resource limits configured (CPU, memory)
- Rate limiting on public endpoints
- Request ID propagation across services

**Success Metrics**:
- Zero downtime deployments in staging
- All services pass Kubernetes health checks
- Configuration documented in .env.example files

---

### US5: Documentation and Runbooks (P2)

**As a** new team member
**I want to** comprehensive documentation
**So that** I can onboard quickly and troubleshoot issues

**Acceptance Criteria**:
- Architecture diagrams updated (sequence, component)
- API documentation complete (OpenAPI/Swagger for REST, proto docs for gRPC)
- Runbooks for common operations: deploy, rollback, scale, debug
- Troubleshooting guide for common issues
- Development setup guide with step-by-step instructions
- Contributing guide with code standards
- Changelog maintained for releases

**Success Metrics**:
- New developer onboarded in <4 hours
- 90% of issues resolved using runbooks

---

## Functional Requirements

### FR1: RBAC Seed Data
- **Admin Role**: Full system access (all permissions)
  - users:read, users:write, users:delete, users:activate
  - catalog:read, catalog:write, catalog:delete
  - buffers:read, buffers:write, buffers:delete
  - orders:read, orders:write, orders:delete, orders:approve
  - reports:read, reports:write
  - settings:read, settings:write

- **Manager Role**: Operational management
  - users:read, users:write
  - catalog:read, catalog:write
  - buffers:read, buffers:write
  - orders:read, orders:write, orders:approve
  - reports:read

- **Analyst Role**: Read-only analysis
  - catalog:read
  - buffers:read
  - orders:read
  - reports:read

- **Viewer Role**: Read-only viewing
  - catalog:read
  - buffers:read
  - orders:read

### FR2: gRPC Reflection
- Enable reflection on all gRPC servers
- Register all services with reflection API
- Document how to use grpcurl for debugging
- Create client helper libraries for common operations

### FR3: Additional Domain Events
- **Auth Service Events**:
  - `user.registered` - New user created
  - `user.verified` - Email verified
  - `user.password_changed` - Password updated
  - `user.activated` - Account activated by admin
  - `user.deactivated` - Account deactivated by admin
  - `role.created`, `role.updated`, `role.deleted`

- **Catalog Service Events**:
  - `supplier.created`, `supplier.updated`, `supplier.deleted`
  - `buffer_profile.created`, `buffer_profile.updated`, `buffer_profile.deleted`
  - `product_supplier.associated`, `product_supplier.removed`

### FR4: Dead Letter Queue Processing
- Consumer service monitors DLQ stream
- Logs failed events with error details
- Retries failed events with exponential backoff
- Alerts on DLQ threshold (>10 events in 1 hour)
- Admin UI to view and manually replay DLQ events (future enhancement)

### FR5: Production Configuration
- Health endpoints: `/health/live`, `/health/ready`
- Metrics endpoints: `/metrics` (Prometheus format)
- Graceful shutdown: Wait for in-flight requests (max 30s)
- Startup validation: Check database, NATS, Redis connectivity
- Environment configs: `.env.dev`, `.env.staging`, `.env.prod`
- Secret injection: Use Kubernetes secrets, not hardcoded

---

## Key Tasks

### RBAC and Audit Logging

#### T001: Create Seed Data Script
**File**: `services/auth-service/scripts/seed-roles.sh`

```bash
#!/bin/bash
# Seed standard roles and permissions

psql $DATABASE_URL << EOF
-- Insert standard roles
INSERT INTO roles (id, organization_id, name, description, level) VALUES
  (gen_random_uuid(), NULL, 'Admin', 'Full system administrator', 1),
  (gen_random_uuid(), NULL, 'Manager', 'Operations manager', 2),
  (gen_random_uuid(), NULL, 'Analyst', 'Read-only analyst', 3),
  (gen_random_uuid(), NULL, 'Viewer', 'Read-only viewer', 4)
ON CONFLICT (name) DO NOTHING;

-- Insert standard permissions
INSERT INTO permissions (id, name, description, resource, action) VALUES
  (gen_random_uuid(), 'users:read', 'Read users', 'users', 'read'),
  (gen_random_uuid(), 'users:write', 'Create/update users', 'users', 'write'),
  -- ... more permissions
ON CONFLICT (name) DO NOTHING;

-- Associate permissions with roles
-- Admin gets all permissions
-- Manager gets operational permissions
-- Analyst gets read-only permissions
-- Viewer gets basic read permissions
EOF
```

#### T002: Implement Audit Logging
**File**: `services/auth-service/internal/infrastructure/adapters/audit/audit_logger.go`

```go
type AuditLogger struct {
    db     *gorm.DB
    logger pkgLogger.Logger
}

type AuditLog struct {
    ID             uuid.UUID
    OrganizationID uuid.UUID
    UserID         uuid.UUID
    Action         string // "permission.check", "role.assign"
    Resource       string
    Result         string // "allowed", "denied"
    IPAddress      string
    UserAgent      string
    Timestamp      time.Time
}

func (a *AuditLogger) LogPermissionCheck(ctx context.Context, userID uuid.UUID, permission string, allowed bool) {
    // Log to database and structured logs
}
```

#### T003: Performance Testing
**File**: `services/auth-service/test/performance/permission_check_bench_test.go`

```go
func BenchmarkCheckPermission(b *testing.B) {
    // Setup
    useCase := setupCheckPermissionUseCase()
    userID := uuid.New()
    permission := "catalog:read"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = useCase.Execute(context.Background(), userID, permission)
    }
}

// Target: <10ms p95 (cached), <50ms p95 (uncached)
```

---

### gRPC Reflection and Examples

#### T004: Enable gRPC Reflection
**File**: `services/auth-service/internal/infrastructure/grpc/server/server.go`

```go
import "google.golang.org/grpc/reflection"

func NewGRPCServer(useCases *UseCases, logger pkgLogger.Logger) *grpc.Server {
    server := grpc.NewServer(/* interceptors */)

    // Register services
    authpb.RegisterAuthServiceServer(server, authServiceHandler)

    // Enable reflection
    reflection.Register(server)

    return server
}
```

#### T005: Create Client Examples
**File**: `examples/grpc-clients/auth-client/main.go`

```go
package main

import (
    "context"
    "log"

    "google.golang.org/grpc"
    authpb "github.com/giia/giia-core-engine/services/auth-service/api/proto/auth/v1"
)

func main() {
    // Connect to Auth service
    conn, err := grpc.Dial("localhost:9091", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := authpb.NewAuthServiceClient(conn)

    // Example 1: Validate Token
    resp, err := client.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{
        Token: "eyJhbGciOiJIUzI1NiIs...",
    })
    if err != nil {
        log.Fatalf("ValidateToken failed: %v", err)
    }
    log.Printf("Token valid: %v", resp.Valid)

    // Example 2: Check Permission
    permResp, err := client.CheckPermission(context.Background(), &authpb.CheckPermissionRequest{
        UserId:     "550e8400-e29b-41d4-a716-446655440000",
        Permission: "catalog:write",
    })
    if err != nil {
        log.Fatalf("CheckPermission failed: %v", err)
    }
    log.Printf("Permission allowed: %v", permResp.Allowed)
}
```

**File**: `examples/grpc-clients/README.md` - Usage instructions

#### T006: Catalog Service Proto Files
**File**: `services/catalog-service/api/proto/catalog/v1/catalog.proto`

```protobuf
syntax = "proto3";

package catalog.v1;

option go_package = "github.com/giia/giia-core-engine/services/catalog-service/api/proto/catalog/v1;catalogpb";

service CatalogService {
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc SearchProducts(SearchProductsRequest) returns (SearchProductsResponse);
  rpc GetSupplier(GetSupplierRequest) returns (GetSupplierResponse);
  rpc ListSuppliers(ListSuppliersRequest) returns (ListSuppliersResponse);
}

message Product {
  string id = 1;
  string organization_id = 2;
  string sku = 3;
  string name = 4;
  string category = 5;
  string status = 6;
  // ... more fields
}

// Request/Response messages...
```

---

### NATS Event System Polish

#### T007: Publish Additional Auth Events
**File**: `services/auth-service/internal/core/usecases/auth/register_user.go` (updated)

```go
// After user registration
uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
    Type:    "user.registered",
    Subject: fmt.Sprintf("auth.users.%s", user.ID.String()),
    Data: map[string]interface{}{
        "user_id":         user.ID.String(),
        "email":           user.Email,
        "organization_id": user.OrganizationID.String(),
        "status":          user.Status,
    },
})

// After email verification
uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
    Type:    "user.verified",
    Subject: fmt.Sprintf("auth.users.%s", user.ID.String()),
    Data: map[string]interface{}{
        "user_id":     user.ID.String(),
        "verified_at": user.VerifiedAt,
    },
})
```

#### T008: Implement DLQ Consumer
**File**: `services/monitoring-service/internal/consumers/dlq_consumer.go` (new service or standalone)

```go
type DLQConsumer struct {
    subscriber pkgEvents.Subscriber
    logger     pkgLogger.Logger
}

func (c *DLQConsumer) Start(ctx context.Context) error {
    return c.subscriber.Subscribe(ctx, "DLQ_EVENTS", "dlq.>", func(event *pkgEvents.Event) error {
        // Log failed event
        c.logger.Error(ctx, nil, "Event failed and sent to DLQ", pkgLogger.Tags{
            "event_type":    event.Type,
            "event_subject": event.Subject,
            "event_data":    fmt.Sprintf("%+v", event.Data),
        })

        // Check if should retry
        retryCount := getRetryCount(event)
        if retryCount < 3 {
            // Retry with exponential backoff
            time.Sleep(time.Duration(math.Pow(2, float64(retryCount))) * time.Second)
            return c.retryEvent(ctx, event)
        }

        // Alert on persistent failure
        if retryCount >= 3 {
            c.alertCriticalFailure(event)
        }

        return nil
    })
}
```

#### T009: Event Schema Validation
**File**: `pkg/events/schema_validator.go`

```go
type EventSchema struct {
    Type          string
    RequiredFields []string
}

var schemas = map[string]EventSchema{
    "user.registered": {
        Type:          "user.registered",
        RequiredFields: []string{"user_id", "email", "organization_id"},
    },
    "product.created": {
        Type:          "product.created",
        RequiredFields: []string{"product_id", "sku", "name"},
    },
}

func ValidateEvent(event *Event) error {
    schema, exists := schemas[event.Type]
    if !exists {
        return fmt.Errorf("unknown event type: %s", event.Type)
    }

    for _, field := range schema.RequiredFields {
        if _, ok := event.Data[field]; !ok {
            return fmt.Errorf("missing required field: %s", field)
        }
    }

    return nil
}
```

---

### Production Readiness

#### T010: Health Check Endpoints
**File**: `services/auth-service/internal/infrastructure/entrypoints/http/health_handlers.go`

```go
type HealthHandlers struct {
    db         *gorm.DB
    nats       *nats.Conn
    redis      *redis.Client
    logger     pkgLogger.Logger
}

// GET /health/live - Kubernetes liveness probe
func (h *HealthHandlers) Liveness(w http.ResponseWriter, r *http.Request) {
    respondJSON(w, http.StatusOK, map[string]string{
        "status": "alive",
    })
}

// GET /health/ready - Kubernetes readiness probe
func (h *HealthHandlers) Readiness(w http.ResponseWriter, r *http.Request) {
    checks := map[string]string{}

    // Check database
    if err := h.db.Exec("SELECT 1").Error; err != nil {
        checks["database"] = "unhealthy"
    } else {
        checks["database"] = "healthy"
    }

    // Check NATS
    if !h.nats.IsConnected() {
        checks["nats"] = "unhealthy"
    } else {
        checks["nats"] = "healthy"
    }

    // Check Redis
    if err := h.redis.Ping(r.Context()).Err(); err != nil {
        checks["redis"] = "unhealthy"
    } else {
        checks["redis"] = "healthy"
    }

    // Determine overall health
    allHealthy := true
    for _, status := range checks {
        if status == "unhealthy" {
            allHealthy = false
            break
        }
    }

    statusCode := http.StatusOK
    if !allHealthy {
        statusCode = http.StatusServiceUnavailable
    }

    respondJSON(w, statusCode, map[string]interface{}{
        "status": map[string]bool{"ready": allHealthy},
        "checks": checks,
    })
}
```

#### T011: Graceful Shutdown
**File**: `services/auth-service/cmd/server/main.go`

```go
func main() {
    // Start HTTP server
    httpServer := &http.Server{
        Addr:    ":8083",
        Handler: router,
    }

    // Start gRPC server
    grpcServer := setupGRPCServer()

    // Handle shutdown signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Start servers in goroutines
    go func() {
        if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error(context.Background(), err, "HTTP server error")
        }
    }()

    go func() {
        if err := grpcServer.Serve(listener); err != nil {
            logger.Error(context.Background(), err, "gRPC server error")
        }
    }()

    // Wait for shutdown signal
    <-quit
    logger.Info(context.Background(), "Shutting down gracefully...")

    // Graceful shutdown with 30s timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Shutdown HTTP server
    if err := httpServer.Shutdown(ctx); err != nil {
        logger.Error(ctx, err, "HTTP server forced to shutdown")
    }

    // Shutdown gRPC server
    grpcServer.GracefulStop()

    logger.Info(context.Background(), "Server exited")
}
```

---

### Documentation

#### T012: Architecture Diagrams
**File**: `docs/architecture/sequence-diagrams.md`

Create diagrams for:
- User registration flow (Frontend → Auth → Email → Database)
- Token validation flow (Catalog → Auth gRPC → Cache → Database)
- Order creation flow (Frontend → Execution → DDMRP → Catalog → Events)
- Buffer calculation flow (DDMRP → Catalog → Database → Events)

#### T013: API Documentation
- Generate OpenAPI/Swagger docs for REST APIs
- Generate proto docs for gRPC APIs using protoc-gen-doc
- Host documentation on service endpoints (e.g., `/docs`)

#### T014: Runbooks
**File**: `docs/runbooks/README.md`

Create runbooks for:
- **Deployment**: How to deploy services to staging/production
- **Rollback**: How to rollback a failed deployment
- **Scaling**: How to scale services horizontally
- **Troubleshooting**: Common issues and solutions
- **Database Migrations**: How to run migrations safely
- **Backup and Restore**: How to backup/restore databases

---

## Non-Functional Requirements

### Performance
- Permission checks: <10ms p95 (cached), <50ms p95 (uncached)
- Health checks: <100ms p95
- Graceful shutdown: <30s for all in-flight requests

### Reliability
- Event delivery: 99.99% (at-least-once semantics)
- Health check availability: 99.9%
- Zero data loss during graceful shutdown

### Security
- Audit logging for all permission checks
- Secret rotation supported (no hardcoded secrets)
- Rate limiting on public endpoints (100 req/min per IP)

### Observability
- All services expose Prometheus metrics
- All critical events logged with structured logging
- Distributed tracing with request IDs

---

## Success Criteria

### Mandatory (Must Have)
- ✅ RBAC seed data with 4 standard roles and all permissions
- ✅ Audit logging for permission checks
- ✅ Performance tests showing <10ms p95 for permission checks
- ✅ gRPC reflection enabled on all services
- ✅ Working gRPC client examples
- ✅ Catalog service proto files defined
- ✅ Additional domain events published (user.registered, user.verified, etc.)
- ✅ DLQ consumer processing failed events
- ✅ Event schema validation
- ✅ Health check endpoints (/health/live, /health/ready)
- ✅ Graceful shutdown handling
- ✅ Environment-specific configs (.env.dev, .env.staging, .env.prod)
- ✅ Architecture diagrams updated
- ✅ API documentation generated
- ✅ Basic runbooks created

### Optional (Nice to Have)
- ⚪ Admin UI for DLQ event replay
- ⚪ Real-time event monitoring dashboard
- ⚪ Advanced rate limiting with token bucket
- ⚪ Comprehensive troubleshooting guide

---

## Out of Scope

- ❌ Advanced monitoring (Grafana dashboards) - Future observability task
- ❌ Distributed tracing implementation - Future observability task
- ❌ Load testing - Future performance task
- ❌ Security penetration testing - Future security task

---

## Dependencies

- **Tasks 11, 12, 13**: Core Phase 2A tasks must be in progress
- **All Phase 1 Services**: Auth, Catalog, Shared Packages
- **Infrastructure**: PostgreSQL, Redis, NATS, Kubernetes

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Seed data conflicts with existing data | Medium | Medium | Idempotent scripts with ON CONFLICT handling |
| Performance targets not met | Medium | Low | Performance testing early, optimize queries |
| DLQ processing complexity | Medium | Medium | Start simple, iterate based on failure patterns |
| Documentation becomes stale | Low | High | Automate doc generation where possible |

---

## Deliverables Checklist

- [ ] RBAC seed data script (seed-roles.sh)
- [ ] Audit logging implementation
- [ ] Performance benchmarks passing (<10ms p95)
- [ ] gRPC reflection enabled
- [ ] Client examples (auth-client, catalog-client)
- [ ] Catalog proto files
- [ ] Additional auth events (user.registered, user.verified, user.password_changed)
- [ ] Additional catalog events (supplier.created, buffer_profile.created)
- [ ] DLQ consumer service
- [ ] Event schema validator
- [ ] Health check endpoints
- [ ] Graceful shutdown handling
- [ ] Environment configs (.env.dev, .env.staging, .env.prod)
- [ ] Architecture sequence diagrams
- [ ] OpenAPI/Swagger docs
- [ ] Proto documentation
- [ ] Deployment runbook
- [ ] Troubleshooting runbook
- [ ] All tests passing (unit, integration, performance)
- [ ] Phase 1 at 100% completion ✅

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)
**Estimated Completion**: 1 week (5 days)
