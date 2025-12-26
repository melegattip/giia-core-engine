package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEventEnvelope(t *testing.T) {
	payload := map[string]interface{}{
		"id":     "test-123",
		"amount": 100.50,
	}

	envelope, err := NewEventEnvelope(
		"execution.purchase_order.created",
		"purchase_order.created",
		"execution-service",
		"org-123",
		payload,
	)

	require.NoError(t, err)
	assert.NotEmpty(t, envelope.ID)
	assert.Equal(t, "execution.purchase_order.created", envelope.Subject)
	assert.Equal(t, "purchase_order.created", envelope.Type)
	assert.Equal(t, "execution-service", envelope.Source)
	assert.Equal(t, "org-123", envelope.OrganizationID)
	assert.Equal(t, "1.0", envelope.SchemaVersion)
	assert.False(t, envelope.Timestamp.IsZero())
	assert.NotEmpty(t, envelope.Payload)
}

func TestNewEventEnvelopeWithCorrelation(t *testing.T) {
	payload := map[string]string{"key": "value"}

	envelope, err := NewEventEnvelopeWithCorrelation(
		"execution.purchase_order.created",
		"purchase_order.created",
		"execution-service",
		"org-123",
		"corr-123",
		"cause-123",
		payload,
	)

	require.NoError(t, err)
	assert.Equal(t, "corr-123", envelope.CorrelationID)
	assert.Equal(t, "cause-123", envelope.CausationID)
}

func TestEventEnvelope_Validate(t *testing.T) {
	tests := []struct {
		name      string
		envelope  *EventEnvelope
		expectErr error
	}{
		{
			name: "valid envelope",
			envelope: &EventEnvelope{
				ID:             "env-123",
				Subject:        "test.subject",
				OrganizationID: "org-123",
				Source:         "test-service",
				Type:           "test.event",
				Timestamp:      time.Now(),
			},
			expectErr: nil,
		},
		{
			name: "missing ID",
			envelope: &EventEnvelope{
				Subject:        "test.subject",
				OrganizationID: "org-123",
				Source:         "test-service",
				Type:           "test.event",
				Timestamp:      time.Now(),
			},
			expectErr: ErrMissingEventID,
		},
		{
			name: "missing subject",
			envelope: &EventEnvelope{
				ID:             "env-123",
				OrganizationID: "org-123",
				Source:         "test-service",
				Type:           "test.event",
				Timestamp:      time.Now(),
			},
			expectErr: ErrMissingSubject,
		},
		{
			name: "missing organization ID",
			envelope: &EventEnvelope{
				ID:        "env-123",
				Subject:   "test.subject",
				Source:    "test-service",
				Type:      "test.event",
				Timestamp: time.Now(),
			},
			expectErr: ErrMissingOrganizationID,
		},
		{
			name: "missing source",
			envelope: &EventEnvelope{
				ID:             "env-123",
				Subject:        "test.subject",
				OrganizationID: "org-123",
				Type:           "test.event",
				Timestamp:      time.Now(),
			},
			expectErr: ErrMissingSource,
		},
		{
			name: "missing type",
			envelope: &EventEnvelope{
				ID:             "env-123",
				Subject:        "test.subject",
				OrganizationID: "org-123",
				Source:         "test-service",
				Timestamp:      time.Now(),
			},
			expectErr: ErrMissingType,
		},
		{
			name: "missing timestamp",
			envelope: &EventEnvelope{
				ID:             "env-123",
				Subject:        "test.subject",
				OrganizationID: "org-123",
				Source:         "test-service",
				Type:           "test.event",
			},
			expectErr: ErrMissingTimestamp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.envelope.Validate()
			if tt.expectErr != nil {
				assert.Equal(t, tt.expectErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEventEnvelope_ToJSON(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"key": "value"})
	envelope := &EventEnvelope{
		ID:             "env-123",
		Subject:        "test.subject",
		CorrelationID:  "corr-123",
		OrganizationID: "org-123",
		Source:         "test-service",
		Type:           "test.event",
		SchemaVersion:  "1.0",
		Timestamp:      time.Now().UTC(),
		Payload:        payload,
	}

	data, err := envelope.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it can be parsed back
	var decoded EventEnvelope
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, envelope.ID, decoded.ID)
}

func TestEventEnvelope_UnmarshalPayload(t *testing.T) {
	type TestPayload struct {
		ID     string  `json:"id"`
		Amount float64 `json:"amount"`
	}

	originalPayload := TestPayload{ID: "test-123", Amount: 100.50}
	payloadBytes, _ := json.Marshal(originalPayload)

	envelope := &EventEnvelope{
		Payload: payloadBytes,
	}

	var decoded TestPayload
	err := envelope.UnmarshalPayload(&decoded)
	require.NoError(t, err)
	assert.Equal(t, "test-123", decoded.ID)
	assert.Equal(t, 100.50, decoded.Amount)
}

func TestFromJSON(t *testing.T) {
	payload, _ := json.Marshal(map[string]string{"key": "value"})
	original := &EventEnvelope{
		ID:             "env-123",
		Subject:        "test.subject",
		CorrelationID:  "corr-123",
		OrganizationID: "org-123",
		Source:         "test-service",
		Type:           "test.event",
		SchemaVersion:  "1.0",
		Timestamp:      time.Now().UTC(),
		Payload:        payload,
	}

	data, _ := json.Marshal(original)

	decoded, err := FromJSON(data)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	assert.Equal(t, original.Subject, decoded.Subject)
	assert.Equal(t, original.Type, decoded.Type)
}

func TestFromJSON_InvalidJSON(t *testing.T) {
	invalidData := []byte("not valid json")

	_, err := FromJSON(invalidData)
	assert.Error(t, err)
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "id",
		Message: "event ID is required",
	}

	assert.Equal(t, "event ID is required", err.Error())
	assert.Equal(t, "id", err.Field)
}

func TestNewEventEnvelope_InvalidPayload(t *testing.T) {
	// Create a channel which cannot be marshaled to JSON
	invalidPayload := make(chan int)

	_, err := NewEventEnvelope(
		"test.subject",
		"test.type",
		"test-service",
		"org-123",
		invalidPayload,
	)

	assert.Error(t, err)
}
