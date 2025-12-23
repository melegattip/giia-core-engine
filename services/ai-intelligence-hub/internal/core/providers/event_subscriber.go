package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
)

type EventSubscriber interface {
	Start(ctx context.Context) error
	Stop() error
}

type EventHandler interface {
	Handle(ctx context.Context, event *events.Event) error
}

type BufferEventHandler interface {
	Handle(ctx context.Context, event *events.Event) error
}

type ExecutionEventHandler interface {
	Handle(ctx context.Context, event *events.Event) error
}

type UserEventHandler interface {
	Handle(ctx context.Context, event *events.Event) error
}

type PatternDetector interface {
	DetectPatterns(ctx context.Context, event *events.Event) ([]Pattern, error)
}

type Pattern struct {
	Type        string
	Events      []*events.Event
	Description string
	Priority    string
}
