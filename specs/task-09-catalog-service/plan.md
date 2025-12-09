# Implementation Plan: Catalog Service with Clean Architecture

**Created**: 2025-12-09
**Spec**: [spec.md](./spec.md)

## Technical Context

### Technologies & Tools
- **Go**: 1.23.4 with Clean Architecture principles
- **Database**: PostgreSQL 16 with GORM ORM, pgx driver
- **HTTP Router**: Chi router v5 (lightweight, middleware-friendly)
- **API Documentation**: swaggo/swag for OpenAPI/Swagger generation
- **Validation**: go-playground/validator v10 for input validation
- **Auth Integration**: gRPC client to Auth service for token validation and permission checks
- **Event Publishing**: NATS Jetstream via pkg/events for domain events
- **Shared Packages**: pkg/config, pkg/logger, pkg/database, pkg/errors, pkg/events

### Design Decisions
- **Clean Architecture layers**: Domain (entities) → Use Cases (business logic) → Adapters (repositories) → Infrastructure (HTTP, DB, gRPC)
- **Repository pattern**: Abstract data access behind interfaces
- **Dependency injection**: Constructor injection for all dependencies
- **Multi-tenancy**: Automatic organization_id filtering via GORM scope
- **Permission enforcement**: Middleware calls Auth gRPC CheckPermission before handling requests
- **Pagination**: Cursor-based pagination for scalability (offset-based as fallback)
- **Soft delete**: Mark records as deleted (status='inactive') instead of hard delete
- **Event sourcing**: Publish domain events (product.created, supplier.updated) to NATS
- **API versioning**: /api/v1/ prefix for all endpoints

## Project Structure

```
services/catalog-service/
├── cmd/
│   └── api/
│       └── main.go                      # Entry point (NEW)
│
├── internal/
│   ├── core/
│   │   ├── domain/                      # Domain entities (NEW)
│   │   │   ├── product.go              # Product entity
│   │   │   ├── supplier.go             # Supplier entity
│   │   │   ├── product_supplier.go     # Many-to-many relationship
│   │   │   ├── buffer_profile.go       # Buffer profile template
│   │   │   └── category.go             # Product category (optional)
│   │   │
│   │   ├── usecases/                    # Business logic (NEW)
│   │   │   ├── product/
│   │   │   │   ├── create_product.go
│   │   │   │   ├── get_product.go
│   │   │   │   ├── update_product.go
│   │   │   │   ├── delete_product.go
│   │   │   │   ├── list_products.go
│   │   │   │   └── search_products.go
│   │   │   │
│   │   │   ├── supplier/
│   │   │   │   ├── create_supplier.go
│   │   │   │   ├── get_supplier.go
│   │   │   │   ├── update_supplier.go
│   │   │   │   ├── delete_supplier.go
│   │   │   │   └── list_suppliers.go
│   │   │   │
│   │   │   └── buffer_profile/
│   │   │       ├── create_profile.go
│   │   │       ├── get_profile.go
│   │   │       ├── update_profile.go
│   │   │       ├── delete_profile.go
│   │   │       └── assign_to_product.go
│   │   │
│   │   └── providers/                   # Interfaces (NEW)
│   │       ├── product_repository.go
│   │       ├── supplier_repository.go
│   │       ├── buffer_profile_repository.go
│   │       ├── auth_service.go         # gRPC Auth client interface
│   │       └── event_publisher.go      # Event publishing interface
│   │
│   └── infrastructure/
│       ├── adapters/
│       │   ├── auth/
│       │   │   ├── grpc_auth_client.go    # Auth gRPC client wrapper (NEW)
│       │   │   └── auth_client_test.go    # (NEW)
│       │   │
│       │   └── events/
│       │       ├── catalog_publisher.go    # Catalog event publisher (NEW)
│       │       └── publisher_test.go       # (NEW)
│       │
│       ├── repositories/
│       │   ├── product_repository.go       # GORM implementation (NEW)
│       │   ├── supplier_repository.go      # GORM implementation (NEW)
│       │   └── buffer_profile_repository.go # GORM implementation (NEW)
│       │
│       ├── entrypoints/
│       │   └── http/
│       │       ├── handlers/
│       │       │   ├── product_handler.go      # Product HTTP handlers (NEW)
│       │       │   ├── supplier_handler.go     # Supplier HTTP handlers (NEW)
│       │       │   └── buffer_profile_handler.go # Buffer profile handlers (NEW)
│       │       │
│       │       ├── middleware/
│       │       │   ├── auth.go                 # JWT validation middleware (NEW)
│       │       │   ├── permission.go           # Permission check middleware (NEW)
│       │       │   ├── tenant.go               # Tenant isolation middleware (NEW)
│       │       │   └── logging.go              # Request logging middleware (NEW)
│       │       │
│       │       ├── dto/
│       │       │   ├── product_dto.go          # Request/response DTOs (NEW)
│       │       │   ├── supplier_dto.go         # (NEW)
│       │       │   └── buffer_profile_dto.go   # (NEW)
│       │       │
│       │       └── router.go                   # HTTP router setup (NEW)
│       │
│       └── persistence/
│           └── migrations/
│               ├── 001_create_products.sql         # (NEW)
│               ├── 002_create_suppliers.sql        # (NEW)
│               ├── 003_create_product_suppliers.sql # (NEW)
│               ├── 004_create_buffer_profiles.sql  # (NEW)
│               └── 005_create_categories.sql       # (NEW)
│
├── docs/
│   ├── API.md                          # API documentation (NEW)
│   ├── ARCHITECTURE.md                 # Clean Architecture explanation (NEW)
│   └── swagger.yaml                    # OpenAPI spec (auto-generated)
│
├── .env.example                        # Environment variables (NEW)
├── Dockerfile                          # Docker build (NEW)
├── go.mod                             # Dependencies (NEW)
└── README.md                           # Service documentation (NEW)
```

