package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Alert struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	AlertType      AlertType
	Severity       AlertSeverity
	ResourceType   string
	ResourceID     uuid.UUID
	Title          string
	Message        string
	Data           map[string]interface{}
	AcknowledgedAt *time.Time
	AcknowledgedBy *uuid.UUID
	ResolvedAt     *time.Time
	ResolvedBy     *uuid.UUID
	CreatedAt      time.Time
}

type AlertType string

const (
	AlertTypePODelayed            AlertType = "po_delayed"
	AlertTypePOLateWarning        AlertType = "po_late_warning"
	AlertTypeBufferRed            AlertType = "buffer_red"
	AlertTypeBufferBelowRed       AlertType = "buffer_below_red"
	AlertTypeBufferStockout       AlertType = "buffer_stockout"
	AlertTypeStockDeviation       AlertType = "stock_deviation"
	AlertTypeObsolescenceRisk     AlertType = "obsolescence_risk"
	AlertTypeExcessInventory      AlertType = "excess_inventory"
	AlertTypeSupplierDelayPattern AlertType = "supplier_delay_pattern"
)

type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

func NewPODelayedAlert(po *PurchaseOrder) *Alert {
	return &Alert{
		ID:             uuid.New(),
		OrganizationID: po.OrganizationID,
		AlertType:      AlertTypePODelayed,
		Severity:       AlertSeverityHigh,
		ResourceType:   "purchase_order",
		ResourceID:     po.ID,
		Title:          "Purchase Order Delayed",
		Message:        fmt.Sprintf("PO %s is delayed by %d days", po.PONumber, po.DelayDays),
		Data: map[string]interface{}{
			"po_number":             po.PONumber,
			"supplier_id":           po.SupplierID.String(),
			"expected_arrival_date": po.ExpectedArrivalDate,
			"delay_days":            po.DelayDays,
		},
		CreatedAt: time.Now(),
	}
}

func NewPOLateWarningAlert(po *PurchaseOrder, daysUntilDue int) *Alert {
	return &Alert{
		ID:             uuid.New(),
		OrganizationID: po.OrganizationID,
		AlertType:      AlertTypePOLateWarning,
		Severity:       AlertSeverityMedium,
		ResourceType:   "purchase_order",
		ResourceID:     po.ID,
		Title:          "Purchase Order Approaching Due Date",
		Message:        fmt.Sprintf("PO %s is due in %d days", po.PONumber, daysUntilDue),
		Data: map[string]interface{}{
			"po_number":             po.PONumber,
			"supplier_id":           po.SupplierID.String(),
			"expected_arrival_date": po.ExpectedArrivalDate,
			"days_until_due":        daysUntilDue,
		},
		CreatedAt: time.Now(),
	}
}

func NewBufferStockoutAlert(orgID, productID uuid.UUID, productName string) *Alert {
	return &Alert{
		ID:             uuid.New(),
		OrganizationID: orgID,
		AlertType:      AlertTypeBufferStockout,
		Severity:       AlertSeverityCritical,
		ResourceType:   "product",
		ResourceID:     productID,
		Title:          "Product Stockout",
		Message:        fmt.Sprintf("Product %s is out of stock", productName),
		Data: map[string]interface{}{
			"product_id":   productID.String(),
			"product_name": productName,
		},
		CreatedAt: time.Now(),
	}
}

func (a *Alert) Acknowledge(userID uuid.UUID) error {
	if a.AcknowledgedAt != nil {
		return NewValidationError("alert already acknowledged")
	}
	now := time.Now()
	a.AcknowledgedAt = &now
	a.AcknowledgedBy = &userID
	return nil
}

func (a *Alert) Resolve(userID uuid.UUID) error {
	if a.ResolvedAt != nil {
		return NewValidationError("alert already resolved")
	}
	now := time.Now()
	a.ResolvedAt = &now
	a.ResolvedBy = &userID
	return nil
}

func (a *Alert) IsActive() bool {
	return a.AcknowledgedAt == nil && a.ResolvedAt == nil
}