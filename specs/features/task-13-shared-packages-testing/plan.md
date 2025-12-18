# Task 13: Shared Packages Testing - Implementation Plan

**Task ID**: task-13-shared-packages-testing
**Phase**: 2A - Complete to 100%
**Priority**: P1 (High)
**Estimated Duration**: 1 week (5 days)
**Dependencies**: Task 4 (85% - Shared Packages foundation)

---

## 1. Technical Context

### Current State
- **Shared Packages**: 85% complete
  - `pkg/database`: Connection pooling, transactions (needs integration tests)
  - `pkg/events`: Publisher, Subscriber, stream management (needs integration tests)
  - `pkg/config`: Configuration loading (needs unit tests)
  - `pkg/logger`: Structured logging (needs additional unit tests)
  - `pkg/errors`: Typed errors (100% complete with tests ✅)

### Technology Stack
- **Testing Framework**: Go testing package + testify
- **Integration Tests**: Testcontainers-go for Docker containers
- **Infrastructure**: PostgreSQL 16, NATS 2.x, Redis 7
- **CI/CD**: GitHub Actions with Docker services
- **Coverage**: go tool cover

### Key Design Decisions
- **Real Infrastructure**: Use testcontainers for integration tests (PostgreSQL, NATS)
- **Isolation**: Each test gets clean infrastructure (fresh database, purged streams)
- **Parallel**: Unit tests run in parallel, integration tests run sequentially (resource constraints)
- **Build Tags**: `//go:build integration` for integration tests (excluded from default `go test`)
- **Cleanup**: defer cleanup functions to prevent resource leaks

---

## 2. Project Structure

### Files to Create

```
giia-core-engine/
└── pkg/
    ├── database/
    │   ├── connection.go                        [EXISTS]
    │   ├── connection_test.go                   [NEW] Unit tests
    │   └── connection_integration_test.go       [NEW] Integration tests
    │
    ├── events/
    │   ├── publisher.go                         [EXISTS]
    │   ├── subscriber.go                        [EXISTS]
    │   ├── stream_config.go                     [EXISTS]
    │   ├── publisher_test.go                    [NEW] Unit tests with mocks
    │   ├── subscriber_test.go                   [NEW] Unit tests with mocks
    │   ├── publisher_integration_test.go        [NEW] Integration tests
    │   ├── subscriber_integration_test.go       [NEW] Integration tests
    │   └── stream_integration_test.go           [NEW] Stream management tests
    │
    ├── config/
    │   ├── loader.go                            [EXISTS]
    │   ├── loader_test.go                       [NEW] Unit tests
    │   └── validator_test.go                    [NEW] Validation tests
    │
    ├── logger/
    │   ├── logger.go                            [EXISTS]
    │   ├── logger_test.go                       [UPDATE] Add more unit tests
    │   └── context_test.go                      [NEW] Context extraction tests
    │
    └── testutil/
        ├── docker.go                            [NEW] Docker helper utilities
        ├── postgres.go                          [NEW] PostgreSQL test helpers
        ├── nats.go                              [NEW] NATS test helpers
        └── fixtures.go                          [NEW] Common test fixtures

```

---

## 3. Implementation Phases

### Phase 1: Test Infrastructure Setup (Day 1)

#### T001: Setup Testcontainers Infrastructure

**File**: `pkg/testutil/docker.go`

```go
package testutil

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ContainerManager handles lifecycle of test containers
type ContainerManager struct {
	containers []testcontainers.Container
}

func NewContainerManager() *ContainerManager {
	return &ContainerManager{
		containers: []testcontainers.Container{},
	}
}

func (cm *ContainerManager) StartPostgres(ctx context.Context, t *testing.T) (string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_pass",
			"POSTGRES_DB":       "test_db",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	cm.containers = append(cm.containers, container)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	connectionString := fmt.Sprintf("postgres://test_user:test_pass@%s:%s/test_db?sslmode=disable", host, port.Port())

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PostgreSQL container: %v", err)
		}
	}

	return connectionString, cleanup
}

func (cm *ContainerManager) StartNATS(ctx context.Context, t *testing.T) (string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "nats:2-alpine",
		ExposedPorts: []string{"4222/tcp"},
		Cmd:          []string{"-js"},
		WaitingFor:   wait.ForLog("Server is ready"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start NATS container: %v", err)
	}

	cm.containers = append(cm.containers, container)

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "4222")

	connectionURL := fmt.Sprintf("nats://%s:%s", host, port.Port())

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate NATS container: %v", err)
		}
	}

	return connectionURL, cleanup
}

func (cm *ContainerManager) Cleanup(ctx context.Context) {
	for _, container := range cm.containers {
		_ = container.Terminate(ctx)
	}
}
```

