# Config Package

Configuration management using Viper with support for environment variables, .env files, and type-safe getters.

## Features

- Load configuration from multiple sources (.env files, environment variables)
- Environment variable overrides
- Type-safe getters (String, Int, Bool, Float64)
- Required key validation
- Hierarchical configuration support

## Installation

```go
import "github.com/giia/giia-core-engine/pkg/config"
```

## Usage

### Basic Configuration

```go
// Initialize config with environment prefix
cfg, err := config.New("GIIA")
if err != nil {
    log.Fatal(err)
}

// Get configuration values
dbHost := cfg.GetString("database.host")
dbPort := cfg.GetInt("database.port")
enableCache := cfg.GetBool("cache.enabled")
timeout := cfg.GetFloat64("timeout")
```

### Validate Required Keys

```go
requiredKeys := []string{
    "database.host",
    "database.port",
    "database.user",
    "database.password",
    "database.name",
}

if err := cfg.Validate(requiredKeys); err != nil {
    log.Fatalf("Missing required configuration: %v", err)
}
```

### Configuration Sources

Configuration is loaded in the following order (later sources override earlier ones):

1. `.env` file in the current directory
2. Environment variables (with optional prefix)

### Environment Variables

```bash
# .env file
DATABASE_HOST=localhost
DATABASE_PORT=5432
LOG_LEVEL=info

# Or export as environment variables
export GIIA_DATABASE_HOST=prod-db.example.com
export GIIA_DATABASE_PORT=5432
export GIIA_LOG_LEVEL=error
```

### Example .env File

```env
# Database Configuration
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=secret
DATABASE_NAME=giia_db
DATABASE_SSL_MODE=disable

# Application Configuration
APP_NAME=giia-service
APP_PORT=8080
LOG_LEVEL=info

# External Services
NATS_URL=nats://localhost:4222
REDIS_URL=redis://localhost:6379
```

## Best Practices

1. **Use environment-specific .env files** (`.env.development`, `.env.production`)
2. **Never commit sensitive values** to version control
3. **Validate required keys on startup** to fail fast
4. **Use consistent naming conventions** (UPPER_SNAKE_CASE for env vars)
5. **Document all configuration keys** in your service README
