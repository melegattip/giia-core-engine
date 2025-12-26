// Package events provides NATS event subscription for the Execution Service.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	// Default subscriber configuration
	defaultMaxDeliver = 5
	defaultAckWait    = 30 * time.Second
	defaultBatchSize  = 100
	defaultFetchWait  = 5 * time.Second
)

// SubscriberConfig contains configuration for the event subscriber.
type SubscriberConfig struct {
	// MaxDeliver is the maximum number of delivery attempts
	MaxDeliver int
	// AckWait is the time to wait for acknowledgment
	AckWait time.Duration
	// BatchSize for pull subscribe
	BatchSize int
	// FetchWait is the timeout for fetch operations
	FetchWait time.Duration
	// ConsumerName is the durable consumer name
	ConsumerName string
}

// DefaultSubscriberConfig returns the default subscriber configuration.
func DefaultSubscriberConfig() *SubscriberConfig {
	return &SubscriberConfig{
		MaxDeliver:   defaultMaxDeliver,
		AckWait:      defaultAckWait,
		BatchSize:    defaultBatchSize,
		FetchWait:    defaultFetchWait,
		ConsumerName: "execution-service",
	}
}

// EventHandler is called when an event is received.
type EventHandler func(ctx context.Context, envelope *EventEnvelope) error

// SubscriberMetrics tracks subscription statistics.
type SubscriberMetrics struct {
	ReceivedCount   int64
	ProcessedCount  int64
	FailedCount     int64
	RedeliveryCount int64
	LastReceivedAt  time.Time
}

// Subscriber provides event subscription capabilities for the Execution Service.
type Subscriber struct {
	js       nats.JetStreamContext
	conn     *nats.Conn
	config   *SubscriberConfig
	handlers map[string]EventHandler
	subs     []*nats.Subscription
	metrics  *SubscriberMetrics
	mu       sync.RWMutex
	running  bool
	stopCh   chan struct{}
}

// NewSubscriber creates a new event subscriber with NATS JetStream.
func NewSubscriber(nc *nats.Conn, config *SubscriberConfig) (*Subscriber, error) {
	if nc == nil {
		return nil, fmt.Errorf("NATS connection is required")
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	if config == nil {
		config = DefaultSubscriberConfig()
	}

	return &Subscriber{
		js:       js,
		conn:     nc,
		config:   config,
		handlers: make(map[string]EventHandler),
		subs:     make([]*nats.Subscription, 0),
		metrics:  &SubscriberMetrics{},
		stopCh:   make(chan struct{}),
	}, nil
}

// RegisterHandler registers a handler for a specific event type pattern.
func (s *Subscriber) RegisterHandler(pattern string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[pattern] = handler
}

// Subscribe subscribes to a subject with push-based delivery.
func (s *Subscriber) Subscribe(ctx context.Context, subject string, handler EventHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, err := s.js.Subscribe(
		subject,
		s.wrapHandler(handler),
		nats.Durable(s.config.ConsumerName),
		nats.ManualAck(),
		nats.MaxDeliver(s.config.MaxDeliver),
		nats.AckWait(s.config.AckWait),
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	s.subs = append(s.subs, sub)
	return nil
}

// SubscribePull subscribes to a subject with pull-based delivery.
// This is preferred for long-running consumers with flow control.
func (s *Subscriber) SubscribePull(ctx context.Context, subject string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, err := s.js.PullSubscribe(
		subject,
		s.config.ConsumerName,
		nats.ManualAck(),
		nats.MaxDeliver(s.config.MaxDeliver),
		nats.AckWait(s.config.AckWait),
	)
	if err != nil {
		return fmt.Errorf("failed to pull subscribe to %s: %w", subject, err)
	}

	s.subs = append(s.subs, sub)

	// Start the pull worker
	go s.pullWorker(ctx, sub)

	return nil
}

// pullWorker continuously fetches and processes messages.
func (s *Subscriber) pullWorker(ctx context.Context, sub *nats.Subscription) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		default:
		}

		msgs, err := sub.Fetch(s.config.BatchSize, nats.MaxWait(s.config.FetchWait))
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			// Log error and continue
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range msgs {
			s.processMessage(ctx, msg)
		}
	}
}