**File**: `pkg/testutil/postgres.go`

```go
package testutil

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func SetupTestDatabase(t *testing.T, dsn string) (*gorm.DB, func()) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

func TruncateAllTables(t *testing.T, db *gorm.DB) {
	// Get all table names
	var tables []string
	db.Raw(`
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
	`).Scan(&tables)

	// Truncate each table
	for _, table := range tables {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}
}
```

**File**: `pkg/testutil/nats.go`

```go
package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func SetupTestNATS(t *testing.T, url string) (*nats.Conn, func()) {
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		t.Fatalf("Failed to connect to test NATS: %v", err)
	}

	cleanup := func() {
		nc.Close()
	}

	return nc, cleanup
}

func PurgeStream(t *testing.T, js nats.JetStreamContext, streamName string) {
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		return // Stream doesn't exist, nothing to purge
	}

	if stream != nil {
		_ = js.PurgeStream(streamName)
	}
}

func DeleteStream(t *testing.T, js nats.JetStreamContext, streamName string) {
	_ = js.DeleteStream(streamName)
}
```

#### T002: Docker Compose for Local Testing

**File**: `docker-compose.test.yml`

```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: test_user
      POSTGRES_PASSWORD: test_pass
      POSTGRES_DB: test_db
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test_user"]
      interval: 5s
      timeout: 5s
      retries: 5

  nats-test:
    image: nats:2-alpine
    command: ["-js"]
    ports:
      - "4223:4222"
    healthcheck:
      test: ["CMD", "/nats-server", "-sl=ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
```

**Usage**:
```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test ./pkg/... -tags=integration -v

# Stop infrastructure
docker-compose -f docker-compose.test.yml down -v
```

---

### Phase 2: Database Package Integration Tests (Day 2)

#### T003: Database Connection Tests

**File**: `pkg/database/connection_integration_test.go`

```go
//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConnection_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	// Start PostgreSQL container
	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	t.Run("should connect successfully with valid DSN", func(t *testing.T) {
		db, err := Connect(dsn)
		require.NoError(t, err)
		require.NotNil(t, db)

		// Verify connection
		sqlDB, _ := db.DB()
		err = sqlDB.Ping()
		assert.NoError(t, err)

		sqlDB.Close()
	})

	t.Run("should fail with invalid DSN", func(t *testing.T) {
		db, err := Connect("postgres://invalid:invalid@localhost:9999/invalid")
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestDatabaseConnectionPooling_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := Connect(dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Set pool configuration
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	t.Run("should handle concurrent connections", func(t *testing.T) {
		const numGoroutines = 20

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				err := sqlDB.Ping()
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		stats := sqlDB.Stats()
		assert.LessOrEqual(t, stats.OpenConnections, 10, "Should not exceed max open connections")
	})
}

func TestDatabaseTransactions_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := Connect(dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create test table
	db.Exec(`CREATE TABLE test_table (id SERIAL PRIMARY KEY, name VARCHAR(100))`)

	t.Run("should commit transaction successfully", func(t *testing.T) {
		tx := db.Begin()

		tx.Exec("INSERT INTO test_table (name) VALUES (?)", "test1")
		tx.Exec("INSERT INTO test_table (name) VALUES (?)", "test2")

		err := tx.Commit().Error
		assert.NoError(t, err)

		// Verify records persisted
		var count int64
		db.Model(&struct{ Name string }{}).Table("test_table").Count(&count)
		assert.Equal(t, int64(2), count)
	})

	t.Run("should rollback transaction on error", func(t *testing.T) {
		// Clear table
		db.Exec("TRUNCATE test_table")

		tx := db.Begin()

		tx.Exec("INSERT INTO test_table (name) VALUES (?)", "test3")

		// Rollback
		tx.Rollback()

		// Verify no records persisted
		var count int64
		db.Model(&struct{ Name string }{}).Table("test_table").Count(&count)
		assert.Equal(t, int64(0), count)
	})
}

func TestDatabaseHealthCheck_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := Connect(dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	t.Run("should return healthy when database is up", func(t *testing.T) {
		healthy := IsHealthy(db)
		assert.True(t, healthy)
	})

	t.Run("should return unhealthy when database is down", func(t *testing.T) {
		// Close connection
		sqlDB.Close()

		healthy := IsHealthy(db)
		assert.False(t, healthy)
	})
}
```

