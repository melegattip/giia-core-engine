package providers

import (
	"context"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.AINotification) error
	GetByID(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) (*domain.AINotification, error)
	List(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID, filters *NotificationFilters) ([]*domain.AINotification, error)
	Update(ctx context.Context, notification *domain.AINotification) error
	Delete(ctx context.Context, id uuid.UUID, organizationID uuid.UUID) error
}

type NotificationFilters struct {
	Types      []domain.NotificationType
	Priorities []domain.NotificationPriority
	Statuses   []domain.NotificationStatus
	Limit      int
	Offset     int
}

type PreferencesRepository interface {
	Create(ctx context.Context, prefs *domain.UserNotificationPreferences) error
	GetByUserID(ctx context.Context, userID uuid.UUID, organizationID uuid.UUID) (*domain.UserNotificationPreferences, error)
	Update(ctx context.Context, prefs *domain.UserNotificationPreferences) error
}
