package providers

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
)

type AIAnalyzer interface {
	Analyze(ctx context.Context, request *AIAnalysisRequest) (*AIAnalysisResponse, error)
}

type AIAnalysisRequest struct {
	Event          *events.Event
	Context        map[string]interface{}
	DDMRPKnowledge []string
	Prompt         string
}

type AIAnalysisResponse struct {
	Summary          string
	FullAnalysis     string
	Reasoning        string
	Recommendations  []AIRecommendation
	ImpactAssessment AIImpactAssessment
	Confidence       float64
}

type AIRecommendation struct {
	Action          string
	Reasoning       string
	ExpectedOutcome string
	Effort          string
	Impact          string
}

type AIImpactAssessment struct {
	RiskLevel            string
	RevenueImpact        float64
	CostImpact           float64
	TimeToImpactHours    int
	AffectedOrders       int
	AffectedProducts     int
}

type RAGKnowledge interface {
	Retrieve(ctx context.Context, query string, topK int) ([]string, error)
	Initialize(ctx context.Context) error
}