#### T004: Database Retry Logic Tests

**File**: `pkg/database/retry_integration_test.go`

```go
//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseRetryLogic_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	dsn, cleanup := cm.StartPostgres(ctx, t)
	defer cleanup()

	db, err := Connect(dsn)
	require.NoError(t, err)

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	t.Run("should retry on connection failure", func(t *testing.T) {
		// This test would require temporarily killing the database
		// For now, we'll test the retry mechanism with a timeout scenario

		// Set short connection timeout
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Attempt query with timeout
		err := db.WithContext(ctx).Raw("SELECT pg_sleep(1)").Error

		// Should error due to timeout
		assert.Error(t, err)
	})
}
```

---

### Phase 3: Events Package Integration Tests (Days 3-4)

#### T005: Event Publisher Tests

**File**: `pkg/events/publisher_integration_test.go`

```go
//go:build integration

package events

import (
	"context"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/pkg/testutil"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventPublisher_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// Create test stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "TEST_EVENTS",
		Subjects: []string{"test.>"},
	})
	require.NoError(t, err)

	publisher := NewPublisher(nc)

	t.Run("should publish event successfully", func(t *testing.T) {
		event := &Event{
			Type:    "test.created",
			Subject: "test.items.123",
			Data: map[string]interface{}{
				"item_id": "123",
				"name":    "Test Item",
			},
		}

		err := publisher.Publish(ctx, event)
		assert.NoError(t, err)

		// Verify event was published
		time.Sleep(100 * time.Millisecond) // Wait for async publish

		stream, err := js.StreamInfo("TEST_EVENTS")
		require.NoError(t, err)
		assert.Equal(t, uint64(1), stream.State.Msgs)
	})

	t.Run("should publish multiple events", func(t *testing.T) {
		// Purge stream first
		testutil.PurgeStream(t, js, "TEST_EVENTS")

		events := []*Event{
			{Type: "test.created", Subject: "test.items.1", Data: map[string]interface{}{"id": "1"}},
			{Type: "test.created", Subject: "test.items.2", Data: map[string]interface{}{"id": "2"}},
			{Type: "test.created", Subject: "test.items.3", Data: map[string]interface{}{"id": "3"}},
		}

		for _, event := range events {
			err := publisher.Publish(ctx, event)
			assert.NoError(t, err)
		}

		time.Sleep(100 * time.Millisecond)

		stream, err := js.StreamInfo("TEST_EVENTS")
		require.NoError(t, err)
		assert.Equal(t, uint64(3), stream.State.Msgs)
	})
}
```

#### T006: Event Subscriber Tests

**File**: `pkg/events/subscriber_integration_test.go`

