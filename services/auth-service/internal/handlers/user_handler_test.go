package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/auth-service/internal/domain"
	"github.com/giia/giia-core-engine/services/auth-service/internal/infrastructure/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.UserResponse, *auth.TokenPair, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*domain.UserResponse), args.Get(1).(*auth.TokenPair), args.Error(2)
}

func (m *MockUserService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.UserResponse, *auth.TokenPair, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*domain.UserResponse), args.Get(1).(*auth.TokenPair), args.Error(2)
}

func (m *MockUserService) Logout(ctx context.Context, userID uint, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *MockUserService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID uint, update *domain.User) (*domain.UserResponse, error) {
	args := m.Called(ctx, userID, update)
	if args.Get(0) == nil {
		return nil, args.Error(2)
	}
	return args.Get(0).(*domain.UserResponse), args.Error(2)
}

func (m *MockUserService) UpdateAvatar(ctx context.Context, userID uint, avatarPath string) error {
	args := m.Called(ctx, userID, avatarPath)
	return args.Error(0)
}

func (m *MockUserService) GetPreferences(ctx context.Context, userID uint) (*domain.Preferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Preferences), args.Error(1)
}

func (m *MockUserService) UpdatePreferences(ctx context.Context, userID uint, prefs *domain.Preferences) error {
	args := m.Called(ctx, userID, prefs)
	return args.Error(0)
}

func (m *MockUserService) GetNotifications(ctx context.Context, userID uint) (*domain.NotificationSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotificationSettings), args.Error(1)
}

func (m *MockUserService) UpdateNotifications(ctx context.Context, userID uint, notif *domain.NotificationSettings) error {
	args := m.Called(ctx, userID, notif)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(ctx context.Context, userID uint, req *domain.ChangePasswordRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockUserService) Setup2FA(ctx context.Context, userID uint) (*auth.TwoFASetup, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TwoFASetup), args.Error(1)
}

func (m *MockUserService) GenerateQRCode(ctx context.Context, userID uint) ([]byte, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockUserService) Enable2FA(ctx context.Context, userID uint, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *MockUserService) Disable2FA(ctx context.Context, userID uint, password string) error {
	args := m.Called(ctx, userID, password)
	return args.Error(0)
}

func (m *MockUserService) Verify2FA(ctx context.Context, userID uint, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *MockUserService) VerifyEmailWithToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserService) RequestPasswordReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func (m *MockUserService) ExportData(ctx context.Context, userID uint) (string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockUserService) DeleteAccount(ctx context.Context, userID uint, password string) error {
	args := m.Called(ctx, userID, password)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create mock service
	mockService := &MockUserService{}
	h := NewUserHandler(mockService)
	r.POST("/users/register", h.Register)

	// Setup mock expectations
	expectedUser := &domain.UserResponse{
		ID:         1,
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Phone:      "+1234567890",
		IsActive:   true,
		IsVerified: false,
	}

	expectedTokens := &auth.TokenPair{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	mockService.On("Register", mock.Anything, mock.AnythingOfType("*domain.RegisterRequest")).Return(expectedUser, expectedTokens, nil)

	payload := map[string]interface{}{
		"email":      "test@example.com",
		"password":   "password123",
		"first_name": "Test",
		"last_name":  "User",
		"phone":      "+1234567890",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/users/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code) // Register should return 201 Created
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["user"])

	// Verify mock was called
	mockService.AssertExpectations(t)
}
