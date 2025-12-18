# Task 12: Catalog Service Integration - Implementation Plan

**Task ID**: task-12-catalog-service-integration
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Duration**: 2 weeks
**Dependencies**: Task 9 (85%), Task 7 (95%), Task 18 (gRPC proto files)

---

## 1. Technical Context

### Current State
- **Catalog Service**: 85% complete
  - Product use cases implemented (Create, Update, Get, List)
  - Database migrations for products table
  - REST API endpoints for products
  - Basic multi-tenancy support
- **Missing**: Supplier/BufferProfile entities, gRPC server, Auth integration

### Technology Stack
- **Language**: Go 1.23.4
- **Architecture**: Clean Architecture (Domain, Use Cases, Infrastructure)
- **Database**: PostgreSQL 16 with GORM
- **gRPC**: Protocol Buffers v3
- **Auth Integration**: gRPC client to Auth service
- **Event Streaming**: NATS JetStream
- **Testing**: testify, httptest, gRPC testing framework

### Key Design Decisions
- **gRPC + REST**: gRPC for service-to-service, REST for frontend
- **Auth Middleware**: Validate tokens and permissions before use case execution
- **Supplier-Product**: Many-to-many with product_suppliers join table
- **Soft Deletes**: Preserve data integrity for referenced entities
- **Multi-tenancy**: organization_id filtering at repository layer

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
└── services/catalog-service/
    ├── api/
    │   └── proto/
    │       └── catalog/
    │           └── v1/
    │               ├── catalog.proto                   [CREATED in Task 18]
    │               ├── catalog.pb.go                   [GENERATED]
    │               └── catalog_grpc.pb.go              [GENERATED]
    │
    ├── internal/
    │   ├── core/
    │   │   ├── domain/
    │   │   │   ├── supplier.go                        [NEW]
    │   │   │   ├── buffer_profile.go                  [NEW]
    │   │   │   ├── product_supplier.go                [NEW]
    │   │   │   └── product.go                         [MODIFY] Add BufferProfileID
    │   │   │
    │   │   ├── providers/
    │   │   │   ├── supplier_repository.go             [NEW]
    │   │   │   ├── buffer_profile_repository.go       [NEW]
    │   │   │   ├── product_supplier_repository.go     [NEW]
    │   │   │   └── auth_service_client.go             [NEW]
    │   │   │
    │   │   └── usecases/
    │   │       ├── supplier/
    │   │       │   ├── create_supplier.go             [NEW]
    │   │       │   ├── update_supplier.go             [NEW]
    │   │       │   ├── delete_supplier.go             [NEW]
    │   │       │   ├── get_supplier.go                [NEW]
    │   │       │   ├── list_suppliers.go              [NEW]
    │   │       │   └── search_suppliers.go            [NEW]
    │   │       │
    │   │       ├── buffer_profile/
    │   │       │   ├── create_buffer_profile.go       [NEW]
    │   │       │   ├── update_buffer_profile.go       [NEW]
    │   │       │   ├── delete_buffer_profile.go       [NEW]
    │   │       │   ├── get_buffer_profile.go          [NEW]
    │   │       │   └── list_buffer_profiles.go        [NEW]
    │   │       │
    │   │       └── product/
    │   │           ├── assign_buffer_profile.go       [NEW]
    │   │           ├── associate_supplier.go          [NEW]
    │   │           └── get_product.go                 [MODIFY] Include suppliers
    │   │
    │   └── infrastructure/
    │       ├── adapters/
    │       │   └── auth/
    │       │       ├── grpc_auth_client.go            [NEW]
    │       │       └── auth_client_mock.go            [NEW]
    │       │
    │       ├── repositories/
    │       │   ├── supplier_repository.go             [NEW]
    │       │   ├── buffer_profile_repository.go       [NEW]
    │       │   └── product_supplier_repository.go     [NEW]
    │       │
    │       ├── entrypoints/
    │       │   ├── http/
    │       │   │   ├── supplier_handlers.go           [NEW]
    │       │   │   └── buffer_profile_handlers.go     [NEW]
    │       │   │
    │       │   └── grpc/
    │       │       ├── server/
    │       │       │   └── server.go                  [NEW]
    │       │       └── handlers/
    │       │           └── catalog_handlers.go        [NEW]
    │       │
    │       └── middlewares/
    │           ├── auth_middleware.go                 [NEW]
    │           └── grpc_auth_interceptor.go           [NEW]
    │
    ├── migrations/
    │   ├── YYYYMMDDHHMMSS_create_suppliers.up.sql     [NEW]
    │   ├── YYYYMMDDHHMMSS_create_buffer_profiles.up.sql [NEW]
    │   ├── YYYYMMDDHHMMSS_create_product_suppliers.up.sql [NEW]
    │   └── YYYYMMDDHHMMSS_add_buffer_profile_to_products.up.sql [NEW]
    │
    ├── test/
    │   ├── integration/
    │   │   ├── supplier_test.go                       [NEW]
    │   │   ├── buffer_profile_test.go                 [NEW]
    │   │   └── grpc_catalog_test.go                   [NEW]
    │   │
    │   └── performance/
    │       └── grpc_benchmark_test.go                 [NEW]
    │
    └── cmd/
        └── server/
            └── main.go                                [MODIFY] Start gRPC server

