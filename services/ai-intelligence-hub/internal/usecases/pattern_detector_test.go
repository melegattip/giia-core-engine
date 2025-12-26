package usecases

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEventStore implements EventStore for testing
type MockEventStore struct {
	mock.Mock
}

func (m *MockEventStore) GetEventsByType(ctx context.Context, orgID uuid.UUID, eventType string, since time.Time) ([]*events.Event, error) {
	args := m.Called(ctx, orgID, eventType, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*events.Event), args.Error(1)
}

func (m *MockEventStore) GetEventsByEntity(ctx context.Context, orgID uuid.UUID, entityType, entityID string, since time.Time) ([]*events.Event, error) {
	args := m.Called(ctx, orgID, entityType, entityID, since)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*events.Event), args.Error(1)
}

func (m *MockEventStore) GetRecentEvents(ctx context.Context, orgID uuid.UUID, limit int) ([]*events.Event, error) {
	args := m.Called(ctx, orgID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*events.Event), args.Error(1)
}

// MockPatternStore implements PatternStore for testing
type MockPatternStore struct {
	mock.Mock
	savedPatterns []*entities.Pattern
}

func (m *MockPatternStore) Save(ctx context.Context, pattern *entities.Pattern) error {
	m.savedPatterns = append(m.savedPatterns, pattern)
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockPatternStore) FindByOrganization(ctx context.Context, orgID uuid.UUID, status entities.PatternStatus) ([]*entities.Pattern, error) {
	args := m.Called(ctx, orgID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Pattern), args.Error(1)
}

func (m *MockPatternStore) FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entities.Pattern, error) {
	args := m.Called(ctx, entityType, entityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Pattern), args.Error(1)
}

func (m *MockPatternStore) Update(ctx context.Context, pattern *entities.Pattern) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// MockAlertService implements AlertService for testing
type MockAlertService struct {
	mock.Mock
	sentAlerts []*entities.Pattern
}

func (m *MockAlertService) SendPatternAlert(ctx context.Context, pattern *entities.Pattern, notification *domain.AINotification) error {
	m.sentAlerts = append(m.sentAlerts, pattern)
	args := m.Called(ctx, pattern, notification)
	return args.Error(0)
}

// MockAIAnalyzer implements AIAnalyzer for testing
type MockAIAnalyzer struct {
	mock.Mock
}

func (m *MockAIAnalyzer) Analyze(ctx context.Context, request *providers.AIAnalysisRequest) (*providers.AIAnalysisResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.AIAnalysisResponse), args.Error(1)
}

// MockLogger implements logger.Logger for testing
type MockLogger struct{}

func (m *MockLogger) Debug(ctx context.Context, msg string, tags logger.Tags)            {}
func (m *MockLogger) Info(ctx context.Context, msg string, tags logger.Tags)             {}
func (m *MockLogger) Warn(ctx context.Context, msg string, tags logger.Tags)             {}
func (m *MockLogger) Error(ctx context.Context, err error, msg string, tags logger.Tags) {}
func (m *MockLogger) Fatal(ctx context.Context, err error, msg string, tags logger.Tags) {}

// Test helper to create events
func createTestEvent(eventType, productID, supplierID string, timestamp time.Time) *events.Event {
	return &events.Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    "test",
		Timestamp: timestamp,
		Data: map[string]interface{}{
			"product_id":  productID,
			"supplier_id": supplierID,
		},
	}
}

func TestNewPatternDetector(t *testing.T) {
	lg := &MockLogger{}

	detector := NewPatternDetector(PatternDetectorConfig{}, lg)

	require.NotNil(t, detector)
	assert.Equal(t, 3, detector.config.RecurringStockoutThreshold)
	assert.Equal(t, 7*24*time.Hour, detector.config.RecurringStockoutWindow)
}

