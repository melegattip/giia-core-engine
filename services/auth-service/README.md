# GIIA Auth Service

Multi-tenant authentication and authorization service built with Clean Architecture principles, supporting JWT-based authentication with refresh tokens and organization-level tenant isolation.

## Features

- ✅ **Multi-tenancy**: Organization-based tenant isolation with automatic query filtering
- 🔐 **JWT Authentication**: Access tokens (15-min) + refresh tokens (7-day)
- 👤 **User Management**: Registration, activation, login, logout
- 🔄 **Token Refresh**: Automatic token renewal without re-authentication
- 📧 **Email Integration**: Activation emails with SMTP support
- 🚦 **Rate Limiting**: Redis-based rate limiting for login/register endpoints
- 🔒 **Security**: bcrypt password hashing, token blacklisting, password complexity validation
- 📊 **Structured Logging**: Zerolog-based JSON logging with context support

## Architecture

This service follows Clean Architecture with clear separation of concerns:

```
services/auth-service/
├── cmd/api/                       # Application entry point
├── internal/
│   ├── core/                      # Business logic layer
│   │   ├── domain/                # Entities and value objects
│   │   ├── usecases/              # Use cases (business logic)
│   │   └── providers/             # Interface contracts
│   └── infrastructure/            # External adapters
│       ├── adapters/              # External service implementations
│       │   ├── jwt/               # JWT token management
│       │   ├── email/             # SMTP email service
│       │   └── rate_limiter/      # Redis rate limiter
│       ├── repositories/          # Data access layer
│       └── entrypoints/           # HTTP handlers & middleware
│           └── http/
│               ├── handlers/      # HTTP request handlers
│               └── middleware/    # Middleware (auth, tenant, rate limit)
├── migrations/                    # Database migrations
└── docs/                         # Documentation

Dependencies on shared packages:
- pkg/config                       # Configuration management
- pkg/logger                       # Structured logging
- pkg/database                     # Database connection
- pkg/errors                       # Typed error system
```

## Database Schema

### Organizations
- `id` (UUID, PK)
- `name`, `slug` (unique)
- `status` (active/inactive/suspended)
- `settings` (JSONB)

### Users (Tenant-isolated)
- `id` (UUID, PK)
- `email` (unique per organization)
- `password` (bcrypt hashed)
- `organization_id` (UUID, FK) - **Tenant isolation key**
- `status` (active/inactive/suspended)
- `first_name`, `last_name`, `phone`, `avatar`
- `last_login_at`

### Refresh Tokens (Tenant-isolated)
- `id` (UUID, PK)
- `user_id` (UUID, FK)
- `token_hash` (SHA-256, indexed)
- `expires_at`
- `revoked` (boolean)

### Activation Tokens
- Similar structure for email activation flow

### Password Reset Tokens
- Similar structure for password reset flow

**Automatic Cleanup**: PostgreSQL function `cleanup_expired_tokens()` runs daily to remove expired tokens.

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "organization_id": "uuid-of-organization"
}

Response: 201 Created
{
  "message": "User registered successfully. Please check your email for activation instructions."
}
```

**Rate Limit**: 3 attempts per 60 minutes per IP

#### Activate Account
```http
POST /api/v1/auth/activate
Content-Type: application/json

{
  "token": "activation-token-from-email"
}

OR

GET /api/v1/auth/activate?token=activation-token-from-email

Response: 200 OK
{
  "message": "Account activated successfully. You can now log in."
}
```

#### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response: 200 OK
Set-Cookie: refresh_token=...; HttpOnly; Max-Age=604800

{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user": {
    "id": "user-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "org-uuid",
    "status": "active"
  }
}
```

**Rate Limit**: 5 attempts per 15 minutes per IP

#### Refresh Token
```http
POST /api/v1/auth/refresh

(Reads refresh_token from HTTP-only cookie or request body)

{
  "refresh_token": "optional-if-not-in-cookie"
}

Response: 200 OK
{
  "access_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### Protected Endpoints (Require Authentication)

All protected endpoints require `Authorization: Bearer <access_token>` header.

#### Logout
```http
POST /api/v1/auth/logout
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

Response: 200 OK
Set-Cookie: refresh_token=; Max-Age=-1

{
  "message": "Logged out successfully"
}
```

## Multi-Tenancy Implementation

### JWT Claims
Access tokens include organization context:
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "organization_id": "org-uuid",
  "roles": ["user"],
  "exp": 1234567890
}
```

### Automatic Tenant Filtering
The `TenantMiddleware` extracts `organization_id` from JWT claims and injects it into the request context. All repository queries automatically filter by organization using GORM scopes:

