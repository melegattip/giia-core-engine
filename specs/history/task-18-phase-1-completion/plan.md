# Task 18: Phase 1 Completion - Implementation Plan

**Task ID**: task-18-phase-1-completion
**Phase**: 1 - Foundation to 100%
**Priority**: P1 (High - Complete Foundation)
**Estimated Duration**: 1 week
**Dependencies**: All Phase 1 tasks (1-10)

---

## 1. Technical Context

### Current State
- **Auth Service**: 95% complete (missing gRPC interceptors, token refresh)
- **Catalog Service**: 85% complete (missing gRPC server, full testing)
- **Integration Testing**: 75% complete (missing end-to-end scenarios)
- **Infrastructure**: Missing Kubernetes manifests, Docker Compose

### Technology Stack
- **Language**: Go 1.23.4
- **Infrastructure**: Kubernetes, Docker, Docker Compose
- **Monitoring**: Prometheus, Grafana
- **CI/CD**: GitHub Actions
- **Documentation**: Markdown, OpenAPI 3.0

### Key Design Decisions
- **gRPC-first**: All inter-service communication via gRPC
- **JWT Authentication**: Token-based auth with refresh mechanism
- **Kubernetes Native**: Deploy to Kubernetes for production
- **Local Dev with Docker Compose**: Easy local environment setup

---

## 2. Implementation Steps

### Phase 1: Auth Service Completion (Days 1-2)

#### T001: gRPC Authentication Interceptors

**File**: `services/auth-service/internal/infrastructure/middleware/grpc_auth_interceptor.go`

```go
package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"giia-core-engine/services/auth-service/internal/core/providers"
)

type AuthInterceptor struct {
	jwtService providers.JWTService
}

func NewAuthInterceptor(jwtService providers.JWTService) *AuthInterceptor {
	return &AuthInterceptor{
		jwtService: jwtService,
	}
}

// UnaryInterceptor validates JWT tokens in unary gRPC calls
func (i *AuthInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public methods
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := extractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "missing authentication token")
		}

		// Validate token
		claims, err := i.jwtService.ValidateToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "organization_id", claims.OrganizationID)
		ctx = context.WithValue(ctx, "roles", claims.Roles)

		return handler(ctx, req)
	}
}

// StreamInterceptor validates JWT tokens in streaming gRPC calls
func (i *AuthInterceptor) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Skip auth for public methods
		if isPublicMethod(info.FullMethod) {
			return handler(srv, ss)
		}

		// Extract token from metadata
		token, err := extractToken(ss.Context())
		if err != nil {
			return status.Error(codes.Unauthenticated, "missing authentication token")
		}

		// Validate token
		claims, err := i.jwtService.ValidateToken(token)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token")
		}

		// Wrap stream with authenticated context
		wrappedStream := &authenticatedStream{
			ServerStream: ss,
			ctx:          context.WithValue(ss.Context(), "user_id", claims.UserID),
		}

		return handler(srv, wrappedStream)
	}
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	token := values[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return "", status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	return strings.TrimPrefix(token, "Bearer "), nil
}

func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/auth.v1.AuthService/Register": true,
		"/auth.v1.AuthService/Login":    true,
		"/auth.v1.AuthService/RefreshToken": true,
	}
	return publicMethods[method]
}

type authenticatedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *authenticatedStream) Context() context.Context {
	return s.ctx
}
```

#### T002: Token Refresh Mechanism

**File**: `services/auth-service/internal/core/usecases/auth/refresh_token.go`

```go
package auth

import (
	"context"
	"time"

	"giia-core-engine/services/auth-service/internal/core/domain"
	"giia-core-engine/services/auth-service/internal/core/providers"
)

type RefreshTokenUseCase struct {
	jwtService providers.JWTService
	userRepo   providers.UserRepository
}

func NewRefreshTokenUseCase(
	jwtService providers.JWTService,
	userRepo providers.UserRepository,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
}

type RefreshTokenInput struct {
	RefreshToken string
}

type RefreshTokenOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, input RefreshTokenInput) (*RefreshTokenOutput, error) {
	// 1. Validate refresh token
	claims, err := uc.jwtService.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.NewUnauthorizedError("invalid refresh token")
	}

	// 2. Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.NewUnauthorizedError("user not found")
	}

	if !user.IsActive() {
		return nil, domain.NewUnauthorizedError("user is inactive")
	}

	// 3. Generate new access token
	accessToken, err := uc.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// 4. Generate new refresh token
	newRefreshToken, err := uc.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(15 * time.Minute.Seconds()), // 15 minutes
	}, nil
}
```