func TestPatternDetector_DetectRecurringStockouts(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New().String()
	lg := &MockLogger{}

	tests := []struct {
		name             string
		events           []*events.Event
		expectedCount    int
		expectedSeverity entities.PatternSeverity
	}{
		{
			name: "detects recurring stockouts with 3+ events",
			events: []*events.Event{
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-1*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-2*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-3*time.Hour)),
			},
			expectedCount:    1,
			expectedSeverity: entities.PatternSeverityLow,
		},
		{
			name: "detects high severity with many events",
			events: []*events.Event{
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-1*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-2*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-3*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-4*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-5*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-6*time.Hour)),
			},
			expectedCount:    1,
			expectedSeverity: entities.PatternSeverityHigh,
		},
		{
			name: "no pattern with insufficient events",
			events: []*events.Event{
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-1*time.Hour)),
				createTestEvent("buffer.critical", productID, "", time.Now().Add(-2*time.Hour)),
			},
			expectedCount: 0,
		},
		{
			name:          "no pattern with empty events",
			events:        []*events.Event{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEventStore := new(MockEventStore)
			mockPatternStore := new(MockPatternStore)
			mockAlertService := new(MockAlertService)

			mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return(tt.events, nil)
			mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return([]*events.Event{}, nil)
			mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return([]*events.Event{}, nil)
			mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return([]*events.Event{}, nil)
			mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
			mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

			detector := NewPatternDetector(PatternDetectorConfig{
				EventStore:   mockEventStore,
				PatternStore: mockPatternStore,
				Alerter:      mockAlertService,
			}, lg)

			patterns, err := detector.DetectPatterns(ctx, orgID)

			require.NoError(t, err)
			assert.Len(t, patterns, tt.expectedCount)

			if tt.expectedCount > 0 {
				assert.Equal(t, entities.PatternTypeRecurringStockout, patterns[0].Type)
				assert.Equal(t, tt.expectedSeverity, patterns[0].Severity)
				assert.Equal(t, len(tt.events), patterns[0].Statistics.Occurrences)
			}
		})
	}
}

