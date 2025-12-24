package postgres

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAlertModel_TableName(t *testing.T) {
	model := AlertModel{}
	assert.Equal(t, "alerts", model.TableName())
}

func TestAlertRepository_ToModel(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()
	alertID := uuid.New()
	orgID := uuid.New()
	resourceID := uuid.New()

	alert := &domain.Alert{
		ID:             alertID,
		OrganizationID: orgID,
		AlertType:      domain.AlertTypePODelayed,
		Severity:       domain.AlertSeverityHigh,
		ResourceType:   "purchase_order",
		ResourceID:     resourceID,
		Title:          "PO Delayed",
		Message:        "Purchase order PO-001 is delayed by 5 days",
		Data: map[string]interface{}{
			"po_number":  "PO-001",
			"delay_days": float64(5),
		},
		CreatedAt: now,
	}

	model, err := repo.toModel(alert)

	require.NoError(t, err)
	assert.Equal(t, alertID, model.ID)
	assert.Equal(t, orgID, model.OrganizationID)
	assert.Equal(t, "po_delayed", model.AlertType)
	assert.Equal(t, "high", model.Severity)
	assert.Equal(t, "purchase_order", model.ResourceType)
	assert.Equal(t, resourceID, model.ResourceID)
	assert.Equal(t, "PO Delayed", model.Title)
	assert.Contains(t, model.Data, "po_number")
}

func TestAlertRepository_ToDomain(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()
	alertID := uuid.New()
	orgID := uuid.New()
	resourceID := uuid.New()
	userID := uuid.New()
	ackTime := now.Add(-1 * time.Hour)

	model := &AlertModel{
		ID:             alertID,
		OrganizationID: orgID,
		AlertType:      "buffer_red",
		Severity:       "critical",
		ResourceType:   "product",
		ResourceID:     resourceID,
		Title:          "Buffer Status Red",
		Message:        "Product SKU-001 is in red buffer zone",
		Data:           `{"product_id":"abc-123","buffer_status":"red"}`,
		AcknowledgedAt: &ackTime,
		AcknowledgedBy: &userID,
		CreatedAt:      now,
	}

	alert, err := repo.toDomain(model)

	require.NoError(t, err)
	assert.Equal(t, alertID, alert.ID)
	assert.Equal(t, orgID, alert.OrganizationID)
	assert.Equal(t, domain.AlertTypeBufferRed, alert.AlertType)
	assert.Equal(t, domain.AlertSeverityCritical, alert.Severity)
	assert.Equal(t, "product", alert.ResourceType)
	assert.Equal(t, resourceID, alert.ResourceID)
	assert.Equal(t, "Buffer Status Red", alert.Title)
	assert.NotNil(t, alert.AcknowledgedAt)
	assert.NotNil(t, alert.AcknowledgedBy)
	assert.Equal(t, userID, *alert.AcknowledgedBy)
	assert.NotNil(t, alert.Data)
	assert.Equal(t, "abc-123", alert.Data["product_id"])
}

func TestAlertRepository_ToModelWithNilData(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()

	alert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		Severity:       domain.AlertSeverityMedium,
		ResourceType:   "purchase_order",
		ResourceID:     uuid.New(),
		Title:          "PO Alert",
		Message:        "Alert message",
		Data:           nil,
		CreatedAt:      now,
	}

	model, err := repo.toModel(alert)

	require.NoError(t, err)
	assert.Equal(t, "", model.Data)
}

func TestAlertRepository_ToDomainWithEmptyData(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()

	model := &AlertModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      "po_delayed",
		Severity:       "medium",
		ResourceType:   "purchase_order",
		ResourceID:     uuid.New(),
		Title:          "PO Alert",
		Message:        "Alert message",
		Data:           "",
		CreatedAt:      now,
	}

	alert, err := repo.toDomain(model)

	require.NoError(t, err)
	assert.Nil(t, alert.Data)
}

func TestAlertRepository_ToDomainWithInvalidJSON(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()

	model := &AlertModel{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      "po_delayed",
		Severity:       "medium",
		ResourceType:   "purchase_order",
		ResourceID:     uuid.New(),
		Title:          "PO Alert",
		Message:        "Alert message",
		Data:           "invalid json {",
		CreatedAt:      now,
	}

	_, err := repo.toDomain(model)

	assert.Error(t, err) // Should fail on invalid JSON
}

func TestAlertRepository_AllAlertTypes(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()

	testCases := []struct {
		domainType domain.AlertType
		modelType  string
	}{
		{domain.AlertTypePODelayed, "po_delayed"},
		{domain.AlertTypePOLateWarning, "po_late_warning"},
		{domain.AlertTypeBufferRed, "buffer_red"},
		{domain.AlertTypeBufferBelowRed, "buffer_below_red"},
		{domain.AlertTypeBufferStockout, "buffer_stockout"},
		{domain.AlertTypeStockDeviation, "stock_deviation"},
		{domain.AlertTypeObsolescenceRisk, "obsolescence_risk"},
		{domain.AlertTypeExcessInventory, "excess_inventory"},
		{domain.AlertTypeSupplierDelayPattern, "supplier_delay_pattern"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.domainType), func(t *testing.T) {
			alert := &domain.Alert{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				AlertType:      tc.domainType,
				Severity:       domain.AlertSeverityMedium,
				ResourceType:   "test",
				ResourceID:     uuid.New(),
				Title:          "Test",
				Message:        "Test",
				CreatedAt:      now,
			}

			model, err := repo.toModel(alert)
			require.NoError(t, err)
			assert.Equal(t, tc.modelType, model.AlertType)
		})
	}
}

func TestAlertRepository_AllSeverityLevels(t *testing.T) {
	repo := &alertRepository{}
	now := time.Now()

	testCases := []struct {
		domainSeverity domain.AlertSeverity
		modelSeverity  string
	}{
		{domain.AlertSeverityInfo, "info"},
		{domain.AlertSeverityLow, "low"},
		{domain.AlertSeverityMedium, "medium"},
		{domain.AlertSeverityHigh, "high"},
		{domain.AlertSeverityCritical, "critical"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.domainSeverity), func(t *testing.T) {
			alert := &domain.Alert{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				AlertType:      domain.AlertTypePODelayed,
				Severity:       tc.domainSeverity,
				ResourceType:   "test",
				ResourceID:     uuid.New(),
				Title:          "Test",
				Message:        "Test",
				CreatedAt:      now,
			}

			model, err := repo.toModel(alert)
			require.NoError(t, err)
			assert.Equal(t, tc.modelSeverity, model.Severity)
		})
	}
}

func BenchmarkAlertToModel(b *testing.B) {
	repo := &alertRepository{}
	now := time.Now()

	alert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		Severity:       domain.AlertSeverityHigh,
		ResourceType:   "purchase_order",
		ResourceID:     uuid.New(),
		Title:          "Test Alert",
		Message:        "Test message",
		Data:           map[string]interface{}{"key": "value"},
		CreatedAt:      now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.toModel(alert)
	}
}
