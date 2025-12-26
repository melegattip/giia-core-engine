// Package handlers provides event handlers for the Execution Service.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/events"
)

// CatalogEventType constants for catalog events.
const (
	EventProductCreated  = "catalog.product.created"
	EventProductUpdated  = "catalog.product.updated"
	EventProductDeleted  = "catalog.product.deleted"
	EventLocationCreated = "catalog.location.created"
	EventLocationUpdated = "catalog.location.updated"
	EventSupplierCreated = "catalog.supplier.created"
	EventSupplierUpdated = "catalog.supplier.updated"
)

// ProductEvent represents a product event from the Catalog Service.
type ProductEvent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	SKU            string    `json:"sku"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Category       string    `json:"category,omitempty"`
	UnitPrice      float64   `json:"unit_price"`
	UOM            string    `json:"uom"`
	IsActive       bool      `json:"is_active"`
	LeadTimeDays   int       `json:"lead_time_days,omitempty"`
	MinOrderQty    float64   `json:"min_order_qty,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LocationEvent represents a location event from the Catalog Service.
type LocationEvent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Address        string    `json:"address,omitempty"`
	IsActive       bool      `json:"is_active"`
	ParentID       string    `json:"parent_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SupplierEvent represents a supplier event from the Catalog Service.
type SupplierEvent struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	ContactEmail   string    `json:"contact_email,omitempty"`
	ContactPhone   string    `json:"contact_phone,omitempty"`
	LeadTimeDays   int       `json:"lead_time_days"`
	MinOrderAmount float64   `json:"min_order_amount,omitempty"`
	Currency       string    `json:"currency,omitempty"`
	IsActive       bool      `json:"is_active"`
	Rating         float64   `json:"rating,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// LocalProductCache provides a local cache interface for products.
type LocalProductCache interface {
	Set(ctx context.Context, key string, value *ProductEvent, ttl time.Duration) error
	Get(ctx context.Context, key string) (*ProductEvent, error)
	Delete(ctx context.Context, key string) error
}

// LocalLocationCache provides a local cache interface for locations.
type LocalLocationCache interface {
	Set(ctx context.Context, key string, value *LocationEvent, ttl time.Duration) error
	Get(ctx context.Context, key string) (*LocationEvent, error)
	Delete(ctx context.Context, key string) error
}

// LocalSupplierCache provides a local cache interface for suppliers.
type LocalSupplierCache interface {
	Set(ctx context.Context, key string, value *SupplierEvent, ttl time.Duration) error
	Get(ctx context.Context, key string) (*SupplierEvent, error)
	Delete(ctx context.Context, key string) error
}

// Logger interface for logging.
type Logger interface {
	Debug(ctx context.Context, msg string, fields map[string]interface{})
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, err error, msg string, fields map[string]interface{})
}

// CatalogHandler handles events from the Catalog Service.
type CatalogHandler struct {
	productCache  LocalProductCache
	locationCache LocalLocationCache
	supplierCache LocalSupplierCache
	logger        Logger
	cacheTTL      time.Duration
}

// CatalogHandlerConfig contains configuration for the handler.
type CatalogHandlerConfig struct {
	CacheTTL time.Duration
}

// DefaultCatalogHandlerConfig returns default configuration.
func DefaultCatalogHandlerConfig() *CatalogHandlerConfig {
	return &CatalogHandlerConfig{
		CacheTTL: 24 * time.Hour,
	}
}

// NewCatalogHandler creates a new catalog event handler.
func NewCatalogHandler(
	productCache LocalProductCache,
	locationCache LocalLocationCache,
	supplierCache LocalSupplierCache,
	logger Logger,
	config *CatalogHandlerConfig,
) *CatalogHandler {
	if config == nil {
		config = DefaultCatalogHandlerConfig()
	}

	return &CatalogHandler{
		productCache:  productCache,
		locationCache: locationCache,
		supplierCache: supplierCache,
		logger:        logger,
		cacheTTL:      config.CacheTTL,
	}
}