---

### Phase 2: Catalog Service gRPC Server (Days 2-3)

#### T003: Protocol Buffers Definition

**File**: `services/catalog-service/api/proto/catalog/v1/catalog.proto`

```protobuf
syntax = "proto3";

package catalog.v1;

option go_package = "giia-core-engine/services/catalog-service/api/proto/catalog/v1;catalogv1";

import "google/protobuf/timestamp.proto";

service CatalogService {
  // Product operations
  rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
  rpc UpdateProduct(UpdateProductRequest) returns (UpdateProductResponse);
  rpc GetProduct(GetProductRequest) returns (GetProductResponse);
  rpc ListProducts(ListProductsRequest) returns (ListProductsResponse);
  rpc DeleteProduct(DeleteProductRequest) returns (DeleteProductResponse);

  // Supplier operations
  rpc CreateSupplier(CreateSupplierRequest) returns (CreateSupplierResponse);
  rpc UpdateSupplier(UpdateSupplierRequest) returns (UpdateSupplierResponse);
  rpc GetSupplier(GetSupplierRequest) returns (GetSupplierResponse);
  rpc ListSuppliers(ListSuppliersRequest) returns (ListSuppliersResponse);

  // Buffer Profile operations
  rpc CreateBufferProfile(CreateBufferProfileRequest) returns (CreateBufferProfileResponse);
  rpc GetBufferProfile(GetBufferProfileRequest) returns (GetBufferProfileResponse);
  rpc ListBufferProfiles(ListBufferProfilesRequest) returns (ListBufferProfilesResponse);

  // Product-Supplier association
  rpc AssociateSupplier(AssociateSupplierRequest) returns (AssociateSupplierResponse);
  rpc GetProductSuppliers(GetProductSuppliersRequest) returns (GetProductSuppliersResponse);
}

message Product {
  string id = 1;
  string organization_id = 2;
  string sku = 3;
  string name = 4;
  string description = 5;
  string category = 6;
  string unit_of_measure = 7;
  double standard_cost = 8;
  google.protobuf.Timestamp last_purchase_date = 9;
  string buffer_profile_id = 10;
  string status = 11;
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}

message Supplier {
  string id = 1;
  string organization_id = 2;
  string code = 3;
  string name = 4;
  string contact_name = 5;
  string contact_email = 6;
  string contact_phone = 7;
  int32 default_lead_time = 8;
  string reliability = 9;
  string status = 10;
  string address = 11;
}

message BufferProfile {
  string id = 1;
  string organization_id = 2;
  string name = 3;
  string description = 4;
  string adu_method = 5;
  string lead_time_category = 6;
  string variability_category = 7;
  double lead_time_factor = 8;
  double variability_factor = 9;
  string status = 10;
}

message CreateProductRequest {
  string organization_id = 1;
  string sku = 2;
  string name = 3;
  string description = 4;
  string category = 5;
  string unit_of_measure = 6;
  double standard_cost = 7;
}

message CreateProductResponse {
  Product product = 1;
}

message GetProductRequest {
  string id = 1;
  string organization_id = 2;
}

message GetProductResponse {
  Product product = 1;
}

// ... (other messages)
```

#### T004: gRPC Server Implementation

**File**: `services/catalog-service/internal/infrastructure/entrypoints/grpc/server.go`

