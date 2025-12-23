package analysis_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/mocks"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/usecases/analysis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeStockoutRiskUseCase_Execute_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	// Prepare test event
	orgID := uuid.New()
	productID := "PROD-ABC-123"
	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: orgID.String(),
		Data: map[string]interface{}{
			"product_id":        productID,
			"current_stock":     50.0,
			"min_buffer":        100.0,
			"daily_consumption": 10.0,
		},
	}

	// Mock RAG knowledge retrieval
	mockKnowledge := []string{
		"DDMRP buffer zone methodology...",
		"Red zone penetration requires immediate action...",
		"Consider emergency replenishment orders...",
	}
	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).
		Return(mockKnowledge, nil)

	// Mock AI analysis response
	mockAIResponse := &providers.AIAnalysisResponse{
		Summary:      "Critical buffer status. Immediate action required.",
		FullAnalysis: "Product is in red zone with 5 days until stockout...",
		Reasoning:    "DDMRP methodology indicates emergency replenishment needed",
		Recommendations: []providers.AIRecommendation{
			{
				Action:          "Place emergency replenishment order",
				Reasoning:       "Current stock insufficient for lead time",
				ExpectedOutcome: "Stockout prevented, buffer restored",
				Effort:          "medium",
				Impact:          "high",
			},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         "critical",
			RevenueImpact:     15000.00,
			CostImpact:        200.00,
			TimeToImpactHours: 120,
			AffectedOrders:    5,
			AffectedProducts:  1,
		},
		Confidence: 0.92,
	}
	mockAI.On("Analyze", ctx, mock.MatchedBy(func(req *providers.AIAnalysisRequest) bool {
		return req.Event.ID == "evt-123" &&
			len(req.DDMRPKnowledge) == 3 &&
			req.Prompt != ""
	})).Return(mockAIResponse, nil)

	// Mock repository create - 5 days until stockout = Medium priority
	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *domain.AINotification) bool {
		return n.OrganizationID == orgID &&
			n.Type == domain.NotificationTypeAlert &&
			n.Priority == domain.NotificationPriorityMedium &&
			n.Title == "Stockout Risk: "+productID &&
			n.Summary == mockAIResponse.Summary &&
			len(n.Recommendations) == 1
	})).Return(nil)

	// Execute
	err := useCase.Execute(ctx, event)

	// Assert
	require.NoError(t, err)
	mockAI.AssertExpectations(t)
	mockRAG.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAnalyzeStockoutRiskUseCase_Execute_NilEvent(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	// Execute
	err := useCase.Execute(ctx, nil)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event cannot be nil")
}

func TestAnalyzeStockoutRiskUseCase_Execute_MissingProductID(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	// Event without product_id
	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"current_stock": 50.0,
		},
	}

	// Execute
	err := useCase.Execute(ctx, event)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing product_id")
}

func TestAnalyzeStockoutRiskUseCase_Execute_RAGFailureDoesNotStopProcess(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	orgID := uuid.New()
	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: orgID.String(),
		Data: map[string]interface{}{
			"product_id":        "PROD-123",
			"current_stock":     50.0,
			"min_buffer":        100.0,
			"daily_consumption": 10.0,
		},
	}

	// Mock RAG to return error
	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).
		Return(nil, assert.AnError)

	// Mock AI should still be called with empty knowledge
	mockAIResponse := &providers.AIAnalysisResponse{
		Summary:      "Analysis without RAG knowledge",
		FullAnalysis: "Product analysis...",
		Reasoning:    "Based on current data...",
		Recommendations: []providers.AIRecommendation{
			{Action: "Test action", Reasoning: "Test reasoning", ExpectedOutcome: "Test outcome", Effort: "low", Impact: "medium"},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         "medium",
			TimeToImpactHours: 48,
		},
		Confidence: 0.75,
	}
	mockAI.On("Analyze", ctx, mock.MatchedBy(func(req *providers.AIAnalysisRequest) bool {
		// Should be called with empty knowledge array
		return len(req.DDMRPKnowledge) == 0
	})).Return(mockAIResponse, nil)

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.AINotification")).Return(nil)

	// Execute
	err := useCase.Execute(ctx, event)

	// Assert - should succeed despite RAG failure
	require.NoError(t, err)
	mockAI.AssertExpectations(t)
	mockRAG.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAnalyzeStockoutRiskUseCase_Execute_AIAnalyzerFailure(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"product_id":    "PROD-123",
			"current_stock": 50.0,
			"min_buffer":    100.0,
		},
	}

	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).
		Return([]string{}, nil)

	// AI analyzer returns error
	mockAI.On("Analyze", ctx, mock.AnythingOfType("*providers.AIAnalysisRequest")).
		Return(nil, assert.AnError)

	// Execute
	err := useCase.Execute(ctx, event)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AI analysis failed")
	mockAI.AssertExpectations(t)
	mockRAG.AssertExpectations(t)
	// Repository should NOT be called
	mockRepo.AssertNotCalled(t, "Create")
}

