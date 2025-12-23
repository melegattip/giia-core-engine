package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type ClaudeClient struct {
	apiKey  string
	model   string
	baseURL string
	logger  logger.Logger
}

func NewClaudeClient(apiKey string, model string, logger logger.Logger) providers.AIAnalyzer {
	if model == "" {
		model = "claude-3-5-sonnet-20250929"
	}
	return &ClaudeClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1/messages",
		logger:  logger,
	}
}

func (c *ClaudeClient) Analyze(ctx context.Context, request *providers.AIAnalysisRequest) (*providers.AIAnalysisResponse, error) {
	if request == nil {
		return nil, errors.NewBadRequest("analysis request cannot be nil")
	}

	prompt := c.buildPrompt(request)

	c.logger.Debug(ctx, "Sending request to Claude API", logger.Tags{
		"model":       c.model,
		"event_type":  request.Event.Type,
		"prompt_size": len(prompt),
	})

	responseText, err := c.callClaudeAPI(ctx, prompt)
	if err != nil {
		return nil, err
	}

	response, err := c.parseResponse(responseText)
	if err != nil {
		return nil, err
	}

	c.logger.Info(ctx, "AI analysis completed", logger.Tags{
		"event_type": request.Event.Type,
		"confidence": response.Confidence,
		"risk_level": response.ImpactAssessment.RiskLevel,
	})

	return response, nil
}

func (c *ClaudeClient) buildPrompt(request *providers.AIAnalysisRequest) string {
	var knowledgeSection string
	if len(request.DDMRPKnowledge) > 0 {
		knowledgeSection = fmt.Sprintf(`
DDMRP METHODOLOGY KNOWLEDGE:
The following documents contain DDMRP best practices and methodologies:

%s
`, strings.Join(request.DDMRPKnowledge, "\n\n---\n\n"))
	}

	contextJSON, _ := json.MarshalIndent(request.Context, "", "  ")

	prompt := fmt.Sprintf(`You are an expert DDMRP (Demand Driven Material Requirements Planning) supply chain consultant.
You're analyzing a real-time event from the GIIA platform to help prevent supply chain disruptions.

ROLE:
Act as a senior supply chain advisor with 15+ years of DDMRP expertise.
Your goal is to provide actionable, specific recommendations that prevent stockouts while optimizing costs.

%s

CURRENT EVENT:
Type: %s
Source: %s
Event ID: %s

RELATED CONTEXT:
%s

ANALYSIS REQUEST:
%s

Provide your analysis in the following JSON format:
{
  "summary": "A clear, concise 1-2 sentence summary for busy executives",
  "full_analysis": "Detailed multi-paragraph analysis covering:
                    1. Current state assessment
                    2. Root cause analysis
                    3. Risk evaluation
                    4. Impact on operations and revenue",
  "reasoning": "Explain the 'why' behind your analysis using DDMRP principles",
  "recommendations": [
    {
      "action": "Specific action to take (be concrete, not vague)",
      "reasoning": "Why this action is recommended",
      "expected_outcome": "What will happen if this action is taken",
      "effort": "low|medium|high",
      "impact": "low|medium|high"
    }
  ],
  "impact_assessment": {
    "risk_level": "low|medium|high|critical",
    "revenue_impact": 0.0,
    "cost_impact": 0.0,
    "time_to_impact_hours": 0,
    "affected_orders": 0,
    "affected_products": 0
  },
  "confidence": 0.95
}

Be specific with numbers and dates. Reference DDMRP methodology explicitly.
Prioritize preventing stockouts over cost optimization.`,
		knowledgeSection,
		request.Event.Type,
		request.Event.Source,
		request.Event.ID,
		string(contextJSON),
		request.Prompt,
	)

	return prompt
}

func (c *ClaudeClient) callClaudeAPI(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		c.logger.Warn(ctx, "Claude API key not configured, using mock response", nil)
		return c.getMockResponse(), nil
	}

	return "", errors.NewInternalServerError("Claude API integration not yet implemented - requires anthropic-sdk-go dependency")
}

