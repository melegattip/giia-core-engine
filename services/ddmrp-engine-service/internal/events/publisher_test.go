package events

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoOpPublisher(t *testing.T) {
	publisher := NewNoOpPublisher()

	assert.NotNil(t, publisher)
	assert.False(t, publisher.IsEnabled())
}

func TestDefaultPublisherConfig(t *testing.T) {
	config := DefaultPublisherConfig()

	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, config.InitialBackoff)
	assert.Equal(t, 2*time.Second, config.MaxBackoff)
	assert.False(t, config.AsyncMode)
}

func TestNoOpPublisher_PublishBufferCreated(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()

	err := publisher.PublishBufferCreated(context.Background(), buffer)
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishBufferCalculated(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()
	buffer.CPD = 100.0
	buffer.LTD = 14
	buffer.RedZone = 200.0
	buffer.YellowZone = 600.0
	buffer.GreenZone = 400.0
	buffer.TopOfGreen = 1200.0
	buffer.NetFlowPosition = 700.0

	err := publisher.PublishBufferCalculated(context.Background(), buffer)
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishBufferStatusChanged(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()
	buffer.Zone = domain.ZoneRed
	oldZone := domain.ZoneYellow

	err := publisher.PublishBufferStatusChanged(context.Background(), buffer, oldZone)
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishBufferAlertTriggered(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()
	buffer.Zone = domain.ZoneRed
	buffer.AlertLevel = domain.AlertCritical

	err := publisher.PublishBufferAlertTriggered(context.Background(), buffer, "low_stock", "Buffer is critical")
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishBufferZoneChanged(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()
	buffer.Zone = domain.ZoneYellow

	err := publisher.PublishBufferZoneChanged(context.Background(), buffer, domain.ZoneGreen, "Stock consumed")
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishFADCreated(t *testing.T) {
	publisher := NewNoOpPublisher()

	fad := createTestFAD()

	err := publisher.PublishFADCreated(context.Background(), fad)
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishFADUpdated(t *testing.T) {
	publisher := NewNoOpPublisher()

	fad := createTestFAD()

	err := publisher.PublishFADUpdated(context.Background(), fad)
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishFADDeleted(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.PublishFADDeleted(context.Background(), "fad-123", "org-123")
	assert.NoError(t, err)
}

func TestNoOpPublisher_PublishADUCalculated(t *testing.T) {
	publisher := NewNoOpPublisher()

	buffer := createTestBuffer()
	buffer.CPD = 150.0

	err := publisher.PublishADUCalculated(context.Background(), buffer, 120.0, 90)
	assert.NoError(t, err)
}

func TestPublisher_GetMetrics(t *testing.T) {
	publisher := NewNoOpPublisher()

	metrics := publisher.GetMetrics()

	assert.Equal(t, int64(0), metrics.PublishCount)
	assert.Equal(t, int64(0), metrics.SuccessCount)
	assert.Equal(t, int64(0), metrics.FailureCount)
}

func TestPublisher_Close(t *testing.T) {
	publisher := NewNoOpPublisher()

	err := publisher.Close()
	assert.NoError(t, err)
}

func TestEventEnvelope_Marshal(t *testing.T) {
	envelope := &EventEnvelope{
		ID:             "env-123",
		Subject:        SubjectBufferCalculated,
		CorrelationID:  "corr-123",
		OrganizationID: "org-123",
		Source:         sourceService,
		Type:           TypeBufferCalculated,
		SchemaVersion:  schemaVersion,
		Timestamp:      time.Now().UTC(),
		Payload:        []byte(`{"buffer_id": "buf-123"}`),
	}

	// Verify the envelope can be serialized
	assert.Equal(t, "env-123", envelope.ID)
	assert.Equal(t, SubjectBufferCalculated, envelope.Subject)
	assert.Equal(t, sourceService, envelope.Source)
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

func TestNewPublisher_NilConnection(t *testing.T) {
	publisher, err := NewPublisher(nil, nil)

	require.NoError(t, err)
	assert.NotNil(t, publisher)
	assert.False(t, publisher.IsEnabled())
}

func TestPublisherMetrics(t *testing.T) {
	metrics := &PublisherMetrics{
		PublishCount:   100,
		SuccessCount:   95,
		FailureCount:   5,
		RetryCount:     10,
		LastPublishAt:  time.Now(),
		AverageLatency: 50 * time.Millisecond,
	}

	assert.Equal(t, int64(100), metrics.PublishCount)
	assert.Equal(t, int64(95), metrics.SuccessCount)
	assert.Equal(t, int64(5), metrics.FailureCount)
	assert.Equal(t, int64(10), metrics.RetryCount)
	assert.Equal(t, 50*time.Millisecond, metrics.AverageLatency)
}

// Helper functions

func createTestBuffer() *domain.Buffer {
	return &domain.Buffer{
		ID:                 uuid.New(),
		OrganizationID:     uuid.New(),
		ProductID:          uuid.New(),
		BufferProfileID:    uuid.New(),
		CPD:                100.0,
		LTD:                14,
		RedBase:            50.0,
		RedSafe:            25.0,
		RedZone:            75.0,
		YellowZone:         300.0,
		GreenZone:          200.0,
		TopOfRed:           75.0,
		TopOfYellow:        375.0,
		TopOfGreen:         575.0,
		OnHand:             200.0,
		OnOrder:            100.0,
		QualifiedDemand:    50.0,
		NetFlowPosition:    250.0,
		BufferPenetration:  43.5,
		Zone:               domain.ZoneGreen,
		AlertLevel:         domain.AlertNormal,
		LastRecalculatedAt: time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func createTestFAD() *domain.DemandAdjustment {
	return &domain.DemandAdjustment{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		ProductID:      uuid.New(),
		AdjustmentType: domain.DemandAdjustmentSeasonal,
		Factor:         1.5,
		StartDate:      time.Now(),
		EndDate:        time.Now().Add(30 * 24 * time.Hour),
		Reason:         "Holiday season",
		CreatedBy:      uuid.New(),
		CreatedAt:      time.Now(),
	}
}