func TestPatternDetector_DetectSupplierDelays(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	supplierID := uuid.New().String()
	lg := &MockLogger{}

	delayEvents := []*events.Event{
		{
			ID:        uuid.New().String(),
			Type:      "delivery.delayed",
			Source:    "test",
			Timestamp: time.Now().Add(-1 * 24 * time.Hour),
			Data: map[string]interface{}{
				"supplier_id":   supplierID,
				"supplier_name": "Test Supplier",
				"delay_days":    float64(3),
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "delivery.delayed",
			Source:    "test",
			Timestamp: time.Now().Add(-3 * 24 * time.Hour),
			Data: map[string]interface{}{
				"supplier_id":   supplierID,
				"supplier_name": "Test Supplier",
				"delay_days":    float64(2),
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "delivery.delayed",
			Source:    "test",
			Timestamp: time.Now().Add(-5 * 24 * time.Hour),
			Data: map[string]interface{}{
				"supplier_id":   supplierID,
				"supplier_name": "Test Supplier",
				"delay_days":    float64(5),
			},
		},
	}

	mockEventStore := new(MockEventStore)
	mockPatternStore := new(MockPatternStore)
	mockAlertService := new(MockAlertService)

	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return(delayEvents, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return([]*events.Event{}, nil)
	mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
	mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore:   mockEventStore,
		PatternStore: mockPatternStore,
		Alerter:      mockAlertService,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	require.NoError(t, err)
	require.Len(t, patterns, 1)
	assert.Equal(t, entities.PatternTypeSupplierDelay, patterns[0].Type)
	assert.Equal(t, 3, patterns[0].Statistics.Occurrences)
	assert.Equal(t, float64(10), patterns[0].Statistics.TotalImpactValue) // 3+2+5 delay days
}

func TestPatternDetector_DetectDemandSpikes(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New().String()
	lg := &MockLogger{}

	demandEvents := []*events.Event{
		{
			ID:        uuid.New().String(),
			Type:      "demand.updated",
			Source:    "test",
			Timestamp: time.Now().Add(-5 * 24 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"quantity":   float64(100),
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "demand.updated",
			Source:    "test",
			Timestamp: time.Now().Add(-3 * 24 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"quantity":   float64(110),
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "demand.updated",
			Source:    "test",
			Timestamp: time.Now().Add(-1 * 24 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"quantity":   float64(180), // 80% increase over average
			},
		},
	}

	mockEventStore := new(MockEventStore)
	mockPatternStore := new(MockPatternStore)
	mockAlertService := new(MockAlertService)

	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return(demandEvents, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return([]*events.Event{}, nil)
	mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
	mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore:   mockEventStore,
		PatternStore: mockPatternStore,
		Alerter:      mockAlertService,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	require.NoError(t, err)
	// Should detect demand spike (180 vs avg of 105 = 71% increase, > 50% threshold)
	require.Len(t, patterns, 1)
	assert.Equal(t, entities.PatternTypeDemandSpike, patterns[0].Type)
}

func TestPatternDetector_DetectBufferPenetrationPatterns(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New().String()
	lg := &MockLogger{}

	bufferEvents := []*events.Event{
		{
			ID:        uuid.New().String(),
			Type:      "buffer.status_changed",
			Source:    "test",
			Timestamp: time.Now().Add(-1 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"from_zone":  "yellow",
				"to_zone":    "red",
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "buffer.status_changed",
			Source:    "test",
			Timestamp: time.Now().Add(-2 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"from_zone":  "green",
				"to_zone":    "red",
			},
		},
		{
			ID:        uuid.New().String(),
			Type:      "buffer.status_changed",
			Source:    "test",
			Timestamp: time.Now().Add(-3 * time.Hour),
			Data: map[string]interface{}{
				"product_id": productID,
				"from_zone":  "yellow",
				"to_zone":    "red",
			},
		},
	}

	mockEventStore := new(MockEventStore)
	mockPatternStore := new(MockPatternStore)
	mockAlertService := new(MockAlertService)

	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return(bufferEvents, nil)
	mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
	mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore:   mockEventStore,
		PatternStore: mockPatternStore,
		Alerter:      mockAlertService,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	require.NoError(t, err)
	require.Len(t, patterns, 1)
	assert.Equal(t, entities.PatternTypeBufferPenetration, patterns[0].Type)
	assert.Equal(t, entities.PatternSeverityHigh, patterns[0].Severity)
}

func TestPatternDetector_WithAIEnrichment(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	productID := uuid.New().String()
	lg := &MockLogger{}

	stockoutEvents := []*events.Event{
		createTestEvent("buffer.critical", productID, "", time.Now().Add(-1*time.Hour)),
		createTestEvent("buffer.critical", productID, "", time.Now().Add(-2*time.Hour)),
		createTestEvent("buffer.critical", productID, "", time.Now().Add(-3*time.Hour)),
	}

	mockEventStore := new(MockEventStore)
	mockPatternStore := new(MockPatternStore)
	mockAlertService := new(MockAlertService)
	mockAIAnalyzer := new(MockAIAnalyzer)

	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return(stockoutEvents, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return([]*events.Event{}, nil)
	mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
	mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

	aiResponse := &providers.AIAnalysisResponse{
		Summary:      "Recurring stockout pattern detected",
		FullAnalysis: "Root cause: Insufficient buffer sizing for this product category",
		Reasoning:    "Based on DDMRP principles, the buffer is undersized",
		Recommendations: []providers.AIRecommendation{
			{
				Action:          "Increase buffer by 20%",
				Reasoning:       "Current buffer insufficient for demand variability",
				ExpectedOutcome: "Reduced stockout frequency by 80%",
				Effort:          "low",
				Impact:          "high",
			},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:     "high",
			RevenueImpact: 15000,
		},
		Confidence: 0.88,
	}
	mockAIAnalyzer.On("Analyze", ctx, mock.Anything).Return(aiResponse, nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore:   mockEventStore,
		PatternStore: mockPatternStore,
		Alerter:      mockAlertService,
		AIAnalyzer:   mockAIAnalyzer,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	require.NoError(t, err)
	require.Len(t, patterns, 1)

	// Verify AI enrichment
	assert.Contains(t, patterns[0].RootCauseAnalysis, "Insufficient buffer sizing")
	assert.Len(t, patterns[0].Recommendations, 1)
	assert.Equal(t, "Increase buffer by 20%", patterns[0].Recommendations[0].Action)
	assert.Equal(t, 0.88, patterns[0].Statistics.ConfidenceScore)
}

func TestPatternDetector_MultipleProducts(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	product1 := uuid.New().String()
	product2 := uuid.New().String()
	lg := &MockLogger{}

	// Both products have recurring stockouts
	stockoutEvents := []*events.Event{
		createTestEvent("buffer.critical", product1, "", time.Now().Add(-1*time.Hour)),
		createTestEvent("buffer.critical", product1, "", time.Now().Add(-2*time.Hour)),
		createTestEvent("buffer.critical", product1, "", time.Now().Add(-3*time.Hour)),
		createTestEvent("buffer.critical", product2, "", time.Now().Add(-1*time.Hour)),
		createTestEvent("buffer.critical", product2, "", time.Now().Add(-2*time.Hour)),
		createTestEvent("buffer.critical", product2, "", time.Now().Add(-3*time.Hour)),
		createTestEvent("buffer.critical", product2, "", time.Now().Add(-4*time.Hour)),
	}

	mockEventStore := new(MockEventStore)
	mockPatternStore := new(MockPatternStore)
	mockAlertService := new(MockAlertService)

	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).Return(stockoutEvents, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).Return([]*events.Event{}, nil)
	mockPatternStore.On("Save", ctx, mock.Anything).Return(nil)
	mockAlertService.On("SendPatternAlert", ctx, mock.Anything, mock.Anything).Return(nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore:   mockEventStore,
		PatternStore: mockPatternStore,
		Alerter:      mockAlertService,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	require.NoError(t, err)
	assert.Len(t, patterns, 2) // Two separate patterns for two products
}

func TestPatternDetector_SeverityCalculation(t *testing.T) {
	lg := &MockLogger{}
	detector := NewPatternDetector(PatternDetectorConfig{}, lg)

	tests := []struct {
		occurrences int
		threshold   int
		expected    entities.PatternSeverity
	}{
		{3, 3, entities.PatternSeverityLow},       // ratio = 1.0
		{5, 3, entities.PatternSeverityMedium},    // ratio = 1.67
		{6, 3, entities.PatternSeverityHigh},      // ratio = 2.0
		{9, 3, entities.PatternSeverityCritical},  // ratio = 3.0
		{12, 3, entities.PatternSeverityCritical}, // ratio = 4.0
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("occurrences_%d_threshold_%d", tt.occurrences, tt.threshold), func(t *testing.T) {
			severity := detector.calculateSeverity(tt.occurrences, tt.threshold)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestPatternDetector_ConfidenceCalculation(t *testing.T) {
	lg := &MockLogger{}
	detector := NewPatternDetector(PatternDetectorConfig{}, lg)

	tests := []struct {
		occurrences int
		threshold   int
		minExpected float64
		maxExpected float64
	}{
		{3, 3, 0.70, 0.71},  // base confidence
		{4, 3, 0.74, 0.76},  // +0.05
		{5, 3, 0.79, 0.81},  // +0.10
		{10, 3, 0.94, 0.96}, // capped near 0.95
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("occurrences_%d", tt.occurrences), func(t *testing.T) {
			confidence := detector.calculateConfidence(tt.occurrences, tt.threshold)
			assert.GreaterOrEqual(t, confidence, tt.minExpected)
			assert.LessOrEqual(t, confidence, tt.maxExpected)
		})
	}
}

func TestPatternDetector_CreatePatternNotification(t *testing.T) {
	lg := &MockLogger{}
	detector := NewPatternDetector(PatternDetectorConfig{}, lg)

	pattern := entities.NewPattern(
		uuid.New(),
		entities.PatternTypeRecurringStockout,
		entities.PatternSeverityCritical,
		"Test Pattern",
		"Test Description",
	)
	pattern.RootCauseAnalysis = "Root cause analysis"
	pattern.Statistics.Occurrences = 5
	pattern.Statistics.ConfidenceScore = 0.85
	pattern.AddRecommendation("Action 1", 1, "Outcome 1", "low", 1000)
	pattern.AddRecommendation("Action 2", 2, "Outcome 2", "medium", 500)

	notification := detector.createPatternNotification(pattern)

	assert.NotNil(t, notification)
	assert.Equal(t, domain.NotificationPriorityCritical, notification.Priority)
	assert.Equal(t, domain.NotificationTypeInsight, notification.Type)
	assert.Equal(t, "Test Pattern", notification.Title)
	assert.Equal(t, "Test Description", notification.Summary)
	assert.Equal(t, "Root cause analysis", notification.FullAnalysis)
	assert.Len(t, notification.Recommendations, 2)

	// Should be sorted by priority
	assert.Equal(t, 1, notification.Recommendations[0].PriorityOrder)
	assert.Equal(t, 2, notification.Recommendations[1].PriorityOrder)
}

func TestPatternDetector_EventStoreError(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	lg := &MockLogger{}

	mockEventStore := new(MockEventStore)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.critical", mock.Anything).
		Return(nil, fmt.Errorf("database error"))
	mockEventStore.On("GetEventsByType", ctx, orgID, "delivery.delayed", mock.Anything).
		Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "demand.updated", mock.Anything).
		Return([]*events.Event{}, nil)
	mockEventStore.On("GetEventsByType", ctx, orgID, "buffer.status_changed", mock.Anything).
		Return([]*events.Event{}, nil)

	detector := NewPatternDetector(PatternDetectorConfig{
		EventStore: mockEventStore,
	}, lg)

	patterns, err := detector.DetectPatterns(ctx, orgID)

	// Should continue despite errors for individual detection types
	require.NoError(t, err)
	assert.Empty(t, patterns)
}

func TestExtractProductID(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "with product_id",
			data:     map[string]interface{}{"product_id": "prod-123"},
			expected: "prod-123",
		},
		{
			name:     "with buffer_id fallback",
			data:     map[string]interface{}{"buffer_id": "buf-456"},
			expected: "buf-456",
		},
		{
			name:     "empty data",
			data:     map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &events.Event{Data: tt.data}
			result := extractProductID(event)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSupplierName(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "with supplier_name",
			data:     map[string]interface{}{"supplier_name": "ACME Corp", "supplier_id": "sup-123"},
			expected: "ACME Corp",
		},
		{
			name:     "fallback to supplier_id",
			data:     map[string]interface{}{"supplier_id": "sup-456"},
			expected: "sup-456",
		},
		{
			name:     "empty data",
			data:     map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &events.Event{Data: tt.data}
			result := extractSupplierName(event)
			assert.Equal(t, tt.expected, result)
		})
	}
}
