package event_processing

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type ExecutionEventHandlerImpl struct {
	notificationRepo providers.NotificationRepository
	logger           logger.Logger
}

func NewExecutionEventHandler(
	notificationRepo providers.NotificationRepository,
	logger logger.Logger,
) providers.ExecutionEventHandler {
	return &ExecutionEventHandlerImpl{
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

func (h *ExecutionEventHandlerImpl) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Info(ctx, "Processing execution event", logger.Tags{
		"event_type": event.Type,
		"event_id":   event.ID,
	})

	return nil
}