## Phased Implementation

### Phase 1: Setup (Foundation)
**Goal**: Initialize service structure and dependencies

- **T001** [Setup]: Create service directory `services/catalog-service/`
- **T002** [Setup]: Initialize `go.mod`:
  ```bash
  cd services/catalog-service
  go mod init github.com/giia/catalog-service
  ```
- **T003** [Setup]: Add dependencies to `go.mod`:
  ```
  gorm.io/gorm v1.25.5
  gorm.io/driver/postgres v1.5.4
  github.com/go-chi/chi/v5 v5.0.11
  github.com/go-playground/validator/v10 v10.16.0
  github.com/swaggo/swag v1.16.2
  github.com/swaggo/http-swagger v1.3.4
  ```
- **T004** [Setup]: Add workspace reference to root `go.work`:
  ```
  use ./services/catalog-service
  ```
- **T005** [Setup]: Create `.env.example`:
  ```
  PORT=8082
  DATABASE_URL=postgresql://giia:giia_dev_password@localhost:5432/giia_dev?search_path=catalog
  REDIS_URL=redis://:giia_redis_password@localhost:6379/1
  NATS_URL=nats://localhost:4222
  AUTH_GRPC_URL=localhost:9091
  LOG_LEVEL=debug
  SERVICE_NAME=catalog-service
  ```
- **T006** [Setup]: Create directory structure (domain, usecases, providers, repositories, handlers)
- **T007** [Setup]: Run `go mod tidy` to download dependencies

**Checkpoint**: Service structure created, dependencies installed

---

### Phase 2: Foundational - Database Schema (Blocking Prerequisites)
**Goal**: Define database tables and domain entities

- **T008** [Foundational]: Create migration `001_create_products.sql`:
  ```sql
  CREATE SCHEMA IF NOT EXISTS catalog;
  SET search_path TO catalog;

  CREATE TABLE products (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      sku VARCHAR(100) NOT NULL,
      name VARCHAR(255) NOT NULL,
      description TEXT,
      category VARCHAR(100),
      unit_of_measure VARCHAR(50) NOT NULL,
      status VARCHAR(20) NOT NULL DEFAULT 'active',
      organization_id UUID NOT NULL,
      buffer_profile_id UUID,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
      UNIQUE(sku, organization_id)
  );

  CREATE INDEX idx_products_organization_id ON products(organization_id);
  CREATE INDEX idx_products_sku ON products(sku);
  CREATE INDEX idx_products_status ON products(status);
  CREATE INDEX idx_products_category ON products(category);
  ```
- **T009** [Foundational]: Create migration `002_create_suppliers.sql`:
  ```sql
  CREATE TABLE suppliers (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      code VARCHAR(100) NOT NULL,
      name VARCHAR(255) NOT NULL,
      lead_time_days INT NOT NULL DEFAULT 0,
      reliability_rating INT DEFAULT 80,
      contact_info JSONB,
      status VARCHAR(20) NOT NULL DEFAULT 'active',
      organization_id UUID NOT NULL,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
      UNIQUE(code, organization_id)
  );

  CREATE INDEX idx_suppliers_organization_id ON suppliers(organization_id);
  CREATE INDEX idx_suppliers_code ON suppliers(code);
  ```
