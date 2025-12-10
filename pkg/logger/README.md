# Logger Package

Structured JSON logging using Zerolog with context support and log level filtering.

## Features

- Structured JSON logging for production
- Log levels: debug, info, warn, error, fatal
- Context-aware logging (request ID extraction)
- Custom tags/fields per log entry
- Configurable output (stdout, file)
- Mock implementation for testing

## Installation

```go
import "github.com/giia/giia-core-engine/pkg/logger"
```

## Usage

### Basic Logging

```go
// Initialize logger
log := logger.New("auth-service", "info")

// Log with tags
log.Info(ctx, "User authenticated successfully", logger.Tags{
    "user_id": userID,
    "method": "oauth2",
})

log.Error(ctx, err, "Failed to save user session", logger.Tags{
    "user_id": userID,
    "session_id": sessionID,
})
```

### Log Levels

```go
// Debug - Detailed information for debugging
log.Debug(ctx, "Validating request payload", logger.Tags{"payload_size": len(data)})

// Info - General informational messages
log.Info(ctx, "Processing user request", logger.Tags{"user_id": 123})

// Warn - Warning messages for potential issues
log.Warn(ctx, "Rate limit approaching threshold", logger.Tags{"requests": 95, "limit": 100})

// Error - Error messages with error context
log.Error(ctx, err, "Database query failed", logger.Tags{"query": "SELECT * FROM users"})

// Fatal - Critical errors that terminate the application
log.Fatal(ctx, err, "Failed to connect to database", logger.Tags{"host": "localhost"})
```

### Context-Aware Logging

```go
// Add request ID to context
ctx = logger.WithRequestID(ctx, "req-123-456")

// Request ID is automatically included in all logs
log.Info(ctx, "Processing request", logger.Tags{"user_id": 123})
// Output: {"level":"info","service":"auth-service","request_id":"req-123-456","user_id":123,"message":"Processing request"}
```

### Console Logger (Development)

```go
// Human-readable console output for local development
log := logger.NewConsoleLogger("auth-service")
```

### File Logging

```go
log, err := logger.NewWithConfig("auth-service", "info", true, "/var/log/app.log")
if err != nil {
    log.Fatal(ctx, err, "Failed to initialize file logger", nil)
}
```

### Log Output Format

```json
{
  "level": "info",
  "service": "auth-service",
  "request_id": "req-123-456",
  "user_id": 12345,
  "timestamp": "2024-01-15T10:30:45Z",
  "message": "User authenticated successfully"
}
```

### Testing with Mocks

```go
mockLogger := new(logger.LoggerMock)
mockLogger.On("Info", mock.Anything, "User created", mock.Anything).Return()

service := NewUserService(mockLogger)
service.CreateUser(ctx, userRequest)

mockLogger.AssertExpectations(t)
```

## Environment Configuration

```bash
# Set log level via environment variable
export LOG_LEVEL=debug   # debug, info, warn, error, fatal

# Service name
export SERVICE_NAME=auth-service
```

## Best Practices

1. **Always pass context** to logging methods for request tracing
2. **Use structured tags** instead of string interpolation
3. **Log errors with context**, not just error messages
4. **Use appropriate log levels** (don't log everything as error)
5. **Include relevant metadata** (user_id, request_id, operation)
6. **Never log sensitive data** (passwords, tokens, PII)
