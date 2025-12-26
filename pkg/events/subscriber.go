package events

import (
	"context"
	"strings"
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

	// Extract stream name from subject (e.g., "auth.>" -> "auth")
	streamName := getStreamNameFromSubject(subject)

	// Try to ensure the stream exists (create if needed)
	if err := s.ensureStream(streamName, subject); err != nil {
		// Log but continue - we'll try to subscribe anyway
		// The stream might be created by another service
	}

	// Try durable subscription first
	sub, err := s.js.Subscribe(
		subject,
		s.wrapHandler(ctx, handler),
		nats.Durable(durableName),
		nats.ManualAck(),
		nats.MaxDeliver(config.MaxDeliver),
		nats.AckWait(config.AckWait),
		nats.DeliverNew(), // Only deliver new messages
	)

	if err != nil {
		// Fallback to simple push-based subscription without durability
		sub, err = s.js.Subscribe(
			subject,
			s.wrapHandler(ctx, handler),
			nats.DeliverNew(),
		)
		if err != nil {
			// Last resort: use core NATS subscription (no JetStream)
			coreSub, coreErr := s.conn.Subscribe(subject, func(msg *nats.Msg) {
				event, parseErr := FromJSON(msg.Data)
				if parseErr != nil {
					return
				}
				_ = handler(ctx, event)
			})
			if coreErr != nil {
				return errors.NewInternalServerError("failed to subscribe to subject: " + coreErr.Error())
			}
			s.subs = append(s.subs, coreSub)
			return nil
		}
	}

	s.subs = append(s.subs, sub)
	return nil
}

// ensureStream creates a stream if it doesn't exist
func (s *NATSSubscriber) ensureStream(streamName, subject string) error {
	// Check if stream exists
	_, err := s.js.StreamInfo(streamName)
	if err == nil {
		return nil // Stream exists
	}

	// Create stream with basic config
	_, err = s.js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subject},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    24 * time.Hour, // Keep messages for 24 hours
		MaxMsgs:   -1,             // No limit on number of messages
		MaxBytes:  -1,             // No limit on size
		Discard:   nats.DiscardOld,
	})

	return err
}

// getStreamNameFromSubject extracts a stream name from a subject pattern
func getStreamNameFromSubject(subject string) string {
	// Remove wildcards and dots to create a clean stream name
	name := subject
	for _, ch := range []string{".>", ".*", ".", ">", "*"} {
		name = strings.ReplaceAll(name, ch, "")
	}
	if name == "" {
		name = "default"
	}
	return strings.ToUpper(name)
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