- **T010** [Foundational]: Create migration `003_create_product_suppliers.sql`:
  ```sql
  CREATE TABLE product_suppliers (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
      supplier_id UUID NOT NULL REFERENCES suppliers(id) ON DELETE CASCADE,
      lead_time_days INT NOT NULL,
      unit_cost DECIMAL(12, 2),
      is_primary_supplier BOOLEAN NOT NULL DEFAULT false,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      UNIQUE(product_id, supplier_id)
  );

  CREATE INDEX idx_product_suppliers_product_id ON product_suppliers(product_id);
  CREATE INDEX idx_product_suppliers_supplier_id ON product_suppliers(supplier_id);
  ```
- **T011** [Foundational]: Create migration `004_create_buffer_profiles.sql`:
  ```sql
  CREATE TABLE buffer_profiles (
      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      name VARCHAR(100) NOT NULL,
      description TEXT,
      lead_time_factor DECIMAL(5, 2) NOT NULL DEFAULT 1.0,
      variability_factor DECIMAL(5, 2) NOT NULL DEFAULT 1.0,
      target_service_level INT NOT NULL DEFAULT 95,
      organization_id UUID NOT NULL,
      created_at TIMESTAMP NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
      UNIQUE(name, organization_id)
  );

  CREATE INDEX idx_buffer_profiles_organization_id ON buffer_profiles(organization_id);
  ```
- **T012** [Foundational]: Run migrations on local PostgreSQL
- **T013** [Foundational]: Create domain entity `internal/core/domain/product.go`:
  ```go
  package domain

  import "time"

  type Product struct {
      ID               string
      SKU              string
      Name             string
      Description      string
      Category         string
      UnitOfMeasure    string
      Status           string
      OrganizationID   string
      BufferProfileID  *string
      Suppliers        []ProductSupplier
      CreatedAt        time.Time
      UpdatedAt        time.Time
  }

  func (p *Product) Validate() error {
      if p.SKU == "" {
          return errors.NewBadRequest("SKU is required")
      }
      if p.Name == "" {
          return errors.NewBadRequest("name is required")
      }
      if p.UnitOfMeasure == "" {
          return errors.NewBadRequest("unit of measure is required")
      }
      return nil
  }

  func (p *Product) IsActive() bool {
      return p.Status == "active"
  }
  ```
- **T014** [Foundational]: Create domain entities for Supplier, ProductSupplier, BufferProfile
- **T015** [Foundational]: Define repository interfaces in `internal/core/providers/`

**Checkpoint**: Database schema created, domain entities defined

---

### Phase 3: User Story 1 - Product Master Data Management (P1)
**Goal**: Implement CRUD operations for products

- **T016** [P][US1]: Write test for product repository Create
- **T017** [US1]: Implement `internal/infrastructure/repositories/product_repository.go`:
  ```go
  type GORMProductRepository struct {
      db *gorm.DB
  }

  func (r *GORMProductRepository) Create(ctx context.Context, product *domain.Product) error {
      // Extract organization_id from context
      orgID := ctx.Value("organization_id").(string)
      product.OrganizationID = orgID

      return r.db.WithContext(ctx).Create(product).Error
  }

  func (r *GORMProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
      var product domain.Product
      err := r.db.WithContext(ctx).
          Scopes(r.tenantScope(ctx)).
          Where("id = ?", id).
          First(&product).Error

      if err == gorm.ErrRecordNotFound {
          return nil, errors.NewNotFound("product not found")
      }
      return &product, err
  }

  func (r *GORMProductRepository) tenantScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
      return func(db *gorm.DB) *gorm.DB {
          orgID := ctx.Value("organization_id").(string)
          return db.Where("organization_id = ?", orgID)
      }
  }
  ```
- **T018** [P][US1]: Write test for create product use case
- **T019** [US1]: Implement `internal/core/usecases/product/create_product.go`:
  ```go
  type CreateProductUseCase struct {
      productRepo   providers.ProductRepository
      eventPublisher providers.EventPublisher
      logger        logger.Logger
  }

  func (uc *CreateProductUseCase) Execute(ctx context.Context, req *CreateProductRequest) (*domain.Product, error) {
      // Validate request
      if err := req.Validate(); err != nil {
          return nil, err
      }

      // Create product entity
      product := &domain.Product{
          SKU:           req.SKU,
          Name:          req.Name,
          Description:   req.Description,
          Category:      req.Category,
          UnitOfMeasure: req.UnitOfMeasure,
          Status:        "active",
      }

      // Validate domain rules
      if err := product.Validate(); err != nil {
          return nil, err
      }

      // Persist to database
      if err := uc.productRepo.Create(ctx, product); err != nil {
          return nil, errors.NewInternalServerError("failed to create product")
      }

      // Publish domain event
      uc.publishProductCreated(ctx, product)

      return product, nil
  }
  ```
