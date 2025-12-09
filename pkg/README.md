# Shared Infrastructure Packages

This directory contains shared Go packages used across all GIIA microservices. These packages provide foundational infrastructure capabilities following Clean Architecture principles.

## Available Packages

### 🔧 [config](./config)
Configuration management using Viper with support for environment variables, .env files, and type-safe getters.

**Key Features:**
- Load from multiple sources (.env, environment variables)
- Environment variable overrides
- Type-safe getters (String, Int, Bool, Float64)
- Required key validation

**Usage:**
```go
cfg, _ := config.New("GIIA")
dbHost := cfg.GetString("database.host")
```

---

### 📝 [logger](./logger)
Structured JSON logging using Zerolog with context support and log level filtering.

**Key Features:**
- Structured JSON output for production
- Context-aware logging (request ID)
- Log levels: debug, info, warn, error, fatal
- Mock implementation for testing

**Usage:**
```go
log := logger.New("auth-service", "info")
log.Info(ctx, "User authenticated", logger.Tags{"user_id": 123})
```

---

### 🗄️ [database](./database)
PostgreSQL connection management using GORM with connection pooling, retry logic, and health checks.

**Key Features:**
- GORM-based connections
- Automatic retry with exponential backoff
- Configurable connection pooling
- Health check support

**Usage:**
```go
db := database.New()
conn, _ := db.Connect(ctx, &database.Config{
    Host: "localhost",
    Port: 5432,
    User: "postgres",
    Password: "secret",
    DatabaseName: "giia_db",
})
```

---

### ⚠️ [errors](./errors)
Typed error system with HTTP status code mapping for consistent API error responses.

**Key Features:**
- Typed error constructors (400, 401, 403, 404, 500, 503)
- Error wrapping with context
- JSON serialization for HTTP responses
- Error code constants

**Usage:**
```go
if userID <= 0 {
    return errors.NewBadRequest("invalid user ID")
}

response := errors.ToHTTPResponse(err)
```

---

### 📡 [events](./events)
NATS Jetstream event publishing and subscription with CloudEvents-like structure.

**Key Features:**
- Publisher and subscriber interfaces
- Automatic retry with exponential backoff
- Durable subscriptions
- At-least-once delivery

**Usage:**
```go
publisher, _ := events.NewPublisher(nc)
event := events.NewEvent("user.created", "auth-service", "org-123", data)
publisher.Publish(ctx, "users.events", event)
```

---

## Package Dependencies

```
errors (no dependencies)
  ↓
config (uses: viper)
  ↓
logger (uses: zerolog, depends on: errors)
  ↓
database (uses: gorm, pgx, depends on: errors)
events (uses: nats, uuid, depends on: errors)
```

## Installation

All packages are part of the GIIA Core Engine Go workspace:

```go
// In your service's go.mod
require (
    github.com/giia/giia-core-engine/pkg/config v0.1.0
    github.com/giia/giia-core-engine/pkg/logger v0.1.0
    github.com/giia/giia-core-engine/pkg/database v0.1.0
    github.com/giia/giia-core-engine/pkg/errors v0.1.0
    github.com/giia/giia-core-engine/pkg/events v0.1.0
)
```

## Common Usage Pattern

Here's how to initialize all packages together in a microservice:

```go
package main

import (
    "context"
    "log"

    "github.com/giia/giia-core-engine/pkg/config"
    "github.com/giia/giia-core-engine/pkg/logger"
    "github.com/giia/giia-core-engine/pkg/database"
    "github.com/giia/giia-core-engine/pkg/events"
)

func main() {
    ctx := context.Background()

    // 1. Load configuration
    cfg, err := config.New("AUTH_SERVICE")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Validate required keys
    requiredKeys := []string{"database.host", "database.port", "nats.url"}
    if err := cfg.Validate(requiredKeys); err != nil {
        log.Fatalf("Missing required config: %v", err)
    }

    // 2. Initialize logger
    logLevel := cfg.GetString("log.level")
    logger := logger.New("auth-service", logLevel)
    logger.Info(ctx, "Starting auth-service", nil)

    // 3. Connect to database
    db := database.New()
    dbConfig := &database.Config{
        Host:         cfg.GetString("database.host"),
        Port:         cfg.GetInt("database.port"),
        User:         cfg.GetString("database.user"),
        Password:     cfg.GetString("database.password"),
        DatabaseName: cfg.GetString("database.name"),
    }
    conn, err := db.Connect(ctx, dbConfig)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to connect to database", nil)
    }
    defer db.Close(conn)

    // 4. Connect to NATS
    natsURL := cfg.GetString("nats.url")
    nc, err := events.ConnectWithDefaults(natsURL)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to connect to NATS", nil)
    }
    defer events.Disconnect(nc)

    // 5. Initialize event publisher
    publisher, err := events.NewPublisher(nc)
    if err != nil {
        logger.Fatal(ctx, err, "Failed to create event publisher", nil)
    }
    defer publisher.Close()

    logger.Info(ctx, "All infrastructure initialized successfully", nil)

    // Start your service...
}
```

## Testing

All packages include mock implementations for testing:

```go
import (
    "github.com/giia/giia-core-engine/pkg/logger"
    "github.com/giia/giia-core-engine/pkg/database"
    "github.com/giia/giia-core-engine/pkg/events"
)

func TestUserService(t *testing.T) {
    // Setup mocks
    mockLogger := new(logger.LoggerMock)
    mockDB := new(database.DatabaseMock)
    mockPublisher := new(events.PublisherMock)

    // Configure expectations
    mockLogger.On("Info", mock.Anything, "User created", mock.Anything).Return()
    mockDB.On("Connect", mock.Anything, mock.Anything).Return(gormDB, nil)
    mockPublisher.On("Publish", mock.Anything, "users.events", mock.Anything).Return(nil)

    // Test your service
    service := NewUserService(mockLogger, mockDB, mockPublisher)
    err := service.CreateUser(ctx, userRequest)

    // Verify
    assert.NoError(t, err)
    mockLogger.AssertExpectations(t)
    mockDB.AssertExpectations(t)
    mockPublisher.AssertExpectations(t)
}
```

## Development Guidelines

### Error Handling
- Always use typed errors from the `errors` package
- Never ignore errors
- Wrap errors with context when propagating

### Logging
- Use structured logging with tags
- Include request_id in context
- Never log sensitive data (passwords, tokens)

### Configuration
- Validate required keys on startup
- Use environment-specific .env files
- Never commit secrets to version control

### Database
- Always use context for timeout support
- Configure connection pools based on workload
- Implement health checks

### Events
- Use durable subscriptions for critical processing
- Make event handlers idempotent
- Include organization_id for multi-tenancy

## Version History

- **v0.1.0** (2024-01-15) - Initial release
  - Config package with Viper
  - Logger package with Zerolog
  - Database package with GORM
  - Errors package with HTTP mapping
  - Events package with NATS Jetstream

## Support

For issues, questions, or contributions, please refer to the main GIIA Core Engine repository.
