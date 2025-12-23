package events

import (
	"context"

	pkgEvents "github.com/giia/giia-core-engine/pkg/events"
	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
	"github.com/nats-io/nats.go"
)

type NATSEventPublisher struct {
	publisher pkgEvents.Publisher
	logger    pkgLogger.Logger
}

func NewNATSEventPublisher(conn *nats.Conn, logger pkgLogger.Logger) *NATSEventPublisher {
	publisher, err := pkgEvents.NewPublisher(conn)
	if err != nil {
		logger.Error(context.Background(), err, "Failed to create NATS publisher", nil)
		return nil
	}

	return &NATSEventPublisher{
		publisher: publisher,
		logger:    logger,
	}
}

func (p *NATSEventPublisher) Publish(ctx context.Context, subject string, event *pkgEvents.Event) error {
	return p.publisher.Publish(ctx, subject, event)
}

func (p *NATSEventPublisher) PublishAsync(ctx context.Context, subject string, event *pkgEvents.Event) error {
	return p.publisher.PublishAsync(ctx, subject, event)
}

func (p *NATSEventPublisher) Close() error {
	return p.publisher.Close()
}

type NoOpEventPublisher struct{}

func NewNoOpEventPublisher() *NoOpEventPublisher {
	return &NoOpEventPublisher{}
}

func (p *NoOpEventPublisher) Publish(ctx context.Context, subject string, event *pkgEvents.Event) error {
	return nil
}

func (p *NoOpEventPublisher) PublishAsync(ctx context.Context, subject string, event *pkgEvents.Event) error {
	return nil
}

func (p *NoOpEventPublisher) Close() error {
	return nil
}
