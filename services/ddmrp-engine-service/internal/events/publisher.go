// Package events provides NATS event publishing for the DDMRP Engine Service.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
)

const (
	// Default retry configuration
	defaultMaxRetries     = 3
	defaultInitialBackoff = 100 * time.Millisecond
	defaultMaxBackoff     = 2 * time.Second

	// Source identifier
	sourceService = "ddmrp-engine-service"

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

// Publisher provides event publishing capabilities for the DDMRP Engine Service.
type Publisher struct {
	js      nats.JetStreamContext
	conn    *nats.Conn
	config  *PublisherConfig
	metrics *PublisherMetrics
	enabled bool
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
func NewNoOpPublisher() *Publisher {
	return &Publisher{
		enabled: false,
		metrics: &PublisherMetrics{},
		config:  DefaultPublisherConfig(),
	}
}

// publish is the internal publish method with retry logic.
func (p *Publisher) publish(ctx context.Context, subject, eventType, organizationID string, payload interface{}) error {
	if !p.enabled {
		return nil
	}

	atomic.AddInt64(&p.metrics.PublishCount, 1)
	start := time.Now()

	envelope, err := p.createEnvelope(subject, eventType, organizationID, payload)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to create event envelope: %w", err)
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	if err := p.publishWithRetry(ctx, subject, data); err != nil {
		atomic.AddInt64(&p.metrics.FailureCount, 1)
		return err
	}

	atomic.AddInt64(&p.metrics.SuccessCount, 1)
	p.metrics.LastPublishAt = time.Now()

	latency := time.Since(start)
	if p.metrics.AverageLatency == 0 {
		p.metrics.AverageLatency = latency
	} else {
		p.metrics.AverageLatency = (p.metrics.AverageLatency + latency) / 2
	}

	return nil
}

// createEnvelope creates a new event envelope.
func (p *Publisher) createEnvelope(subject, eventType, organizationID string, payload interface{}) (*EventEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	return &EventEnvelope{
		ID:             id,
		Subject:        subject,
		CorrelationID:  id,
		OrganizationID: organizationID,
		Source:         sourceService,
		Type:           eventType,
		SchemaVersion:  schemaVersion,
		Timestamp:      time.Now().UTC(),
		Payload:        payloadBytes,
	}, nil
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

// PublishBufferCreated publishes a buffer created event.
func (p *Publisher) PublishBufferCreated(ctx context.Context, buffer *domain.Buffer) error {
	event := &BufferCreatedEvent{
		BufferID:       buffer.ID.String(),
		OrganizationID: buffer.OrganizationID.String(),
		ProductID:      buffer.ProductID.String(),
		ProfileType:    buffer.BufferProfileID.String(),
		CreatedAt:      buffer.CreatedAt,
	}
	return p.publish(ctx, SubjectBufferCreated, TypeBufferCreated, buffer.OrganizationID.String(), event)
}

// PublishBufferUpdated publishes a buffer updated event.
func (p *Publisher) PublishBufferUpdated(ctx context.Context, buffer *domain.Buffer) error {
	event := &BufferUpdatedEvent{
		BufferID:       buffer.ID.String(),
		OrganizationID: buffer.OrganizationID.String(),
		ProductID:      buffer.ProductID.String(),
		UpdatedAt:      buffer.UpdatedAt,
	}
	return p.publish(ctx, SubjectBufferUpdated, TypeBufferUpdated, buffer.OrganizationID.String(), event)
}

// PublishBufferCalculated publishes a buffer calculation event.
func (p *Publisher) PublishBufferCalculated(ctx context.Context, buffer *domain.Buffer) error {
	event := &BufferCalculatedEvent{
		BufferID:        buffer.ID.String(),
		OrganizationID:  buffer.OrganizationID.String(),
		ProductID:       buffer.ProductID.String(),
		DLT:             buffer.LTD,
		ADU:             buffer.CPD,
		RedZoneBase:     buffer.RedBase,
		RedZoneSafety:   buffer.RedSafe,
		RedZone:         buffer.RedZone,
		YellowZone:      buffer.YellowZone,
		GreenZone:       buffer.GreenZone,
		CPD:             buffer.CPD,
		TOG:             buffer.TopOfGreen,
		TOR:             buffer.TopOfRed,
		TOY:             buffer.TopOfYellow,
		OnHandQty:       buffer.OnHand,
		OpenPOQty:       buffer.OnOrder,
		OpenSOQty:       buffer.QualifiedDemand,
		NetFlowPosition: buffer.NetFlowPosition,
		CalculatedAt:    time.Now().UTC(),
	}
	return p.publish(ctx, SubjectBufferCalculated, TypeBufferCalculated, buffer.OrganizationID.String(), event)
}

// PublishBufferStatusChanged publishes a buffer status change event.
func (p *Publisher) PublishBufferStatusChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType) error {
	event := &BufferStatusChangedEvent{
		BufferID:       buffer.ID.String(),
		OrganizationID: buffer.OrganizationID.String(),
		ProductID:      buffer.ProductID.String(),
		OldZone:        string(oldZone),
		NewZone:        string(buffer.Zone),
		NewNFP:         buffer.NetFlowPosition,
		AlertLevel:     string(buffer.AlertLevel),
		TOG:            buffer.TopOfGreen,
		TOY:            buffer.TopOfYellow,
		TOR:            buffer.TopOfRed,
		ChangedAt:      time.Now().UTC(),
	}
	return p.publish(ctx, SubjectBufferStatusChanged, TypeBufferStatusChanged, buffer.OrganizationID.String(), event)
}

// PublishBufferAlertTriggered publishes a buffer alert event.
func (p *Publisher) PublishBufferAlertTriggered(ctx context.Context, buffer *domain.Buffer, alertType, message string) error {
	event := &BufferAlertTriggeredEvent{
		BufferID:       buffer.ID.String(),
		OrganizationID: buffer.OrganizationID.String(),
		ProductID:      buffer.ProductID.String(),
		AlertType:      alertType,
		AlertLevel:     string(buffer.AlertLevel),
		Zone:           string(buffer.Zone),
		NFP:            buffer.NetFlowPosition,
		TOG:            buffer.TopOfGreen,
		TOY:            buffer.TopOfYellow,
		TOR:            buffer.TopOfRed,
		Message:        message,
		TriggeredAt:    time.Now().UTC(),
	}

	// Calculate replenishment quantity if in red zone
	if buffer.Zone == domain.ZoneRed {
		event.ReplenishmentQty = buffer.TopOfGreen - buffer.NetFlowPosition
		event.SuggestedOrderQty = event.ReplenishmentQty
	}

	return p.publish(ctx, SubjectBufferAlertTriggered, TypeBufferAlertTriggered, buffer.OrganizationID.String(), event)
}

// PublishBufferZoneChanged publishes a buffer zone change event.
func (p *Publisher) PublishBufferZoneChanged(ctx context.Context, buffer *domain.Buffer, oldZone domain.ZoneType, reason string) error {
	event := &BufferZoneChangedEvent{
		BufferStatusChangedEvent: BufferStatusChangedEvent{
			BufferID:       buffer.ID.String(),
			OrganizationID: buffer.OrganizationID.String(),
			ProductID:      buffer.ProductID.String(),
			OldZone:        string(oldZone),
			NewZone:        string(buffer.Zone),
			NewNFP:         buffer.NetFlowPosition,
			AlertLevel:     string(buffer.AlertLevel),
			TOG:            buffer.TopOfGreen,
			TOY:            buffer.TopOfYellow,
			TOR:            buffer.TopOfRed,
			ChangedAt:      time.Now().UTC(),
		},
		TransitionReason: reason,
	}
	return p.publish(ctx, SubjectBufferZoneChanged, TypeBufferZoneChanged, buffer.OrganizationID.String(), event)
}

// PublishFADCreated publishes a FAD created event.
func (p *Publisher) PublishFADCreated(ctx context.Context, fad *domain.DemandAdjustment) error {
	event := &FADCreatedEvent{
		FADID:          fad.ID.String(),
		OrganizationID: fad.OrganizationID.String(),
		ProductID:      fad.ProductID.String(),
		AdjustmentType: string(fad.AdjustmentType),
		Factor:         fad.Factor,
		StartDate:      fad.StartDate,
		EndDate:        fad.EndDate,
		Reason:         fad.Reason,
		CreatedBy:      fad.CreatedBy.String(),
		CreatedAt:      fad.CreatedAt,
	}
	return p.publish(ctx, SubjectFADCreated, TypeFADCreated, fad.OrganizationID.String(), event)
}

// PublishFADUpdated publishes a FAD updated event.
func (p *Publisher) PublishFADUpdated(ctx context.Context, fad *domain.DemandAdjustment) error {
	event := &FADUpdatedEvent{
		FADID:          fad.ID.String(),
		OrganizationID: fad.OrganizationID.String(),
		ProductID:      fad.ProductID.String(),
		UpdatedAt:      time.Now().UTC(),
	}
	return p.publish(ctx, SubjectFADUpdated, TypeFADUpdated, fad.OrganizationID.String(), event)
}

// PublishFADDeleted publishes a FAD deleted event.
func (p *Publisher) PublishFADDeleted(ctx context.Context, fadID, orgID string) error {
	event := &FADDeletedEvent{
		FADID:          fadID,
		OrganizationID: orgID,
		DeletedAt:      time.Now().UTC(),
	}
	return p.publish(ctx, SubjectFADDeleted, TypeFADDeleted, orgID, event)
}

// PublishADUCalculated publishes an ADU calculation event.
func (p *Publisher) PublishADUCalculated(ctx context.Context, buffer *domain.Buffer, previousADU float64, periodDays int) error {
	event := &ADUCalculatedEvent{
		BufferID:       buffer.ID.String(),
		OrganizationID: buffer.OrganizationID.String(),
		ProductID:      buffer.ProductID.String(),
		ADU:            buffer.CPD,
		PreviousADU:    previousADU,
		PeriodDays:     periodDays,
		CalculatedAt:   time.Now().UTC(),
	}
	return p.publish(ctx, SubjectADUCalculated, TypeADUCalculated, buffer.OrganizationID.String(), event)
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