- **T020** [P][US1]: Write test for list products with pagination
- **T021** [US1]: Implement `internal/core/usecases/product/list_products.go`:
  ```go
  type ListProductsRequest struct {
      Page     int    // Default 1
      PageSize int    // Default 20, max 100
      Category string // Optional filter
      Status   string // Optional filter
  }

  func (uc *ListProductsUseCase) Execute(ctx context.Context, req *ListProductsRequest) (*PaginatedProductsResponse, error) {
      // Apply defaults
      if req.Page < 1 {
          req.Page = 1
      }
      if req.PageSize < 1 || req.PageSize > 100 {
          req.PageSize = 20
      }

      // Build filters
      filters := make(map[string]interface{})
      if req.Category != "" {
          filters["category"] = req.Category
      }
      if req.Status != "" {
          filters["status"] = req.Status
      }

      // Query repository
      products, total, err := uc.productRepo.List(ctx, filters, req.Page, req.PageSize)
      if err != nil {
          return nil, err
      }

      return &PaginatedProductsResponse{
          Products:   products,
          Page:       req.Page,
          PageSize:   req.PageSize,
          TotalCount: total,
          TotalPages: (total + req.PageSize - 1) / req.PageSize,
      }, nil
  }
  ```
- **T022** [US1]: Implement update, delete, get use cases
- **T023** [P][US1]: Write test for HTTP handler POST /api/v1/products
- **T024** [US1]: Implement `internal/infrastructure/entrypoints/http/handlers/product_handler.go`:
  ```go
  type ProductHandler struct {
      createUC usecases.CreateProductUseCase
      listUC   usecases.ListProductsUseCase
      getUC    usecases.GetProductUseCase
      updateUC usecases.UpdateProductUseCase
      deleteUC usecases.DeleteProductUseCase
      logger   logger.Logger
  }

  // @Summary Create product
  // @Tags products
  // @Accept json
  // @Produce json
  // @Param request body CreateProductRequest true "Product data"
  // @Success 201 {object} ProductResponse
  // @Failure 400 {object} ErrorResponse
  // @Router /api/v1/products [post]
  func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
      var req dto.CreateProductRequest
      if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
          h.respondError(w, errors.NewBadRequest("invalid request body"))
          return
      }

      product, err := h.createUC.Execute(r.Context(), &req)
      if err != nil {
          h.respondError(w, err)
          return
      }

      h.respondJSON(w, http.StatusCreated, product)
  }
  ```
- **T025** [US1]: Implement remaining HTTP handlers (GET, PUT, DELETE, LIST)
- **T026** [P][US1]: Write integration test: Full CRUD flow via HTTP API

**Checkpoint**: Product CRUD operations work via REST API

---

### Phase 4: User Story 2 - Supplier Management (P1)
**Goal**: Implement CRUD operations for suppliers and product-supplier relationships

- **T027** [P][US2]: Write test for supplier repository
- **T028** [US2]: Implement `internal/infrastructure/repositories/supplier_repository.go` (similar to product)
- **T029** [P][US2]: Write test for create supplier use case
- **T030** [US2]: Implement supplier use cases (create, get, update, delete, list)
- **T031** [P][US2]: Write test for product-supplier association
- **T032** [US2]: Implement `AssociateSupplier` use case:
  ```go
  func (uc *AssociateSupplierUseCase) Execute(ctx context.Context, req *AssociateSupplierRequest) error {
      // Validate product exists
      product, err := uc.productRepo.GetByID(ctx, req.ProductID)
      if err != nil {
          return err
      }

      // Validate supplier exists
      supplier, err := uc.supplierRepo.GetByID(ctx, req.SupplierID)
      if err != nil {
          return err
      }

      // Create association
      association := &domain.ProductSupplier{
          ProductID:         req.ProductID,
          SupplierID:        req.SupplierID,
          LeadTimeDays:      req.LeadTimeDays,
          UnitCost:          req.UnitCost,
          IsPrimarySupplier: req.IsPrimarySupplier,
      }

      return uc.productRepo.AssociateSupplier(ctx, association)
  }
  ```
