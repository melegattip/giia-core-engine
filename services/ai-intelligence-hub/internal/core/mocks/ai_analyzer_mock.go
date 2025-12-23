package mocks

import (
	"context"

	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/stretchr/testify/mock"
)

// MockAIAnalyzer is a mock implementation of providers.AIAnalyzer
type MockAIAnalyzer struct {
	mock.Mock
}

func NewMockAIAnalyzer() *MockAIAnalyzer {
	return &MockAIAnalyzer{}
}

func (m *MockAIAnalyzer) Analyze(ctx context.Context, request *providers.AIAnalysisRequest) (*providers.AIAnalysisResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*providers.AIAnalysisResponse), args.Error(1)
}

// MockRAGKnowledge is a mock implementation of providers.RAGKnowledge
type MockRAGKnowledge struct {
	mock.Mock
}

func NewMockRAGKnowledge() *MockRAGKnowledge {
	return &MockRAGKnowledge{}
}

func (m *MockRAGKnowledge) Retrieve(ctx context.Context, query string, topK int) ([]string, error) {
	args := m.Called(ctx, query, topK)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRAGKnowledge) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
