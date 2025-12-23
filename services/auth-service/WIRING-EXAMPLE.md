# Auth Service - Dependency Wiring Example

This document shows how to wire the new multi-tenant auth components in `cmd/api/main.go`.

## Complete main.go Example

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// Shared packages
	pkgConfig "github.com/melegattip/giia-core-engine/pkg/config"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"

	// Domain
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"

	// Use cases
	authUseCases "github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/auth"

	// Infrastructure
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/handlers"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/middleware"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/infrastructure/repositories"
)

func main() {
	ctx := context.Background()

	// 1. Load Configuration
	cfg, err := pkgConfig.New("AUTH_SERVICE")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required configuration
	requiredKeys := []string{
		"database.host",
		"database.port",
		"database.user",
		"database.password",
		"database.name",
		"redis.host",
		"redis.port",
		"jwt.secret",
	}
	if err := cfg.Validate(requiredKeys); err != nil {
		log.Fatalf("Missing required config: %v", err)
	}

	// 2. Initialize Logger
	logLevel := cfg.GetString("log.level")
	if logLevel == "" {
		logLevel = "info"
	}
	logger := pkgLogger.New("auth-service", logLevel)
	logger.Info(ctx, "Starting auth-service", nil)

	// 3. Initialize Database (PostgreSQL with GORM)
	dsn := buildDatabaseDSN(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		logger.Fatal(ctx, err, "Failed to connect to database", nil)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal(ctx, err, "Failed to get database instance", nil)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Auto-migrate tables (development only - use migrations in production)
	if err := db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.RefreshToken{},
		&domain.PasswordResetToken{},
		&domain.ActivationToken{},
	); err != nil {
		logger.Fatal(ctx, err, "Failed to run migrations", nil)
	}

	logger.Info(ctx, "Database connected successfully", nil)

	// 4. Initialize Redis
	redisHost := cfg.GetString("redis.host")
	redisPort := cfg.GetString("redis.port")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: cfg.GetString("redis.password"),
		DB:       cfg.GetInt("redis.db"),
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal(ctx, err, "Failed to connect to Redis", nil)
	}
	logger.Info(ctx, "Redis connected successfully", nil)

	// 5. Initialize JWT Manager
	jwtSecret := cfg.GetString("jwt.secret")
	accessExpiry := 15 * time.Minute  // 15 minutes
	refreshExpiry := 7 * 24 * time.Hour // 7 days
	jwtManager := jwt.NewJWTManager(jwtSecret, accessExpiry, refreshExpiry, "auth-service")

	// 6. Initialize Repositories
	orgRepo := repositories.NewOrganizationRepository(db)
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(redisClient, db)

	// 7. Initialize Use Cases
	loginUseCase := authUseCases.NewLoginUseCase(userRepo, tokenRepo, jwtManager, logger)
	registerUseCase := authUseCases.NewRegisterUseCase(userRepo, orgRepo, tokenRepo, logger)
	refreshTokenUseCase := authUseCases.NewRefreshTokenUseCase(userRepo, tokenRepo, jwtManager, logger)
	logoutUseCase := authUseCases.NewLogoutUseCase(tokenRepo, jwtManager, logger)

	// 8. Initialize HTTP Handlers
	authHandler := handlers.NewAuthHandler(
		loginUseCase,
		registerUseCase,
		refreshTokenUseCase,
		logoutUseCase,
		logger,
	)

	// 9. Initialize Middleware
	tenantMiddleware := middleware.NewTenantMiddleware(jwtManager)

	// 10. Setup Gin Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	api := r.Group("/api/v1")

	// Public auth endpoints (no authentication required)
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// Protected auth endpoints (authentication required)
	authProtected := api.Group("/auth")
	authProtected.Use(tenantMiddleware.ExtractTenantContext())
	{
		authProtected.POST("/logout", authHandler.Logout)
	}

	// Protected user endpoints
	usersProtected := api.Group("/users")
	usersProtected.Use(tenantMiddleware.ExtractTenantContext())
	{
		// Add user endpoints here (profile, etc.)
	}

	// 11. Start HTTP Server
	serverAddr := cfg.GetString("server.addr")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info(ctx, "Starting HTTP server", pkgLogger.Tags{
			"addr": serverAddr,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, err, "Failed to start server", nil)
		}
	}()

	// 12. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "Shutting down server...", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, err, "Server forced to shutdown", nil)
	} else {
		logger.Info(ctx, "Server shutdown gracefully", nil)
	}

	// Close database connection
	if sqlDB != nil {
		sqlDB.Close()
	}

	// Close Redis connection
	if redisClient != nil {
		redisClient.Close()
	}
}

func buildDatabaseDSN(cfg pkgConfig.Config) string {
	host := cfg.GetString("database.host")
	port := cfg.GetString("database.port")
	user := cfg.GetString("database.user")
	password := cfg.GetString("database.password")
	dbName := cfg.GetString("database.name")
	sslMode := cfg.GetString("database.sslmode")

	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)
}
```

## Environment Variables

Create a `.env` file with:

```bash
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=giia_db
DATABASE_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Server
SERVER_ADDR=:8080
LOG_LEVEL=info
```

## Testing the API

### 1. Create an Organization
```bash
# First, seed a default organization via SQL or create one
psql -h localhost -U postgres -d giia_db -c "
INSERT INTO organizations (name, slug, status)
VALUES ('Test Company', 'test-company', 'active')
ON CONFLICT (slug) DO NOTHING;
"
```

### 2. Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "organization_id": "<org-uuid-from-database>"
  }'
```

### 3. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@test.com",
    "password": "SecurePass123!"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@test.com",
    "organization_id": "org-uuid",
    ...
  }
}
```

### 4. Use Access Token
```bash
curl -X GET http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer <access-token>"
```

### 5. Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "<refresh-token-from-cookie-or-response>"
  }'
```

### 6. Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <access-token>"
```

## Multi-Tenancy Verification

1. Create two organizations
2. Create users in each organization
3. User A logs in → receives JWT with org_a_id
4. User A tries to access User B's data → Should get 403 Forbidden
5. Verify database queries include `WHERE organization_id = org_a_id`

## Next Steps

1. Run database migrations from `internal/infrastructure/persistence/migrations/`
2. Update existing `cmd/api/main.go` with new wiring
3. Test all endpoints
4. Add email service for activation/password reset
5. Add rate limiting middleware
6. Add comprehensive integration tests
