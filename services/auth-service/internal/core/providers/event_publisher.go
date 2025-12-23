package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
)

type EventPublisher interface {
	Publish(ctx context.Context, subject string, event *events.Event) error
	PublishAsync(ctx context.Context, subject string, event *events.Event) error
	Close() error
}
