package events

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, subject string, event *Event) error
	PublishAsync(ctx context.Context, subject string, event *Event) error
	Close() error
}

type NATSPublisher struct {
	js   nats.JetStreamContext
	conn *nats.Conn
}

func NewPublisher(nc *nats.Conn) (*NATSPublisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &NATSPublisher{
		js:   js,
		conn: nc,
	}, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, subject string, event *Event) error {
	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	err = retryPublish(ctx, func() error {
		_, err := p.js.Publish(subject, data)
		return err
	})

	if err != nil {
		return fmt.Errorf("failed to publish event after retries: %w", err)
	}

	return nil
}

func (p *NATSPublisher) PublishAsync(ctx context.Context, subject string, event *Event) error {
	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	_, err = p.js.PublishAsync(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event async: %w", err)
	}

	return nil
}

func (p *NATSPublisher) Close() error {
	return Disconnect(p.conn)
}

func retryPublish(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	const initialBackoff = 100 * time.Millisecond

	var err error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("publish cancelled: %w", ctx.Err())
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	return fmt.Errorf("publish failed after %d attempts: %w", maxRetries, err)
}
