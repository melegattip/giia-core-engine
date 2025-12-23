package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/auth-service/internal/core/providers"
)

type tokenRepository struct {
	redis *redis.Client
	db    *gorm.DB
}

func NewTokenRepository(redis *redis.Client, db *gorm.DB) providers.TokenRepository {
	return &tokenRepository{
		redis: redis,
		db:    db,
	}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *tokenRepository) StoreRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *tokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked", true).Error
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked = ?", userID, false).
		Update("revoked", true).Error
}

func (r *tokenRepository) StorePasswordResetToken(ctx context.Context, token *domain.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *tokenRepository) GetPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	var token domain.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.PasswordResetToken{}).
		Where("token_hash = ?", tokenHash).
		Update("used", true).Error
}

func (r *tokenRepository) StoreActivationToken(ctx context.Context, token *domain.ActivationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *tokenRepository) GetActivationToken(ctx context.Context, tokenHash string) (*domain.ActivationToken, error) {
	var token domain.ActivationToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND used = ? AND expires_at > ?", tokenHash, false, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) MarkActivationTokenUsed(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).
		Model(&domain.ActivationToken{}).
		Where("token_hash = ?", tokenHash).
		Update("used", true).Error
}

func (r *tokenRepository) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("blacklist:%s", hashToken(token))
	return r.redis.Set(ctx, key, "1", ttl).Err()
}

func (r *tokenRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", hashToken(token))
	result, err := r.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

type RefreshTokenData struct {
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (r *tokenRepository) storeRefreshTokenInRedis(ctx context.Context, tokenHash string, userID uuid.UUID, expiresAt time.Time) error {
	data := RefreshTokenData{
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	ttl := time.Until(expiresAt)
	return r.redis.Set(ctx, key, jsonData, ttl).Err()
}

func (r *tokenRepository) getRefreshTokenFromRedis(ctx context.Context, tokenHash string) (*RefreshTokenData, error) {
	key := fmt.Sprintf("refresh_token:%s", tokenHash)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var tokenData RefreshTokenData
	if err := json.Unmarshal([]byte(data), &tokenData); err != nil {
		return nil, err
	}

	return &tokenData, nil
}
