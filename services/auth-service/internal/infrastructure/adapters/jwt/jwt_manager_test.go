package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

func TestNewJWTManager(t *testing.T) {
	// Given
	givenSecretKey := "test-secret-key"
	givenAccessExpiry := 15 * time.Minute
	givenRefreshExpiry := 24 * time.Hour
	givenIssuer := "auth-service"

	// When
	manager := NewJWTManager(givenSecretKey, givenAccessExpiry, givenRefreshExpiry, givenIssuer)

	// Then
	assert.NotNil(t, manager)
	assert.IsType(t, &JWTManager{}, manager)
}

func TestJWTManager_GenerateAccessToken_WithValidInput_ReturnsToken(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRoles := []string{"admin", "editor"}

	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	// When
	token, err := manager.GenerateAccessToken(givenUserID, givenOrgID, givenEmail, givenRoles)

	// Then
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_GenerateAccessToken_TokenContainsCorrectClaims(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRoles := []string{"admin", "editor"}
	givenIssuer := "auth-service"

	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, givenIssuer)

	// When
	tokenString, err := manager.GenerateAccessToken(givenUserID, givenOrgID, givenEmail, givenRoles)
	assert.NoError(t, err)

	// Parse token to verify claims
	token, parseErr := jwt.ParseWithClaims(tokenString, &providers.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, parseErr)
	claims, ok := token.Claims.(*providers.Claims)
	assert.True(t, ok)

	// Then
	assert.Equal(t, givenUserID.String(), claims.UserID)
	assert.Equal(t, givenOrgID.String(), claims.OrganizationID)
	assert.Equal(t, givenEmail, claims.Email)
	assert.Equal(t, givenRoles, claims.Roles)
	assert.Equal(t, givenIssuer, claims.Issuer)
	assert.Equal(t, givenUserID.String(), claims.Subject)
	assert.NotEmpty(t, claims.ID)
}

func TestJWTManager_GenerateAccessToken_TokenExpiresAtCorrectTime(t *testing.T) {
	// Given
	givenAccessExpiry := 15 * time.Minute
	manager := NewJWTManager("test-secret", givenAccessExpiry, 24*time.Hour, "auth-service")

	beforeGeneration := time.Now()

	// When
	tokenString, err := manager.GenerateAccessToken(uuid.New(), uuid.New(), "user@example.com", []string{})
	assert.NoError(t, err)

	afterGeneration := time.Now()

	// Parse token to verify expiry
	token, parseErr := jwt.ParseWithClaims(tokenString, &providers.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, parseErr)
	claims, ok := token.Claims.(*providers.Claims)
	assert.True(t, ok)

	// Then
	// The expiry should be approximately 15 minutes from now (with some tolerance for execution time)
	expectedExpiryMin := beforeGeneration.Add(givenAccessExpiry).Add(-1 * time.Second) // 1 second tolerance
	expectedExpiryMax := afterGeneration.Add(givenAccessExpiry).Add(1 * time.Second)   // 1 second tolerance

	assert.True(t, !claims.ExpiresAt.Time.Before(expectedExpiryMin),
		"Expiry time %v should not be before %v", claims.ExpiresAt.Time, expectedExpiryMin)
	assert.True(t, !claims.ExpiresAt.Time.After(expectedExpiryMax),
		"Expiry time %v should not be after %v", claims.ExpiresAt.Time, expectedExpiryMax)
}

func TestJWTManager_GenerateRefreshToken_WithValidInput_ReturnsToken(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	// When
	token, err := manager.GenerateRefreshToken(givenUserID)

	// Then
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_GenerateRefreshToken_TokenContainsCorrectClaims(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenIssuer := "auth-service"
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, givenIssuer)

	// When
	tokenString, err := manager.GenerateRefreshToken(givenUserID)
	assert.NoError(t, err)

	// Parse token to verify claims
	token, parseErr := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, parseErr)
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	assert.True(t, ok)

	// Then
	assert.Equal(t, givenUserID.String(), claims.Subject)
	assert.Equal(t, givenIssuer, claims.Issuer)
	assert.NotEmpty(t, claims.ID)
}

func TestJWTManager_GenerateRefreshToken_TokenExpiresAtCorrectTime(t *testing.T) {
	// Given
	givenRefreshExpiry := 24 * time.Hour
	manager := NewJWTManager("test-secret", 15*time.Minute, givenRefreshExpiry, "auth-service")

	beforeGeneration := time.Now()

	// When
	tokenString, err := manager.GenerateRefreshToken(uuid.New())
	assert.NoError(t, err)

	afterGeneration := time.Now()

	// Parse token to verify expiry
	token, parseErr := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	assert.NoError(t, parseErr)
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	assert.True(t, ok)

	// Then
	// The expiry should be approximately 24 hours from now (with some tolerance for execution time)
	expectedExpiryMin := beforeGeneration.Add(givenRefreshExpiry).Add(-1 * time.Second) // 1 second tolerance
	expectedExpiryMax := afterGeneration.Add(givenRefreshExpiry).Add(1 * time.Second)   // 1 second tolerance

	assert.True(t, !claims.ExpiresAt.Time.Before(expectedExpiryMin),
		"Expiry time %v should not be before %v", claims.ExpiresAt.Time, expectedExpiryMin)
	assert.True(t, !claims.ExpiresAt.Time.After(expectedExpiryMax),
		"Expiry time %v should not be after %v", claims.ExpiresAt.Time, expectedExpiryMax)
}

