package event_processing

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type UserEventHandlerImpl struct {
	notificationRepo providers.NotificationRepository
	logger           logger.Logger
}

func NewUserEventHandler(
	notificationRepo providers.NotificationRepository,
	logger logger.Logger,
) providers.UserEventHandler {
	return &UserEventHandlerImpl{
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

func (h *UserEventHandlerImpl) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Info(ctx, "Processing user event", logger.Tags{
		"event_type": event.Type,
		"event_id":   event.ID,
	})

	return nil
}
