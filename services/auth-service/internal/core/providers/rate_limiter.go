package providers

import (
	"context"
	"time"
)

type RateLimiter interface {
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
	ResetRateLimit(ctx context.Context, key string) error
}