// Handle processes a catalog event.
func (h *CatalogHandler) Handle(ctx context.Context, envelope *events.EventEnvelope) error {
	h.logger.Debug(ctx, "Processing catalog event", map[string]interface{}{
		"event_id":   envelope.ID,
		"event_type": envelope.Type,
		"org_id":     envelope.OrganizationID,
	})

	switch envelope.Type {
	case EventProductCreated, EventProductUpdated:
		return h.handleProductEvent(ctx, envelope)
	case EventProductDeleted:
		return h.handleProductDeleted(ctx, envelope)
	case EventLocationCreated, EventLocationUpdated:
		return h.handleLocationEvent(ctx, envelope)
	case EventSupplierCreated, EventSupplierUpdated:
		return h.handleSupplierEvent(ctx, envelope)
	default:
		h.logger.Debug(ctx, "Ignoring unhandled catalog event type", map[string]interface{}{
			"event_type": envelope.Type,
		})
		return nil
	}
}

// handleProductEvent handles product created/updated events.
func (h *CatalogHandler) handleProductEvent(ctx context.Context, envelope *events.EventEnvelope) error {
	var product ProductEvent
	if err := json.Unmarshal(envelope.Payload, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product event: %w", err)
	}

	if h.productCache != nil {
		cacheKey := fmt.Sprintf("product:%s:%s", envelope.OrganizationID, product.ID)
		if err := h.productCache.Set(ctx, cacheKey, &product, h.cacheTTL); err != nil {
			h.logger.Warn(ctx, "Failed to cache product", map[string]interface{}{
				"product_id": product.ID,
				"error":      err.Error(),
			})
		}
	}

	h.logger.Info(ctx, "Product event processed", map[string]interface{}{
		"product_id":   product.ID,
		"product_sku":  product.SKU,
		"product_name": product.Name,
		"event_type":   envelope.Type,
	})

	return nil
}

// handleProductDeleted handles product deleted events.
func (h *CatalogHandler) handleProductDeleted(ctx context.Context, envelope *events.EventEnvelope) error {
	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal product deleted event: %w", err)
	}

	if h.productCache != nil {
		cacheKey := fmt.Sprintf("product:%s:%s", envelope.OrganizationID, payload.ID)
		if err := h.productCache.Delete(ctx, cacheKey); err != nil {
			h.logger.Warn(ctx, "Failed to remove product from cache", map[string]interface{}{
				"product_id": payload.ID,
				"error":      err.Error(),
			})
		}
	}

	h.logger.Info(ctx, "Product deleted event processed", map[string]interface{}{
		"product_id": payload.ID,
	})

	return nil
}

// handleLocationEvent handles location created/updated events.
func (h *CatalogHandler) handleLocationEvent(ctx context.Context, envelope *events.EventEnvelope) error {
	var location LocationEvent
	if err := json.Unmarshal(envelope.Payload, &location); err != nil {
		return fmt.Errorf("failed to unmarshal location event: %w", err)
	}

	if h.locationCache != nil {
		cacheKey := fmt.Sprintf("location:%s:%s", envelope.OrganizationID, location.ID)
		if err := h.locationCache.Set(ctx, cacheKey, &location, h.cacheTTL); err != nil {
			h.logger.Warn(ctx, "Failed to cache location", map[string]interface{}{
				"location_id": location.ID,
				"error":       err.Error(),
			})
		}
	}

	h.logger.Info(ctx, "Location event processed", map[string]interface{}{
		"location_id":   location.ID,
		"location_code": location.Code,
		"location_name": location.Name,
		"event_type":    envelope.Type,
	})

	return nil
}

// handleSupplierEvent handles supplier created/updated events.
func (h *CatalogHandler) handleSupplierEvent(ctx context.Context, envelope *events.EventEnvelope) error {
	var supplier SupplierEvent
	if err := json.Unmarshal(envelope.Payload, &supplier); err != nil {
		return fmt.Errorf("failed to unmarshal supplier event: %w", err)
	}

	if h.supplierCache != nil {
		cacheKey := fmt.Sprintf("supplier:%s:%s", envelope.OrganizationID, supplier.ID)
		if err := h.supplierCache.Set(ctx, cacheKey, &supplier, h.cacheTTL); err != nil {
			h.logger.Warn(ctx, "Failed to cache supplier", map[string]interface{}{
				"supplier_id": supplier.ID,
				"error":       err.Error(),
			})
		}
	}

	h.logger.Info(ctx, "Supplier event processed", map[string]interface{}{
		"supplier_id":   supplier.ID,
		"supplier_code": supplier.Code,
		"supplier_name": supplier.Name,
		"event_type":    envelope.Type,
	})

	return nil
}

// GetSubscriptionSubjects returns the subjects this handler subscribes to.
func (h *CatalogHandler) GetSubscriptionSubjects() []string {
	return []string{
		"catalog.product.>",
		"catalog.location.>",
		"catalog.supplier.>",
	}
}