```go
package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	catalogv1 "giia-core-engine/services/catalog-service/api/proto/catalog/v1"
	"giia-core-engine/services/catalog-service/internal/core/usecases/product"
	"giia-core-engine/services/catalog-service/internal/core/usecases/supplier"
	"giia-core-engine/services/catalog-service/internal/core/usecases/buffer_profile"
)

type CatalogServer struct {
	catalogv1.UnimplementedCatalogServiceServer
	createProductUC       *product.CreateProductUseCase
	getProductUC          *product.GetProductUseCase
	listProductsUC        *product.ListProductsUseCase
	createSupplierUC      *supplier.CreateSupplierUseCase
	getSupplierUC         *supplier.GetSupplierUseCase
	createBufferProfileUC *buffer_profile.CreateBufferProfileUseCase
	getBufferProfileUC    *buffer_profile.GetBufferProfileUseCase
}

func NewCatalogServer(
	createProductUC *product.CreateProductUseCase,
	getProductUC *product.GetProductUseCase,
	listProductsUC *product.ListProductsUseCase,
	createSupplierUC *supplier.CreateSupplierUseCase,
	getSupplierUC *supplier.GetSupplierUseCase,
	createBufferProfileUC *buffer_profile.CreateBufferProfileUseCase,
	getBufferProfileUC *buffer_profile.GetBufferProfileUseCase,
) *CatalogServer {
	return &CatalogServer{
		createProductUC:       createProductUC,
		getProductUC:          getProductUC,
		listProductsUC:        listProductsUC,
		createSupplierUC:      createSupplierUC,
		getSupplierUC:         getSupplierUC,
		createBufferProfileUC: createBufferProfileUC,
		getBufferProfileUC:    getBufferProfileUC,
	}
}

func (s *CatalogServer) CreateProduct(ctx context.Context, req *catalogv1.CreateProductRequest) (*catalogv1.CreateProductResponse, error) {
	// Extract user context
	orgID := ctx.Value("organization_id").(string)

	// Execute use case
	input := product.CreateProductInput{
		OrganizationID: uuid.MustParse(orgID),
		SKU:            req.Sku,
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		UnitOfMeasure:  req.UnitOfMeasure,
		StandardCost:   req.StandardCost,
	}

	result, err := s.createProductUC.Execute(ctx, input)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to proto
	return &catalogv1.CreateProductResponse{
		Product: toProtoProduct(result),
	}, nil
}

func (s *CatalogServer) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	productID := uuid.MustParse(req.Id)
	orgID := uuid.MustParse(req.OrganizationId)

	result, err := s.getProductUC.Execute(ctx, product.GetProductInput{
		ProductID:      productID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &catalogv1.GetProductResponse{
		Product: toProtoProduct(result),
	}, nil
}

// ... (other methods)

func toProtoProduct(p *domain.Product) *catalogv1.Product {
	proto := &catalogv1.Product{
		Id:             p.ID.String(),
		OrganizationId: p.OrganizationID.String(),
		Sku:            p.SKU,
		Name:           p.Name,
		Description:    p.Description,
		Category:       p.Category,
		UnitOfMeasure:  p.UnitOfMeasure,
		StandardCost:   p.StandardCost,
		Status:         string(p.Status),
		CreatedAt:      timestamppb.New(p.CreatedAt),
		UpdatedAt:      timestamppb.New(p.UpdatedAt),
	}

	if p.LastPurchaseDate != nil {
		proto.LastPurchaseDate = timestamppb.New(*p.LastPurchaseDate)
	}

	if p.BufferProfileID != nil {
		proto.BufferProfileId = p.BufferProfileID.String()
	}

	return proto
}
```

---

### Phase 3: Infrastructure as Code (Days 3-4)

#### T005: Kubernetes Manifests

**File**: `k8s/auth-service/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: giia
  labels:
    app: auth-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: giia/auth-service:latest
        ports:
        - containerPort: 50051
          name: grpc
        - containerPort: 8080
          name: http
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-service-secrets
              key: jwt-secret
        - name: REDIS_URL
          valueFrom:
            configMapKeyRef:
              name: auth-service-config
              key: redis-url
        - name: NATS_URL
          valueFrom:
            configMapKeyRef:
              name: auth-service-config
              key: nats-url
        livenessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: giia
spec:
  selector:
    app: auth-service
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  - port: 8080
    targetPort: 8080
    name: http
  type: ClusterIP
```

