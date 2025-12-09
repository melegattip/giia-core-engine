package events

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

type EventHandler func(ctx context.Context, event *Event) error

type Subscriber interface {
	Subscribe(ctx context.Context, subject string, handler EventHandler) error
	SubscribeDurable(ctx context.Context, subject, durableName string, handler EventHandler) error
	Close() error
}

type NATSSubscriber struct {
	js   nats.JetStreamContext
	conn *nats.Conn
	subs []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn) (*NATSSubscriber, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &NATSSubscriber{
		js:   js,
		conn: nc,
		subs: make([]*nats.Subscription, 0),
	}, nil
}

func (s *NATSSubscriber) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		event, err := FromJSON(msg.Data)
		if err != nil {
			msg.Nak()
			return
		}

		if err := handler(ctx, event); err != nil {
			msg.Nak()
			return
		}

		msg.Ack()
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *NATSSubscriber) SubscribeDurable(ctx context.Context, subject, durableName string, handler EventHandler) error {
	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		event, err := FromJSON(msg.Data)
		if err != nil {
			msg.Nak()
			return
		}

		if err := handler(ctx, event); err != nil {
			msg.Nak()
			return
		}

		msg.Ack()
	}, nats.Durable(durableName), nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe to subject %s with durable %s: %w", subject, durableName, err)
	}

	s.subs = append(s.subs, sub)
	return nil
}

func (s *NATSSubscriber) Close() error {
	for _, sub := range s.subs {
		if err := sub.Drain(); err != nil {
			return fmt.Errorf("failed to drain subscription: %w", err)
		}
	}

	return Disconnect(s.conn)
}
