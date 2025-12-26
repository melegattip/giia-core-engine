package handlers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockReplenishmentService implements ReplenishmentService for testing.
type MockReplenishmentService struct {
	createCalled    bool
	evaluateCalled  bool
	shouldReplenish bool
	quantity        float64
}

func (m *MockReplenishmentService) CreateReplenishmentOrder(ctx context.Context, req *ReplenishmentRequest) (*ReplenishmentResponse, error) {
	m.createCalled = true
	return &ReplenishmentResponse{
		OrderID:     "order-123",
		OrderType:   "purchase_order",
		OrderNumber: "PO-2024-001",
		Status:      "created",
		CreatedAt:   time.Now(),
	}, nil
}

func (m *MockReplenishmentService) EvaluateReplenishment(ctx context.Context, productID, locationID string) (*ReplenishmentRecommendation, error) {
	m.evaluateCalled = true
	return &ReplenishmentRecommendation{
		ShouldReplenish: m.shouldReplenish,
		Quantity:        m.quantity,
		Priority:        "high",
		Reason:          "Buffer in red zone",
		SupplierID:      "sup-123",
	}, nil
}

func TestNewDDMRPHandler(t *testing.T) {
	logger := &MockLogger{}
	replenishmentSvc := &MockReplenishmentService{}

	handler := NewDDMRPHandler(replenishmentSvc, logger, nil)

	assert.NotNil(t, handler)
}

func TestDefaultDDMRPHandlerConfig(t *testing.T) {
	config := DefaultDDMRPHandlerConfig()

	assert.False(t, config.AutoReplenishment)
}

func TestDDMRPHandler_HandleBufferCalculated(t *testing.T) {
	logger := &MockLogger{}
	handler := NewDDMRPHandler(nil, logger, nil)

	payload, err := json.Marshal(BufferCalculatedEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		CPD:            100.0,
		TOG:            1200.0,
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		Subject:        "ddmrp.buffer.calculated",
		OrganizationID: "org-123",
		Source:         "ddmrp-engine-service",
		Type:           EventBufferCalculated,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)
	assert.Len(t, logger.infoLogs, 1)
}

func TestDDMRPHandler_HandleBufferStatusChanged(t *testing.T) {
	logger := &MockLogger{}
	handler := NewDDMRPHandler(nil, logger, nil)

	payload, err := json.Marshal(BufferStatusChangedEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		OldZone:        "yellow",
		NewZone:        "red",
		NFP:            100.0,
		AlertLevel:     "warning",
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		Subject:        "ddmrp.buffer.status_changed",
		OrganizationID: "org-123",
		Source:         "ddmrp-engine-service",
		Type:           EventBufferStatusChanged,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)
	// Should log info and warn (entering red zone)
	assert.Len(t, logger.infoLogs, 1)
	assert.Len(t, logger.warnLogs, 1)
}