- **T033** [US2]: Implement supplier HTTP handlers
- **T034** [P][US2]: Write test for GET /api/v1/products/:id?include=suppliers
- **T035** [US2]: Implement eager loading of suppliers in GetProduct:
  ```go
  func (r *GORMProductRepository) GetByIDWithSuppliers(ctx context.Context, id string) (*domain.Product, error) {
      var product domain.Product
      err := r.db.WithContext(ctx).
          Scopes(r.tenantScope(ctx)).
          Preload("Suppliers").
          Preload("Suppliers.Supplier").
          Where("id = ?", id).
          First(&product).Error

      if err == gorm.ErrRecordNotFound {
          return nil, errors.NewNotFound("product not found")
      }
      return &product, err
  }
  ```
- **T036** [P][US2]: Write integration test: Create product, create supplier, associate them, query

**Checkpoint**: Supplier management and product-supplier relationships work

---

### Phase 5: User Story 3 - Buffer Profile Templates (P2)
**Goal**: Implement buffer profile CRUD and assignment to products

- **T037** [P][US3]: Write test for buffer profile repository
- **T038** [US3]: Implement `internal/infrastructure/repositories/buffer_profile_repository.go`
- **T039** [P][US3]: Write test for create buffer profile use case
- **T040** [US3]: Implement buffer profile use cases (create, get, update, delete, list)
- **T041** [P][US3]: Write test for assign profile to product
- **T042** [US3]: Implement `AssignBufferProfileUseCase`:
  ```go
  func (uc *AssignBufferProfileUseCase) Execute(ctx context.Context, productID, profileID string) error {
      // Validate product exists
      product, err := uc.productRepo.GetByID(ctx, productID)
      if err != nil {
          return err
      }

      // Validate profile exists and belongs to same organization
      profile, err := uc.profileRepo.GetByID(ctx, profileID)
      if err != nil {
          return err
      }

      // Update product's buffer_profile_id
      product.BufferProfileID = &profileID
      if err := uc.productRepo.Update(ctx, product); err != nil {
          return errors.NewInternalServerError("failed to assign profile")
      }

      // Publish event
      uc.eventPublisher.PublishBufferProfileAssigned(ctx, product, profile)

      return nil
  }
  ```
- **T043** [P][US3]: Write test for bulk assignment
- **T044** [US3]: Implement `AssignProfileToMultipleProducts` use case (batch operation)
- **T045** [US3]: Implement buffer profile HTTP handlers
- **T046** [P][US3]: Write integration test: Create profile, assign to 10 products, verify all updated

**Checkpoint**: Buffer profile templates work and can be assigned to products

---

### Phase 6: Authentication and Authorization Integration
**Goal**: Integrate with Auth service via gRPC

- **T047** [P]: Write test for gRPC Auth client wrapper
- **T048**: Implement `internal/infrastructure/adapters/auth/grpc_auth_client.go`:
  ```go
  type GRPCAuthClient struct {
      conn   *grpc.ClientConn
      client authpb.AuthServiceClient
      logger logger.Logger
  }

  func NewGRPCAuthClient(authURL string, logger logger.Logger) (*GRPCAuthClient, error) {
      conn, err := grpc.Dial(authURL, grpc.WithInsecure())
      if err != nil {
          return nil, err
      }

      return &GRPCAuthClient{
          conn:   conn,
          client: authpb.NewAuthServiceClient(conn),
          logger: logger,
      }, nil
  }

  func (c *GRPCAuthClient) ValidateToken(ctx context.Context, token string) (*domain.UserInfo, error) {
      resp, err := c.client.ValidateToken(ctx, &authpb.ValidateTokenRequest{
          Token: token,
      })
      if err != nil {
          return nil, errors.NewUnauthorized("invalid token")
      }

      return &domain.UserInfo{
          UserID:         resp.User.UserId,
          OrganizationID: resp.User.OrganizationId,
          Email:          resp.User.Email,
          Roles:          resp.User.Roles,
      }, nil
  }

  func (c *GRPCAuthClient) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
      resp, err := c.client.CheckPermission(ctx, &authpb.CheckPermissionRequest{
          UserId:     userID,
          Permission: permission,
      })
      if err != nil {
          return false, err
      }
      return resp.Allowed, nil
  }
  ```
