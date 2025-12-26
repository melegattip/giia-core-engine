// Package handlers provides event handlers for the Execution Service.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/events"
)

// DDMRP event type constants.
const (
	EventBufferCalculated     = "ddmrp.buffer.calculated"
	EventBufferStatusChanged  = "ddmrp.buffer.status_changed"
	EventBufferAlertTriggered = "ddmrp.buffer.alert_triggered"
	EventBufferZoneChanged    = "ddmrp.buffer.zone_changed"
)

// BufferCalculatedEvent represents a buffer calculation event from DDMRP.
type BufferCalculatedEvent struct {
	BufferID       string    `json:"buffer_id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	CPD            float64   `json:"cpd"`
	RedZone        float64   `json:"red_zone"`
	YellowZone     float64   `json:"yellow_zone"`
	GreenZone      float64   `json:"green_zone"`
	TOG            float64   `json:"tog"`
	CalculatedAt   time.Time `json:"calculated_at"`
}

// BufferStatusChangedEvent represents a buffer status change event.
type BufferStatusChangedEvent struct {
	BufferID       string    `json:"buffer_id"`
	OrganizationID string    `json:"organization_id"`
	ProductID      string    `json:"product_id"`
	LocationID     string    `json:"location_id"`
	OldZone        string    `json:"old_zone"`
	NewZone        string    `json:"new_zone"`
	AlertLevel     string    `json:"alert_level"`
	NFP            float64   `json:"nfp"`
	ChangedAt      time.Time `json:"changed_at"`
}

// BufferAlertTriggeredEvent represents a buffer alert event from DDMRP.
type BufferAlertTriggeredEvent struct {
	BufferID         string            `json:"buffer_id"`
	OrganizationID   string            `json:"organization_id"`
	ProductID        string            `json:"product_id"`
	LocationID       string            `json:"location_id"`
	AlertType        string            `json:"alert_type"`
	AlertLevel       string            `json:"alert_level"`
	Zone             string            `json:"zone"`
	NFP              float64           `json:"nfp"`
	TOG              float64           `json:"tog"`
	ReplenishmentQty float64           `json:"replenishment_qty,omitempty"`
	Message          string            `json:"message"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	TriggeredAt      time.Time         `json:"triggered_at"`
}

// ReplenishmentService handles automatic replenishment based on DDMRP alerts.
type ReplenishmentService interface {
	CreateReplenishmentOrder(ctx context.Context, req *ReplenishmentRequest) (*ReplenishmentResponse, error)
	EvaluateReplenishment(ctx context.Context, productID, locationID string) (*ReplenishmentRecommendation, error)
}

// ReplenishmentRequest represents a request to create a replenishment order.
type ReplenishmentRequest struct {
	OrganizationID string
	ProductID      string
	LocationID     string
	Quantity       float64
	Priority       string
	SourceBufferID string
	TriggerEvent   string
}

// ReplenishmentResponse represents the result of creating a replenishment order.
type ReplenishmentResponse struct {
	OrderID     string
	OrderType   string
	OrderNumber string
	Status      string
	CreatedAt   time.Time
}

// ReplenishmentRecommendation represents a replenishment suggestion.
type ReplenishmentRecommendation struct {
	ShouldReplenish bool
	Quantity        float64
	Priority        string
	Reason          string
	SupplierID      string
}

// DDMRPHandler handles events from the DDMRP Engine Service.
type DDMRPHandler struct {
	replenishmentSvc  ReplenishmentService
	logger            Logger
	autoReplenishment bool
}

// DDMRPHandlerConfig contains configuration for the handler.
type DDMRPHandlerConfig struct {
	// AutoReplenishment enables automatic PO creation for critical alerts
	AutoReplenishment bool
}

// DefaultDDMRPHandlerConfig returns default configuration.
func DefaultDDMRPHandlerConfig() *DDMRPHandlerConfig {
	return &DDMRPHandlerConfig{
		AutoReplenishment: false, // Disabled by default for safety
	}
}

// NewDDMRPHandler creates a new DDMRP event handler.
func NewDDMRPHandler(
	replenishmentSvc ReplenishmentService,
	logger Logger,
	config *DDMRPHandlerConfig,
) *DDMRPHandler {
	if config == nil {
		config = DefaultDDMRPHandlerConfig()
	}

	return &DDMRPHandler{
		replenishmentSvc:  replenishmentSvc,
		logger:            logger,
		autoReplenishment: config.AutoReplenishment,
	}
}

// Handle processes a DDMRP event.
func (h *DDMRPHandler) Handle(ctx context.Context, envelope *events.EventEnvelope) error {
	h.logger.Debug(ctx, "Processing DDMRP event", map[string]interface{}{
		"event_id":   envelope.ID,
		"event_type": envelope.Type,
		"org_id":     envelope.OrganizationID,
	})

	switch envelope.Type {
	case EventBufferCalculated:
		return h.handleBufferCalculated(ctx, envelope)
	case EventBufferStatusChanged:
		return h.handleBufferStatusChanged(ctx, envelope)
	case EventBufferAlertTriggered:
		return h.handleBufferAlertTriggered(ctx, envelope)
	case EventBufferZoneChanged:
		return h.handleBufferZoneChanged(ctx, envelope)
	default:
		h.logger.Debug(ctx, "Ignoring unhandled DDMRP event type", map[string]interface{}{
			"event_type": envelope.Type,
		})
		return nil
	}
}