```

---

## 3. Implementation Phases

### Phase 1: Database Schema and Domain Entities (Days 1-2)

#### T001: Create Database Migrations

**File**: `services/catalog-service/migrations/000005_create_suppliers.up.sql`

```sql
-- Suppliers table
CREATE TABLE IF NOT EXISTS suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    contact_name VARCHAR(200),
    contact_email VARCHAR(200),
    contact_phone VARCHAR(50),
    default_lead_time INTEGER NOT NULL DEFAULT 7,
    reliability VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    address TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT uq_supplier_code_org UNIQUE (organization_id, code),
    CONSTRAINT chk_supplier_reliability CHECK (reliability IN ('high', 'medium', 'low'))
);

CREATE INDEX idx_suppliers_org ON suppliers(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_suppliers_status ON suppliers(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_suppliers_code ON suppliers(code) WHERE deleted_at IS NULL;
```

**File**: `services/catalog-service/migrations/000006_create_buffer_profiles.up.sql`

```sql
-- Buffer Profiles table
CREATE TABLE IF NOT EXISTS buffer_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    adu_method VARCHAR(20) NOT NULL DEFAULT 'average', -- 'average', 'exponential', 'weighted'
    lead_time_category VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'long', 'medium', 'short'
    variability_category VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'high', 'medium', 'low'
    lead_time_factor DECIMAL(5,2) NOT NULL DEFAULT 0.5,
    variability_factor DECIMAL(5,2) NOT NULL DEFAULT 0.5,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT chk_adu_method CHECK (adu_method IN ('average', 'exponential', 'weighted')),
    CONSTRAINT chk_lead_time_category CHECK (lead_time_category IN ('long', 'medium', 'short')),
    CONSTRAINT chk_variability_category CHECK (variability_category IN ('high', 'medium', 'low')),
    CONSTRAINT chk_lead_time_factor CHECK (lead_time_factor > 0),
    CONSTRAINT chk_variability_factor CHECK (variability_factor > 0)
);

CREATE INDEX idx_buffer_profiles_org ON buffer_profiles(organization_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_buffer_profiles_status ON buffer_profiles(status) WHERE deleted_at IS NULL;
```

**File**: `services/catalog-service/migrations/000007_create_product_suppliers.up.sql`

```sql
-- Product-Supplier association table (many-to-many)
CREATE TABLE IF NOT EXISTS product_suppliers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
    is_primary_supplier BOOLEAN NOT NULL DEFAULT false,
    lead_time_days INTEGER NOT NULL DEFAULT 7,
    unit_cost DECIMAL(15,2),
    min_order_quantity INTEGER DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_product_supplier UNIQUE (product_id, supplier_id)
);

CREATE INDEX idx_product_suppliers_product ON product_suppliers(product_id);
CREATE INDEX idx_product_suppliers_supplier ON product_suppliers(supplier_id);
CREATE INDEX idx_product_suppliers_primary ON product_suppliers(is_primary_supplier) WHERE is_primary_supplier = true;
```

**File**: `services/catalog-service/migrations/000008_add_buffer_profile_to_products.up.sql`

```sql
-- Add buffer_profile_id and last_purchase_date to products table
ALTER TABLE products
ADD COLUMN buffer_profile_id UUID REFERENCES buffer_profiles(id) ON DELETE SET NULL,
ADD COLUMN last_purchase_date TIMESTAMP;

CREATE INDEX idx_products_buffer_profile ON products(buffer_profile_id);
```

#### T002: Define Domain Entities

**File**: `services/catalog-service/internal/core/domain/supplier.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	Code            string
	Name            string
	ContactName     string
	ContactEmail    string
	ContactPhone    string
	DefaultLeadTime int
	Reliability     SupplierReliability // [NEW] For supply variability in buffer calculations
	Status          SupplierStatus
	Address         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type SupplierReliability string

const (
	SupplierReliabilityHigh   SupplierReliability = "high"   // Variability Low (B): 0-40%
	SupplierReliabilityMedium SupplierReliability = "medium" // Variability Medium (M): 41-60%
	SupplierReliabilityLow    SupplierReliability = "low"    // Variability High (A): 61-100%
)

func (r SupplierReliability) IsValid() bool {
	switch r {
	case SupplierReliabilityHigh, SupplierReliabilityMedium, SupplierReliabilityLow:
		return true
	}
	return false
}

type SupplierStatus string

const (
	SupplierStatusActive   SupplierStatus = "active"
	SupplierStatusInactive SupplierStatus = "inactive"
)

func (s SupplierStatus) IsValid() bool {
	switch s {
	case SupplierStatusActive, SupplierStatusInactive:
		return true
	}
	return false
}

func NewSupplier(orgID uuid.UUID, code, name string, leadTime int) (*Supplier, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if code == "" {
		return nil, NewValidationError("code is required")
	}
	if name == "" {
		return nil, NewValidationError("name is required")
	}
	if leadTime <= 0 {
		return nil, NewValidationError("lead time must be positive")
	}

	return &Supplier{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Code:            code,
		Name:            name,
		DefaultLeadTime: leadTime,
		Status:          SupplierStatusActive,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}
```

**File**: `services/catalog-service/internal/core/domain/buffer_profile.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

type BufferProfile struct {
	ID                    uuid.UUID
	OrganizationID        uuid.UUID
	Name                  string
	Description           string
	ADUMethod             ADUMethod
	LeadTimeCategory      LeadTimeCategory      // [NEW] "long", "medium", "short"
	VariabilityCategory   VariabilityCategory   // [NEW] "high", "medium", "low"
	LeadTimeFactor        float64               // %LT from matrix (0.2 to 0.7)
	VariabilityFactor     float64               // %CV from matrix (0.25 to 1.0)
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

func (m ADUMethod) IsValid() bool {
	switch m {
	case ADUMethodAverage, ADUMethodExponential, ADUMethodWeighted:
		return true
	}
	return false
}

type LeadTimeCategory string

const (
	LeadTimeLong   LeadTimeCategory = "long"    // >60 days
	LeadTimeMedium LeadTimeCategory = "medium"  // 15-60 days
	LeadTimeShort  LeadTimeCategory = "short"   // <15 days
)

func (c LeadTimeCategory) IsValid() bool {
	switch c {
	case LeadTimeLong, LeadTimeMedium, LeadTimeShort:
		return true
	}
	return false
}

type VariabilityCategory string

const (
	VariabilityHigh   VariabilityCategory = "high"    // 61-100% coefficient
	VariabilityMedium VariabilityCategory = "medium"  // 41-60% coefficient
	VariabilityLow    VariabilityCategory = "low"     // 0-40% coefficient
)

func (c VariabilityCategory) IsValid() bool {
	switch c {
	case VariabilityHigh, VariabilityMedium, VariabilityLow:
		return true
	}
	return false
}

type BufferProfileStatus string

const (
	BufferProfileStatusActive   BufferProfileStatus = "active"
	BufferProfileStatusInactive BufferProfileStatus = "inactive"
)

func (s BufferProfileStatus) IsValid() bool {
	switch s {
	case BufferProfileStatusActive, BufferProfileStatusInactive:
		return true
	}
	return false
}

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

func GetBufferFactors(leadTimeCategory LeadTimeCategory, variabilityCategory VariabilityCategory) (leadTimeFactor, variabilityFactor float64, err error) {
	if !leadTimeCategory.IsValid() {
		return 0, 0, NewValidationError("invalid lead time category")
	}
	if !variabilityCategory.IsValid() {
		return 0, 0, NewValidationError("invalid variability category")
	}

	variabilityFactor = BufferProfileMatrix[leadTimeCategory][variabilityCategory]

	// Lead time factors based on category
	leadTimeFactors := map[LeadTimeCategory]float64{
		LeadTimeLong:   0.70,
		LeadTimeMedium: 0.50,
		LeadTimeShort:  0.20,
	}

	leadTimeFactor = leadTimeFactors[leadTimeCategory]
	return leadTimeFactor, variabilityFactor, nil
}

func NewBufferProfile(orgID uuid.UUID, name string, method ADUMethod, leadTimeCategory LeadTimeCategory, variabilityCategory VariabilityCategory) (*BufferProfile, error) {
	if orgID == uuid.Nil {
		return nil, NewValidationError("organization_id is required")
	}
	if name == "" {
		return nil, NewValidationError("name is required")
	}
	if !method.IsValid() {
		return nil, NewValidationError("invalid ADU method")
	}

	leadTimeFactor, variabilityFactor, err := GetBufferFactors(leadTimeCategory, variabilityCategory)
	if err != nil {
		return nil, err
	}

	return &BufferProfile{
		ID:                  uuid.New(),
		OrganizationID:      orgID,
		Name:                name,
		ADUMethod:           method,
		LeadTimeCategory:    leadTimeCategory,
		VariabilityCategory: variabilityCategory,
		LeadTimeFactor:      leadTimeFactor,
		VariabilityFactor:   variabilityFactor,
		Status:              BufferProfileStatusActive,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}, nil
}
```

**File**: `services/catalog-service/internal/core/domain/product_supplier.go`

```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

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

func NewProductSupplier(productID, supplierID uuid.UUID, leadTime int) (*ProductSupplier, error) {
	if productID == uuid.Nil {
		return nil, NewValidationError("product_id is required")
	}
	if supplierID == uuid.Nil {
		return nil, NewValidationError("supplier_id is required")
	}
	if leadTime <= 0 {
		return nil, NewValidationError("lead time must be positive")
	}

	return &ProductSupplier{
		ID:                uuid.New(),
		ProductID:         productID,
		SupplierID:        supplierID,
		IsPrimarySupplier: false,
		LeadTimeDays:      leadTime,
		MinOrderQuantity:  1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}, nil
}
```

---

### Phase 2: Auth Service Integration (Days 3-4)

#### T003: Create Auth Service gRPC Client

**File**: `services/catalog-service/internal/core/providers/auth_service_client.go`

```go
package providers

import (
	"context"

	"github.com/google/uuid"
)

type AuthServiceClient interface {
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*User, error)
}

type TokenValidationResult struct {
	Valid          bool
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Email          string
	Error          string
}

type User struct {
	ID             uuid.UUID
	Email          string
	Name           string
	OrganizationID uuid.UUID
	Status         string
	Roles          []string
}
```

**File**: `services/catalog-service/internal/infrastructure/adapters/auth/grpc_auth_client.go`

```go
package auth

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/giia/giia-core-engine/services/auth-service/api/proto/auth/v1"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/google/uuid"
)

type grpcAuthClient struct {
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
	logger logger.Logger
}

func NewGRPCAuthClient(authServiceURL string, logger logger.Logger) (providers.AuthServiceClient, error) {
	// Connect to Auth service with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		authServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := authpb.NewAuthServiceClient(conn)

	logger.Info(context.Background(), "Connected to Auth service", logger.Tags{
		"url": authServiceURL,
	})

	return &grpcAuthClient{
		client: client,
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *grpcAuthClient) ValidateToken(ctx context.Context, token string) (*providers.TokenValidationResult, error) {
	resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		c.logger.Error(ctx, err, "Failed to validate token")
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	if !resp.Valid {
		return &providers.TokenValidationResult{
			Valid: false,
			Error: resp.Error,
		}, nil
	}

	userID, err := uuid.Parse(resp.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	orgID, err := uuid.Parse(resp.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("invalid organization_id format: %w", err)
	}

	return &providers.TokenValidationResult{
		Valid:          true,
		UserID:         userID,
		OrganizationID: orgID,
		Email:          resp.Email,
	}, nil
}

func (c *grpcAuthClient) CheckPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	resp, err := c.client.CheckPermission(ctx, &authpb.CheckPermissionRequest{
		UserId:     userID.String(),
		Permission: permission,
	})
	if err != nil {
		c.logger.Error(ctx, err, "Failed to check permission", logger.Tags{
			"user_id":    userID.String(),
			"permission": permission,
		})
		return false, fmt.Errorf("auth service error: %w", err)
	}

	return resp.Allowed, nil
}

func (c *grpcAuthClient) GetUser(ctx context.Context, userID uuid.UUID) (*providers.User, error) {
	resp, err := c.client.GetUser(ctx, &authpb.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		c.logger.Error(ctx, err, "Failed to get user", logger.Tags{
			"user_id": userID.String(),
		})
		return nil, fmt.Errorf("auth service error: %w", err)
	}

	parsedUserID, _ := uuid.Parse(resp.User.Id)
	parsedOrgID, _ := uuid.Parse(resp.User.OrganizationId)

	return &providers.User{
		ID:             parsedUserID,
		Email:          resp.User.Email,
		Name:           resp.User.Name,
		OrganizationID: parsedOrgID,
		Status:         resp.User.Status,
		Roles:          resp.User.Roles,
	}, nil
}

func (c *grpcAuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
```

#### T004: Implement Auth Middleware

**File**: `services/catalog-service/internal/infrastructure/middlewares/auth_middleware.go`

```go
package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	authClient providers.AuthServiceClient
	logger     logger.Logger
}

func NewAuthMiddleware(authClient providers.AuthServiceClient, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
		logger:     logger,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, r, errors.NewUnauthorizedRequest("missing authorization header"))
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.respondError(w, r, errors.NewUnauthorizedRequest("invalid authorization header format"))
			return
		}

		// Validate token with Auth service
		result, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			m.logger.Error(r.Context(), err, "Auth service error")
			m.respondError(w, r, errors.NewInternalServerError("authentication service unavailable"))
			return
		}

		if !result.Valid {
			m.respondError(w, r, errors.NewUnauthorizedRequest("invalid token"))
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", result.UserID)
		ctx = context.WithValue(ctx, "organization_id", result.OrganizationID)
		ctx = context.WithValue(ctx, "email", result.Email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from context (set by Authenticate middleware)
			userID, ok := r.Context().Value("user_id").(uuid.UUID)
			if !ok {
				m.respondError(w, r, errors.NewUnauthorizedRequest("user not authenticated"))
				return
			}

			// Check permission
			allowed, err := m.authClient.CheckPermission(r.Context(), userID, permission)
			if err != nil {
				m.logger.Error(r.Context(), err, "Permission check failed", logger.Tags{
					"user_id":    userID.String(),
					"permission": permission,
				})
				m.respondError(w, r, errors.NewInternalServerError("authorization service unavailable"))
				return
			}

			if !allowed {
				m.respondError(w, r, errors.NewForbiddenRequest("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) respondError(w http.ResponseWriter, r *http.Request, err error) {
	// Use existing error response utility
	// This would be implemented based on your error handling pattern
	w.Header().Set("Content-Type", "application/json")

	statusCode := http.StatusInternalServerError
	if apiErr, ok := err.(errors.APIError); ok {
		statusCode = apiErr.StatusCode()
	}

	w.WriteHeader(statusCode)
	// Write JSON error response
}
```

**File**: `services/catalog-service/internal/infrastructure/middlewares/grpc_auth_interceptor.go`

```go
package middlewares

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
)

type GRPCAuthInterceptor struct {
	authClient providers.AuthServiceClient
	logger     logger.Logger
}

func NewGRPCAuthInterceptor(authClient providers.AuthServiceClient, logger logger.Logger) *GRPCAuthInterceptor {
	return &GRPCAuthInterceptor{
		authClient: authClient,
		logger:     logger,
	}
}

func (i *GRPCAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		token := strings.TrimPrefix(authHeaders[0], "Bearer ")

		// Validate token
		result, err := i.authClient.ValidateToken(ctx, token)
		if err != nil {
			i.logger.Error(ctx, err, "Token validation failed")
			return nil, status.Error(codes.Internal, "authentication service unavailable")
		}

		if !result.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add user info to context
		ctx = context.WithValue(ctx, "user_id", result.UserID)
		ctx = context.WithValue(ctx, "organization_id", result.OrganizationID)
		ctx = context.WithValue(ctx, "email", result.Email)

		return handler(ctx, req)
	}
}
```

---

### Phase 3: Supplier Use Cases (Days 5-6)

#### T005: Implement Supplier Repository

**File**: `services/catalog-service/internal/core/providers/supplier_repository.go`

```go
package providers

import (
	"context"

	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
)

type SupplierRepository interface {
	Create(ctx context.Context, supplier *domain.Supplier) error
	Update(ctx context.Context, supplier *domain.Supplier) error
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.Supplier, error)
	GetByCode(ctx context.Context, code string, organizationID uuid.UUID) (*domain.Supplier, error)
	List(ctx context.Context, organizationID uuid.UUID, page, pageSize int, status domain.SupplierStatus) ([]*domain.Supplier, int, error)
	Search(ctx context.Context, organizationID uuid.UUID, query string) ([]*domain.Supplier, error)
}
```

**File**: `services/catalog-service/internal/infrastructure/repositories/supplier_repository.go`

```go
package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type supplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) providers.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to create supplier: %v", err))
	}
	return nil
}

