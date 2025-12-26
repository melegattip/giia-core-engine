package events

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
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
	if nc == nil {
		return nil, errors.NewBadRequest("NATS connection is required")
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, errors.NewInternalServerError("failed to create JetStream context")
	}

	return &NATSPublisher{
		js:   js,
		conn: nc,
	}, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, subject string, event *Event) error {
	if subject == "" {
		return errors.NewBadRequest("subject is required")
	}

	if event == nil {
		return errors.NewBadRequest("event is required")
	}

	if err := event.Validate(); err != nil {
		return err
	}

	data, err := event.ToJSON()
	if err != nil {
		return errors.NewInternalServerError("failed to serialize event")
	}

	err = retryPublish(ctx, func() error {
		_, err := p.js.Publish(subject, data)
		return err
	})

	if err != nil {
		return errors.NewInternalServerError("failed to publish event after retries")
	}

	return nil
}

func (p *NATSPublisher) PublishAsync(ctx context.Context, subject string, event *Event) error {
	if subject == "" {
		return errors.NewBadRequest("subject is required")
	}

	if event == nil {
		return errors.NewBadRequest("event is required")
	}

	if err := event.Validate(); err != nil {
		return err
	}

	data, err := event.ToJSON()
	if err != nil {
		return errors.NewInternalServerError("failed to serialize event")
	}

	_, err = p.js.PublishAsync(subject, data)
	if err != nil {
		return errors.NewInternalServerError("failed to publish event async")
	}

	return nil
}

func (p *NATSPublisher) Close() error {
	return Disconnect(p.conn)
}

func retryPublish(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	const initialBackoff = 1 * time.Second

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
				return errors.NewInternalServerError("publish cancelled")
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}

	return err
}