func TestDDMRPHandler_HandleBufferAlertTriggered_NoAutoReplenishment(t *testing.T) {
	logger := &MockLogger{}
	replenishmentSvc := &MockReplenishmentService{}

	// Auto-replenishment disabled by default
	handler := NewDDMRPHandler(replenishmentSvc, logger, nil)

	payload, err := json.Marshal(BufferAlertTriggeredEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		AlertType:      "low_stock",
		AlertLevel:     "critical",
		Zone:           "red",
		NFP:            50.0,
		TOG:            1200.0,
		Message:        "Critical low stock",
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		Subject:        "ddmrp.buffer.alert_triggered",
		OrganizationID: "org-123",
		Source:         "ddmrp-engine-service",
		Type:           EventBufferAlertTriggered,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	// Should not have called replenishment service
	assert.False(t, replenishmentSvc.createCalled)
	assert.False(t, replenishmentSvc.evaluateCalled)
}

func TestDDMRPHandler_HandleBufferAlertTriggered_WithAutoReplenishment(t *testing.T) {
	logger := &MockLogger{}
	replenishmentSvc := &MockReplenishmentService{
		shouldReplenish: true,
		quantity:        500.0,
	}

	config := &DDMRPHandlerConfig{
		AutoReplenishment: true,
	}
	handler := NewDDMRPHandler(replenishmentSvc, logger, config)

	payload, err := json.Marshal(BufferAlertTriggeredEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		AlertType:      "low_stock",
		AlertLevel:     "critical",
		Zone:           "red",
		NFP:            50.0,
		TOG:            1200.0,
		Message:        "Critical low stock",
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		Subject:        "ddmrp.buffer.alert_triggered",
		OrganizationID: "org-123",
		Source:         "ddmrp-engine-service",
		Type:           EventBufferAlertTriggered,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	// Should have called replenishment service
	assert.True(t, replenishmentSvc.evaluateCalled)
	assert.True(t, replenishmentSvc.createCalled)
}

func TestDDMRPHandler_HandleBufferAlertTriggered_NoReplenishmentNeeded(t *testing.T) {
	logger := &MockLogger{}
	replenishmentSvc := &MockReplenishmentService{
		shouldReplenish: false,
		quantity:        0,
	}

	config := &DDMRPHandlerConfig{
		AutoReplenishment: true,
	}
	handler := NewDDMRPHandler(replenishmentSvc, logger, config)

	payload, err := json.Marshal(BufferAlertTriggeredEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		AlertLevel:     "critical",
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		OrganizationID: "org-123",
		Type:           EventBufferAlertTriggered,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	assert.True(t, replenishmentSvc.evaluateCalled)
	assert.False(t, replenishmentSvc.createCalled) // Should not create order if not needed
}

func TestDDMRPHandler_HandleBufferZoneChanged(t *testing.T) {
	logger := &MockLogger{}
	handler := NewDDMRPHandler(nil, logger, nil)

	payload, err := json.Marshal(BufferStatusChangedEvent{
		BufferID:       "buf-123",
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		OldZone:        "green",
		NewZone:        "yellow",
	})
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		OrganizationID: "org-123",
		Type:           EventBufferZoneChanged,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)
	assert.Len(t, logger.infoLogs, 1)
}

func TestDDMRPHandler_HandleUnknownEventType(t *testing.T) {
	logger := &MockLogger{}
	handler := NewDDMRPHandler(nil, logger, nil)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		OrganizationID: "org-123",
		Type:           "ddmrp.unknown.event",
		Payload:        []byte("{}"),
		Timestamp:      time.Now(),
	}

	err := handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)
}

func TestDDMRPHandler_HandleInvalidPayload(t *testing.T) {
	logger := &MockLogger{}
	handler := NewDDMRPHandler(nil, logger, nil)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		OrganizationID: "org-123",
		Type:           EventBufferCalculated,
		Payload:        []byte("invalid json"),
		Timestamp:      time.Now(),
	}

	err := handler.Handle(context.Background(), envelope)
	assert.Error(t, err)
}

func TestDDMRPHandler_GetSubscriptionSubjects(t *testing.T) {
	handler := NewDDMRPHandler(nil, &MockLogger{}, nil)

	subjects := handler.GetSubscriptionSubjects()

	assert.Len(t, subjects, 1)
	assert.Contains(t, subjects, "ddmrp.buffer.>")
}

func TestReplenishmentRequest_Fields(t *testing.T) {
	req := &ReplenishmentRequest{
		OrganizationID: "org-123",
		ProductID:      "prod-123",
		LocationID:     "loc-123",
		Quantity:       500.0,
		Priority:       "high",
		SourceBufferID: "buf-123",
		TriggerEvent:   "evt-123",
	}

	assert.Equal(t, "org-123", req.OrganizationID)
	assert.Equal(t, 500.0, req.Quantity)
	assert.Equal(t, "high", req.Priority)
}

func TestReplenishmentResponse_Fields(t *testing.T) {
	resp := &ReplenishmentResponse{
		OrderID:     "order-123",
		OrderType:   "purchase_order",
		OrderNumber: "PO-2024-001",
		Status:      "created",
		CreatedAt:   time.Now(),
	}

	assert.Equal(t, "order-123", resp.OrderID)
	assert.Equal(t, "purchase_order", resp.OrderType)
	assert.Equal(t, "created", resp.Status)
}

func TestReplenishmentRecommendation_Fields(t *testing.T) {
	rec := &ReplenishmentRecommendation{
		ShouldReplenish: true,
		Quantity:        1000.0,
		Priority:        "critical",
		Reason:          "Below red zone",
		SupplierID:      "sup-123",
	}

	assert.True(t, rec.ShouldReplenish)
	assert.Equal(t, 1000.0, rec.Quantity)
	assert.Equal(t, "sup-123", rec.SupplierID)
}
