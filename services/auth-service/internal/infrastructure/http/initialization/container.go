package initialization

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/role"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/user"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/cache"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/email"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/events"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/time_manager"
	httpServer "github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/entrypoints/http/handlers"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/repositories"
)

type HTTPContainer struct {
	Server *httpServer.Server
	Logger pkgLogger.Logger
}

type HTTPConfig struct {
	Port             string
	JWTSecretKey     string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string
	SMTPHost         string
	SMTPPort         string
	SMTPUsername     string
	SMTPPassword     string
	SMTPFrom         string
	BaseURL          string
	DB               *gorm.DB
	RedisClient      *redis.Client
	NATSConn         *nats.Conn
	Logger           pkgLogger.Logger
}

func InitializeHTTPServer(cfg *HTTPConfig) (*HTTPContainer, error) {
	jwtManager := jwt.NewJWTManager(
		cfg.JWTSecretKey,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
		cfg.JWTIssuer,
	)

	userRepo := repositories.NewUserRepository(cfg.DB)
	orgRepo := repositories.NewOrganizationRepository(cfg.DB)
	roleRepo := repositories.NewRoleRepository(cfg.DB)
	permissionRepo := repositories.NewPermissionRepository(cfg.DB)
	tokenRepo := repositories.NewTokenRepository(cfg.RedisClient, cfg.DB)

	permissionCache := cache.NewRedisPermissionCache(cfg.RedisClient, cfg.Logger)
	timeManager := time_manager.NewTimeManager()

	var eventPublisher *events.NATSEventPublisher
	if cfg.NATSConn != nil {
		eventPublisher = events.NewNATSEventPublisher(cfg.NATSConn, cfg.Logger)
	}

	emailService := email.NewSMTPEmailService(&email.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUsername,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
	}, cfg.Logger)

	loginUseCase := auth.NewLoginUseCase(
		userRepo,
		tokenRepo,
		jwtManager,
		eventPublisher,
		timeManager,
		cfg.Logger,
	)

	registerUseCase := auth.NewRegisterUseCase(
		userRepo,
		orgRepo,
		tokenRepo,
		eventPublisher,
		timeManager,
		cfg.Logger,
	)

	refreshTokenUseCase := auth.NewRefreshTokenUseCase(
		userRepo,
		tokenRepo,
		jwtManager,
		cfg.Logger,
	)

	logoutUseCase := auth.NewLogoutUseCase(
		tokenRepo,
		jwtManager,
		cfg.Logger,
	)

	activateAccountUseCase := auth.NewActivateAccountUseCase(
		userRepo,
		tokenRepo,
		emailService,
		cfg.Logger,
	)

	requestPasswordResetUseCase := auth.NewRequestPasswordResetUseCase(
		userRepo,
		tokenRepo,
		emailService,
		cfg.Logger,
	)

	confirmPasswordResetUseCase := auth.NewConfirmPasswordResetUseCase(
		userRepo,
		tokenRepo,
		cfg.Logger,
	)

	activateUserUseCase := user.NewActivateUserUseCase(
		userRepo,
		permissionRepo,
		eventPublisher,
		timeManager,
		cfg.Logger,
	)

	deactivateUserUseCase := user.NewDeactivateUserUseCase(
		userRepo,
		permissionRepo,
		eventPublisher,
		timeManager,
		cfg.Logger,
	)

	resolveInheritanceUC := rbac.NewResolveInheritanceUseCase(roleRepo, permissionRepo, cfg.Logger)
	getUserPermissionsUC := rbac.NewGetUserPermissionsUseCase(
		roleRepo,
		resolveInheritanceUC,
		permissionCache,
		cfg.Logger,
	)
	checkPermissionUC := rbac.NewCheckPermissionUseCase(getUserPermissionsUC, cfg.Logger)
	batchCheckUC := rbac.NewBatchCheckPermissionsUseCase(checkPermissionUC, cfg.Logger)

	createRoleUseCase := role.NewCreateRoleUseCase(roleRepo, permissionRepo, cfg.Logger)
	updateRoleUseCase := role.NewUpdateRoleUseCase(roleRepo, permissionRepo, permissionCache, cfg.Logger)
	deleteRoleUseCase := role.NewDeleteRoleUseCase(roleRepo, permissionCache, cfg.Logger)
	assignRoleUseCase := role.NewAssignRoleUseCase(roleRepo, userRepo, permissionCache, eventPublisher, timeManager, cfg.Logger)

	authHandler := handlers.NewAuthHandler(
		loginUseCase,
		registerUseCase,
		refreshTokenUseCase,
		logoutUseCase,
		activateAccountUseCase,
		requestPasswordResetUseCase,
		confirmPasswordResetUseCase,
		cfg.Logger,
	)

	userHandler := handlers.NewUserHandler(
		activateUserUseCase,
		deactivateUserUseCase,
		cfg.Logger,
	)

	roleHandler := handlers.NewRoleHandler(
		assignRoleUseCase,
		createRoleUseCase,
		updateRoleUseCase,
		deleteRoleUseCase,
		cfg.Logger,
	)

	permissionHandler := handlers.NewPermissionHandler(
		checkPermissionUC,
		batchCheckUC,
		cfg.Logger,
	)

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("jwtManager", jwtManager)
		c.Next()
	})

	httpServer.SetupRoutes(router, &httpServer.RouteConfig{
		AuthHandler:       authHandler,
		UserHandler:       userHandler,
		RoleHandler:       roleHandler,
		PermissionHandler: permissionHandler,
		Logger:            cfg.Logger,
	})

	server := httpServer.NewServer(cfg.Port, router, cfg.Logger)

	return &HTTPContainer{
		Server: server,
		Logger: cfg.Logger,
	}, nil
}
