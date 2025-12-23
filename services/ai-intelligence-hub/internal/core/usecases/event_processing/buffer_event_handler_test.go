package event_processing_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/mocks"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/usecases/event_processing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBufferEventHandler_Handle_BelowMinimum(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAnalyzer := mocks.NewMockStockoutRiskAnalyzer()
	testLogger := logger.New("test", "debug")

	handler := event_processing.NewBufferEventHandler(mockAnalyzer, testLogger)

	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"product_id":    "PROD-123",
			"current_stock": 50.0,
			"min_buffer":    100.0,
		},
	}

	// Mock analyzer should be called for below_minimum events
	mockAnalyzer.On("Execute", ctx, event).Return(nil)

	// Execute
	err := handler.Handle(ctx, event)

	// Assert
	require.NoError(t, err)
	mockAnalyzer.AssertExpectations(t)
}

func TestBufferEventHandler_Handle_OtherBufferEvents(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"Buffer created", "buffer.created"},
		{"Buffer updated", "buffer.updated"},
		{"Buffer deleted", "buffer.deleted"},
		{"Buffer green zone", "buffer.in_green_zone"},
		{"Buffer yellow zone", "buffer.in_yellow_zone"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			mockAnalyzer := mocks.NewMockStockoutRiskAnalyzer()
			testLogger := logger.New("test", "debug")

			handler := event_processing.NewBufferEventHandler(mockAnalyzer, testLogger)

			event := &events.Event{
				ID:             "evt-other",
				Type:           tt.eventType,
				OrganizationID: uuid.New().String(),
				Data:           map[string]interface{}{},
			}

			// Execute - should not call analyzer for non-below_minimum events
			err := handler.Handle(ctx, event)

			// Assert
			require.NoError(t, err)
			// Analyzer should NOT be called
			mockAnalyzer.AssertNotCalled(t, "Execute")
		})
	}
}

func TestBufferEventHandler_Handle_AnalyzerError(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAnalyzer := mocks.NewMockStockoutRiskAnalyzer()
	testLogger := logger.New("test", "debug")

	handler := event_processing.NewBufferEventHandler(mockAnalyzer, testLogger)

	event := &events.Event{
		ID:             "evt-error",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"product_id": "PROD-ERROR",
		},
	}

	// Mock analyzer returns error
	mockAnalyzer.On("Execute", ctx, event).Return(assert.AnError)

	// Execute
	err := handler.Handle(ctx, event)

	// Assert - error should be propagated
	require.Error(t, err)
	mockAnalyzer.AssertExpectations(t)
}

func TestBufferEventHandler_Handle_NilEvent(t *testing.T) {
	// Note: Current implementation does not handle nil events gracefully
	// This would cause a panic. In future, we might want to add nil checking.
	t.Skip("Handler does not currently handle nil events - would panic on event.Type access")
}

func TestBufferEventHandler_MultipleEvents(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAnalyzer := mocks.NewMockStockoutRiskAnalyzer()
	testLogger := logger.New("test", "debug")

	handler := event_processing.NewBufferEventHandler(mockAnalyzer, testLogger)

	// Create multiple events
	events := []*events.Event{
		{
			ID:             "evt-1",
			Type:           "buffer.below_minimum",
			OrganizationID: uuid.New().String(),
			Data:           map[string]interface{}{"product_id": "PROD-1"},
		},
		{
			ID:             "evt-2",
			Type:           "buffer.in_green_zone",
			OrganizationID: uuid.New().String(),
			Data:           map[string]interface{}{"product_id": "PROD-2"},
		},
		{
			ID:             "evt-3",
			Type:           "buffer.below_minimum",
			OrganizationID: uuid.New().String(),
			Data:           map[string]interface{}{"product_id": "PROD-3"},
		},
	}

	// Only below_minimum events should trigger analyzer
	mockAnalyzer.On("Execute", ctx, events[0]).Return(nil)
	mockAnalyzer.On("Execute", ctx, events[2]).Return(nil)

	// Execute all events
	for _, event := range events {
		err := handler.Handle(ctx, event)
		require.NoError(t, err)
	}

	// Assert - analyzer called exactly twice (for the two below_minimum events)
	mockAnalyzer.AssertExpectations(t)
	mockAnalyzer.AssertNumberOfCalls(t, "Execute", 2)
}