// handleBufferCalculated handles buffer calculation events.
func (h *DDMRPHandler) handleBufferCalculated(ctx context.Context, envelope *events.EventEnvelope) error {
	var event BufferCalculatedEvent
	if err := json.Unmarshal(envelope.Payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal buffer calculated event: %w", err)
	}

	h.logger.Info(ctx, "Buffer calculated event processed", map[string]interface{}{
		"buffer_id":  event.BufferID,
		"product_id": event.ProductID,
		"cpd":        event.CPD,
		"tog":        event.TOG,
	})

	// Could trigger inventory analysis or update local metrics
	return nil
}

// handleBufferStatusChanged handles buffer status change events.
func (h *DDMRPHandler) handleBufferStatusChanged(ctx context.Context, envelope *events.EventEnvelope) error {
	var event BufferStatusChangedEvent
	if err := json.Unmarshal(envelope.Payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal buffer status changed event: %w", err)
	}

	h.logger.Info(ctx, "Buffer status changed", map[string]interface{}{
		"buffer_id":   event.BufferID,
		"product_id":  event.ProductID,
		"old_zone":    event.OldZone,
		"new_zone":    event.NewZone,
		"alert_level": event.AlertLevel,
		"nfp":         event.NFP,
	})

	// If entering red zone, might want to flag related orders
	if event.NewZone == "red" && event.OldZone != "red" {
		h.logger.Warn(ctx, "Product entered red zone, may need expedited orders", map[string]interface{}{
			"product_id":  event.ProductID,
			"location_id": event.LocationID,
			"nfp":         event.NFP,
		})
	}

	return nil
}

// handleBufferAlertTriggered handles buffer alert events.
func (h *DDMRPHandler) handleBufferAlertTriggered(ctx context.Context, envelope *events.EventEnvelope) error {
	var event BufferAlertTriggeredEvent
	if err := json.Unmarshal(envelope.Payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal buffer alert triggered event: %w", err)
	}

	h.logger.Info(ctx, "Buffer alert triggered", map[string]interface{}{
		"buffer_id":     event.BufferID,
		"product_id":    event.ProductID,
		"alert_type":    event.AlertType,
		"alert_level":   event.AlertLevel,
		"zone":          event.Zone,
		"nfp":           event.NFP,
		"replenish_qty": event.ReplenishmentQty,
	})

	// Auto-replenishment if enabled and alert is critical
	if h.autoReplenishment && h.replenishmentSvc != nil && event.AlertLevel == "critical" {
		return h.triggerAutoReplenishment(ctx, &event, envelope)
	}

	return nil
}

// triggerAutoReplenishment creates an automatic replenishment order.
func (h *DDMRPHandler) triggerAutoReplenishment(ctx context.Context, event *BufferAlertTriggeredEvent, envelope *events.EventEnvelope) error {
	// First evaluate if replenishment is recommended
	recommendation, err := h.replenishmentSvc.EvaluateReplenishment(ctx, event.ProductID, event.LocationID)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to evaluate replenishment", map[string]interface{}{
			"product_id":  event.ProductID,
			"location_id": event.LocationID,
		})
		return nil // Don't fail the event processing
	}

	if !recommendation.ShouldReplenish {
		h.logger.Debug(ctx, "Replenishment not recommended", map[string]interface{}{
			"product_id": event.ProductID,
			"reason":     recommendation.Reason,
		})
		return nil
	}

	// Create the replenishment order
	req := &ReplenishmentRequest{
		OrganizationID: event.OrganizationID,
		ProductID:      event.ProductID,
		LocationID:     event.LocationID,
		Quantity:       recommendation.Quantity,
		Priority:       recommendation.Priority,
		SourceBufferID: event.BufferID,
		TriggerEvent:   envelope.ID,
	}

	resp, err := h.replenishmentSvc.CreateReplenishmentOrder(ctx, req)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to create replenishment order", map[string]interface{}{
			"product_id":  event.ProductID,
			"location_id": event.LocationID,
		})
		return nil // Don't fail the event processing
	}

	h.logger.Info(ctx, "Auto-replenishment order created", map[string]interface{}{
		"order_id":     resp.OrderID,
		"order_number": resp.OrderNumber,
		"product_id":   event.ProductID,
		"quantity":     recommendation.Quantity,
		"trigger":      "ddmrp_buffer_alert",
	})

	return nil
}

// handleBufferZoneChanged handles buffer zone change events.
func (h *DDMRPHandler) handleBufferZoneChanged(ctx context.Context, envelope *events.EventEnvelope) error {
	// Similar to status changed, but specifically for zone transitions
	var event BufferStatusChangedEvent
	if err := json.Unmarshal(envelope.Payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal buffer zone changed event: %w", err)
	}

	h.logger.Info(ctx, "Buffer zone changed", map[string]interface{}{
		"buffer_id":  event.BufferID,
		"product_id": event.ProductID,
		"old_zone":   event.OldZone,
		"new_zone":   event.NewZone,
	})

	return nil
}

// GetSubscriptionSubjects returns the subjects this handler subscribes to.
func (h *DDMRPHandler) GetSubscriptionSubjects() []string {
	return []string{
		"ddmrp.buffer.>",
	}
}
