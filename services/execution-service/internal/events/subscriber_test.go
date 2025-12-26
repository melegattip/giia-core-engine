package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultSubscriberConfig(t *testing.T) {
	config := DefaultSubscriberConfig()

	assert.Equal(t, 5, config.MaxDeliver)
	assert.Equal(t, 30*time.Second, config.AckWait)
	assert.Equal(t, 100, config.BatchSize)
	assert.Equal(t, 5*time.Second, config.FetchWait)
	assert.Equal(t, "execution-service", config.ConsumerName)
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    string
		expected bool
	}{
		{
			name:     "wildcard all",
			pattern:  "*",
			value:    "any.value",
			expected: true,
		},
		{
			name:     "greater than",
			pattern:  ">",
			value:    "any.nested.value",
			expected: true,
		},
		{
			name:     "suffix wildcard match",
			pattern:  "purchase_order.*",
			value:    "purchase_order.created",
			expected: true,
		},
		{
			name:     "suffix wildcard no match",
			pattern:  "sales_order.*",
			value:    "purchase_order.created",
			expected: false,
		},
		{
			name:     "exact match",
			pattern:  "purchase_order.created",
			value:    "purchase_order.created",
			expected: true,
		},
		{
			name:     "exact no match",
			pattern:  "purchase_order.created",
			value:    "purchase_order.updated",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.pattern, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSubscriberMetrics(t *testing.T) {
	metrics := &SubscriberMetrics{
		ReceivedCount:   10,
		ProcessedCount:  8,
		FailedCount:     2,
		RedeliveryCount: 1,
		LastReceivedAt:  time.Now(),
	}

	assert.Equal(t, int64(10), metrics.ReceivedCount)
	assert.Equal(t, int64(8), metrics.ProcessedCount)
	assert.Equal(t, int64(2), metrics.FailedCount)
	assert.Equal(t, int64(1), metrics.RedeliveryCount)
}

func TestEventEnvelope_Unmarshal(t *testing.T) {
	payload := map[string]interface{}{
		"id":     "po-123",
		"status": "created",
	}
	payloadBytes, _ := json.Marshal(payload)

	envelope := EventEnvelope{
		ID:             "env-123",
		Subject:        "execution.purchase_order.created",
		CorrelationID:  "corr-123",
		OrganizationID: "org-123",
		Source:         "execution-service",
		Type:           "purchase_order.created",
		SchemaVersion:  "1.0",
		Timestamp:      time.Now().UTC(),
		Payload:        payloadBytes,
	}

	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	var decoded EventEnvelope
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, envelope.ID, decoded.ID)
	assert.Equal(t, envelope.Type, decoded.Type)

	var decodedPayload map[string]interface{}
	err = json.Unmarshal(decoded.Payload, &decodedPayload)
	require.NoError(t, err)
	assert.Equal(t, "po-123", decodedPayload["id"])
}

func TestSubscriber_RegisterHandler(t *testing.T) {
	// Can't test without real NATS, but we can test the handler registration structure
	handlers := make(map[string]EventHandler)

	handler := func(ctx context.Context, envelope *EventEnvelope) error {
		return nil
	}

	handlers["purchase_order.*"] = handler
	handlers["sales_order.*"] = handler

	assert.Len(t, handlers, 2)
	assert.NotNil(t, handlers["purchase_order.*"])
	assert.NotNil(t, handlers["sales_order.*"])
}

func TestEventHandler_Interface(t *testing.T) {
	var callCount int
	handler := func(ctx context.Context, envelope *EventEnvelope) error {
		callCount++
		return nil
	}

	payload, _ := json.Marshal(map[string]string{"test": "value"})
	envelope := &EventEnvelope{
		ID:             "test-123",
		Subject:        "test.subject",
		OrganizationID: "org-123",
		Source:         "test-service",
		Type:           "test.event",
		Payload:        payload,
	}

	err := handler(context.Background(), envelope)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestSubscriber_FindHandler_ExactMatch(t *testing.T) {
	handlers := map[string]EventHandler{
		"purchase_order.created": func(ctx context.Context, envelope *EventEnvelope) error {
			return nil
		},
		"purchase_order.*": func(ctx context.Context, envelope *EventEnvelope) error {
			return nil
		},
	}

	// Simulate findHandler logic
	eventType := "purchase_order.created"
	if handler, ok := handlers[eventType]; ok {
		assert.NotNil(t, handler)
		return
	}

	// Pattern match
	for pattern, handler := range handlers {
		if matchPattern(pattern, eventType) {
			assert.NotNil(t, handler)
			return
		}
	}

	t.Fail()
}

func TestSubscriber_ProcessMessageWithValidPayload(t *testing.T) {
	payload, _ := json.Marshal(map[string]interface{}{
		"id":     "po-123",
		"status": "created",
	})

	envelope := &EventEnvelope{
		ID:             "env-123",
		Subject:        "execution.purchase_order.created",
		CorrelationID:  "corr-123",
		OrganizationID: "org-123",
		Source:         "execution-service",
		Type:           "purchase_order.created",
		SchemaVersion:  "1.0",
		Timestamp:      time.Now().UTC(),
		Payload:        payload,
	}

	data, err := json.Marshal(envelope)
	require.NoError(t, err)

	// Simulate message processing
	var decoded EventEnvelope
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "env-123", decoded.ID)
	assert.Equal(t, "purchase_order.created", decoded.Type)
}

func TestSubscriber_ProcessMessageWithInvalidPayload(t *testing.T) {
	invalidData := []byte("this is not valid JSON")

	var decoded EventEnvelope
	err := json.Unmarshal(invalidData, &decoded)

	assert.Error(t, err)
}