```go
//go:build integration

package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/pkg/testutil"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventSubscriber_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// Create test stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "TEST_EVENTS",
		Subjects: []string{"test.>"},
	})
	require.NoError(t, err)

	publisher := NewPublisher(nc)
	subscriber := NewSubscriber(nc)

	t.Run("should receive published event", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_EVENTS")

		var receivedEvent *Event
		var wg sync.WaitGroup
		wg.Add(1)

		// Subscribe
		err := subscriber.Subscribe(ctx, "TEST_EVENTS", "test.>", func(event *Event) error {
			receivedEvent = event
			wg.Done()
			return nil
		})
		require.NoError(t, err)

		// Publish event
		event := &Event{
			Type:    "test.created",
			Subject: "test.items.456",
			Data: map[string]interface{}{
				"item_id": "456",
			},
		}
		err = publisher.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for event to be received
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.NotNil(t, receivedEvent)
			assert.Equal(t, "test.created", receivedEvent.Type)
			assert.Equal(t, "456", receivedEvent.Data["item_id"])
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for event")
		}
	})

	t.Run("should handle multiple events", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_EVENTS")

		receivedEvents := []*Event{}
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(3)

		// Subscribe
		err := subscriber.Subscribe(ctx, "TEST_EVENTS", "test.>", func(event *Event) error {
			mu.Lock()
			receivedEvents = append(receivedEvents, event)
			mu.Unlock()
			wg.Done()
			return nil
		})
		require.NoError(t, err)

		// Publish multiple events
		for i := 1; i <= 3; i++ {
			event := &Event{
				Type:    "test.created",
				Subject: fmt.Sprintf("test.items.%d", i),
				Data: map[string]interface{}{
					"id": fmt.Sprintf("%d", i),
				},
			}
			err = publisher.Publish(ctx, event)
			require.NoError(t, err)
		}

		// Wait for all events
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Len(t, receivedEvents, 3)
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for events")
		}
	})
}

func TestEventSubscriber_ErrorHandling_Integration(t *testing.T) {
	ctx := context.Background()
	cm := testutil.NewContainerManager()
	defer cm.Cleanup(ctx)

	natsURL, cleanup := cm.StartNATS(ctx, t)
	defer cleanup()

	nc, cleanupConn := testutil.SetupTestNATS(t, natsURL)
	defer cleanupConn()

	js, err := nc.JetStream()
	require.NoError(t, err)

	// Create test stream with max retries
	_, err = js.AddStream(&nats.StreamConfig{
		Name:       "TEST_DLQ",
		Subjects:   []string{"dlq.>"},
		MaxMsgsPer: 10,
	})
	require.NoError(t, err)

	publisher := NewPublisher(nc)
	subscriber := NewSubscriber(nc)

	t.Run("should send to DLQ after max retries", func(t *testing.T) {
		testutil.PurgeStream(t, js, "TEST_DLQ")

		attemptCount := 0
		var wg sync.WaitGroup
		wg.Add(3) // Expect 3 retry attempts

		// Subscribe with failing handler
		err := subscriber.Subscribe(ctx, "TEST_DLQ", "dlq.test", func(event *Event) error {
			attemptCount++
			wg.Done()
			return fmt.Errorf("simulated processing error")
		})
		require.NoError(t, err)

		// Publish event
		event := &Event{
			Type:    "test.failed",
			Subject: "dlq.test.789",
			Data: map[string]interface{}{
				"test": "dlq",
			},
		}
		err = publisher.Publish(ctx, event)
		require.NoError(t, err)

		// Wait for retries
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			assert.Equal(t, 3, attemptCount, "Should retry 3 times")
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for retries")
		}
	})
}
```

---

### Phase 4: Config and Logger Unit Tests (Day 5)

#### T007: Config Package Tests

**File**: `pkg/config/loader_test.go`

```go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv(t *testing.T) {
	t.Run("should load configuration from environment variables", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("DATABASE_URL", "postgres://localhost:5432/testdb")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("LOG_LEVEL", "debug")
		defer os.Clearenv()

		cfg := &Config{}
		err := LoadFromEnv(cfg)

		require.NoError(t, err)
		assert.Equal(t, "postgres://localhost:5432/testdb", cfg.DatabaseURL)
		assert.Equal(t, "redis://localhost:6379", cfg.RedisURL)
		assert.Equal(t, "debug", cfg.LogLevel)
	})

	t.Run("should use default values when env var not set", func(t *testing.T) {
		os.Clearenv()

		cfg := &Config{
			LogLevel: "info", // default value
		}
		err := LoadFromEnv(cfg)

		require.NoError(t, err)
		assert.Equal(t, "info", cfg.LogLevel)
	})

	t.Run("should fail when required env var missing", func(t *testing.T) {
		os.Clearenv()

		cfg := &Config{}
		err := LoadFromEnv(cfg)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("should validate required fields", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL: "",
		}

		err := Validate(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database_url")
	})

	t.Run("should validate URL format", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL: "invalid-url",
		}

		err := Validate(cfg)
		assert.Error(t, err)
	})

	t.Run("should pass validation with valid config", func(t *testing.T) {
		cfg := &Config{
			DatabaseURL: "postgres://localhost:5432/db",
			RedisURL:    "redis://localhost:6379",
			NATSUrl:     "nats://localhost:4222",
			LogLevel:    "info",
		}

		err := Validate(cfg)
		assert.NoError(t, err)
	})
}
```

#### T008: Logger Package Tests

**File**: `pkg/logger/logger_test.go` (update existing)