func TestAnalyzeStockoutRiskUseCase_Execute_RepositoryFailure(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"product_id":    "PROD-123",
			"current_stock": 50.0,
			"min_buffer":    100.0,
		},
	}

	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).Return([]string{}, nil)

	mockAIResponse := &providers.AIAnalysisResponse{
		Summary:      "Test",
		FullAnalysis: "Test",
		Reasoning:    "Test",
		Recommendations: []providers.AIRecommendation{
			{Action: "Test", Reasoning: "Test", ExpectedOutcome: "Test", Effort: "low", Impact: "high"},
		},
		ImpactAssessment: providers.AIImpactAssessment{RiskLevel: "high", TimeToImpactHours: 24},
		Confidence:       0.8,
	}
	mockAI.On("Analyze", ctx, mock.AnythingOfType("*providers.AIAnalysisRequest")).
		Return(mockAIResponse, nil)

	// Repository fails
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.AINotification")).
		Return(assert.AnError)

	// Execute
	err := useCase.Execute(ctx, event)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create notification")
	mockRepo.AssertExpectations(t)
}

func TestAnalyzeStockoutRiskUseCase_DeterminePriority(t *testing.T) {
	tests := []struct {
		name              string
		daysUntilStockout float64
		expectedPriority  domain.NotificationPriority
	}{
		{"Critical - less than 2 days", 1.5, domain.NotificationPriorityCritical},
		{"Critical - exactly 2 days", 2.0, domain.NotificationPriorityHigh},
		{"High - less than 5 days", 3.0, domain.NotificationPriorityHigh},
		{"High - exactly 5 days", 5.0, domain.NotificationPriorityMedium},
		{"Medium - less than 10 days", 7.0, domain.NotificationPriorityMedium},
		{"Medium - exactly 10 days", 10.0, domain.NotificationPriorityLow},
		{"Low - more than 10 days", 15.0, domain.NotificationPriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx := context.Background()
			mockAI := mocks.NewMockAIAnalyzer()
			mockRAG := mocks.NewMockRAGKnowledge()
			mockRepo := mocks.NewMockNotificationRepository()
			testLogger := logger.New("test", "debug")

			useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

			event := &events.Event{
				ID:             "evt-test",
				Type:           "buffer.below_minimum",
				OrganizationID: uuid.New().String(),
				Data: map[string]interface{}{
					"product_id":        "PROD-TEST",
					"current_stock":     tt.daysUntilStockout * 10,
					"min_buffer":        100.0,
					"daily_consumption": 10.0,
				},
			}

			mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).Return([]string{}, nil)

			mockAIResponse := &providers.AIAnalysisResponse{
				Summary:         "Test",
				FullAnalysis:    "Test",
				Reasoning:       "Test",
				Recommendations: []providers.AIRecommendation{{Action: "Test", Reasoning: "Test", ExpectedOutcome: "Test", Effort: "low", Impact: "low"}},
				ImpactAssessment: providers.AIImpactAssessment{
					RiskLevel:         "medium",
					TimeToImpactHours: 24,
				},
				Confidence: 0.8,
			}
			mockAI.On("Analyze", ctx, mock.AnythingOfType("*providers.AIAnalysisRequest")).Return(mockAIResponse, nil)

			// Capture the created notification to check priority
			var capturedNotif *domain.AINotification
			mockRepo.On("Create", ctx, mock.MatchedBy(func(n *domain.AINotification) bool {
				capturedNotif = n
				return true
			})).Return(nil)

			// Execute
			err := useCase.Execute(ctx, event)
			require.NoError(t, err)

			// Assert priority
			assert.Equal(t, tt.expectedPriority, capturedNotif.Priority)
		})
	}
}

func TestAnalyzeStockoutRiskUseCase_Execute_TimeToImpactConversion(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	event := &events.Event{
		ID:             "evt-123",
		Type:           "buffer.below_minimum",
		OrganizationID: uuid.New().String(),
		Data: map[string]interface{}{
			"product_id":    "PROD-123",
			"current_stock": 50.0,
			"min_buffer":    100.0,
		},
	}

	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).Return([]string{}, nil)

	// AI returns 48 hours time to impact
	mockAIResponse := &providers.AIAnalysisResponse{
		Summary:      "Test",
		FullAnalysis: "Test",
		Reasoning:    "Test",
		Recommendations: []providers.AIRecommendation{
			{Action: "Test", Reasoning: "Test", ExpectedOutcome: "Test", Effort: "low", Impact: "high"},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         "critical",
			TimeToImpactHours: 48,
		},
		Confidence: 0.9,
	}
	mockAI.On("Analyze", ctx, mock.AnythingOfType("*providers.AIAnalysisRequest")).Return(mockAIResponse, nil)

	var capturedNotif *domain.AINotification
	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *domain.AINotification) bool {
		capturedNotif = n
		return true
	})).Return(nil)

	// Execute
	err := useCase.Execute(ctx, event)
	require.NoError(t, err)

	// Assert time to impact was correctly converted
	require.NotNil(t, capturedNotif.Impact.TimeToImpact)
	assert.Equal(t, 48*time.Hour, *capturedNotif.Impact.TimeToImpact)
}