**File**: `k8s/catalog-service/deployment.yaml`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: catalog-service
  namespace: giia
  labels:
    app: catalog-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: catalog-service
  template:
    metadata:
      labels:
        app: catalog-service
    spec:
      containers:
      - name: catalog-service
        image: giia/catalog-service:latest
        ports:
        - containerPort: 50052
          name: grpc
        - containerPort: 8081
          name: http
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: catalog-service-secrets
              key: database-url
        - name: AUTH_SERVICE_URL
          value: "auth-service:50051"
        - name: NATS_URL
          valueFrom:
            configMapKeyRef:
              name: catalog-service-config
              key: nats-url
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: catalog-service
  namespace: giia
spec:
  selector:
    app: catalog-service
  ports:
  - port: 50052
    targetPort: 50052
    name: grpc
  - port: 8081
    targetPort: 8081
    name: http
  type: ClusterIP
```

**File**: `k8s/infrastructure/postgres.yaml`

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: giia
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secrets
              key: password
        - name: POSTGRES_DB
          value: giia
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: giia
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
  clusterIP: None
```

#### T006: Docker Compose for Local Development

**File**: `docker-compose.yaml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: giia
      POSTGRES_USER: giia
      POSTGRES_PASSWORD: giia_local_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U giia"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  nats:
    image: nats:2-alpine
    ports:
      - "4222:4222"
      - "8222:8222"
    command: "-js -m 8222"
    healthcheck:
      test: ["CMD", "wget", "-q", "-O-", "http://localhost:8222/healthz"]
      interval: 10s
      timeout: 5s
      retries: 5

  auth-service:
    build:
      context: .
      dockerfile: services/auth-service/Dockerfile
    ports:
      - "50051:50051"
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://giia:giia_local_password@postgres:5432/giia?sslmode=disable
      REDIS_URL: redis:6379
      NATS_URL: nats://nats:4222
      JWT_SECRET: local_jwt_secret_key_change_in_production
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      nats:
        condition: service_healthy

  catalog-service:
    build:
      context: .
      dockerfile: services/catalog-service/Dockerfile
    ports:
      - "50052:50052"
      - "8081:8081"
    environment:
      DATABASE_URL: postgres://giia:giia_local_password@postgres:5432/giia?sslmode=disable
      AUTH_SERVICE_URL: auth-service:50051
      NATS_URL: nats://nats:4222
    depends_on:
      postgres:
        condition: service_healthy
      nats:
        condition: service_healthy
      auth-service:
        condition: service_started

volumes:
  postgres_data:
```

---

### Phase 4: Testing & Documentation (Days 4-5)

#### T007: Integration Tests

**File**: `tests/integration/auth_catalog_flow_test.go`

```go
package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	authv1 "giia-core-engine/services/auth-service/api/proto/auth/v1"
	catalogv1 "giia-core-engine/services/catalog-service/api/proto/catalog/v1"
)

func TestAuthCatalogFlow_CompleteScenario(t *testing.T) {
	// Setup
	authConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	assert.NoError(t, err)
	defer authConn.Close()

	catalogConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	assert.NoError(t, err)
	defer catalogConn.Close()

	authClient := authv1.NewAuthServiceClient(authConn)
	catalogClient := catalogv1.NewCatalogServiceClient(catalogConn)

	ctx := context.Background()

	// 1. Register user
	registerResp, err := authClient.Register(ctx, &authv1.RegisterRequest{
		Email:          "test@example.com",
		Password:       "SecurePass123!",
		OrganizationId: "org-123",
	})
	assert.NoError(t, err)
	assert.NotNil(t, registerResp.AccessToken)

	// 2. Create authenticated context with token
	authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+registerResp.AccessToken)

	// 3. Create product with authenticated context
	createProductResp, err := catalogClient.CreateProduct(authCtx, &catalogv1.CreateProductRequest{
		OrganizationId: "org-123",
		Sku:            "PROD-001",
		Name:           "Test Product",
		Category:       "Electronics",
		UnitOfMeasure:  "unit",
		StandardCost:   100.0,
	})
	assert.NoError(t, err)
	assert.NotNil(t, createProductResp.Product)
	assert.Equal(t, "PROD-001", createProductResp.Product.Sku)

	// 4. Get product
	getProductResp, err := catalogClient.GetProduct(authCtx, &catalogv1.GetProductRequest{
		Id:             createProductResp.Product.Id,
		OrganizationId: "org-123",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Test Product", getProductResp.Product.Name)
}
```

