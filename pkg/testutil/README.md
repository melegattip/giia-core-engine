# Test Utilities Package

Shared utilities for integration and unit testing across all GIIA shared packages.

## Overview

This package provides Docker container management and helper functions for integration testing with real infrastructure (PostgreSQL, NATS JetStream, etc.).

## Components

### Docker Container Management

**File**: `docker.go`

Manages lifecycle of test containers using testcontainers-go:

```go
import "github.com/melegattip/giia-core-engine/pkg/testutil"

ctx := context.Background()
cm := testutil.NewContainerManager()
defer cm.Cleanup(ctx)

// Start PostgreSQL
dsn, cleanup := cm.StartPostgres(ctx, t)
defer cleanup()

// Start NATS
natsURL, cleanup := cm.StartNATS(ctx, t)
defer cleanup()
```

### PostgreSQL Helpers

**File**: `postgres.go`

Helper functions for PostgreSQL testing:

```go
// Setup test database connection
db, cleanup := testutil.SetupTestDatabase(t, dsn)
defer cleanup()

// Create test table
testutil.CreateTestTable(t, db, "test_table")

// Truncate all tables
testutil.TruncateAllTables(t, db)

// Count records
count := testutil.CountRecords(t, db, "test_table")

// Drop test table
testutil.DropTestTable(t, db, "test_table")
```

### NATS JetStream Helpers

**File**: `nats.go`

Helper functions for NATS JetStream testing:

```go
// Setup NATS connection
nc, cleanup := testutil.SetupTestNATS(t, natsURL)
defer cleanup()

js, _ := nc.JetStream()

// Create test stream
testutil.CreateTestStream(t, js, "TEST_STREAM", []string{"test.>"})

// Purge stream
testutil.PurgeStream(t, js, "TEST_STREAM")

// Delete stream
testutil.DeleteStream(t, js, "TEST_STREAM")

// Get stream info
info := testutil.GetStreamInfo(t, js, "TEST_STREAM")

// Wait for messages
success := testutil.WaitForMessages(t, js, "TEST_STREAM", 5, 3*time.Second)
```

## Usage in Tests

### Integration Test Example

```go
//go:build integration

package mypackage

import (
    "context"
    "testing"

    "github.com/melegattip/giia-core-engine/pkg/testutil"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyFeature_Integration(t *testing.T) {
    ctx := context.Background()
    cm := testutil.NewContainerManager()
    defer cm.Cleanup(ctx)

    // Start required infrastructure
    dsn, cleanup := cm.StartPostgres(ctx, t)
    defer cleanup()

    // Your test logic here
    db, dbCleanup := testutil.SetupTestDatabase(t, dsn)
    defer dbCleanup()

    // ... test code ...
}
```

## Running Tests

### With Docker Compose (Recommended)

```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./pkg/... -tags=integration -v

# Stop infrastructure
docker-compose -f docker-compose.test.yml down -v
```

### With Testcontainers (Automatic)

Integration tests automatically start required Docker containers:

```bash
# Just run the tests
go test ./pkg/database -tags=integration -v
go test ./pkg/events -tags=integration -v
```

**Note**: Docker must be running for testcontainers to work.

## Dependencies

- `github.com/testcontainers/testcontainers-go v0.34.0` - Docker container management
- `github.com/stretchr/testify v1.10.0` - Test assertions
- `gorm.io/gorm` - ORM for database testing
- `github.com/nats-io/nats.go` - NATS client for event testing

## Best Practices

1. **Always clean up**: Use defer statements for cleanup functions
2. **Isolate tests**: Each test should get fresh infrastructure
3. **Proper timeouts**: Set appropriate timeouts for container startup
4. **Build tags**: Use `//go:build integration` for integration tests
5. **Parallel execution**: Unit tests can run in parallel, integration tests should be sequential

## Troubleshooting

### Docker not available
```
Error: Cannot connect to the Docker daemon
```
**Solution**: Ensure Docker Desktop is running

### Port conflicts
```
Error: Port 5432 is already in use
```
**Solution**: Stop conflicting services or use Docker Compose with mapped ports

### Container startup timeout
```
Error: Waiting for container timed out
```
**Solution**: Increase timeout or check Docker resource limits

## Examples

See these packages for usage examples:
- `pkg/database/database_integration_test.go`
- `pkg/events/publisher_integration_test.go`
- `pkg/events/subscriber_integration_test.go`
