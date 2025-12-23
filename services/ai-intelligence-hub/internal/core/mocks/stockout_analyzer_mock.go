package mocks

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/stretchr/testify/mock"
)

// MockStockoutRiskAnalyzer is a mock implementation of the StockoutRiskAnalyzer interface
type MockStockoutRiskAnalyzer struct {
	mock.Mock
}

func NewMockStockoutRiskAnalyzer() *MockStockoutRiskAnalyzer {
	return &MockStockoutRiskAnalyzer{}
}

func (m *MockStockoutRiskAnalyzer) Execute(ctx context.Context, event *events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
