package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	pkgErrors "github.com/melegattip/giia-core-engine/pkg/errors"
)

type JWTService interface {
	GenerateTokens(userID uint, email string) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (*Claims, error)
	GenerateEmailVerificationToken(userID uint, email string) (string, error)
	GeneratePasswordResetToken(userID uint, email string) (string, error)
	ValidateEmailVerificationToken(tokenString string) (*Claims, error)
	ValidatePasswordResetToken(tokenString string) (*Claims, error)
}

type jwtService struct {
	secretKey     string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

type Claims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

func NewJWTService(secretKey string, accessExpiry, refreshExpiry time.Duration, issuer string) JWTService {
	return &jwtService{
		secretKey:     secretKey,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

func (j *jwtService) GenerateTokens(userID uint, email string) (*TokenPair, error) {
	// Generate access token
	accessToken, accessExpiry, err := j.generateToken(userID, email, "access", j.accessExpiry)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to generate access token")
	}

	// Generate refresh token
	refreshToken, _, err := j.generateToken(userID, email, "refresh", j.refreshExpiry)
	if err != nil {
		return nil, pkgErrors.NewInternalServerError("failed to generate refresh token")
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "access")
}

func (j *jwtService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "refresh")
}

func (j *jwtService) GenerateEmailVerificationToken(userID uint, email string) (string, error) {
	token, _, err := j.generateToken(userID, email, "email_verification", 24*time.Hour)
	return token, err
}

func (j *jwtService) GeneratePasswordResetToken(userID uint, email string) (string, error) {
	token, _, err := j.generateToken(userID, email, "password_reset", 1*time.Hour)
	return token, err
}

func (j *jwtService) ValidateEmailVerificationToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "email_verification")
}

func (j *jwtService) ValidatePasswordResetToken(tokenString string) (*Claims, error) {
	return j.validateToken(tokenString, "password_reset")
}

func (j *jwtService) generateToken(userID uint, email, tokenType string, expiry time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(expiry)
	jwtID := uuid.New().String()

	claims := Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jwtID,
			Issuer:    j.issuer,
			Subject:   fmt.Sprintf("%d", userID),
			Audience:  []string{"users-service"},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", time.Time{}, pkgErrors.NewInternalServerError("failed to sign token")
	}

	return tokenString, expiresAt, nil
}

func (j *jwtService) validateToken(tokenString, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, pkgErrors.NewUnauthorized("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, pkgErrors.NewUnauthorized("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, pkgErrors.NewUnauthorized("invalid token claims")
	}

	if claims.TokenType != expectedType {
		return nil, pkgErrors.NewUnauthorized("invalid token type")
	}

	return claims, nil
}