func (r *supplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", supplier.ID, supplier.OrganizationID).
		Updates(supplier)

	if result.Error != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to update supplier: %v", result.Error))
	}

	if result.RowsAffected == 0 {
		return errors.NewResourceNotFound("supplier not found")
	}

	return nil
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error {
	// Soft delete
	result := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ?", id, organizationID).
		Delete(&domain.Supplier{})

	if result.Error != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to delete supplier: %v", result.Error))
	}

	if result.RowsAffected == 0 {
		return errors.NewResourceNotFound("supplier not found")
	}

	return nil
}

func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier

	err := r.db.WithContext(ctx).
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL", id, organizationID).
		First(&supplier).Error

	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewResourceNotFound("supplier not found")
	}
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get supplier: %v", err))
	}

	return &supplier, nil
}

func (r *supplierRepository) GetByCode(ctx context.Context, code string, organizationID uuid.UUID) (*domain.Supplier, error) {
	var supplier domain.Supplier

	err := r.db.WithContext(ctx).
		Where("code = ? AND organization_id = ? AND deleted_at IS NULL", code, organizationID).
		First(&supplier).Error

	if err == gorm.ErrRecordNotFound {
		return nil, errors.NewResourceNotFound("supplier not found")
	}
	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to get supplier: %v", err))
	}

	return &supplier, nil
}