// processMessage processes a single message.
func (s *Subscriber) processMessage(ctx context.Context, msg *nats.Msg) {
	atomic.AddInt64(&s.metrics.ReceivedCount, 1)
	s.metrics.LastReceivedAt = time.Now()

	// Check for redelivery
	meta, err := msg.Metadata()
	if err == nil && meta.NumDelivered > 1 {
		atomic.AddInt64(&s.metrics.RedeliveryCount, 1)
	}

	// Parse envelope
	var envelope EventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		atomic.AddInt64(&s.metrics.FailedCount, 1)
		_ = msg.Nak()
		return
	}

	// Find and execute handler
	handler := s.findHandler(envelope.Type)
	if handler == nil {
		// No handler registered, still ack to prevent redelivery
		_ = msg.Ack()
		return
	}

	// Execute handler
	if err := handler(ctx, &envelope); err != nil {
		atomic.AddInt64(&s.metrics.FailedCount, 1)
		_ = msg.Nak()
		return
	}

	atomic.AddInt64(&s.metrics.ProcessedCount, 1)
	_ = msg.Ack()
}

// findHandler finds a handler for the given event type.
func (s *Subscriber) findHandler(eventType string) EventHandler {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Exact match first
	if handler, ok := s.handlers[eventType]; ok {
		return handler
	}

	// Pattern match (e.g., "purchase_order.*" matches "purchase_order.created")
	for pattern, handler := range s.handlers {
		if matchPattern(pattern, eventType) {
			return handler
		}
	}

	return nil
}

// matchPattern performs simple wildcard matching.
func matchPattern(pattern, value string) bool {
	if pattern == "*" || pattern == ">" {
		return true
	}

	// Handle suffix wildcard
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}

	return pattern == value
}

// wrapHandler wraps a handler to process NATS messages.
func (s *Subscriber) wrapHandler(handler EventHandler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		s.processMessageWithHandler(context.Background(), msg, handler)
	}
}

// processMessageWithHandler processes a message with a specific handler.
func (s *Subscriber) processMessageWithHandler(ctx context.Context, msg *nats.Msg, handler EventHandler) {
	atomic.AddInt64(&s.metrics.ReceivedCount, 1)
	s.metrics.LastReceivedAt = time.Now()

	var envelope EventEnvelope
	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		atomic.AddInt64(&s.metrics.FailedCount, 1)
		_ = msg.Nak()
		return
	}

	if err := handler(ctx, &envelope); err != nil {
		atomic.AddInt64(&s.metrics.FailedCount, 1)
		_ = msg.Nak()
		return
	}

	atomic.AddInt64(&s.metrics.ProcessedCount, 1)
	_ = msg.Ack()
}

// Start starts the subscriber with all registered subjects.
func (s *Subscriber) Start(ctx context.Context, subjects []string) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("subscriber already running")
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	for _, subject := range subjects {
		if err := s.SubscribePull(ctx, subject); err != nil {
			s.Stop()
			return err
		}
	}

	return nil
}

// Stop stops the subscriber and drains all subscriptions.
func (s *Subscriber) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.stopCh)

	var lastErr error
	for _, sub := range s.subs {
		if err := sub.Drain(); err != nil {
			lastErr = err
		}
	}

	s.subs = make([]*nats.Subscription, 0)
	return lastErr
}

// GetMetrics returns the current subscriber metrics.
func (s *Subscriber) GetMetrics() SubscriberMetrics {
	return SubscriberMetrics{
		ReceivedCount:   atomic.LoadInt64(&s.metrics.ReceivedCount),
		ProcessedCount:  atomic.LoadInt64(&s.metrics.ProcessedCount),
		FailedCount:     atomic.LoadInt64(&s.metrics.FailedCount),
		RedeliveryCount: atomic.LoadInt64(&s.metrics.RedeliveryCount),
		LastReceivedAt:  s.metrics.LastReceivedAt,
	}
}

// IsRunning returns whether the subscriber is running.
func (s *Subscriber) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// Close closes the subscriber and its connection.
func (s *Subscriber) Close() error {
	if err := s.Stop(); err != nil {
		return err
	}
	if s.conn != nil {
		s.conn.Close()
	}
	return nil
}
