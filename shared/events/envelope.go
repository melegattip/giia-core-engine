// Package events provides shared event types and utilities for NATS messaging.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventEnvelope is the standard wrapper for all events across the platform.
// It provides consistent metadata for event routing, correlation, and auditing.
type EventEnvelope struct {
	// ID is the unique identifier for this event
	ID string `json:"id"`

	// Subject is the NATS subject the event was published to
	Subject string `json:"subject"`

	// CorrelationID links related events together (e.g., request-response chains)
	CorrelationID string `json:"correlation_id,omitempty"`

	// CausationID references the event that caused this event
	CausationID string `json:"causation_id,omitempty"`

	// OrganizationID is the tenant identifier for multi-tenancy
	OrganizationID string `json:"organization_id"`

	// Source identifies the service that published the event
	Source string `json:"source"`

	// Type is the event type (e.g., "purchase_order.created")
	Type string `json:"type"`

	// SchemaVersion for backward compatibility
	SchemaVersion string `json:"schema_version"`

	// Timestamp when the event was created
	Timestamp time.Time `json:"timestamp"`

	// Payload contains the actual event data
	Payload json.RawMessage `json:"payload"`
}

// NewEventEnvelope creates a new event envelope with default values.
func NewEventEnvelope(subject, eventType, source, organizationID string, payload interface{}) (*EventEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &EventEnvelope{
		ID:             uuid.New().String(),
		Subject:        subject,
		OrganizationID: organizationID,
		Source:         source,
		Type:           eventType,
		SchemaVersion:  "1.0",
		Timestamp:      time.Now().UTC(),
		Payload:        payloadBytes,
	}, nil
}

// NewEventEnvelopeWithCorrelation creates an event envelope with correlation tracking.
func NewEventEnvelopeWithCorrelation(subject, eventType, source, organizationID, correlationID, causationID string, payload interface{}) (*EventEnvelope, error) {
	envelope, err := NewEventEnvelope(subject, eventType, source, organizationID, payload)
	if err != nil {
		return nil, err
	}
	envelope.CorrelationID = correlationID
	envelope.CausationID = causationID
	return envelope, nil
}

// Validate checks that the envelope has all required fields.
func (e *EventEnvelope) Validate() error {
	if e.ID == "" {
		return ErrMissingEventID
	}
	if e.Subject == "" {
		return ErrMissingSubject
	}
	if e.OrganizationID == "" {
		return ErrMissingOrganizationID
	}
	if e.Source == "" {
		return ErrMissingSource
	}
	if e.Type == "" {
		return ErrMissingType
	}
	if e.Timestamp.IsZero() {
		return ErrMissingTimestamp
	}
	return nil
}

// ToJSON serializes the envelope to JSON.
func (e *EventEnvelope) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// UnmarshalPayload decodes the payload into the provided target.
func (e *EventEnvelope) UnmarshalPayload(target interface{}) error {
	return json.Unmarshal(e.Payload, target)
}

// FromJSON deserializes an envelope from JSON.
func FromJSON(data []byte) (*EventEnvelope, error) {
	var envelope EventEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// Error types for envelope validation.
var (
	ErrMissingEventID        = &ValidationError{Field: "id", Message: "event ID is required"}
	ErrMissingSubject        = &ValidationError{Field: "subject", Message: "subject is required"}
	ErrMissingOrganizationID = &ValidationError{Field: "organization_id", Message: "organization ID is required"}
	ErrMissingSource         = &ValidationError{Field: "source", Message: "source is required"}
	ErrMissingType           = &ValidationError{Field: "type", Message: "type is required"}
	ErrMissingTimestamp      = &ValidationError{Field: "timestamp", Message: "timestamp is required"}
)

// ValidationError represents a validation error with field context.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
