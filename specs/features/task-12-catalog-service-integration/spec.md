# Task 12: Catalog Service Integration - Specification

**Task ID**: task-12-catalog-service-integration
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Completion**: 15% remaining work on catalog-service
**Dependencies**: Task 9 (85% complete), Task 7 (95% complete - gRPC proto files)

---

## Overview

Complete the catalog-service by implementing gRPC endpoints for inter-service communication, Supplier and BufferProfile use cases, integration with Auth service for authentication/authorization, and comprehensive testing. This task brings catalog-service from 85% to 100% completion.

---

## User Scenarios

### US1: gRPC Catalog Service API (P1)

**As a** microservice (DDMRP Engine, Execution Service)
**I want to** access catalog data via gRPC
**So that** I can retrieve product and supplier information for calculations and orders

**Acceptance Criteria**:
- gRPC service definition with proto files
- GetProduct RPC returns product by ID
- ListProducts RPC returns paginated products
- SearchProducts RPC with filters (SKU, category, status)
- GetSupplier RPC returns supplier by ID
- ListSuppliers RPC returns paginated suppliers
- GetBufferProfile RPC returns buffer profile by ID
- All RPCs enforce multi-tenancy (organization_id)
- All RPCs return structured errors

**Success Metrics**:
- <50ms p50 for GetProduct
- <100ms p50 for List operations
- 100% gRPC endpoint coverage

---

### US2: Supplier Management (P1)

**As a** catalog manager
**I want to** manage supplier information
**So that** I can track product sources and lead times

**Acceptance Criteria**:
- Create supplier with name, contact, lead time
- Update supplier information
- Delete supplier (soft delete if has products)
- List suppliers with pagination
- Search suppliers by name or code
- Associate suppliers with products (many-to-many)
- Multi-tenancy: Suppliers scoped to organization

**Success Metrics**:
- <2s p95 for CRUD operations
- <5s p95 for list/search operations
- 100% use case test coverage

---

### US3: Buffer Profile Management (P2)

**As a** DDMRP planner
**I want to** manage buffer profiles
**So that** I can configure buffer calculation parameters for product categories

**Acceptance Criteria**:
- Create buffer profile with ADU method, lead time factor, variability factor
- Update buffer profile parameters
- Delete buffer profile (soft delete if assigned to products)
- List buffer profiles with pagination
- Assign buffer profile to products
- Multi-tenancy: Buffer profiles scoped to organization

**Success Metrics**:
- <2s p95 for CRUD operations
- 100% use case test coverage

---

### US4: Authentication and Authorization Integration (P1)

**As a** catalog service
**I want to** validate user authentication and permissions via Auth service gRPC
**So that** I can enforce access control on catalog operations

**Acceptance Criteria**:
- gRPC client for Auth service (ValidateToken, CheckPermission)
- Middleware validates JWT tokens before processing requests
- Permission checks: `catalog:read`, `catalog:write`, `catalog:delete`
- Multi-tenancy enforcement: Users can only access their organization's data
- Graceful handling of Auth service unavailability

**Success Metrics**:
- <10ms overhead for token validation
- <5ms overhead for permission checks (cached)
- 100% endpoint coverage with auth checks

---

### US5: Comprehensive Testing (P1)

**As a** developer
**I want to** comprehensive test coverage for catalog-service
**So that** I can confidently deploy to production

**Acceptance Criteria**:
- Unit tests for all use cases (80%+ coverage)
- Integration tests with real PostgreSQL database
- gRPC integration tests with mock clients
- End-to-end tests for complete flows
- Performance tests for high-load scenarios

**Success Metrics**:
- 80%+ overall test coverage
- 95%+ coverage for use cases
- All tests pass in CI/CD pipeline

---

## Functional Requirements

### FR1: gRPC Service Definition
- Define catalog.proto with all RPC methods
- Generate Go code with protoc
- Implement all RPC methods in gRPC server
- Error handling with gRPC status codes
- Request/response validation

### FR2: Supplier Use Cases
- **CreateSupplier**: Validate name, code, lead time; enforce uniqueness
- **UpdateSupplier**: Validate changes; prevent conflicts
- **DeleteSupplier**: Soft delete if has products; hard delete otherwise
- **GetSupplier**: Return supplier by ID; enforce multi-tenancy
- **ListSuppliers**: Pagination, sorting, filtering
- **SearchSuppliers**: Full-text search by name/code

