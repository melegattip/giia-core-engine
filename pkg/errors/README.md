# Errors Package

Typed error system with HTTP status code mapping for consistent API error responses.

## Features

- Typed error constructors for common HTTP status codes
- Error wrapping with context preservation
- JSON serialization for HTTP responses
- Error code constants for standardization

## Installation

```go
import "github.com/giia/giia-core-engine/pkg/errors"
```

## Usage

### Creating Typed Errors

```go
// Client errors (4xx)
err := errors.NewBadRequest("invalid user ID format")
err := errors.NewUnauthorized("authentication required")
err := errors.NewForbidden("insufficient permissions")
err := errors.NewNotFound("user not found")

// Server errors (5xx)
err := errors.NewInternalServerError("database connection failed")
err := errors.NewServiceUnavailable("service temporarily unavailable")
```

### Error Wrapping

```go
result, err := repository.GetUser(ctx, userID)
if err != nil {
    return errors.Wrap(err, "failed to fetch user from repository")
}
```

### HTTP Response Serialization

```go
func errorHandler(w http.ResponseWriter, err error) {
    response := errors.ToHTTPResponse(err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteStatus(response.StatusCode)
    json.NewEncoder(w).Encode(response)
}
```

### Error Response Format

```json
{
  "status_code": 404,
  "error_code": "NOT_FOUND",
  "message": "user not found"
}
```

## Error Codes

| Error Code | HTTP Status | Constructor |
|------------|-------------|-------------|
| `BAD_REQUEST` | 400 | `NewBadRequest()` |
| `UNAUTHORIZED` | 401 | `NewUnauthorized()` |
| `FORBIDDEN` | 403 | `NewForbidden()` |
| `NOT_FOUND` | 404 | `NewNotFound()` |
| `INTERNAL_SERVER_ERROR` | 500 | `NewInternalServerError()` |
| `SERVICE_UNAVAILABLE` | 503 | `NewServiceUnavailable()` |

## Best Practices

1. **Use specific error constructors** instead of generic `fmt.Errorf`
2. **Wrap errors with context** when propagating up the call stack
3. **Validate early** and return typed errors immediately
4. **Log the full error chain** but return sanitized messages to clients
