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

// MockLogger implements the Logger interface for testing.
type MockLogger struct {
	debugLogs []logEntry
	infoLogs  []logEntry
	warnLogs  []logEntry
	errorLogs []logEntry
}

type logEntry struct {
	msg    string
	fields map[string]interface{}
}

func (m *MockLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	m.debugLogs = append(m.debugLogs, logEntry{msg: msg, fields: fields})
}

func (m *MockLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	m.infoLogs = append(m.infoLogs, logEntry{msg: msg, fields: fields})
}

func (m *MockLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	m.warnLogs = append(m.warnLogs, logEntry{msg: msg, fields: fields})
}

func (m *MockLogger) Error(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	m.errorLogs = append(m.errorLogs, logEntry{msg: msg, fields: fields})
}

// MockProductCache implements LocalProductCache for testing.
type MockProductCache struct {
	data map[string]*ProductEvent
}

func NewMockProductCache() *MockProductCache {
	return &MockProductCache{data: make(map[string]*ProductEvent)}
}

func (m *MockProductCache) Set(ctx context.Context, key string, value *ProductEvent, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockProductCache) Get(ctx context.Context, key string) (*ProductEvent, error) {
	return m.data[key], nil
}

func (m *MockProductCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// MockLocationCache implements LocalLocationCache for testing.
type MockLocationCache struct {
	data map[string]*LocationEvent
}

func NewMockLocationCache() *MockLocationCache {
	return &MockLocationCache{data: make(map[string]*LocationEvent)}
}

func (m *MockLocationCache) Set(ctx context.Context, key string, value *LocationEvent, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockLocationCache) Get(ctx context.Context, key string) (*LocationEvent, error) {
	return m.data[key], nil
}

func (m *MockLocationCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

// MockSupplierCache implements LocalSupplierCache for testing.
type MockSupplierCache struct {
	data map[string]*SupplierEvent
}

func NewMockSupplierCache() *MockSupplierCache {
	return &MockSupplierCache{data: make(map[string]*SupplierEvent)}
}

func (m *MockSupplierCache) Set(ctx context.Context, key string, value *SupplierEvent, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *MockSupplierCache) Get(ctx context.Context, key string) (*SupplierEvent, error) {
	return m.data[key], nil
}

func (m *MockSupplierCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestNewCatalogHandler(t *testing.T) {
	logger := &MockLogger{}
	productCache := NewMockProductCache()
	locationCache := NewMockLocationCache()
	supplierCache := NewMockSupplierCache()

	handler := NewCatalogHandler(productCache, locationCache, supplierCache, logger, nil)

	assert.NotNil(t, handler)
}

func TestDefaultCatalogHandlerConfig(t *testing.T) {
	config := DefaultCatalogHandlerConfig()

	assert.Equal(t, 24*time.Hour, config.CacheTTL)
}

func TestCatalogHandler_HandleProductCreated(t *testing.T) {
	logger := &MockLogger{}
	productCache := NewMockProductCache()
	locationCache := NewMockLocationCache()
	supplierCache := NewMockSupplierCache()

	handler := NewCatalogHandler(productCache, locationCache, supplierCache, logger, nil)

	productEvent := ProductEvent{
		ID:             "prod-123",
		OrganizationID: "org-123",
		SKU:            "SKU-001",
		Name:           "Test Product",
		UnitPrice:      99.99,
		UOM:            "EA",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	payload, err := json.Marshal(productEvent)
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-123",
		Subject:        "catalog.product.created",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventProductCreated,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	// Verify product was cached
	cachedProduct, err := productCache.Get(context.Background(), "product:org-123:prod-123")
	require.NoError(t, err)
	assert.NotNil(t, cachedProduct)
	assert.Equal(t, "SKU-001", cachedProduct.SKU)

	// Verify logging
	assert.Len(t, logger.infoLogs, 1)
}

func TestCatalogHandler_HandleProductUpdated(t *testing.T) {
	logger := &MockLogger{}
	productCache := NewMockProductCache()

	handler := NewCatalogHandler(productCache, nil, nil, logger, nil)

	productEvent := ProductEvent{
		ID:             "prod-123",
		OrganizationID: "org-123",
		SKU:            "SKU-001",
		Name:           "Updated Product",
		UnitPrice:      149.99,
		UOM:            "EA",
		IsActive:       true,
		UpdatedAt:      time.Now(),
	}

	payload, err := json.Marshal(productEvent)
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-124",
		Subject:        "catalog.product.updated",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventProductUpdated,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	cachedProduct, _ := productCache.Get(context.Background(), "product:org-123:prod-123")
	assert.Equal(t, "Updated Product", cachedProduct.Name)
	assert.Equal(t, 149.99, cachedProduct.UnitPrice)
}

func TestCatalogHandler_HandleProductDeleted(t *testing.T) {
	logger := &MockLogger{}
	productCache := NewMockProductCache()

	// Pre-populate cache
	productCache.data["product:org-123:prod-123"] = &ProductEvent{
		ID:  "prod-123",
		SKU: "SKU-001",
	}

	handler := NewCatalogHandler(productCache, nil, nil, logger, nil)

	deletePayload := struct {
		ID string `json:"id"`
	}{ID: "prod-123"}

	payload, err := json.Marshal(deletePayload)
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-125",
		Subject:        "catalog.product.deleted",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventProductDeleted,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	cachedProduct, _ := productCache.Get(context.Background(), "product:org-123:prod-123")
	assert.Nil(t, cachedProduct)
}

func TestCatalogHandler_HandleLocationEvent(t *testing.T) {
	logger := &MockLogger{}
	locationCache := NewMockLocationCache()

	handler := NewCatalogHandler(nil, locationCache, nil, logger, nil)

	locationEvent := LocationEvent{
		ID:             "loc-123",
		OrganizationID: "org-123",
		Code:           "WH-01",
		Name:           "Main Warehouse",
		Type:           "warehouse",
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	payload, err := json.Marshal(locationEvent)
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-126",
		Subject:        "catalog.location.created",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventLocationCreated,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	cachedLocation, _ := locationCache.Get(context.Background(), "location:org-123:loc-123")
	assert.NotNil(t, cachedLocation)
	assert.Equal(t, "WH-01", cachedLocation.Code)
}

func TestCatalogHandler_HandleSupplierEvent(t *testing.T) {
	logger := &MockLogger{}
	supplierCache := NewMockSupplierCache()

	handler := NewCatalogHandler(nil, nil, supplierCache, logger, nil)

	supplierEvent := SupplierEvent{
		ID:             "sup-123",
		OrganizationID: "org-123",
		Code:           "SUP-01",
		Name:           "Acme Supplies",
		LeadTimeDays:   7,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	payload, err := json.Marshal(supplierEvent)
	require.NoError(t, err)

	envelope := &events.EventEnvelope{
		ID:             "env-127",
		Subject:        "catalog.supplier.created",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventSupplierCreated,
		Payload:        payload,
		Timestamp:      time.Now(),
	}

	err = handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	cachedSupplier, _ := supplierCache.Get(context.Background(), "supplier:org-123:sup-123")
	assert.NotNil(t, cachedSupplier)
	assert.Equal(t, "Acme Supplies", cachedSupplier.Name)
}

func TestCatalogHandler_HandleUnknownEventType(t *testing.T) {
	logger := &MockLogger{}
	handler := NewCatalogHandler(nil, nil, nil, logger, nil)

	envelope := &events.EventEnvelope{
		ID:             "env-128",
		Subject:        "catalog.unknown.type",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           "catalog.unknown.type",
		Payload:        []byte("{}"),
		Timestamp:      time.Now(),
	}

	err := handler.Handle(context.Background(), envelope)
	assert.NoError(t, err)

	// Should log debug message about ignoring event
	assert.Len(t, logger.debugLogs, 2) // Initial debug + ignore message
}

func TestCatalogHandler_HandleInvalidPayload(t *testing.T) {
	logger := &MockLogger{}
	productCache := NewMockProductCache()

	handler := NewCatalogHandler(productCache, nil, nil, logger, nil)

	envelope := &events.EventEnvelope{
		ID:             "env-129",
		Subject:        "catalog.product.created",
		OrganizationID: "org-123",
		Source:         "catalog-service",
		Type:           EventProductCreated,
		Payload:        []byte("invalid json"),
		Timestamp:      time.Now(),
	}

	err := handler.Handle(context.Background(), envelope)
	assert.Error(t, err)
}

func TestCatalogHandler_GetSubscriptionSubjects(t *testing.T) {
	handler := NewCatalogHandler(nil, nil, nil, &MockLogger{}, nil)

	subjects := handler.GetSubscriptionSubjects()

	assert.Len(t, subjects, 3)
	assert.Contains(t, subjects, "catalog.product.>")
	assert.Contains(t, subjects, "catalog.location.>")
	assert.Contains(t, subjects, "catalog.supplier.>")
}
