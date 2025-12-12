package rate_limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
)

type redisRateLimiter struct {
	client *redis.Client
	logger pkgLogger.Logger
}

func NewRedisRateLimiter(client *redis.Client, logger pkgLogger.Logger) providers.RateLimiter {
	return &redisRateLimiter{
		client: client,
		logger: logger,
	}
}

func (r *redisRateLimiter) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error) {
	redisKey := fmt.Sprintf("rate_limit:%s", key)

	count, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		r.logger.Error(ctx, err, "Failed to increment rate limit counter", pkgLogger.Tags{
			"key": key,
		})
		return true, 0, err
	}

	if count == 1 {
		r.client.Expire(ctx, redisKey, window)
	}

	if count > int64(limit) {
		ttl, err := r.client.TTL(ctx, redisKey).Result()
		if err != nil {
			r.logger.Error(ctx, err, "Failed to get TTL for rate limit", pkgLogger.Tags{
				"key": key,
			})
			return false, window, nil
		}

		r.logger.Warn(ctx, "Rate limit exceeded", pkgLogger.Tags{
			"key":         key,
			"count":       count,
			"limit":       limit,
			"retry_after": ttl.Seconds(),
		})

		return false, ttl, nil
	}

	return true, 0, nil
}

func (r *redisRateLimiter) ResetRateLimit(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("rate_limit:%s", key)
	return r.client.Del(ctx, redisKey).Err()
}