- **T049** [P]: Write test for auth middleware
- **T050**: Implement `internal/infrastructure/entrypoints/http/middleware/auth.go`:
  ```go
  func AuthMiddleware(authClient providers.AuthService) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              // Extract token from Authorization header
              authHeader := r.Header.Get("Authorization")
              if authHeader == "" {
                  http.Error(w, "missing authorization header", http.StatusUnauthorized)
                  return
              }

              token := strings.TrimPrefix(authHeader, "Bearer ")

              // Validate token via gRPC
              userInfo, err := authClient.ValidateToken(r.Context(), token)
              if err != nil {
                  http.Error(w, "invalid token", http.StatusUnauthorized)
                  return
              }

              // Inject user info into context
              ctx := context.WithValue(r.Context(), "user_id", userInfo.UserID)
              ctx = context.WithValue(ctx, "organization_id", userInfo.OrganizationID)
              ctx = context.WithValue(ctx, "user_email", userInfo.Email)

              next.ServeHTTP(w, r.WithContext(ctx))
          })
      }
  }
  ```
- **T051** [P]: Write test for permission middleware
- **T052**: Implement `internal/infrastructure/entrypoints/http/middleware/permission.go`:
  ```go
  func RequirePermission(authClient providers.AuthService, permission string) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              userID := r.Context().Value("user_id").(string)

              // Check permission via gRPC
              allowed, err := authClient.CheckPermission(r.Context(), userID, permission)
              if err != nil || !allowed {
                  http.Error(w, "permission denied", http.StatusForbidden)
                  return
              }

              next.ServeHTTP(w, r)
          })
      }
  }
  ```
- **T053**: Implement tenant isolation middleware (inject organization_id into GORM scope)
- **T054**: Apply middlewares to router:
  ```go
  r.Use(middleware.AuthMiddleware(authClient))
  r.Use(middleware.TenantMiddleware())

  r.Route("/api/v1/products", func(r chi.Router) {
      r.With(middleware.RequirePermission(authClient, "catalog:products:read")).Get("/", productHandler.List)
      r.With(middleware.RequirePermission(authClient, "catalog:products:write")).Post("/", productHandler.Create)
      r.With(middleware.RequirePermission(authClient, "catalog:products:write")).Put("/{id}", productHandler.Update)
      r.With(middleware.RequirePermission(authClient, "catalog:products:delete")).Delete("/{id}", productHandler.Delete)
  })
  ```
- **T055** [P]: Write integration test: Request with invalid token → 401
- **T056** [P]: Write integration test: Request without permission → 403
- **T057** [P]: Write integration test: User from Org A cannot access Org B's products

**Checkpoint**: Authentication and authorization working via Auth gRPC service

---

### Phase 7: Event Publishing (NATS Integration)
**Goal**: Publish domain events to NATS Jetstream

- **T058** [P]: Write test for catalog event publisher
- **T059**: Implement `internal/infrastructure/adapters/events/catalog_publisher.go`:
  ```go
  type CatalogEventPublisher struct {
      publisher events.Publisher
      logger    logger.Logger
  }

  func (p *CatalogEventPublisher) PublishProductCreated(ctx context.Context, product *domain.Product) error {
      event := events.NewEvent(
          "product.created",
          "catalog-service",
          product.OrganizationID,
          map[string]interface{}{
              "product_id": product.ID,
              "sku":        product.SKU,
              "name":       product.Name,
              "category":   product.Category,
          },
      )

      return p.publisher.Publish(ctx, event)
  }

  func (p *CatalogEventPublisher) PublishProductUpdated(ctx context.Context, product *domain.Product) error {
      // Similar implementation
  }

  func (p *CatalogEventPublisher) PublishSupplierCreated(ctx context.Context, supplier *domain.Supplier) error {
      // Similar implementation
  }
  ```
- **T060**: Update use cases to publish events after successful operations
- **T061** [P]: Write integration test: Create product → verify event in CATALOG_EVENTS stream
- **T062**: Document all catalog events in `docs/EVENTS.md`:
  - catalog.product.created
  - catalog.product.updated
  - catalog.product.deleted
  - catalog.supplier.created
  - catalog.supplier.updated
  - catalog.buffer_profile.assigned

**Checkpoint**: Domain events published to NATS Jetstream

---

### Phase 8: User Story 4 - Product Search and Filtering (P3)
**Goal**: Implement search and filtering

