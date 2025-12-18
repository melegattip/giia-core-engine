# Integration Tests

This directory contains end-to-end integration tests for the GIIA platform services.

## Overview

The integration tests verify the complete user journey across multiple services, including:

- **Auth Service**: User registration, login, token management
- **Catalog Service**: Product CRUD operations with authentication
- **Service-to-Service Communication**: Auth → Catalog flow with JWT tokens

## Prerequisites

Before running integration tests, ensure you have:

1. **Docker & Docker Compose** installed
2. **Go 1.23+** installed
3. All services running via Docker Compose

## Running Integration Tests

### Step 1: Start Services with Docker Compose

From the project root directory:

```bash
# Start all infrastructure and services
docker compose up -d

# Verify all services are healthy
docker compose ps

# Check service logs if needed
docker compose logs auth-service
docker compose logs catalog-service
```

Wait for all services to be healthy (usually 30-60 seconds).

### Step 2: Run Integration Tests

```bash
# From the project root
cd tests/integration

# Download dependencies
go mod download

# Run all integration tests
go test -v ./...

# Run with race detection
go test -v -race ./...

# Run specific test
go test -v -run TestAuthCatalogFlow_CompleteUserJourney
```

### Step 3: Stop Services

```bash
# From project root
docker compose down

# Clean up volumes (optional - will delete all data)
docker compose down -v
```

## Test Coverage

### Auth-Catalog Flow Test

The `TestAuthCatalogFlow_CompleteUserJourney` test covers:

1. **User Registration** - Create a new user account
2. **User Login** - Authenticate and receive JWT tokens
3. **Create Product** - Create product with valid authentication
4. **Get Product** - Retrieve product details
5. **Unauthorized Access** - Verify requests without tokens fail
6. **Invalid Token** - Verify requests with invalid tokens fail
7. **List Products** - Retrieve paginated product list
8. **Update Product** - Modify product details
9. **Search Products** - Full-text search functionality
10. **Delete Product** - Soft delete product
11. **Verify Deletion** - Confirm deleted product is inaccessible
12. **Token Refresh** - Validate token refresh mechanism

## Test Structure

```
tests/integration/
├── README.md                    # This file
├── go.mod                       # Go module definition
├── auth_catalog_flow_test.go   # Auth-Catalog integration tests
└── ...                          # Additional test files
```

## Writing New Integration Tests

When adding new integration tests:

1. **Naming Convention**: Use `*_test.go` suffix
2. **Test Functions**: Prefix with `Test` (e.g., `TestServiceFlow`)
3. **Helper Functions**: Create reusable helpers for HTTP requests
4. **Cleanup**: Always clean up test data when possible
5. **Skip in Short Mode**: Add `if testing.Short() { t.Skip() }` for long-running tests

### Example Test Template

```go
func TestNewServiceFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    t.Run("1_DescriptiveTestCase", func(t *testing.T) {
        // Arrange
        // Act
        // Assert
    })
}
```

## Environment Variables

The tests expect services to be accessible at:

- **Auth Service**: `http://localhost:8083`
- **Catalog Service**: `http://localhost:8082`

To override these, modify the constants in the test files.

## Troubleshooting

### Services Not Accessible

```bash
# Check if services are running
docker compose ps

# Check service health
curl http://localhost:8083/health
curl http://localhost:8082/health

# Check service logs
docker compose logs -f auth-service
docker compose logs -f catalog-service
```

### Database Connection Errors

```bash
# Verify PostgreSQL is running
docker compose ps postgres

# Check PostgreSQL logs
docker compose logs postgres

# Restart PostgreSQL
docker compose restart postgres
```

### Tests Fail with "Connection Refused"

Ensure all services have finished starting:

```bash
# Wait for health checks to pass
docker compose ps

# All services should show "Up (healthy)"
```

### Clean Slate

If tests are failing due to stale data:

```bash
# Stop and remove all containers and volumes
docker compose down -v

# Restart everything
docker compose up -d

# Wait for services to be ready
sleep 30

# Run tests
cd tests/integration && go test -v ./...
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'

      - name: Start services
        run: docker compose up -d

      - name: Wait for services
        run: |
          timeout 60 bash -c 'until curl -f http://localhost:8083/health; do sleep 2; done'
          timeout 60 bash -c 'until curl -f http://localhost:8082/health; do sleep 2; done'

      - name: Run integration tests
        run: |
          cd tests/integration
          go test -v ./...

      - name: Stop services
        if: always()
        run: docker compose down -v
```

## Best Practices

1. **Isolation**: Each test should be independent
2. **Idempotency**: Tests should be repeatable
3. **Cleanup**: Always clean up resources
4. **Realistic Data**: Use realistic test data
5. **Error Handling**: Test both success and failure paths
6. **Documentation**: Document complex test scenarios

## Contributing

When adding integration tests:

1. Follow the existing test structure
2. Add comprehensive test cases
3. Document any special setup requirements
4. Ensure tests are idempotent
5. Update this README with new test coverage

## Additional Resources

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Project README](../../README.md)