#### T008: API Documentation

**File**: `docs/API.md`

```markdown
# GIIA API Documentation

## Overview

GIIA provides both REST and gRPC APIs for all services. This document covers the main endpoints and usage patterns.

## Authentication

All API calls (except registration and login) require authentication via JWT token.

### Headers
```
Authorization: Bearer <access_token>
```

## Auth Service

### gRPC API

**Service**: `auth.v1.AuthService`
**Address**: `localhost:50051`

#### Register
```protobuf
rpc Register(RegisterRequest) returns (RegisterResponse);

message RegisterRequest {
  string email = 1;
  string password = 2;
  string organization_id = 3;
}

message RegisterResponse {
  string access_token = 1;
  string refresh_token = 2;
  int64 expires_in = 3;
}
```

#### Login
```protobuf
rpc Login(LoginRequest) returns (LoginResponse);
```

#### Refresh Token
```protobuf
rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
```

## Catalog Service

### gRPC API

**Service**: `catalog.v1.CatalogService`
**Address**: `localhost:50052`

#### Create Product
```protobuf
rpc CreateProduct(CreateProductRequest) returns (CreateProductResponse);
```

#### Get Product
```protobuf
rpc GetProduct(GetProductRequest) returns (GetProductResponse);
```

## Error Codes

| gRPC Code | Description |
|-----------|-------------|
| UNAUTHENTICATED | Missing or invalid authentication token |
| PERMISSION_DENIED | User lacks required permissions |
| NOT_FOUND | Resource not found |
| ALREADY_EXISTS | Resource already exists |
| INVALID_ARGUMENT | Invalid input parameters |
| INTERNAL | Internal server error |
```

---

## 3. Success Criteria

### Mandatory
- ✅ Auth service gRPC interceptors implemented and tested
- ✅ Token refresh mechanism working
- ✅ Catalog service gRPC server complete
- ✅ 90%+ test coverage on Auth and Catalog services
- ✅ Integration tests passing
- ✅ Kubernetes manifests deployable
- ✅ Docker Compose local environment functional
- ✅ API documentation complete

### Verification Checklist
- [ ] `make test` passes with 90%+ coverage
- [ ] `docker-compose up` starts all services
- [ ] Integration tests pass against Docker Compose
- [ ] Kubernetes deployment succeeds: `kubectl apply -k k8s/`
- [ ] gRPC health checks pass in Kubernetes
- [ ] API documentation accurate and complete

---

## 4. Testing Strategy

### Unit Tests
- Auth interceptor validation logic
- Token refresh use case
- gRPC server handlers
- Coverage: 90%+

### Integration Tests
- End-to-end user registration → product creation flow
- Cross-service authentication
- Event publishing and consumption
- gRPC client-server communication

### Manual Testing
- Test with grpcurl
- Test with Postman (gRPC support)
- Kubernetes deployment verification

---

## 5. Deployment Commands

### Local Development
```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f auth-service

# Run migrations
docker-compose exec auth-service /app/migrate up

# Stop all services
docker-compose down
```

### Kubernetes
```bash
# Create namespace
kubectl create namespace giia

# Apply secrets
kubectl apply -f k8s/secrets/

# Deploy infrastructure
kubectl apply -f k8s/infrastructure/

# Deploy services
kubectl apply -f k8s/auth-service/
kubectl apply -f k8s/catalog-service/

# Check status
kubectl get pods -n giia
kubectl logs -f -n giia deployment/auth-service
```

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
