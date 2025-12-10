package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/auth-service/internal/core/providers"
	"github.com/redis/go-redis/v9"
)

type redisPermissionCache struct {
	client *redis.Client
	logger pkgLogger.Logger
}

func NewRedisPermissionCache(client *redis.Client, logger pkgLogger.Logger) providers.PermissionCache {
	return &redisPermissionCache{
		client: client,
		logger: logger,
	}
}

func (c *redisPermissionCache) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	key := fmt.Sprintf("user:%s:permissions", userID)

	data, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		c.logger.Error(ctx, err, "Failed to get user permissions from cache", pkgLogger.Tags{
			"user_id": userID,
		})
		return nil, err
	}

	var permissions []string
	if err := json.Unmarshal([]byte(data), &permissions); err != nil {
		c.logger.Error(ctx, err, "Failed to unmarshal permissions from cache", pkgLogger.Tags{
			"user_id": userID,
		})
		return nil, err
	}

	c.logger.Debug(ctx, "Cache hit for user permissions", pkgLogger.Tags{
		"user_id":          userID,
		"permissions_count": len(permissions),
	})

	return permissions, nil
}

func (c *redisPermissionCache) SetUserPermissions(ctx context.Context, userID string, permissions []string, ttl time.Duration) error {
	key := fmt.Sprintf("user:%s:permissions", userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		c.logger.Error(ctx, err, "Failed to marshal permissions for cache", pkgLogger.Tags{
			"user_id": userID,
		})
		return err
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error(ctx, err, "Failed to set user permissions in cache", pkgLogger.Tags{
			"user_id": userID,
			"ttl":     ttl.String(),
		})
		return err
	}

	c.logger.Debug(ctx, "Cached user permissions", pkgLogger.Tags{
		"user_id":          userID,
		"permissions_count": len(permissions),
		"ttl":              ttl.String(),
	})

	return nil
}

func (c *redisPermissionCache) InvalidateUserPermissions(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:%s:permissions", userID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		c.logger.Error(ctx, err, "Failed to invalidate user permissions cache", pkgLogger.Tags{
			"user_id": userID,
		})
		return err
	}

	c.logger.Debug(ctx, "Invalidated user permissions cache", pkgLogger.Tags{
		"user_id": userID,
	})

	return nil
}

func (c *redisPermissionCache) InvalidateUsersWithRole(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = fmt.Sprintf("user:%s:permissions", userID)
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error(ctx, err, "Failed to invalidate permissions cache for multiple users", pkgLogger.Tags{
			"user_count": len(userIDs),
		})
		return err
	}

	c.logger.Debug(ctx, "Invalidated permissions cache for multiple users", pkgLogger.Tags{
		"user_count": len(userIDs),
	})

	return nil
}