### FR3: BufferProfile Use Cases
- **CreateBufferProfile**: Validate ADU method, lead time factor, variability factor
- **UpdateBufferProfile**: Validate changes
- **DeleteBufferProfile**: Soft delete if assigned to products
- **GetBufferProfile**: Return by ID; enforce multi-tenancy
- **ListBufferProfiles**: Pagination and filtering
- **AssignBufferProfile**: Link to products

### FR4: Product-Supplier Association
- Many-to-many relationship
- Primary supplier designation
- Lead time per supplier-product pair
- Product availability tracking

### FR5: Auth Service Integration
- gRPC client initialization and connection pooling
- Middleware for token validation on all HTTP/gRPC endpoints
- Permission checking before sensitive operations
- Cache auth responses for performance (5-minute TTL)
- Fallback behavior for Auth service outages (deny by default)

---

## Key Entities

### Supplier
```go
type Supplier struct {
    ID              uuid.UUID
    OrganizationID  uuid.UUID
    Code            string    // Unique supplier code (e.g., "SUP-001")
    Name            string
    ContactName     string
    ContactEmail    string
    ContactPhone    string
    DefaultLeadTime int       // Days
    Reliability     SupplierReliability // [NEW] For supply variability in buffer calculations
    Status          SupplierStatus // "active", "inactive"
    Address         string
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time // Soft delete
}

type SupplierReliability string

const (
    SupplierReliabilityHigh   SupplierReliability = "high"   // Variability Low (B): 0-40%
    SupplierReliabilityMedium SupplierReliability = "medium" // Variability Medium (M): 41-60%
    SupplierReliabilityLow    SupplierReliability = "low"    // Variability High (A): 61-100%
)

type SupplierStatus string

const (
    SupplierStatusActive   SupplierStatus = "active"
    SupplierStatusInactive SupplierStatus = "inactive"
)
```

### BufferProfile
```go
type BufferProfile struct {
    ID                    uuid.UUID
    OrganizationID        uuid.UUID
    Name                  string
    Description           string
    ADUMethod             ADUMethod // "average", "exponential", "weighted"
    LeadTimeCategory      LeadTimeCategory // "long", "medium", "short" [UPDATED]
    VariabilityCategory   VariabilityCategory // "high", "medium", "low" [UPDATED]
    LeadTimeFactor        float64   // %LT from matrix (0.2 to 0.7) [UPDATED]
    VariabilityFactor     float64   // %CV from matrix (0.25 to 1.0) [UPDATED]
    Status                BufferProfileStatus
    CreatedAt             time.Time
    UpdatedAt             time.Time
    DeletedAt             *time.Time
}

type ADUMethod string

const (
    ADUMethodAverage      ADUMethod = "average"
    ADUMethodExponential  ADUMethod = "exponential"
    ADUMethodWeighted     ADUMethod = "weighted"
)

type LeadTimeCategory string

const (
    LeadTimeLong   LeadTimeCategory = "long"    // >60 days
    LeadTimeMedium LeadTimeCategory = "medium"  // 15-60 days
    LeadTimeShort  LeadTimeCategory = "short"   // <15 days
)

type VariabilityCategory string

const (
    VariabilityHigh   VariabilityCategory = "high"    // 61-100% coefficient
    VariabilityMedium VariabilityCategory = "medium"  // 41-60% coefficient
    VariabilityLow    VariabilityCategory = "low"     // 0-40% coefficient
)

type BufferProfileStatus string

const (
    BufferProfileStatusActive   BufferProfileStatus = "active"
    BufferProfileStatusInactive BufferProfileStatus = "inactive"
)

// Buffer Profile Matrix - relates Lead Time category with Variability
// Returns %CV (variability coefficient) based on matrix
var BufferProfileMatrix = map[LeadTimeCategory]map[VariabilityCategory]float64{
    LeadTimeLong: {
        VariabilityHigh:   1.00,
        VariabilityMedium: 0.75,
        VariabilityLow:    0.50,
    },
    LeadTimeMedium: {
        VariabilityHigh:   0.75,
        VariabilityMedium: 0.50,
        VariabilityLow:    0.25,
    },
    LeadTimeShort: {
        VariabilityHigh:   0.50,
        VariabilityMedium: 0.25,
        VariabilityLow:    0.25,
    },
}
```

