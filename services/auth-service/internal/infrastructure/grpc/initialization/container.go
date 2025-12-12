package initialization

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/cache"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/adapters/jwt"
	grpcServer "github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/grpc/server"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/repositories"
)

type GRPCContainer struct {
	Server *grpcServer.GRPCServer
	Logger pkgLogger.Logger
}

type GRPCConfig struct {
	Port             string
	JWTSecretKey     string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string
	DB               *gorm.DB
	RedisClient      *redis.Client
	Logger           pkgLogger.Logger
}

func InitializeGRPCServer(cfg *GRPCConfig) (*GRPCContainer, error) {
	jwtManager := jwt.NewJWTManager(
		cfg.JWTSecretKey,
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
		cfg.JWTIssuer,
	)

	userRepo := repositories.NewUserRepository(cfg.DB)
	roleRepo := repositories.NewRoleRepository(cfg.DB)
	permissionRepo := repositories.NewPermissionRepository(cfg.DB)

	permissionCache := cache.NewRedisPermissionCache(cfg.RedisClient, cfg.Logger)

	resolveInheritanceUC := rbac.NewResolveInheritanceUseCase(roleRepo, permissionRepo, cfg.Logger)
	getUserPermissionsUC := rbac.NewGetUserPermissionsUseCase(
		roleRepo,
		resolveInheritanceUC,
		permissionCache,
		cfg.Logger,
	)
	checkPermissionUC := rbac.NewCheckPermissionUseCase(getUserPermissionsUC, cfg.Logger)
	batchCheckUC := rbac.NewBatchCheckPermissionsUseCase(checkPermissionUC, cfg.Logger)

	validateTokenUC := auth.NewValidateTokenUseCase(userRepo, jwtManager, cfg.Logger)

	server, err := grpcServer.NewGRPCServer(
		cfg.Port,
		validateTokenUC,
		checkPermissionUC,
		batchCheckUC,
		getUserPermissionsUC,
		userRepo,
		cfg.DB,
		cfg.RedisClient,
		cfg.Logger,
	)
	if err != nil {
		return nil, err
	}

	return &GRPCContainer{
		Server: server,
		Logger: cfg.Logger,
	}, nil
}