func (r *supplierRepository) List(ctx context.Context, organizationID uuid.UUID, page, pageSize int, status domain.SupplierStatus) ([]*domain.Supplier, int, error) {
	var suppliers []*domain.Supplier
	var total int64

	query := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	if err := query.Model(&domain.Supplier{}).Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError(fmt.Sprintf("failed to count suppliers: %v", err))
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&suppliers).Error; err != nil {
		return nil, 0, errors.NewInternalServerError(fmt.Sprintf("failed to list suppliers: %v", err))
	}

	return suppliers, int(total), nil
}

func (r *supplierRepository) Search(ctx context.Context, organizationID uuid.UUID, query string) ([]*domain.Supplier, error) {
	var suppliers []*domain.Supplier

	searchPattern := fmt.Sprintf("%%%s%%", query)

	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted_at IS NULL", organizationID).
		Where("name ILIKE ? OR code ILIKE ?", searchPattern, searchPattern).
		Find(&suppliers).Error

	if err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to search suppliers: %v", err))
	}

	return suppliers, nil
}
```

#### T006: Implement Supplier Use Cases

**File**: `services/catalog-service/internal/core/usecases/supplier/create_supplier.go`

```go
package supplier

import (
	"context"
	"fmt"

	"github.com/giia/giia-core-engine/pkg/errors"
	pkgEvents "github.com/giia/giia-core-engine/pkg/events"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateSupplierUseCase struct {
	repository     providers.SupplierRepository
	eventPublisher pkgEvents.Publisher
	logger         logger.Logger
}

func NewCreateSupplierUseCase(
	repository providers.SupplierRepository,
	eventPublisher pkgEvents.Publisher,
	logger logger.Logger,
) *CreateSupplierUseCase {
	return &CreateSupplierUseCase{
		repository:     repository,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

type CreateSupplierInput struct {
	OrganizationID  uuid.UUID
	Code            string
	Name            string
	ContactName     string
	ContactEmail    string
	ContactPhone    string
	DefaultLeadTime int
	Address         string
}

func (uc *CreateSupplierUseCase) Execute(ctx context.Context, input CreateSupplierInput) (*domain.Supplier, error) {
	// Validate input
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}
	if input.Code == "" {
		return nil, errors.NewBadRequest("code is required")
	}
	if input.Name == "" {
		return nil, errors.NewBadRequest("name is required")
	}
	if input.DefaultLeadTime <= 0 {
		return nil, errors.NewBadRequest("default_lead_time must be positive")
	}

	// Check if supplier code already exists
	existing, _ := uc.repository.GetByCode(ctx, input.Code, input.OrganizationID)
	if existing != nil {
		return nil, errors.NewBadRequest(fmt.Sprintf("supplier with code '%s' already exists", input.Code))
	}

	// Create supplier
	supplier, err := domain.NewSupplier(input.OrganizationID, input.Code, input.Name, input.DefaultLeadTime)
	if err != nil {
		return nil, err
	}

	supplier.ContactName = input.ContactName
	supplier.ContactEmail = input.ContactEmail
	supplier.ContactPhone = input.ContactPhone
	supplier.Address = input.Address

	// Save to repository
	if err := uc.repository.Create(ctx, supplier); err != nil {
		uc.logger.Error(ctx, err, "Failed to create supplier")
		return nil, err
	}

	// Publish event
	if err := uc.eventPublisher.Publish(ctx, &pkgEvents.Event{
		Type:    "supplier.created",
		Subject: fmt.Sprintf("catalog.suppliers.%s", supplier.ID.String()),
		Data: map[string]interface{}{
			"supplier_id":     supplier.ID.String(),
			"code":            supplier.Code,
			"name":            supplier.Name,
			"organization_id": supplier.OrganizationID.String(),
		},
	}); err != nil {
		uc.logger.Error(ctx, err, "Failed to publish supplier.created event")
	}

	uc.logger.Info(ctx, "Supplier created", logger.Tags{
		"supplier_id": supplier.ID.String(),
		"code":        supplier.Code,
	})

	return supplier, nil
}
```

Similar pattern for:
- `update_supplier.go`
- `delete_supplier.go`
- `get_supplier.go`
- `list_suppliers.go`
- `search_suppliers.go`

---

### Phase 4: Buffer Profile Use Cases (Day 7)

Similar implementation to Supplier use cases:
- Repository interface and implementation
- CreateBufferProfile, UpdateBufferProfile, DeleteBufferProfile, GetBufferProfile, ListBufferProfiles use cases

---

### Phase 5: gRPC Server Implementation (Days 8-9)

#### T007: Implement gRPC Handlers

**File**: `services/catalog-service/internal/infrastructure/grpc/handlers/catalog_handlers.go`

```go
package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalogpb "github.com/giia/giia-core-engine/services/catalog-service/api/proto/catalog/v1"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/product"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/usecases/supplier"
	"github.com/google/uuid"
)

type CatalogServiceHandler struct {
	catalogpb.UnimplementedCatalogServiceServer
	getProductUC     *product.GetProductUseCase
	listProductsUC   *product.ListProductsUseCase
	getSupplierUC    *supplier.GetSupplierUseCase
	listSuppliersUC  *supplier.ListSuppliersUseCase
}

func NewCatalogServiceHandler(
	getProductUC *product.GetProductUseCase,
	listProductsUC *product.ListProductsUseCase,
	getSupplierUC *supplier.GetSupplierUseCase,
	listSuppliersUC *supplier.ListSuppliersUseCase,
) *CatalogServiceHandler {
	return &CatalogServiceHandler{
		getProductUC:    getProductUC,
		listProductsUC:  listProductsUC,
		getSupplierUC:   getSupplierUC,
		listSuppliersUC: listSuppliersUC,
	}
}

func (h *CatalogServiceHandler) GetProduct(ctx context.Context, req *catalogpb.GetProductRequest) (*catalogpb.GetProductResponse, error) {
	// Get organization_id from context (set by auth interceptor)
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "organization_id not found in context")
	}

	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product ID format")
	}

	product, err := h.getProductUC.Execute(ctx, productID, orgID)
	if err != nil {
		// Map domain errors to gRPC status codes
		return nil, mapErrorToGRPCStatus(err)
	}

	return &catalogpb.GetProductResponse{
		Product: &catalogpb.Product{
			Id:             product.ID.String(),
			OrganizationId: product.OrganizationID.String(),
			Sku:            product.SKU,
			Name:           product.Name,
			Description:    product.Description,
			Category:       product.Category,
			UnitOfMeasure:  product.UnitOfMeasure,
			StandardCost:   product.StandardCost,
			Status:         string(product.Status),
			CreatedAt:      timestamppb.New(product.CreatedAt),
			UpdatedAt:      timestamppb.New(product.UpdatedAt),
		},
	}, nil
}

