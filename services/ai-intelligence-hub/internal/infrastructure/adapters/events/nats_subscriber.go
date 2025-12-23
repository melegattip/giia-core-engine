package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type NATSEventSubscriber struct {
	subscriber events.Subscriber
	handlers   map[string]providers.EventHandler
	logger     logger.Logger
}

func NewNATSEventSubscriber(
	subscriber events.Subscriber,
	bufferHandler providers.BufferEventHandler,
	executionHandler providers.ExecutionEventHandler,
	userHandler providers.UserEventHandler,
	logger logger.Logger,
) providers.EventSubscriber {
	return &NATSEventSubscriber{
		subscriber: subscriber,
		handlers: map[string]providers.EventHandler{
			"buffer":    bufferHandler,
			"execution": executionHandler,
			"user":      userHandler,
		},
		logger: logger,
	}
}

func (s *NATSEventSubscriber) Start(ctx context.Context) error {
	subjects := []string{
		"auth.>",
		"catalog.>",
		"ddmrp.>",
		"execution.>",
		"analytics.>",
	}

	for _, subject := range subjects {
		if err := s.subscribeToSubject(ctx, subject); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}
	}

	s.logger.Info(ctx, "AI Intelligence Hub event subscriber started", logger.Tags{
		"subjects": strings.Join(subjects, ", "),
	})

	return nil
}

func (s *NATSEventSubscriber) subscribeToSubject(ctx context.Context, subject string) error {
	config := &events.SubscriberConfig{
		MaxDeliver: 5,
		AckWait:    30 * time.Second,
	}

	return s.subscriber.SubscribeDurableWithConfig(
		ctx,
		subject,
		"ai-intelligence-hub-consumer",
		config,
		s.handleEvent,
	)
}

func (s *NATSEventSubscriber) handleEvent(ctx context.Context, event *events.Event) error {
	s.logger.Debug(ctx, "Received event", logger.Tags{
		"event_type": event.Type,
		"event_id":   event.ID,
		"source":     event.Source,
	})

	go func() {
		handler := s.getHandler(event.Type)
		if handler == nil {
			s.logger.Debug(ctx, "No handler for event type", logger.Tags{
				"event_type": event.Type,
			})
			return
		}

		if err := handler.Handle(ctx, event); err != nil {
			s.logger.Error(ctx, err, "Failed to handle event", logger.Tags{
				"event_type": event.Type,
				"event_id":   event.ID,
			})
		}
	}()

	return nil
}

func (s *NATSEventSubscriber) getHandler(eventType string) providers.EventHandler {
	for prefix, handler := range s.handlers {
		if strings.HasPrefix(eventType, prefix) {
			return handler
		}
	}
	return nil
}

func (s *NATSEventSubscriber) Stop() error {
	s.logger.Info(context.Background(), "Stopping AI Intelligence Hub event subscriber", nil)
	return nil
}
