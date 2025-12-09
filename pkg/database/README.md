# Database Package

PostgreSQL database connection management using GORM with connection pooling, retry logic, and health checks.

## Features

- GORM-based database connections
- Automatic retry with exponential backoff (max 5 retries)
- Configurable connection pooling
- Health check support
- Graceful connection closure
- Mock implementation for testing

## Installation

```go
import "github.com/giia/giia-core-engine/pkg/database"
```

## Usage

### Basic Connection

```go
db := database.New()

config := &database.Config{
    Host:         "localhost",
    Port:         5432,
    User:         "postgres",
    Password:     "secret",
    DatabaseName: "giia_db",
    SSLMode:      "disable",
}

conn, err := db.Connect(context.Background(), config)
if err != nil {
    log.Fatal(err)
}
defer db.Close(conn)
```

### Connection with Custom Pool Settings

```go
config := &database.Config{
    Host:            "localhost",
    Port:            5432,
    User:            "postgres",
    Password:        "secret",
    DatabaseName:    "giia_db",
    SSLMode:         "require",
    MaxOpenConns:    50,           // Default: 25
    MaxIdleConns:    10,           // Default: 5
    ConnMaxLifetime: 10 * time.Minute, // Default: 5 minutes
}

conn, err := db.Connect(ctx, config)
```

### Connection with DSN String

```go
dsn := "host=localhost port=5432 user=postgres password=secret dbname=giia_db sslmode=disable"
conn, err := database.ConnectWithDSN(context.Background(), dsn)
if err != nil {
    log.Fatal(err)
}
```

### Health Check

```go
db := database.New()

if err := db.HealthCheck(ctx, conn); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

### Using Health Checker

```go
healthChecker := database.NewHealthChecker()

if err := healthChecker.Check(ctx, conn); err != nil {
    log.Printf("Health check failed: %v", err)
}
```

### Retry Logic

The package automatically retries failed connections up to 5 times with exponential backoff:

- Initial backoff: 1 second
- Maximum backoff: 30 seconds
- Backoff multiplier: 2x per retry

### Testing with Mocks

```go
mockDB := new(database.DatabaseMock)
mockDB.On("Connect", mock.Anything, mock.Anything).Return(gormDB, nil)
mockDB.On("HealthCheck", mock.Anything, gormDB).Return(nil)

service := NewUserService(mockDB)
```

## Configuration

### Environment Variables

```bash
# Database Configuration
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USER=postgres
export DATABASE_PASSWORD=secret
export DATABASE_NAME=giia_db
export DATABASE_SSL_MODE=disable

# Connection Pool
export DATABASE_MAX_OPEN_CONNS=25
export DATABASE_MAX_IDLE_CONNS=5
export DATABASE_CONN_MAX_LIFETIME=5m
```

### Loading from Config Package

```go
import (
    "github.com/giia/giia-core-engine/pkg/config"
    "github.com/giia/giia-core-engine/pkg/database"
)

cfg, _ := config.New("GIIA")

dbConfig := &database.Config{
    Host:         cfg.GetString("database.host"),
    Port:         cfg.GetInt("database.port"),
    User:         cfg.GetString("database.user"),
    Password:     cfg.GetString("database.password"),
    DatabaseName: cfg.GetString("database.name"),
    SSLMode:      cfg.GetString("database.ssl_mode"),
}

db := database.New()
conn, err := db.Connect(ctx, dbConfig)
```

## Best Practices

1. **Always use context** for timeout and cancellation support
2. **Configure connection pools** based on workload (don't use defaults blindly)
3. **Implement health checks** in your readiness probes
4. **Close connections gracefully** on shutdown
5. **Use transactions** for multi-step operations
6. **Monitor slow queries** and optimize accordingly
7. **Retry transient errors** but fail fast on permanent errors

## Connection Pool Recommendations

| Workload | Max Open | Max Idle | Lifetime |
|----------|----------|----------|----------|
| Light (<100 QPS) | 10 | 2 | 5m |
| Medium (100-1000 QPS) | 25 | 5 | 5m |
| Heavy (>1000 QPS) | 50-100 | 10-20 | 10m |