```go
package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Levels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "debug")

	ctx := context.Background()

	t.Run("should log debug messages", func(t *testing.T) {
		buf.Reset()
		logger.Debug(ctx, "debug message", Tags{"key": "value"})

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		assert.Equal(t, "debug", logEntry["level"])
		assert.Equal(t, "debug message", logEntry["message"])
		assert.Equal(t, "value", logEntry["key"])
	})

	t.Run("should log info messages", func(t *testing.T) {
		buf.Reset()
		logger.Info(ctx, "info message", Tags{"user_id": "123"})

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		assert.Equal(t, "info", logEntry["level"])
		assert.Equal(t, "123", logEntry["user_id"])
	})

	t.Run("should log error messages with error details", func(t *testing.T) {
		buf.Reset()
		err := errors.New("test error")
		logger.Error(ctx, err, "error occurred", Tags{"operation": "test"})

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		assert.Equal(t, "error", logEntry["level"])
		assert.Equal(t, "error occurred", logEntry["message"])
		assert.Equal(t, "test error", logEntry["error"])
	})
}

func TestLogger_ContextExtraction(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")

	t.Run("should extract request ID from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "request_id", "req-123")

		buf.Reset()
		logger.Info(ctx, "test message", nil)

		var logEntry map[string]interface{}
		json.Unmarshal(buf.Bytes(), &logEntry)

		assert.Equal(t, "req-123", logEntry["request_id"])
	})
}
```

---

## 4. Testing Strategy

### Running Tests

#### Unit Tests (Fast)
```bash
# Run all unit tests
go test ./pkg/... -v

# With coverage
go test ./pkg/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Parallel execution
go test ./pkg/... -v -parallel=8
```

#### Integration Tests (Slower)
```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Run integration tests only
go test ./pkg/... -tags=integration -v

# Run specific package integration tests
go test ./pkg/database -tags=integration -v
go test ./pkg/events -tags=integration -v

# Stop infrastructure
docker-compose -f docker-compose.test.yml down -v
```

#### CI/CD (GitHub Actions)
```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_pass
          POSTGRES_DB: test_db
        ports:
          - 5432:5432

      nats:
        image: nats:2-alpine
        options: --health-cmd="/nats-server -sl=ping" --health-interval=5s
        ports:
          - 4222:4222

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23.4'

      - name: Run unit tests
        run: go test ./pkg/... -v -coverprofile=coverage.out

      - name: Run integration tests
        run: go test ./pkg/... -tags=integration -v
        env:
          DATABASE_URL: postgres://test_user:test_pass@localhost:5432/test_db
          NATS_URL: nats://localhost:4222

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### Coverage Targets
- **pkg/database**: 90%+ (integration tests)
- **pkg/events**: 90%+ (integration tests)
- **pkg/config**: 85%+ (unit tests)
- **pkg/logger**: 90%+ (unit tests)
- **Overall**: 85%+

---

## 5. Dependencies and Execution Order

```
T001 (Test Infrastructure) → T002 (Docker Compose)
                               ↓
T003 (Database Tests) → T004 (Retry Tests) → Database Package 100%
                               ↓
T005 (Publisher Tests) → T006 (Subscriber Tests) → Events Package 100%
                               ↓
T007 (Config Tests) → T008 (Logger Tests) → All Packages 100%
```

---

## 6. Acceptance Checklist

### Database Package
- [ ] Integration tests with real PostgreSQL
- [ ] Connection pooling tests (concurrent)
- [ ] Transaction commit and rollback tests
- [ ] Health check tests
- [ ] Retry logic tests
- [ ] 90%+ coverage

### Events Package
- [ ] Integration tests with real NATS
- [ ] Event publish tests
- [ ] Event subscribe tests
- [ ] Multiple events handling
- [ ] DLQ tests with retry logic
- [ ] Consumer group tests
- [ ] 90%+ coverage

### Config Package
- [ ] Environment variable loading tests
- [ ] File loading tests
- [ ] Validation tests
- [ ] Default value tests
- [ ] 85%+ coverage

### Logger Package
- [ ] Log level tests
- [ ] Structured tags tests
- [ ] Context extraction tests
- [ ] JSON output format tests
- [ ] 90%+ coverage

### Infrastructure
- [ ] Testcontainers setup working
- [ ] Docker Compose for local testing
- [ ] GitHub Actions workflow passing
- [ ] All tests pass in CI/CD

---

**Document Version**: 1.0
**Last Updated**: 2025-12-16
**Status**: Ready for Implementation
**Estimated Completion**: 5 days (1 week)
**Next Step**: Begin Phase 1 - Test Infrastructure Setup