### ProductSupplier (Association)
```go
type ProductSupplier struct {
    ID                uuid.UUID
    ProductID         uuid.UUID
    SupplierID        uuid.UUID
    IsPrimarySupplier bool
    LeadTimeDays      int
    UnitCost          float64
    MinOrderQuantity  int
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

### Product (Updated)
```go
type Product struct {
    ID              uuid.UUID
    OrganizationID  uuid.UUID
    SKU             string
    Name            string
    Description     string
    Category        string
    UnitOfMeasure   string
    StandardCost    float64      // [NEW] For inventory valuation and KPIs
    LastPurchaseDate *time.Time  // [NEW] For obsolescence and Days in Inventory calculation
    BufferProfileID *uuid.UUID   // [NEW] Link to buffer profile
    Status          ProductStatus
    CreatedAt       time.Time
    UpdatedAt       time.Time
    DeletedAt       *time.Time
}

type ProductStatus string

const (
    ProductStatusActive       ProductStatus = "active"
    ProductStatusInactive     ProductStatus = "inactive"
    ProductStatusDiscontinued ProductStatus = "discontinued"
)
```

---

## Non-Functional Requirements

### Performance
- gRPC GetProduct: <50ms p50, <100ms p95
- gRPC ListProducts: <100ms p50, <200ms p95
- REST CRUD operations: <2s p95
- Auth token validation: <10ms p95 (with caching)

### Reliability
- Auth service unavailability should not crash catalog service
- Database transactions for data consistency
- Retry logic for transient failures (3 attempts with exponential backoff)

### Security
- All endpoints require valid JWT token
- Permission checks enforce least privilege
- Multi-tenancy strictly enforced (organization_id filtering)
- SQL injection prevention (parameterized queries)

### Observability
- Log all gRPC requests/responses
- Log auth validation attempts
- Metrics: grpc_request_count, auth_validation_duration, db_query_duration
- Distributed tracing with request IDs

---

## Success Criteria

### Mandatory (Must Have)
- ✅ gRPC proto files defined for Catalog service
- ✅ All gRPC RPCs implemented and tested
- ✅ Supplier use cases (Create, Update, Delete, Get, List, Search)
- ✅ BufferProfile use cases (Create, Update, Delete, Get, List)
- ✅ Product-Supplier association working
- ✅ Auth service gRPC client integrated
- ✅ Token validation middleware on all endpoints
- ✅ Permission checks on sensitive operations
- ✅ Unit tests 80%+ coverage
- ✅ Integration tests with database
- ✅ Multi-tenancy enforced everywhere

### Optional (Nice to Have)
- ⚪ Bulk operations (bulk create/update suppliers)
- ⚪ Import/export suppliers from CSV
- ⚪ Supplier performance metrics
- ⚪ Buffer profile templates library

---

## Out of Scope

- ❌ Supplier rating and reviews - Future task
- ❌ Supplier contract management - Future task
- ❌ Automated lead time updates - Future task
- ❌ Purchase order management - Execution service responsibility
- ❌ Supplier portal for self-service - Future task

---

## Dependencies

- **Task 9**: Catalog service at 85% (Product use cases implemented)
- **Task 7**: gRPC proto files and implementation at 95% (Auth service)
- **Task 5**: Auth service at 95% (for gRPC integration)
- **Shared Packages**: pkg/events, pkg/database, pkg/logger, pkg/errors

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Auth service coupling | High | Medium | Implement circuit breaker, fallback to deny |
| gRPC version incompatibility | Medium | Low | Pin protobuf and gRPC versions |
| Performance degradation from auth checks | Medium | Medium | Cache auth responses, optimize queries |
| Complex many-to-many relationships | Medium | Medium | Thorough testing of associations |
| Multi-tenancy data leaks | Critical | Low | Comprehensive tests, code reviews |

---

## References

- [Task 9 Spec](../task-09-catalog-service/spec.md) - Catalog service foundation
- [Task 7 Spec](../task-07-grpc-server/spec.md) - gRPC implementation
- [Auth Service Proto](../../services/auth-service/api/proto/auth/v1/auth.proto) - Auth gRPC API
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Planning
**Next Step**: Create implementation plan (plan.md)