- **T063** [P][US4]: Write test for product search by name/SKU
- **T064** [US4]: Implement `SearchProducts` use case:
  ```go
  func (uc *SearchProductsUseCase) Execute(ctx context.Context, query string, filters map[string]interface{}) ([]*domain.Product, error) {
      // Search by SKU or name (ILIKE)
      return uc.productRepo.Search(ctx, query, filters)
  }
  ```
- **T065** [US4]: Implement repository method with full-text search:
  ```go
  func (r *GORMProductRepository) Search(ctx context.Context, query string, filters map[string]interface{}) ([]*domain.Product, error) {
      db := r.db.WithContext(ctx).Scopes(r.tenantScope(ctx))

      if query != "" {
          db = db.Where("sku ILIKE ? OR name ILIKE ?", "%"+query+"%", "%"+query+"%")
      }

      for key, value := range filters {
          db = db.Where(key+" = ?", value)
      }

      var products []*domain.Product
      err := db.Find(&products).Error
      return products, err
  }
  ```
- **T066** [US4]: Implement search HTTP endpoint `GET /api/v1/products/search?q=widget&category=electronics`
- **T067** [P][US4]: Write integration test: Search returns matching products

**Checkpoint**: Product search and filtering work

---

### Phase 9: Polish (Documentation, Testing, Optimization)
**Goal**: Production-ready service

- **T068** [Polish]: Install Swagger for API documentation:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  swag init -g cmd/api/main.go
  ```
- **T069** [Polish]: Add Swagger annotations to all HTTP handlers
- **T070** [Polish]: Generate OpenAPI spec: `swag init`
- **T071** [Polish]: Serve Swagger UI at `/swagger/index.html`
- **T072** [Polish]: Create comprehensive README.md:
  - Service overview
  - Architecture diagram
  - API endpoints
  - Environment variables
  - Local development guide
  - Testing guide
- **T073** [Polish]: Create `docs/ARCHITECTURE.md` explaining Clean Architecture layers
- **T074** [Polish]: Create Dockerfile:
  ```dockerfile
  FROM golang:1.23-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN CGO_ENABLED=0 go build -o catalog-service ./cmd/api

  FROM alpine:latest
  RUN apk --no-cache add ca-certificates
  WORKDIR /root/
  COPY --from=builder /app/catalog-service .
  EXPOSE 8082
  CMD ["./catalog-service"]
  ```
- **T075** [Polish]: Add health check endpoint `/health`:
  - Check database connectivity
  - Check Auth gRPC connectivity
  - Check NATS connectivity
  - Return 200 OK or 503 Service Unavailable
- **T076** [Polish]: Add metrics endpoint `/metrics` with Prometheus:
  - HTTP request counter (status, method, endpoint)
  - Request duration histogram
  - Active database connections gauge
- **T077** [Polish]: Implement request logging middleware:
  - Log every request (method, path, status, duration)
  - Include request_id for tracing
- **T078** [Polish]: Handle edge case: Duplicate SKU detection
  - Return 409 Conflict with clear error message
- **T079** [Polish]: Handle edge case: Delete product with active buffers
  - Soft delete (set status='inactive')
  - Prevent hard delete if referenced
- **T080** [Polish]: Implement pagination cursor optimization for large datasets
- **T081** [Polish]: Add database indexes for performance:
  - Index on (organization_id, sku) for fast lookups
  - Index on (organization_id, category) for filtering
  - Index on (organization_id, status) for active products
- **T082** [Polish]: Run golangci-lint and fix issues
- **T083** [Polish]: Generate test coverage report, ensure >80%:
  ```bash
  go test ./... -coverprofile=coverage.out
  go tool cover -html=coverage.out
  ```
- **T084** [Polish]: Load test: 1,000 concurrent requests to list products
- **T085** [Polish]: Benchmark API latency:
  - Single product GET: <200ms (p95)
  - Product list (20 items): <500ms (p95)
  - Product creation: <300ms (p95)
- **T086** [Polish]: Create Postman collection for API testing
- **T087** [Polish]: Update main README with link to Catalog service documentation

**Checkpoint**: Service production-ready with complete documentation and tests

---

## Task Dependencies

```mermaid
graph TD
    T001[Create service dir] --> T002[Init go.mod]
    T002 --> T008[Create migrations]
    T008 --> T013[Create entities]
    T013 --> T016[Product repo tests]
    T016 --> T017[Product repo impl]
    T017 --> T019[Create product use case]
    T019 --> T024[Product HTTP handlers]

    T024 --> T027[Supplier repo tests]
    T027 --> T030[Supplier use cases]
    T030 --> T033[Supplier handlers]

    T033 --> T037[Buffer profile tests]
    T037 --> T040[Buffer profile use cases]
    T040 --> T045[Buffer profile handlers]

    T045 --> T047[Auth client tests]
    T047 --> T050[Auth middleware]
    T050 --> T052[Permission middleware]
    T052 --> T054[Apply middlewares]

    T054 --> T058[Event publisher tests]
    T058 --> T060[Integrate events]

    T060 --> T063[Search tests]
    T063 --> T066[Search endpoint]

    T066 --> T068[Swagger setup]
    T068 --> T075[Health checks]
    T075 --> T083[Test coverage]
