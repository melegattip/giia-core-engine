# Catalog Service

The Catalog Service is a microservice responsible for managing product catalogs, suppliers, and buffer profile templates in the GIIA DDMRP system.

## Features

- ✅ Product Master Data Management (CRUD operations)
- ✅ Supplier Management
- ✅ Product-Supplier Relationships
- ✅ Buffer Profile Templates
- ✅ Product Search and Filtering
- ✅ Multi-tenant support via Organization ID
- ✅ Event Publishing to NATS JetStream
- ✅ Clean Architecture with domain-driven design
- ✅ RESTful HTTP API with Chi router

## Architecture

This service follows **Clean Architecture** principles with clear separation of concerns:

```
├── cmd/server/              # Application entry point
├── internal/
│   ├── core/               # Domain logic (entities, use cases, interfaces)
│   │   ├── domain/        # Business entities
│   │   ├── usecases/      # Business logic
│   │   └── providers/     # Interface contracts
│   └── infrastructure/    # External adapters
│       ├── repositories/  # Data access (GORM)
│       ├── adapters/      # Event publishers
│       ├── entrypoints/   # HTTP handlers & router
│       └── config/        # Configuration
```

## Prerequisites

- Go 1.24.0 or later
- PostgreSQL 16+
- NATS Server with JetStream enabled
- Access to shared packages (pkg/*)

## Configuration

Create a `.env` file based on `.env.example`:

```bash
# Server Configuration
HOST=localhost
PORT=8082
ENVIRONMENT=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=giia
DB_PASSWORD=giia_dev_password
DB_NAME=giia_dev
DB_SCHEMA=catalog
DB_SSLMODE=disable

# NATS Configuration
NATS_URL=nats://localhost:4222

# Logging Configuration
LOG_LEVEL=debug
LOG_FORMAT=json
```

## Database Setup

The service will automatically run GORM AutoMigrate on startup. For production, use the SQL migrations in `internal/infrastructure/persistence/migrations/`:

```bash
# Run migrations manually (if needed)
psql -h localhost -U giia -d giia_dev -f internal/infrastructure/persistence/migrations/001_create_products.sql
psql -h localhost -U giia -d giia_dev -f internal/infrastructure/persistence/migrations/002_create_suppliers.sql
psql -h localhost -U giia -d giia_dev -f internal/infrastructure/persistence/migrations/003_create_product_suppliers.sql
psql -h localhost -U giia -d giia_dev -f internal/infrastructure/persistence/migrations/004_create_buffer_profiles.sql
```

## Running the Service

### Development

```bash
# Build the service
go build -o bin/catalog-service ./cmd/server

# Run the service
./bin/catalog-service

# Or run directly
go run ./cmd/server/main.go
```

### Using Docker

```bash
# Build Docker image
docker build -t catalog-service:latest .

# Run container
docker run -p 8082:8082 --env-file .env catalog-service:latest
```

## API Documentation

### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "service": "catalog-service",
  "checks": {
    "database": "healthy"
  }
}
```

### Products

#### Create Product

```http
POST /api/v1/products
Content-Type: application/json
X-Organization-ID: <organization-uuid>

{
  "sku": "WIDGET-001",
  "name": "Premium Widget",
  "description": "High-quality widget for industrial use",
  "category": "electronics",
  "unit_of_measure": "EA",
  "buffer_profile_id": "profile-uuid" // optional
}
```

#### Get Product

```http
GET /api/v1/products/{id}
X-Organization-ID: <organization-uuid>

# Include suppliers
GET /api/v1/products/{id}?include=suppliers
```

#### Update Product

```http
PUT /api/v1/products/{id}
Content-Type: application/json
X-Organization-ID: <organization-uuid>

{
  "name": "Updated Product Name",
  "description": "Updated description",
  "status": "inactive"
}
```

#### Delete Product (Soft Delete)

```http
DELETE /api/v1/products/{id}
X-Organization-ID: <organization-uuid>
```

#### List Products

```http
GET /api/v1/products?page=1&page_size=20&category=electronics&status=active
X-Organization-ID: <organization-uuid>
```

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `page_size` (int, default: 20, max: 100) - Items per page
- `category` (string, optional) - Filter by category
- `status` (string, optional) - Filter by status (active, inactive, discontinued)

#### Search Products

```http
GET /api/v1/products/search?q=widget&category=electronics&page=1&page_size=20
X-Organization-ID: <organization-uuid>
```

**Query Parameters:**
- `q` (string) - Search query (searches SKU and name)
- `page` (int, default: 1) - Page number
- `page_size` (int, default: 20, max: 100) - Items per page
- `category` (string, optional) - Filter by category
- `status` (string, optional) - Filter by status

## Domain Events

The service publishes events to NATS JetStream on the `CATALOG_EVENTS` stream:

### Product Events

- `catalog.product.created` - When a product is created
- `catalog.product.updated` - When a product is updated
- `catalog.product.deleted` - When a product is deleted (soft delete)

### Supplier Events

- `catalog.supplier.created` - When a supplier is created
- `catalog.supplier.updated` - When a supplier is updated
- `catalog.supplier.deleted` - When a supplier is deleted

### Buffer Profile Events

- `catalog.buffer_profile.assigned` - When a buffer profile is assigned to a product

## Multi-Tenancy

All API requests require an `X-Organization-ID` header to identify the organization. Data is automatically scoped to the organization to ensure tenant isolation.

```http
X-Organization-ID: 550e8400-e29b-41d4-a716-446655440000
```

## Error Handling

The service uses typed errors from the shared `pkg/errors` package. All errors are returned with appropriate HTTP status codes:

- `400 Bad Request` - Invalid input or validation errors
- `404 Not Found` - Resource not found
- `409 Conflict` - Duplicate resource (e.g., SKU already exists)
- `500 Internal Server Error` - Server-side errors

**Error Response Format:**
```json
{
  "error_code": "BAD_REQUEST",
  "message": "SKU is required"
}
```

## Development Guidelines

This service follows the GIIA project's development guidelines defined in `/CLAUDE.md`:

- ✅ Clean Architecture with dependency inversion
- ✅ Typed errors (no `fmt.Errorf`)
- ✅ Structured logging with context
- ✅ GORM for database operations
- ✅ Multi-tenant scoping on all queries
- ✅ Domain event publishing
- ✅ Comprehensive input validation

## Testing

```bash
# Run unit tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests with race detection
go test ./... -race
```

## Dependencies

### Internal Dependencies
- `pkg/config` - Configuration management
- `pkg/database` - Database connection utilities
- `pkg/errors` - Typed error handling
- `pkg/events` - NATS event publishing
- `pkg/logger` - Structured logging

### External Dependencies
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/google/uuid` - UUID generation
- `gorm.io/gorm` - ORM for database operations
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variable loading

## License

Proprietary - GIIA Project

## Contributors

- GIIA Development Team