func TestAnalyzeStockoutRiskUseCase_Execute_CompleteNotificationStructure(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAI := mocks.NewMockAIAnalyzer()
	mockRAG := mocks.NewMockRAGKnowledge()
	mockRepo := mocks.NewMockNotificationRepository()
	testLogger := logger.New("test", "debug")

	useCase := analysis.NewAnalyzeStockoutRiskUseCase(mockAI, mockRAG, mockRepo, testLogger)

	orgID := uuid.New()
	productID := "PROD-XYZ-999"
	event := &events.Event{
		ID:             "evt-complete-test",
		Type:           "buffer.below_minimum",
		OrganizationID: orgID.String(),
		Data: map[string]interface{}{
			"product_id":        productID,
			"current_stock":     15.0, // 1.5 days until stockout = Critical priority
			"min_buffer":        100.0,
			"daily_consumption": 10.0,
		},
	}

	mockRAG.On("Retrieve", ctx, mock.AnythingOfType("string"), 3).
		Return([]string{"DDMRP knowledge 1", "DDMRP knowledge 2"}, nil)

	mockAIResponse := &providers.AIAnalysisResponse{
		Summary:      "Critical stockout risk detected",
		FullAnalysis: "Detailed analysis of the buffer situation with calculations",
		Reasoning:    "DDMRP methodology indicates immediate replenishment needed",
		Recommendations: []providers.AIRecommendation{
			{
				Action:          "Emergency order for 200 units",
				Reasoning:       "Cover lead time demand plus safety stock",
				ExpectedOutcome: "Prevent stockout and restore green zone",
				Effort:          "high",
				Impact:          "critical",
			},
			{
				Action:          "Expedite existing PO #12345",
				Reasoning:       "Reduce time to stockout",
				ExpectedOutcome: "Additional 100 units arrive 3 days earlier",
				Effort:          "medium",
				Impact:          "high",
			},
		},
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         "critical",
			RevenueImpact:     50000.00,
			CostImpact:        1500.00,
			TimeToImpactHours: 60,
			AffectedOrders:    12,
			AffectedProducts:  1,
		},
		Confidence: 0.95,
	}
	mockAI.On("Analyze", ctx, mock.AnythingOfType("*providers.AIAnalysisRequest")).Return(mockAIResponse, nil)

	var capturedNotif *domain.AINotification
	mockRepo.On("Create", ctx, mock.MatchedBy(func(n *domain.AINotification) bool {
		capturedNotif = n
		return true
	})).Return(nil)

	// Execute
	err := useCase.Execute(ctx, event)
	require.NoError(t, err)

	// Assert complete notification structure
	assert.Equal(t, orgID, capturedNotif.OrganizationID)
	assert.Equal(t, domain.NotificationTypeAlert, capturedNotif.Type)
	assert.Equal(t, domain.NotificationPriorityCritical, capturedNotif.Priority)
	assert.Equal(t, "Stockout Risk: "+productID, capturedNotif.Title)
	assert.Equal(t, "Critical stockout risk detected", capturedNotif.Summary)
	assert.Equal(t, "Detailed analysis of the buffer situation with calculations", capturedNotif.FullAnalysis)
	assert.Equal(t, "DDMRP methodology indicates immediate replenishment needed", capturedNotif.Reasoning)

	// Check impact assessment
	assert.Equal(t, "critical", capturedNotif.Impact.RiskLevel)
	assert.Equal(t, 50000.00, capturedNotif.Impact.RevenueImpact)
	assert.Equal(t, 1500.00, capturedNotif.Impact.CostImpact)
	assert.Equal(t, 12, capturedNotif.Impact.AffectedOrders)
	assert.Equal(t, 1, capturedNotif.Impact.AffectedProducts)
	require.NotNil(t, capturedNotif.Impact.TimeToImpact)
	assert.Equal(t, 60*time.Hour, *capturedNotif.Impact.TimeToImpact)

	// Check recommendations
	require.Len(t, capturedNotif.Recommendations, 2)
	assert.Equal(t, "Emergency order for 200 units", capturedNotif.Recommendations[0].Action)
	assert.Equal(t, "Cover lead time demand plus safety stock", capturedNotif.Recommendations[0].Reasoning)
	assert.Equal(t, "high", capturedNotif.Recommendations[0].Effort)
	assert.Equal(t, "critical", capturedNotif.Recommendations[0].Impact)
	assert.Equal(t, 1, capturedNotif.Recommendations[0].PriorityOrder)

	assert.Equal(t, "Expedite existing PO #12345", capturedNotif.Recommendations[1].Action)
	assert.Equal(t, 2, capturedNotif.Recommendations[1].PriorityOrder)

	// Check source events and related entities
	assert.Equal(t, []string{"evt-complete-test"}, capturedNotif.SourceEvents)
	assert.Equal(t, map[string][]string{"product_ids": {productID}}, capturedNotif.RelatedEntities)

	// Check initial status
	assert.Equal(t, domain.NotificationStatusUnread, capturedNotif.Status)
}
