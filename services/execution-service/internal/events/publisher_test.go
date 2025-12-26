package events

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockNATSConn is a mock NATS connection for testing.
type MockNATSConn struct {
	mu       sync.Mutex
	messages []MockMessage
	closed   bool
}

type MockMessage struct {
	Subject string
	Data    []byte
}

// MockJetStreamContext is a mock JetStream context.
type MockJetStreamContext struct {
	mu       sync.Mutex
	messages []MockMessage
	failNext bool
}

func (m *MockJetStreamContext) Publish(subject string, data []byte) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failNext {
		m.failNext = false
		return nil, assert.AnError
	}

	m.messages = append(m.messages, MockMessage{Subject: subject, Data: data})
	return nil, nil
}

func (m *MockJetStreamContext) PublishAsync(subject string, data []byte) (interface{}, error) {
	return m.Publish(subject, data)
}

func (m *MockJetStreamContext) GetMessages() []MockMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockMessage{}, m.messages...)
}

func (m *MockJetStreamContext) SetFailNext() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNext = true
}

func TestNewNoOpPublisher(t *testing.T) {
	publisher := NewNoOpPublisher()

	assert.NotNil(t, publisher)
	assert.False(t, publisher.IsEnabled())
}

func TestNoOpPublisher_Publish(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.Publish(context.Background(), "test.subject", "test.type", "org-123", map[string]string{"key": "value"})

	assert.NoError(t, err)

	metrics := publisher.GetMetrics()
	assert.Equal(t, int64(0), metrics.PublishCount)
}

func TestDefaultPublisherConfig(t *testing.T) {
	config := DefaultPublisherConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 2*time.Second, config.MaxBackoff)
	assert.False(t, config.AsyncMode)
}

func TestEventEnvelope_Marshal(t *testing.T) {
	payload := map[string]interface{}{
		"id":     "test-123",
		"amount": 100.50,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	envelope := &EventEnvelope{
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
	assert.Equal(t, envelope.Subject, decoded.Subject)
	assert.Equal(t, envelope.CorrelationID, decoded.CorrelationID)
	assert.Equal(t, envelope.OrganizationID, decoded.OrganizationID)
	assert.Equal(t, envelope.Source, decoded.Source)
	assert.Equal(t, envelope.Type, decoded.Type)
}

func TestPublisherMetrics(t *testing.T) {
	publisher := NewNoOpPublisher()

	metrics := publisher.GetMetrics()

	assert.Equal(t, int64(0), metrics.PublishCount)
	assert.Equal(t, int64(0), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount)
	assert.Equal(t, int64(0), metrics.RetryCount)
}

func TestMinDuration(t *testing.T) {
	tests := []struct {
		a, b, expected time.Duration
	}{
		{1 * time.Second, 2 * time.Second, 1 * time.Second},
		{2 * time.Second, 1 * time.Second, 1 * time.Second},
		{1 * time.Second, 1 * time.Second, 1 * time.Second},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		assert.Equal(t, tt.expected, result)
	}
}

func TestPurchaseOrderEvent_Marshal(t *testing.T) {
	event := PurchaseOrderEvent{
		ID:             "po-123",
		OrganizationID: "org-123",
		PONumber:       "PO-2024-001",
		SupplierID:     "sup-123",
		SupplierName:   "Acme Corp",
		Status:         "pending",
		TotalAmount:    1000.50,
		Currency:       "USD",
		ItemCount:      5,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded PurchaseOrderEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.PONumber, decoded.PONumber)
	assert.Equal(t, event.TotalAmount, decoded.TotalAmount)
}

func TestSalesOrderEvent_Marshal(t *testing.T) {
	event := SalesOrderEvent{
		ID:             "so-123",
		OrganizationID: "org-123",
		SONumber:       "SO-2024-001",
		CustomerID:     "cust-123",
		CustomerName:   "Test Customer",
		Status:         "shipped",
		TotalAmount:    500.00,
		Currency:       "USD",
		ItemCount:      3,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded SalesOrderEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.SONumber, decoded.SONumber)
	assert.Equal(t, event.Status, decoded.Status)
}

func TestInventoryEvent_Marshal(t *testing.T) {
	event := InventoryEvent{
		ID:              "txn-123",
		OrganizationID:  "org-123",
		ProductID:       "prod-123",
		ProductSKU:      "SKU-001",
		LocationID:      "loc-123",
		TransactionType: "receipt",
		Quantity:        100.0,
		PreviousBalance: 50.0,
		NewBalance:      150.0,
		ReferenceType:   "purchase_order",
		ReferenceID:     "po-123",
		CreatedAt:       time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded InventoryEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.Quantity, decoded.Quantity)
	assert.Equal(t, event.NewBalance, decoded.NewBalance)
}

func TestAlertEvent_Marshal(t *testing.T) {
	event := AlertEvent{
		ID:             "alert-123",
		OrganizationID: "org-123",
		AlertType:      "low_stock",
		Severity:       "high",
		Status:         "active",
		Title:          "Low Stock Alert",
		Message:        "Product SKU-001 is below reorder point",
		ResourceType:   "product",
		ResourceID:     "prod-123",
		Metadata: map[string]string{
			"current_qty":   "10",
			"reorder_point": "50",
		},
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(event)
	require.NoError(t, err)

	var decoded AlertEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, event.ID, decoded.ID)
	assert.Equal(t, event.AlertType, decoded.AlertType)
	assert.Equal(t, event.Severity, decoded.Severity)
	assert.Equal(t, event.Metadata["current_qty"], decoded.Metadata["current_qty"])
}

func TestPublisher_Close(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.Close()
	assert.NoError(t, err)
}

func TestPublisher_PublishWithCorrelation(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.PublishWithCorrelation(
		context.Background(),
		"test.subject",
		"test.type",
		"org-123",
		"corr-123",
		"cause-123",
		map[string]string{"key": "value"},
	)

	assert.NoError(t, err)
}

func TestPublisher_PublishAsync(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.PublishAsync(
		context.Background(),
		"test.subject",
		"test.type",
		"org-123",
		map[string]string{"key": "value"},
	)

	assert.NoError(t, err)
}
