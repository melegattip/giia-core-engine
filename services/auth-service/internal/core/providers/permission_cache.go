package providers

import (
	"context"
	"time"
)

type PermissionCache interface {
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	SetUserPermissions(ctx context.Context, userID string, permissions []string, ttl time.Duration) error
	InvalidateUserPermissions(ctx context.Context, userID string) error
	InvalidateUsersWithRole(ctx context.Context, userIDs []string) error
}