```go
// Automatically applied to all queries
query.Scopes(TenantScope(orgID))
```

This ensures complete data isolation between organizations without requiring explicit filtering in business logic.

## Environment Variables

Create a `.env` file in the service root:

```bash
# Server
PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=giia_auth
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key-min-32-chars
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h  # 7 days

# Email (SMTP)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@giia.com

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

## Quick Start

### 1. Install Dependencies

```bash
# From project root
cd services/auth-service
go mod download
```

### 2. Start Infrastructure

```bash
# Using Docker Compose (recommended)
docker-compose up -d postgres redis

# Or install PostgreSQL and Redis locally
```

### 3. Run Database Migrations

```bash
# Apply migrations in order
psql -U postgres -d giia_auth -f migrations/001_create_organizations.sql
psql -U postgres -d giia_auth -f migrations/002_add_org_to_users.sql
psql -U postgres -d giia_auth -f migrations/003_create_refresh_tokens.sql
psql -U postgres -d giia_auth -f migrations/004_create_password_reset_tokens.sql
psql -U postgres -d giia_auth -f migrations/005_create_activation_tokens.sql
```

**Note**: Migration 001 creates a default organization with ID `00000000-0000-0000-0000-000000000001` for testing.

### 4. Configure Environment

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 5. Run the Service

```bash
# Development mode
go run cmd/api/main.go

# Or build and run
go build -o bin/auth-service cmd/api/main.go
./bin/auth-service
```

The service will start on `http://localhost:8080`

### 6. Test the API

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test1234!",
    "first_name": "Test",
    "last_name": "User",
    "organization_id": "00000000-0000-0000-0000-000000000001"
  }'

# Check email for activation token, then activate
curl -X POST http://localhost:8080/api/v1/auth/activate \
  -H "Content-Type: application/json" \
  -d '{"token": "activation-token-from-email"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test1234!"
  }'
```

## Development

### Running Tests

```bash
# Run all tests
go test ./... -count=1

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./internal/core/usecases/auth/... -v
```

### Code Quality

```bash
# Run linters
golangci-lint run

# Format code
gofmt -w .

# Pre-commit checks
pre-commit run --all-files
```

## Security Considerations

### Password Requirements
- Minimum 8 characters
- Must contain uppercase letter
- Must contain lowercase letter
- Must contain digit
- Must contain special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Token Security
- **Access tokens**: Short-lived (15 minutes), included in JWT claims
- **Refresh tokens**: Longer-lived (7 days), stored hashed in database
- **Token blacklist**: Revoked access tokens stored in Redis with TTL
- **Token rotation**: Each refresh generates a new access token

### Rate Limiting
- **Login**: 5 attempts per 15 minutes per IP
- **Register**: 3 attempts per 60 minutes per IP
- Configurable per endpoint via middleware

### Data Protection
- Passwords hashed with bcrypt (cost 12)
- Tokens stored as SHA-256 hashes
- SQL injection prevention via GORM parameterized queries
- XSS protection via input validation

## Troubleshooting

### Common Issues

**Database connection fails**
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Verify connection
psql -U postgres -h localhost -d giia_auth

# Check environment variables
echo $DB_HOST $DB_PORT
```

**Redis connection fails**
```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 ping
```

**Email not sending**
- Verify SMTP credentials
- Check firewall/network settings
- For Gmail: Enable "Less secure app access" or use App Password
- Check logs for detailed error messages

**Rate limit always triggered**
```bash
# Clear rate limits in Redis
redis-cli KEYS "rate_limit:*" | xargs redis-cli DEL
```

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics (if Prometheus enabled)
```bash
curl http://localhost:8080/metrics
```

### Logs
Structured JSON logs are written to stdout. Use your preferred log aggregation tool (ELK, Datadog, etc.)

Example log entry:
```json
{
  "level": "info",
  "timestamp": "2025-12-09T10:30:00Z",
  "message": "User logged in successfully",
  "user_id": "uuid",
  "organization_id": "uuid",
  "email": "user@example.com"
}
```

## Project Status

See [TASK-05-PROGRESS.md](./TASK-05-PROGRESS.md) for detailed implementation progress and [IMPLEMENTATION-SUMMARY.md](./IMPLEMENTATION-SUMMARY.md) for technical summary.

## Related Documentation

- [Wiring Example](./WIRING-EXAMPLE.md) - Complete dependency injection setup
- [Shared Packages](../../pkg/README.md) - Common infrastructure packages
- [Development Guidelines](../../CLAUDE.md) - Project coding standards

## License

Part of GIIA Core Engine - Internal Use Only