```

**Critical Path**: T001 → T013 → T019 → T050 → T060 → T083

## Execution Strategy

### Recommended Order (Single Developer)
**Week 1: Foundation**
- Day 1-2: Phase 1-2 (Setup + Database schema)
- Day 3-5: Phase 3 (Product CRUD)

**Week 2: Core Features**
- Day 1-2: Phase 4 (Supplier management)
- Day 3: Phase 5 (Buffer profiles)
- Day 4-5: Phase 6 (Auth integration)

**Week 3: Events and Polish**
- Day 1: Phase 7 (Event publishing)
- Day 2: Phase 8 (Search and filtering)
- Day 3-5: Phase 9 (Polish and testing)

### Team of 3 Developers
**Week 1-2**
- Dev A: Phase 1-3 (Setup + Products)
- Dev B: Phase 4-5 (Suppliers + Buffer profiles)
- Dev C: Phase 6-7 (Auth integration + Events)

**Week 3**
- Dev A: Phase 8 (Search)
- Dev B: Phase 9 (Documentation)
- Dev C: Phase 9 (Testing and optimization)

**Total: 12-15 days single developer, 8-10 days with 3 developers**

## Testing Strategy

### Unit Testing
- Mock repositories for use case tests
- Mock Auth client for middleware tests
- Mock event publisher for use case tests
- Target: >80% coverage per package

### Integration Testing
- Test with real PostgreSQL (docker-compose)
- Test with Auth gRPC service
- Test with NATS Jetstream
- Test full API flows (create → read → update → delete)

### Performance Testing
- Load test: 1,000 concurrent requests
- Benchmark API endpoints (latency <500ms p95)
- Test pagination with 10,000 products
- Test search with 100,000 products

### Manual Testing Checklist
- [ ] Create product via API
- [ ] List products with pagination
- [ ] Update product details
- [ ] Delete product (soft delete)
- [ ] Create supplier and associate with product
- [ ] Assign buffer profile to product
- [ ] Search products by name/SKU
- [ ] Request with invalid token → 401
- [ ] Request without permission → 403
- [ ] User from Org A cannot access Org B's data
- [ ] Events published to NATS after operations
- [ ] Health check returns correct status
- [ ] Swagger UI accessible

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Auth gRPC service unavailable | High | Circuit breaker pattern, fallback to cached permissions, graceful degradation |
| NATS unavailable during event publish | Medium | Don't block operations, log errors, retry async |
| Database performance degradation | Medium | Add indexes, optimize queries, implement caching |
| Duplicate SKU across organizations | Low | Unique constraint on (sku, organization_id), return 409 Conflict |
| Large pagination offset performance | Medium | Implement cursor-based pagination for large datasets |
| Cross-tenant data leak | Critical | Test tenant isolation thoroughly, automated tests, code review |

## Rollout Plan

1. **PR 1**: Foundation (Phase 1-2) - Setup and database schema
2. **PR 2**: Products (Phase 3) - Product CRUD operations
3. **PR 3**: Suppliers (Phase 4) - Supplier management and relationships
4. **PR 4**: Buffer Profiles (Phase 5) - Buffer profile templates
5. **PR 5**: Auth Integration (Phase 6) - Authentication and authorization
6. **PR 6**: Events (Phase 7) - NATS event publishing
7. **PR 7**: Search (Phase 8) - Search and filtering
8. **PR 8**: Polish (Phase 9) - Documentation, testing, optimization

Each PR independently reviewable and deployable.

---

**Estimated Effort**: 12-15 days (single developer), 8-10 days (team of 3)
**Complexity**: Medium-High (full Clean Architecture + multiple integrations)
**Blockers**: Task 5 (Auth service), Task 6 (RBAC), Task 7 (gRPC server), Task 8 (NATS Jetstream)
**Enables**: DDMRP service (depends on product catalog), Execution service (depends on products and suppliers)
