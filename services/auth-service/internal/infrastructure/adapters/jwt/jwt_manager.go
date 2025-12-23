package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type JWTManager struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

func (j *JWTManager) GenerateAccessToken(userID, orgID uuid.UUID, email string, roles []string) (string, error) {
	now := time.Now()
	claims := &providers.Claims{
		UserID:         userID.String(),
		Email:          email,
		OrganizationID: orgID.String(),
		Roles:          roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", pkgErrors.NewInternalServerError("failed to sign access token")
	}
	return signedToken, nil
}

func (j *JWTManager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		Issuer:    j.issuer,
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", pkgErrors.NewInternalServerError("failed to sign refresh token")
	}
	return signedToken, nil
}

func (j *JWTManager) ValidateAccessToken(tokenString string) (*providers.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &providers.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, pkgErrors.NewUnauthorized("invalid token signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, pkgErrors.NewUnauthorized("invalid or expired token")
	}

	claims, ok := token.Claims.(*providers.Claims)
	if !ok || !token.Valid {
		return nil, pkgErrors.NewUnauthorized("invalid token claims")
	}

	return claims, nil
}

func (j *JWTManager) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, pkgErrors.NewUnauthorized("invalid token signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, pkgErrors.NewUnauthorized("invalid or expired refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, pkgErrors.NewUnauthorized("invalid refresh token claims")
	}

	return claims, nil
}

func (j *JWTManager) GetAccessExpiry() time.Duration {
	return j.accessExpiry
}

func (j *JWTManager) GetRefreshExpiry() time.Duration {
	return j.refreshExpiry
}
