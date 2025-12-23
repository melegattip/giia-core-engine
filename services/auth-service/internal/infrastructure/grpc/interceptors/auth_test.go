package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	pkgLogger "github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateAccessToken(userID, orgID uuid.UUID, email string, roles []string) (string, error) {
	args := m.Called(userID, orgID, email, roles)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) ValidateAccessToken(tokenString string) (*providers.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.Claims), args.Error(1)
}

func (m *MockJWTManager) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.RegisteredClaims), args.Error(1)
}

func (m *MockJWTManager) GetAccessExpiry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockJWTManager) GetRefreshExpiry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func TestAuthInterceptor_UnaryInterceptor_PublicMethod(t *testing.T) {
	mockJWT := new(MockJWTManager)
	logger := pkgLogger.New("test", "info")
	interceptor := NewAuthInterceptor(mockJWT, logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/ValidateToken",
	}

	resp, err := interceptor.UnaryInterceptor()(context.Background(), nil, info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
	mockJWT.AssertNotCalled(t, "ValidateAccessToken")
}

func TestAuthInterceptor_UnaryInterceptor_MissingToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	logger := pkgLogger.New("test", "info")
	interceptor := NewAuthInterceptor(mockJWT, logger)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/CheckPermission",
	}

	resp, err := interceptor.UnaryInterceptor()(context.Background(), nil, info, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, statusErr.Code())
}

func TestAuthInterceptor_UnaryInterceptor_InvalidToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	logger := pkgLogger.New("test", "info")
	interceptor := NewAuthInterceptor(mockJWT, logger)

	mockJWT.On("ValidateAccessToken", "invalid_token").Return(nil, pkgErrors.NewUnauthorized("invalid token"))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/CheckPermission",
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer invalid_token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.UnaryInterceptor()(ctx, nil, info, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, statusErr.Code())
	mockJWT.AssertExpectations(t)
}

func TestAuthInterceptor_UnaryInterceptor_ValidToken(t *testing.T) {
	mockJWT := new(MockJWTManager)
	logger := pkgLogger.New("test", "info")
	interceptor := NewAuthInterceptor(mockJWT, logger)

	userID := uuid.New().String()
	orgID := uuid.New().String()
	claims := &providers.Claims{
		UserID:         userID,
		Email:          "test@example.com",
		OrganizationID: orgID,
		Roles:          []string{"admin"},
	}

	mockJWT.On("ValidateAccessToken", "valid_token").Return(claims, nil)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		ctxUserID := ctx.Value(contextKeyUserID).(string)
		ctxOrgID := ctx.Value(contextKeyOrganizationID).(string)
		ctxEmail := ctx.Value(contextKeyEmail).(string)

		assert.Equal(t, userID, ctxUserID)
		assert.Equal(t, orgID, ctxOrgID)
		assert.Equal(t, "test@example.com", ctxEmail)

		return "success", nil
	}

	info := &grpc.UnaryServerInfo{
		FullMethod: "/auth.v1.AuthService/CheckPermission",
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer valid_token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	resp, err := interceptor.UnaryInterceptor()(ctx, nil, info, handler)

	assert.NoError(t, err)
	assert.Equal(t, "success", resp)
	mockJWT.AssertExpectations(t)
}

func TestExtractToken_Success(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "Bearer test_token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractToken(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "test_token", token)
}

func TestExtractToken_MissingMetadata(t *testing.T) {
	ctx := context.Background()

	token, err := extractToken(ctx)

	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestExtractToken_MissingAuthorizationHeader(t *testing.T) {
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractToken(ctx)

	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestExtractToken_InvalidFormat(t *testing.T) {
	md := metadata.New(map[string]string{
		"authorization": "Basic test_token",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	token, err := extractToken(ctx)

	assert.Empty(t, token)
	assert.Error(t, err)
}

func TestIsPublicMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		expected bool
	}{
		{
			name:     "ValidateToken is public",
			method:   "/auth.v1.AuthService/ValidateToken",
			expected: true,
		},
		{
			name:     "Health Check is public",
			method:   "/grpc.health.v1.Health/Check",
			expected: true,
		},
		{
			name:     "CheckPermission is not public",
			method:   "/auth.v1.AuthService/CheckPermission",
			expected: false,
		},
		{
			name:     "Unknown method is not public",
			method:   "/service.v1.Service/Method",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPublicMethod(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}