func (c *ClaudeClient) getMockResponse() string {
	return `{
  "summary": "Critical buffer status detected. Immediate action required to prevent stockout.",
  "full_analysis": "The buffer has fallen below the minimum threshold, indicating a critical shortage situation. Based on DDMRP methodology, this represents a red zone penetration that requires urgent intervention. Current stock levels are insufficient to cover average daily demand through the replenishment lead time. Without immediate action, stockout is imminent within 24-48 hours, which would result in lost revenue and customer dissatisfaction.",
  "reasoning": "DDMRP buffer zones are designed to provide visual signals for inventory management. Red zone penetration (below minimum buffer) indicates that safety stock has been consumed and the item is at high risk of stockout. The buffer calculation methodology suggests that minimum buffer should cover average demand during lead time plus variability protection. Current status shows this protection has been breached.",
  "recommendations": [
    {
      "action": "Place emergency replenishment order immediately",
      "reasoning": "Current stock level is below minimum buffer threshold and insufficient to cover demand until next scheduled order arrival",
      "expected_outcome": "Stockout prevented, buffer restored to safe levels within lead time",
      "effort": "medium",
      "impact": "high"
    },
    {
      "action": "Review and recalculate buffer levels",
      "reasoning": "Frequent red zone penetration suggests buffer may be undersized for current demand patterns",
      "expected_outcome": "More appropriate buffer levels that reduce stockout frequency",
      "effort": "medium",
      "impact": "high"
    }
  ],
  "impact_assessment": {
    "risk_level": "critical",
    "revenue_impact": 15000.00,
    "cost_impact": 200.00,
    "time_to_impact_hours": 36,
    "affected_orders": 5,
    "affected_products": 1
  },
  "confidence": 0.92
}`
}

func (c *ClaudeClient) parseResponse(responseText string) (*providers.AIAnalysisResponse, error) {
	responseText = strings.TrimSpace(responseText)

	start := strings.Index(responseText, "{")
	end := strings.LastIndex(responseText, "}")
	if start == -1 || end == -1 {
		return nil, errors.NewInternalServerError("invalid JSON response from AI")
	}

	jsonText := responseText[start : end+1]

	var parsed struct {
		Summary          string `json:"summary"`
		FullAnalysis     string `json:"full_analysis"`
		Reasoning        string `json:"reasoning"`
		Recommendations  []struct {
			Action          string `json:"action"`
			Reasoning       string `json:"reasoning"`
			ExpectedOutcome string `json:"expected_outcome"`
			Effort          string `json:"effort"`
			Impact          string `json:"impact"`
		} `json:"recommendations"`
		ImpactAssessment struct {
			RiskLevel         string  `json:"risk_level"`
			RevenueImpact     float64 `json:"revenue_impact"`
			CostImpact        float64 `json:"cost_impact"`
			TimeToImpactHours int     `json:"time_to_impact_hours"`
			AffectedOrders    int     `json:"affected_orders"`
			AffectedProducts  int     `json:"affected_products"`
		} `json:"impact_assessment"`
		Confidence float64 `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return nil, errors.NewInternalServerError(fmt.Sprintf("failed to parse AI response: %v", err))
	}

	if parsed.Summary == "" {
		return nil, errors.NewInternalServerError("AI response missing summary")
	}

	response := &providers.AIAnalysisResponse{
		Summary:      parsed.Summary,
		FullAnalysis: parsed.FullAnalysis,
		Reasoning:    parsed.Reasoning,
		Confidence:   parsed.Confidence,
		ImpactAssessment: providers.AIImpactAssessment{
			RiskLevel:         parsed.ImpactAssessment.RiskLevel,
			RevenueImpact:     parsed.ImpactAssessment.RevenueImpact,
			CostImpact:        parsed.ImpactAssessment.CostImpact,
			TimeToImpactHours: parsed.ImpactAssessment.TimeToImpactHours,
			AffectedOrders:    parsed.ImpactAssessment.AffectedOrders,
			AffectedProducts:  parsed.ImpactAssessment.AffectedProducts,
		},
	}

	for _, rec := range parsed.Recommendations {
		response.Recommendations = append(response.Recommendations, providers.AIRecommendation{
			Action:          rec.Action,
			Reasoning:       rec.Reasoning,
			ExpectedOutcome: rec.ExpectedOutcome,
			Effort:          rec.Effort,
			Impact:          rec.Impact,
		})
	}

	return response, nil
}