func TestJWTManager_ValidateAccessToken_WithValidToken_ReturnsClaims(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	givenOrgID := uuid.New()
	givenEmail := "user@example.com"
	givenRoles := []string{"admin"}
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	tokenString, err := manager.GenerateAccessToken(givenUserID, givenOrgID, givenEmail, givenRoles)
	assert.NoError(t, err)

	// When
	claims, validateErr := manager.ValidateAccessToken(tokenString)

	// Then
	assert.NoError(t, validateErr)
	assert.NotNil(t, claims)
	assert.Equal(t, givenUserID.String(), claims.UserID)
	assert.Equal(t, givenOrgID.String(), claims.OrganizationID)
	assert.Equal(t, givenEmail, claims.Email)
	assert.Equal(t, givenRoles, claims.Roles)
}

func TestJWTManager_ValidateAccessToken_WithExpiredToken_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", -1*time.Hour, 24*time.Hour, "auth-service") // Negative expiry = already expired

	tokenString, err := manager.GenerateAccessToken(uuid.New(), uuid.New(), "user@example.com", []string{})
	assert.NoError(t, err)

	// When
	claims, validateErr := manager.ValidateAccessToken(tokenString)

	// Then
	assert.Error(t, validateErr)
	assert.Nil(t, claims)
	assert.Contains(t, validateErr.Error(), "invalid or expired token")
}

func TestJWTManager_ValidateAccessToken_WithInvalidSignature_ReturnsError(t *testing.T) {
	// Given
	manager1 := NewJWTManager("secret-1", 15*time.Minute, 24*time.Hour, "auth-service")
	manager2 := NewJWTManager("secret-2", 15*time.Minute, 24*time.Hour, "auth-service")

	tokenString, err := manager1.GenerateAccessToken(uuid.New(), uuid.New(), "user@example.com", []string{})
	assert.NoError(t, err)

	// When
	claims, validateErr := manager2.ValidateAccessToken(tokenString) // Different secret

	// Then
	assert.Error(t, validateErr)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateAccessToken_WithMalformedToken_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")
	givenMalformedToken := "not-a-valid-jwt-token"

	// When
	claims, err := manager.ValidateAccessToken(givenMalformedToken)

	// Then
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid or expired token")
}

func TestJWTManager_ValidateAccessToken_WithEmptyToken_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	// When
	claims, err := manager.ValidateAccessToken("")

	// Then
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateRefreshToken_WithValidToken_ReturnsClaims(t *testing.T) {
	// Given
	givenUserID := uuid.New()
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	tokenString, err := manager.GenerateRefreshToken(givenUserID)
	assert.NoError(t, err)

	// When
	claims, validateErr := manager.ValidateRefreshToken(tokenString)

	// Then
	assert.NoError(t, validateErr)
	assert.NotNil(t, claims)
	assert.Equal(t, givenUserID.String(), claims.Subject)
}

func TestJWTManager_ValidateRefreshToken_WithExpiredToken_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, -1*time.Hour, "auth-service") // Negative expiry = already expired

	tokenString, err := manager.GenerateRefreshToken(uuid.New())
	assert.NoError(t, err)

	// When
	claims, validateErr := manager.ValidateRefreshToken(tokenString)

	// Then
	assert.Error(t, validateErr)
	assert.Nil(t, claims)
	assert.Contains(t, validateErr.Error(), "invalid or expired refresh token")
}

func TestJWTManager_ValidateRefreshToken_WithInvalidSignature_ReturnsError(t *testing.T) {
	// Given
	manager1 := NewJWTManager("secret-1", 15*time.Minute, 24*time.Hour, "auth-service")
	manager2 := NewJWTManager("secret-2", 15*time.Minute, 24*time.Hour, "auth-service")

	tokenString, err := manager1.GenerateRefreshToken(uuid.New())
	assert.NoError(t, err)

	// When
	claims, validateErr := manager2.ValidateRefreshToken(tokenString) // Different secret

	// Then
	assert.Error(t, validateErr)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateRefreshToken_WithMalformedToken_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")
	givenMalformedToken := "not-a-valid-jwt-token"

	// When
	claims, err := manager.ValidateRefreshToken(givenMalformedToken)

	// Then
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")
}

func TestJWTManager_GetAccessExpiry_ReturnsCorrectDuration(t *testing.T) {
	// Given
	givenAccessExpiry := 30 * time.Minute
	manager := NewJWTManager("test-secret", givenAccessExpiry, 24*time.Hour, "auth-service")

	// When
	expiry := manager.GetAccessExpiry()

	// Then
	assert.Equal(t, givenAccessExpiry, expiry)
}

func TestJWTManager_GetRefreshExpiry_ReturnsCorrectDuration(t *testing.T) {
	// Given
	givenRefreshExpiry := 48 * time.Hour
	manager := NewJWTManager("test-secret", 15*time.Minute, givenRefreshExpiry, "auth-service")

	// When
	expiry := manager.GetRefreshExpiry()

	// Then
	assert.Equal(t, givenRefreshExpiry, expiry)
}

func TestJWTManager_ValidateAccessToken_WithWrongSigningMethod_ReturnsError(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")

	// Create token with wrong signing method (RS256 instead of HS256)
	// This would normally require RSA keys, but we can test with a malformed header
	givenTokenWithWrongMethod := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.invalid"

	// When
	claims, err := manager.ValidateAccessToken(givenTokenWithWrongMethod)

	// Then
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_TokensAreUnique_WhenGeneratedMultipleTimes(t *testing.T) {
	// Given
	manager := NewJWTManager("test-secret", 15*time.Minute, 24*time.Hour, "auth-service")
	givenUserID := uuid.New()

	// When
	token1, err1 := manager.GenerateAccessToken(givenUserID, uuid.New(), "user@example.com", []string{})
	token2, err2 := manager.GenerateAccessToken(givenUserID, uuid.New(), "user@example.com", []string{})

	// Then
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, token1, token2, "Tokens should be unique even for same user")
}
