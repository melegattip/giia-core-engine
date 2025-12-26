// Package events provides NATS event publishing for the Execution Service.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	// Default retry configuration
	defaultMaxRetries     = 3
	defaultInitialBackoff = 100 * time.Millisecond
	defaultMaxBackoff     = 2 * time.Second

	// Source identifier
	sourceService = "execution-service"

	// Schema version
	schemaVersion = "1.0"
)

// PublisherConfig contains configuration for the event publisher.
type PublisherConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	AsyncMode      bool
}

// DefaultPublisherConfig returns the default publisher configuration.
func DefaultPublisherConfig() *PublisherConfig {
	return &PublisherConfig{
		MaxRetries:     defaultMaxRetries,
		InitialBackoff: defaultInitialBackoff,
		MaxBackoff:     defaultMaxBackoff,
		AsyncMode:      false,
	}
}

// PublisherMetrics tracks publishing statistics.
type PublisherMetrics struct {
	PublishCount   int64
	SuccessCount   int64
	FailureCount   int64
	RetryCount     int64
	LastPublishAt  time.Time
	AverageLatency time.Duration
}

// Publisher provides event publishing capabilities for the Execution Service.
type Publisher struct {
	js      nats.JetStreamContext
	conn    *nats.Conn
	config  *PublisherConfig
	metrics *PublisherMetrics
	enabled bool
}

// EventEnvelope is the standard wrapper for all events.
type EventEnvelope struct {
	ID             string          `json:"id"`
	Subject        string          `json:"subject"`
	CorrelationID  string          `json:"correlation_id,omitempty"`
	CausationID    string          `json:"causation_id,omitempty"`
	OrganizationID string          `json:"organization_id"`
	Source         string          `json:"source"`
	Type           string          `json:"type"`
	SchemaVersion  string          `json:"schema_version"`
	Timestamp      time.Time       `json:"timestamp"`
	Payload        json.RawMessage `json:"payload"`
}

// NewPublisher creates a new event publisher with NATS JetStream.
func NewPublisher(nc *nats.Conn, config *PublisherConfig) (*Publisher, error) {
	if nc == nil {
		return &Publisher{enabled: false, metrics: &PublisherMetrics{}}, nil
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	if config == nil {
		config = DefaultPublisherConfig()
	}

	return &Publisher{
		js:      js,
		conn:    nc,
		config:  config,
		metrics: &PublisherMetrics{},
		enabled: true,
	}, nil
}

// NewNoOpPublisher creates a publisher that doesn't actually publish events.
// Useful for testing or when NATS is not available.
func NewNoOpPublisher() *Publisher {
	return &Publisher{
		enabled: false,
		metrics: &PublisherMetrics{},
		config:  DefaultPublisherConfig(),
	}
}

// Publish publishes an event to the specified subject with retry logic.
func (p *Publisher) Publish(ctx context.Context, subject, eventType, organizationID string, payload interface{}) error {
	return p.PublishWithCorrelation(ctx, subject, eventType, organizationID, "", "", payload)
}

// PublishWithCorrelation publishes an event with correlation and causation IDs.
func (p *Publisher) PublishWithCorrelation(ctx context.Context, subject, eventType, organizationID, correlationID, causationID string, payload interface{}) error {
	if !p.enabled {
		return nil
	}

	atomic.AddInt64(&p.metrics.PublishCount, 1)
	start := time.Now()

	// Create envelope
	envelope, err := p.createEnvelope(subject, eventType, organizationID, correlationID, causationID, payload)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to create event envelope: %w", err)
	}

	// Serialize
	data, err := json.Marshal(envelope)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Publish with retry
	if err := p.publishWithRetry(ctx, subject, data); err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return err
	}

	atomic.AddInt64(&p.metrics.SuccessCount, 1)
	p.metrics.LastPublishAt = time.Now()

	// Update average latency (simple moving average approximation)
	latency := time.Since(start)
	if p.metrics.AverageLatency == 0 {
		p.metrics.AverageLatency = latency
	} else {
		p.metrics.AverageLatency = (p.metrics.AverageLatency + latency) / 2
	}

	return nil
}

// PublishAsync publishes an event asynchronously without waiting for acknowledgment.
func (p *Publisher) PublishAsync(ctx context.Context, subject, eventType, organizationID string, payload interface{}) error {
	if !p.enabled {
		return nil
	}

	atomic.AddInt64(&p.metrics.PublishCount, 1)

	envelope, err := p.createEnvelope(subject, eventType, organizationID, "", "", payload)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to create event envelope: %w", err)
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	_, err = p.js.PublishAsync(subject, data)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to publish async: %w", err)
	}

	atomic.AddInt64(&p.metrics.SuccessCount, 1)
	p.metrics.LastPublishAt = time.Now()

	return nil
}

// createEnvelope creates a new event envelope.
func (p *Publisher) createEnvelope(subject, eventType, organizationID, correlationID, causationID string, payload interface{}) (*EventEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	envelope := &EventEnvelope{
		ID:             uuid.New().String(),
		Subject:        subject,
		CorrelationID:  correlationID,
		CausationID:    causationID,
		OrganizationID: organizationID,
		Source:         sourceService,
		Type:           eventType,
		SchemaVersion:  schemaVersion,
		Timestamp:      time.Now().UTC(),
		Payload:        payloadBytes,
	}

	// Auto-generate correlation ID if not provided
	if envelope.CorrelationID == "" {
		envelope.CorrelationID = envelope.ID
	}

	return envelope, nil
}

// publishWithRetry attempts to publish with exponential backoff.
func (p *Publisher) publishWithRetry(ctx context.Context, subject string, data []byte) error {
	var lastErr error
	backoff := p.config.InitialBackoff

	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := p.js.Publish(subject, data)
		if err == nil {
			return nil
		}

		lastErr = err
		atomic.AddInt64(&p.metrics.RetryCount, 1)

		if attempt < p.config.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff = min(backoff*2, p.config.MaxBackoff)
			}
		}
	}

	return fmt.Errorf("failed to publish after %d retries: %w", p.config.MaxRetries, lastErr)
}

// GetMetrics returns the current publisher metrics.
func (p *Publisher) GetMetrics() PublisherMetrics {
	return PublisherMetrics{
		PublishCount:   atomic.LoadInt64(&p.metrics.PublishCount),
		SuccessCount:   atomic.LoadInt64(&p.metrics.SuccessCount),
		FailureCount:   atomic.LoadInt64(&p.metrics.FailureCount),
		RetryCount:     atomic.LoadInt64(&p.metrics.RetryCount),
		LastPublishAt:  p.metrics.LastPublishAt,
		AverageLatency: p.metrics.AverageLatency,
	}
}

// IsEnabled returns whether the publisher is enabled.
func (p *Publisher) IsEnabled() bool {
	return p.enabled
}

// Close closes the publisher and its connection.
func (p *Publisher) Close() error {
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// min returns the minimum of two durations.
func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