func (h *CatalogServiceHandler) ListProducts(ctx context.Context, req *catalogpb.ListProductsRequest) (*catalogpb.ListProductsResponse, error) {
	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "organization_id not found in context")
	}

	products, total, err := h.listProductsUC.Execute(ctx, product.ListProductsInput{
		OrganizationID: orgID,
		Status:         req.Status,
		Page:           int(req.Page),
		PageSize:       int(req.PageSize),
	})
	if err != nil {
		return nil, mapErrorToGRPCStatus(err)
	}

	pbProducts := make([]*catalogpb.Product, len(products))
	for i, p := range products {
		pbProducts[i] = &catalogpb.Product{
			Id:             p.ID.String(),
			OrganizationId: p.OrganizationID.String(),
			Sku:            p.SKU,
			Name:           p.Name,
			Description:    p.Description,
			Category:       p.Category,
			Status:         string(p.Status),
			CreatedAt:      timestamppb.New(p.CreatedAt),
			UpdatedAt:      timestamppb.New(p.UpdatedAt),
		}
	}

	return &catalogpb.ListProductsResponse{
		Products: pbProducts,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Similar implementations for GetSupplier, ListSuppliers, etc.

func mapErrorToGRPCStatus(err error) error {
	// Map pkg/errors types to gRPC status codes
	switch err.(type) {
	case *errors.BadRequestError:
		return status.Error(codes.InvalidArgument, err.Error())
	case *errors.ResourceNotFoundError:
		return status.Error(codes.NotFound, err.Error())
	case *errors.UnauthorizedRequestError:
		return status.Error(codes.Unauthenticated, err.Error())
	case *errors.ForbiddenRequestError:
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
```

#### T008: Initialize gRPC Server

**File**: `services/catalog-service/cmd/server/main.go` (modifications)

```go
// Add gRPC server initialization
grpcServer := setupGRPCServer(useCases, authClient, logger)
grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
if err != nil {
	logger.Error(context.Background(), err, "Failed to create gRPC listener")
	os.Exit(1)
}

// Start gRPC server in goroutine
go func() {
	logger.Info(context.Background(), fmt.Sprintf("Starting gRPC server on port %s", cfg.GRPCPort))
	if err := grpcServer.Serve(grpcListener); err != nil {
		logger.Error(context.Background(), err, "gRPC server error")
	}
}()

// Include gRPC server in graceful shutdown
defer grpcServer.GracefulStop()
```

---

### Phase 6: REST API Handlers (Day 10)

HTTP handlers for Supplier and BufferProfile endpoints following existing Product handlers pattern.

---

### Phase 7: Testing (Days 11-14)

#### Unit Tests
- All use cases with 80%+ coverage
- Repository methods with mock database
- Domain entity validation

#### Integration Tests
- Database operations with real PostgreSQL (Docker)
- gRPC client-server communication
- Auth service integration

#### Performance Tests
```go
func BenchmarkGetProductGRPC(b *testing.B) {
	// Setup gRPC server and client
	// Target: <50ms p50
}
```

---

## 4. Testing Strategy

### Unit Tests (80%+ Coverage)
```bash
go test ./services/catalog-service/internal/core/usecases/... -v -coverprofile=coverage.out
go test ./services/catalog-service/internal/infrastructure/repositories/... -v
```

### Integration Tests
```bash
docker-compose up -d postgres
go test ./services/catalog-service/test/integration/... -v -tags=integration
```

### Performance Tests
```bash
go test ./services/catalog-service/test/performance/... -bench=. -benchtime=10s
```

---

## 5. Dependencies and Execution Order

```
T001 (Database Migrations) → T002 (Domain Entities) → Phase 1 Complete
                                     ↓
T003 (Auth gRPC Client) → T004 (Auth Middleware) → Phase 2 Complete
                                     ↓
T005 (Supplier Repo) → T006 (Supplier Use Cases) → Phase 3 Complete
                                     ↓
T007 (gRPC Handlers) → T008 (gRPC Server) → Phase 5 Complete
                                     ↓
Testing (Phase 7) → Catalog Service 100% Complete
```

---

## 6. Acceptance Checklist

### gRPC Implementation
- [ ] catalog.proto defined with all RPCs
- [ ] Proto files generated successfully
- [ ] gRPC server running on dedicated port
- [ ] All RPCs implemented and tested
- [ ] gRPC reflection enabled

### Auth Integration
- [ ] gRPC client connects to Auth service
- [ ] Token validation working
- [ ] Permission checks working
- [ ] Auth middleware applied to all endpoints
- [ ] gRPC interceptor validates tokens

### Supplier Management
- [ ] Create supplier use case
- [ ] Update supplier use case
- [ ] Delete supplier (soft delete)
- [ ] Get supplier by ID
- [ ] List suppliers with pagination
- [ ] Search suppliers by name/code
- [ ] Supplier events published

### Buffer Profile Management
- [ ] Create buffer profile
- [ ] Update buffer profile
- [ ] Delete buffer profile
- [ ] Get buffer profile
- [ ] List buffer profiles
- [ ] Assign to products

### Testing
- [ ] Unit tests 80%+ coverage
- [ ] Integration tests pass
- [ ] gRPC integration tests pass
- [ ] Performance benchmarks meet targets
- [ ] All linters pass

### Production Readiness
- [ ] Multi-tenancy enforced
- [ ] Error handling complete
- [ ] Logging comprehensive
- [ ] Metrics exposed
- [ ] Documentation complete

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
**Estimated Completion**: 2 weeks
**Next Step**: Begin Phase 1 - Database Schema and Domain Entities
