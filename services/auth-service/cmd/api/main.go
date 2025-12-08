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
	"github.com/joho/godotenv"

	"users-service/internal/handlers"
	"users-service/internal/infrastructure/auth"
	"users-service/internal/infrastructure/config"
	"users-service/internal/infrastructure/middleware"
	"users-service/internal/repository"
	"users-service/internal/usecases"
	"users-service/pkg/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Log the database configuration
	log.Printf("🔧 [Database] Config: Host=%s, Port=%s, User=%s, DBName=%s, SSLMode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.DBName, cfg.Database.SSLMode)

	// Initialize database
	dbConfig := database.Config{
		DSN:             cfg.GetDatabaseDSN(),
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 2 * time.Minute,
	}

	log.Printf("🔧 [Database] DSN: %s", dbConfig.DSN)

	// Adjust connection pool settings based on pool mode
	if cfg.Database.PoolMode == "transaction" {
		// For transaction pooling, we can use more connections
		// since the pooler handles the actual database connections
		dbConfig.MaxOpenConns = 50
		dbConfig.MaxIdleConns = 10
		log.Printf("🔧 [Database] Using transaction pooling mode (MaxOpenConns: %d, MaxIdleConns: %d)",
			dbConfig.MaxOpenConns, dbConfig.MaxIdleConns)
	} else {
		// Use standard connection pool settings
		dbConfig.MaxOpenConns = 25
		dbConfig.MaxIdleConns = 5
		log.Printf("🔧 [Database] Using standard connection pool (MaxOpenConns: %d, MaxIdleConns: %d)",
			dbConfig.MaxOpenConns, dbConfig.MaxIdleConns)
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize services
	jwtService := auth.NewJWTService(
		cfg.JWT.SecretKey,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
		cfg.JWT.Issuer,
	)
	passwordService := auth.NewPasswordService(cfg.Security.PasswordMinLength)
	twoFAService := auth.NewTwoFAService(cfg.JWT.Issuer)

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize use cases
	userService := usecases.NewUserService(
		userRepo,
		jwtService,
		passwordService,
		twoFAService,
		cfg.Security.MaxLoginAttempts,
		cfg.Security.LockoutDuration,
	)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup Gin
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware (basic)
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, x-caller-id")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", userHandler.HealthCheck)

	// API v1 group
	api := r.Group("/api/v1")

	// Serve static files (avatars) - publicly accessible
	api.Static("/uploads", "./uploads")

	// Public endpoints (no authentication required)
	public := api.Group("/users")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
		public.POST("/check-2fa", userHandler.Check2FA)
		public.POST("/refresh", userHandler.Refresh)
		public.GET("/verify-email/:token", userHandler.VerifyEmail)
		public.POST("/request-password-reset", userHandler.RequestPasswordReset)
		public.POST("/reset-password", userHandler.ResetPassword)
	}

	// Auth endpoints (for compatibility with main API)
	auth := api.Group("/auth")
	{
		auth.POST("/login", userHandler.Login)
		auth.POST("/register", userHandler.Register)
		auth.POST("/check-2fa", userHandler.Check2FA)
		auth.POST("/refresh", userHandler.Refresh)
		auth.PUT("/change-password", userHandler.ChangePassword)
	}

	// Protected endpoints (authentication required)
	protected := api.Group("/users")
	protected.Use(authMiddleware.RequireAuth())
	{
		// Profile management
		protected.GET("/profile", userHandler.GetProfile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.POST("/logout", userHandler.Logout)
		protected.POST("/avatar", userHandler.UploadAvatar)

		// Preferences
		protected.GET("/preferences", userHandler.GetPreferences)
		protected.PUT("/preferences", userHandler.UpdatePreferences)

		// Notification settings
		protected.GET("/notifications/settings", userHandler.GetNotifications)
		protected.PUT("/notifications/settings", userHandler.UpdateNotifications)

		// Security
		protected.PUT("/security/change-password", userHandler.ChangePassword)

		// 2FA endpoints
		twoFA := protected.Group("/security/2fa")
		{
			twoFA.POST("/setup", userHandler.Setup2FA)
			twoFA.POST("/enable", userHandler.Enable2FA)
			twoFA.POST("/disable", userHandler.Disable2FA)
			twoFA.POST("/verify", userHandler.Verify2FA)
		}

		// Data management
		protected.POST("/export", userHandler.ExportData)
		protected.DELETE("", userHandler.DeleteAccount)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 [Users Service] Starting server on %s", cfg.GetServerAddr())
		log.Printf("🌍 [Users Service] Environment: %s", cfg.Server.Environment)
		log.Printf("📊 [Users Service] Health check: http://%s/health", cfg.GetServerAddr())

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 [Users Service] Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ [Users Service] Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ [Users Service] Server shutdown gracefully")
	}
}
