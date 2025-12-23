package domain_test

import (
	"testing"
	"time"

	"github.com/giia/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewPODelayedAlert_WithDelayedPO_CreatesAlert(t *testing.T) {
	givenPO := &domain.PurchaseOrder{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		PONumber:       "PO-001",
		SupplierID:     uuid.New(),
		DelayDays:      5,
		ExpectedArrivalDate: time.Now().AddDate(0, 0, -5),
	}

	alert := domain.NewPODelayedAlert(givenPO)

	assert.NotNil(t, alert)
	assert.Equal(t, givenPO.OrganizationID, alert.OrganizationID)
	assert.Equal(t, domain.AlertTypePODelayed, alert.AlertType)
	assert.Equal(t, domain.AlertSeverityHigh, alert.Severity)
	assert.Equal(t, "purchase_order", alert.ResourceType)
	assert.Equal(t, givenPO.ID, alert.ResourceID)
	assert.Contains(t, alert.Message, "PO-001")
	assert.Contains(t, alert.Message, "delayed by")
	assert.NotNil(t, alert.Data)
}

func TestNewPOLateWarningAlert_WithApproachingDueDate_CreatesAlert(t *testing.T) {
	givenPO := &domain.PurchaseOrder{
		ID:                  uuid.New(),
		OrganizationID:      uuid.New(),
		PONumber:            "PO-002",
		SupplierID:          uuid.New(),
		ExpectedArrivalDate: time.Now().AddDate(0, 0, 2),
	}
	givenDaysUntilDue := 2

	alert := domain.NewPOLateWarningAlert(givenPO, givenDaysUntilDue)

	assert.NotNil(t, alert)
	assert.Equal(t, givenPO.OrganizationID, alert.OrganizationID)
	assert.Equal(t, domain.AlertTypePOLateWarning, alert.AlertType)
	assert.Equal(t, domain.AlertSeverityMedium, alert.Severity)
	assert.Equal(t, "purchase_order", alert.ResourceType)
	assert.Equal(t, givenPO.ID, alert.ResourceID)
	assert.Contains(t, alert.Message, "PO-002")
	assert.Contains(t, alert.Message, "due in")
}

func TestNewBufferStockoutAlert_WithStockout_CreatesAlert(t *testing.T) {
	givenOrgID := uuid.New()
	givenProductID := uuid.New()
	givenProductName := "Widget A"

	alert := domain.NewBufferStockoutAlert(givenOrgID, givenProductID, givenProductName)

	assert.NotNil(t, alert)
	assert.Equal(t, givenOrgID, alert.OrganizationID)
	assert.Equal(t, domain.AlertTypeBufferStockout, alert.AlertType)
	assert.Equal(t, domain.AlertSeverityCritical, alert.Severity)
	assert.Equal(t, "product", alert.ResourceType)
	assert.Equal(t, givenProductID, alert.ResourceID)
	assert.Contains(t, alert.Message, "Widget A")
	assert.Contains(t, alert.Message, "out of stock")
}

func TestAlert_Acknowledge_WithUnacknowledgedAlert_AcknowledgesAlert(t *testing.T) {
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
	}

	err := givenAlert.Acknowledge(givenUserID)

	assert.NoError(t, err)
	assert.NotNil(t, givenAlert.AcknowledgedAt)
	assert.NotNil(t, givenAlert.AcknowledgedBy)
	assert.Equal(t, givenUserID, *givenAlert.AcknowledgedBy)
}

func TestAlert_Acknowledge_WithAlreadyAcknowledgedAlert_ReturnsError(t *testing.T) {
	givenUserID := uuid.New()
	givenTime := time.Now()
	givenPreviousUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		AcknowledgedAt: &givenTime,
		AcknowledgedBy: &givenPreviousUserID,
	}

	err := givenAlert.Acknowledge(givenUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert already acknowledged")
	assert.Equal(t, givenPreviousUserID, *givenAlert.AcknowledgedBy)
}

func TestAlert_Resolve_WithUnresolvedAlert_ResolvesAlert(t *testing.T) {
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
	}

	err := givenAlert.Resolve(givenUserID)

	assert.NoError(t, err)
	assert.NotNil(t, givenAlert.ResolvedAt)
	assert.NotNil(t, givenAlert.ResolvedBy)
	assert.Equal(t, givenUserID, *givenAlert.ResolvedBy)
}

func TestAlert_Resolve_WithAlreadyResolvedAlert_ReturnsError(t *testing.T) {
	givenUserID := uuid.New()
	givenTime := time.Now()
	givenPreviousUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		ResolvedAt:     &givenTime,
		ResolvedBy:     &givenPreviousUserID,
	}

	err := givenAlert.Resolve(givenUserID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert already resolved")
	assert.Equal(t, givenPreviousUserID, *givenAlert.ResolvedBy)
}

func TestAlert_IsActive_WhenNotAcknowledgedOrResolved_ReturnsTrue(t *testing.T) {
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
	}

	result := givenAlert.IsActive()

	assert.True(t, result)
}

func TestAlert_IsActive_WhenAcknowledged_ReturnsFalse(t *testing.T) {
	givenTime := time.Now()
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		AcknowledgedAt: &givenTime,
		AcknowledgedBy: &givenUserID,
	}

	result := givenAlert.IsActive()

	assert.False(t, result)
}

func TestAlert_IsActive_WhenResolved_ReturnsFalse(t *testing.T) {
	givenTime := time.Now()
	givenUserID := uuid.New()
	givenAlert := &domain.Alert{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		AlertType:      domain.AlertTypePODelayed,
		ResolvedAt:     &givenTime,
		ResolvedBy:     &givenUserID,
	}

	result := givenAlert.IsActive()

	assert.False(t, result)
}