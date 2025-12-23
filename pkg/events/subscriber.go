package events

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/nats-io/nats.go"
)

type EventHandler func(ctx context.Context, event *Event) error

type SubscriberConfig struct {
	MaxDeliver int
	AckWait    time.Duration
}

type Subscriber interface {
	Subscribe(ctx context.Context, subject string, handler EventHandler) error
	SubscribeDurable(ctx context.Context, subject, durableName string, handler EventHandler) error
	SubscribeDurableWithConfig(ctx context.Context, subject, durableName string, config *SubscriberConfig, handler EventHandler) error
	Close() error
}

type NATSSubscriber struct {
	js   nats.JetStreamContext
	conn *nats.Conn
	subs []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn) (*NATSSubscriber, error) {
	if nc == nil {
		return nil, errors.NewBadRequest("NATS connection is required")
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, errors.NewInternalServerError("failed to create JetStream context")
	}

	return &NATSSubscriber{
		js:   js,
		conn: nc,
		subs: make([]*nats.Subscription, 0),
	}, nil
}

func (s *NATSSubscriber) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	if subject == "" {
		return errors.NewBadRequest("subject is required")
	}

	if handler == nil {
		return errors.NewBadRequest("handler is required")
	}

	sub, err := s.js.Subscribe(subject, s.wrapHandler(ctx, handler))
	if err != nil {
		return errors.NewInternalServerError("failed to subscribe to subject")
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *NATSSubscriber) SubscribeDurable(ctx context.Context, subject, durableName string, handler EventHandler) error {
	config := &SubscriberConfig{
		MaxDeliver: 5,
		AckWait:    30 * time.Second,
	}
	return s.SubscribeDurableWithConfig(ctx, subject, durableName, config, handler)
}

func (s *NATSSubscriber) SubscribeDurableWithConfig(ctx context.Context, subject, durableName string, config *SubscriberConfig, handler EventHandler) error {
	if subject == "" {
		return errors.NewBadRequest("subject is required")
	}

	if durableName == "" {
		return errors.NewBadRequest("durable name is required")
	}

	if handler == nil {
		return errors.NewBadRequest("handler is required")
	}

	if config == nil {
		config = &SubscriberConfig{
			MaxDeliver: 5,
			AckWait:    30 * time.Second,
		}
	}

	sub, err := s.js.Subscribe(
		subject,
		s.wrapHandler(ctx, handler),
		nats.Durable(durableName),
		nats.ManualAck(),
		nats.MaxDeliver(config.MaxDeliver),
		nats.AckWait(config.AckWait),
	)

	if err != nil {
		return errors.NewInternalServerError("failed to subscribe to subject with durable")
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *NATSSubscriber) wrapHandler(ctx context.Context, handler EventHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		event, err := FromJSON(msg.Data)
		if err != nil {
			_ = msg.Nak()
			return
		}

		if err := handler(ctx, event); err != nil {
			_ = msg.Nak()
			return
		}

		_ = msg.Ack()
	}
}

func (s *NATSSubscriber) Close() error {
	for _, sub := range s.subs {
		if err := sub.Drain(); err != nil {
			return errors.NewInternalServerError("failed to drain subscription")
		}
	}

	return Disconnect(s.conn)
}
