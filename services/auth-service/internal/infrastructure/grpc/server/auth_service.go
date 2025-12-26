package server

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	authv1 "github.com/melegattip/giia-core-engine/services/auth-service/api/proto/gen/go/auth/v1"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/auth"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/usecases/rbac"
)

type AuthServiceServer struct {
	authv1.UnimplementedAuthServiceServer
	validateTokenUC      *auth.ValidateTokenUseCase
	checkPermissionUC    *rbac.CheckPermissionUseCase
	batchCheckUC         *rbac.BatchCheckPermissionsUseCase
	getUserPermissionsUC *rbac.GetUserPermissionsUseCase
	userRepo             providers.UserRepository
	logger               pkgLogger.Logger
}

func NewAuthServiceServer(
	validateTokenUC *auth.ValidateTokenUseCase,
	checkPermissionUC *rbac.CheckPermissionUseCase,
	batchCheckUC *rbac.BatchCheckPermissionsUseCase,
	getUserPermissionsUC *rbac.GetUserPermissionsUseCase,
	userRepo providers.UserRepository,
	logger pkgLogger.Logger,
) *AuthServiceServer {
	return &AuthServiceServer{
		validateTokenUC:      validateTokenUC,
		checkPermissionUC:    checkPermissionUC,
		batchCheckUC:         batchCheckUC,
		getUserPermissionsUC: getUserPermissionsUC,
		userRepo:             userRepo,
		logger:               logger,
	}
}

func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	result, err := s.validateTokenUC.Execute(ctx, req.Token)
	if err != nil {
		s.logger.Error(ctx, err, "Token validation failed", pkgLogger.Tags{})
		return nil, status.Error(codes.Internal, "token validation failed")
	}

	if !result.Valid {
		return &authv1.ValidateTokenResponse{
			Valid:  false,
			Reason: "invalid or expired token",
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		Valid: true,
		User: &authv1.UserInfo{
			UserId:         result.UserID,
			OrganizationId: result.OrganizationID.String(),
			Email:          result.Email,
			Roles:          result.Roles,
		},
		ExpiresAt: result.ExpiresAt,
	}, nil
}

func (s *AuthServiceServer) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Permission == "" {
		return nil, status.Error(codes.InvalidArgument, "permission is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	allowed, err := s.checkPermissionUC.Execute(ctx, userID, req.Permission)
	if err != nil {
		s.logger.Error(ctx, err, "Permission check failed", pkgLogger.Tags{
			"user_id":    req.UserId,
			"permission": req.Permission,
		})
		return nil, status.Error(codes.Internal, "permission check failed")
	}

	reason := "allowed"
	if !allowed {
		reason = "permission denied"
	}

	return &authv1.CheckPermissionResponse{
		Allowed: allowed,
		Reason:  reason,
	}, nil
}

func (s *AuthServiceServer) BatchCheckPermissions(ctx context.Context, req *authv1.BatchCheckPermissionsRequest) (*authv1.BatchCheckPermissionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.Permissions) == 0 {
		return nil, status.Error(codes.InvalidArgument, "permissions list cannot be empty")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	resultsMap, err := s.batchCheckUC.Execute(ctx, userID, req.Permissions)
	if err != nil {
		s.logger.Error(ctx, err, "Batch permission check failed", pkgLogger.Tags{
			"user_id":           req.UserId,
			"permissions_count": len(req.Permissions),
		})
		return nil, status.Error(codes.Internal, "batch permission check failed")
	}

	results := make([]bool, len(req.Permissions))
	for i, perm := range req.Permissions {
		results[i] = resultsMap[perm]
	}

	return &authv1.BatchCheckPermissionsResponse{
		Results: results,
	}, nil
}

func (s *AuthServiceServer) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to get user", pkgLogger.Tags{
			"user_id": req.UserId,
		})
		return nil, translateError(err)
	}

	if req.OrganizationId != "" {
		orgID, err := uuid.Parse(req.OrganizationId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid organization_id format")
		}
		if user.OrganizationID != orgID {
			return nil, status.Error(codes.PermissionDenied, "user belongs to different organization")
		}
	}

	return &authv1.GetUserResponse{
		User: &authv1.UserInfo{
			UserId:         user.IDString(),
			OrganizationId: user.OrganizationID.String(),
			Email:          user.Email,
			Name:           user.FirstName + " " + user.LastName,
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			Status:         string(user.Status),
		},
	}, nil
}

func translateError(err error) error {
	if customErr, ok := err.(*pkgErrors.CustomError); ok {
		switch customErr.ErrorCode {
		case "BAD_REQUEST":
			return status.Error(codes.InvalidArgument, customErr.Message)
		case "UNAUTHORIZED":
			return status.Error(codes.Unauthenticated, customErr.Message)
		case "FORBIDDEN":
			return status.Error(codes.PermissionDenied, customErr.Message)
		case "NOT_FOUND":
			return status.Error(codes.NotFound, customErr.Message)
		default:
			return status.Error(codes.Internal, customErr.Message)
		}
	}
	return status.Error(codes.Internal, "internal server error")
}
