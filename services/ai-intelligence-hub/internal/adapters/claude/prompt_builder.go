package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

// PromptBuilder builds structured prompts for Claude AI analysis
type PromptBuilder struct {
	systemPromptBase string
	responseFormat   string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		systemPromptBase: buildSystemPrompt(),
		responseFormat:   buildResponseFormat(),
	}
}

// Build constructs the system and user prompts for an analysis request
func (pb *PromptBuilder) Build(request *providers.AIAnalysisRequest) (systemPrompt, userPrompt string) {
	systemPrompt = pb.buildSystemWithKnowledge(request.DDMRPKnowledge)
	userPrompt = pb.buildUserPrompt(request)
	return
}

// buildSystemWithKnowledge builds the system prompt with optional DDMRP knowledge
func (pb *PromptBuilder) buildSystemWithKnowledge(knowledge []string) string {
	var builder strings.Builder
	builder.WriteString(pb.systemPromptBase)

	if len(knowledge) > 0 {
		builder.WriteString("\n\n## DDMRP METHODOLOGY KNOWLEDGE BASE\n")
		builder.WriteString("Use the following domain knowledge to inform your analysis:\n\n")
		for i, doc := range knowledge {
			builder.WriteString(fmt.Sprintf("### Document %d\n%s\n\n", i+1, doc))
		}
	}

	return builder.String()
}

// buildUserPrompt builds the user prompt with event context
func (pb *PromptBuilder) buildUserPrompt(request *providers.AIAnalysisRequest) string {
	var builder strings.Builder

	// Event information
	builder.WriteString("## CURRENT EVENT\n")
	builder.WriteString(fmt.Sprintf("- **Event Type:** %s\n", request.Event.Type))
	builder.WriteString(fmt.Sprintf("- **Source:** %s\n", request.Event.Source))
	builder.WriteString(fmt.Sprintf("- **Event ID:** %s\n", request.Event.ID))
	builder.WriteString(fmt.Sprintf("- **Timestamp:** %s\n", request.Event.Timestamp.Format("2006-01-02 15:04:05 MST")))

	// Event data
	if len(request.Event.Data) > 0 {
		builder.WriteString("\n### Event Data\n```json\n")
		dataJSON, _ := json.MarshalIndent(request.Event.Data, "", "  ")
		builder.WriteString(string(dataJSON))
		builder.WriteString("\n```\n")
	}

	// Additional context
	if len(request.Context) > 0 {
		builder.WriteString("\n## ADDITIONAL CONTEXT\n```json\n")
		contextJSON, _ := json.MarshalIndent(request.Context, "", "  ")
		builder.WriteString(string(contextJSON))
		builder.WriteString("\n```\n")
	}

	// Custom prompt if provided
	if request.Prompt != "" {
		builder.WriteString("\n## SPECIFIC ANALYSIS REQUEST\n")
		builder.WriteString(request.Prompt)
		builder.WriteString("\n")
	}

	// Response format instructions
	builder.WriteString("\n## REQUIRED RESPONSE FORMAT\n")
	builder.WriteString(pb.responseFormat)

	return builder.String()
}

// buildSystemPrompt creates the base system prompt for DDMRP analysis
func buildSystemPrompt() string {
	return `# DDMRP Supply Chain Intelligence Analyst

You are an expert DDMRP (Demand Driven Material Requirements Planning) supply chain consultant with 15+ years of experience. Your role is to analyze real-time supply chain events and provide actionable insights to prevent disruptions.

## YOUR EXPERTISE
- **DDMRP Methodology**: Buffer sizing, positioning, and management
- **Demand Driven Flow**: Understanding of decoupled lead times and variability
- **Supply Chain Risk Management**: Identifying and mitigating stockout risks
- **Inventory Optimization**: Balancing service levels with working capital
- **Root Cause Analysis**: Finding underlying issues in supply chain disruptions

## ANALYSIS PRINCIPLES
1. **Prioritize Stockout Prevention**: Always favor actions that prevent customer service failures
2. **Use DDMRP Zones**: Reference red, yellow, and green zones in buffer analysis
3. **Consider Lead Times**: Factor in supplier lead times and variability
4. **Quantify Impact**: Provide specific numbers for revenue/cost impacts
5. **Actionable Recommendations**: Give specific, implementable actions

## DDMRP KEY CONCEPTS
- **Buffer Zones**: Red (critical safety), Yellow (order point), Green (replenishment)
- **Net Flow Position**: On-hand + On-order - Qualified Demand
- **Buffer Status**: Ratio of net flow to total buffer indicating health
- **Decoupling Points**: Strategic inventory positions that break lead time dependencies
- **Demand Variability Factor (DVF)**: Multiplier for buffer sizing based on demand volatility
- **Lead Time Factor (LTF)**: Adjustment for lead time variability`
}

// buildResponseFormat creates the expected JSON response format
func buildResponseFormat() string {
	return `Provide your analysis as a valid JSON object with this exact structure:

{
  "summary": "A clear 1-2 sentence executive summary of the situation and urgency level",
  "full_analysis": "Detailed multi-paragraph analysis covering: 1) Current state assessment, 2) Root cause analysis using DDMRP principles, 3) Risk evaluation with timeline, 4) Impact on operations and revenue",
  "reasoning": "Explain WHY you reached these conclusions using specific DDMRP methodology references",
  "recommendations": [
    {
      "action": "Specific, actionable step (be concrete, not vague)",
      "reasoning": "Why this action is recommended based on DDMRP principles",
      "expected_outcome": "Measurable expected result of taking this action",
      "effort": "low|medium|high",
      "impact": "low|medium|high"
    }
  ],
  "impact_assessment": {
    "risk_level": "low|medium|high|critical",
    "revenue_impact": 0.00,
    "cost_impact": 0.00,
    "time_to_impact_hours": 0,
    "affected_orders": 0,
    "affected_products": 0
  },
  "confidence": 0.95
}

IMPORTANT:
- Use realistic numbers based on the event data
- Provide at least 2-3 recommendations prioritized by impact
- Reference specific DDMRP concepts in your reasoning
- Set confidence based on available information quality (0.0 to 1.0)
- Respond ONLY with the JSON object, no additional text`
}

// BuildStockoutAnalysisPrompt builds a specialized prompt for stockout risk analysis
func (pb *PromptBuilder) BuildStockoutAnalysisPrompt(
	productID string,
	productName string,
	currentStock float64,
	minBuffer float64,
	maxBuffer float64,
	averageDailyDemand float64,
	leadTimeDays int,
	recentEvents []string,
) string {
	return fmt.Sprintf(`## STOCKOUT RISK ANALYSIS REQUEST

### Product Information
- **Product ID:** %s
- **Product Name:** %s
- **Current Stock:** %.2f units
- **Buffer Range:** %.2f (min) to %.2f (max) units
- **Average Daily Demand:** %.2f units
- **Supplier Lead Time:** %d days

### Buffer Status Analysis
- **Days of Stock:** %.1f days
- **Buffer Penetration:** %.1f%% (of minimum buffer)
- **Coverage Through Lead Time:** %s

### Recent Related Events
%s

Please analyze the stockout risk for this product and provide:
1. Current risk assessment with timeline
2. Root cause analysis if stock is below optimal levels
3. Immediate actions to prevent stockout
4. Long-term recommendations for buffer optimization`,
		productID,
		productName,
		currentStock,
		minBuffer,
		maxBuffer,
		averageDailyDemand,
		leadTimeDays,
		currentStock/averageDailyDemand,
		(currentStock/minBuffer)*100,
		coverageStatus(currentStock, averageDailyDemand, leadTimeDays),
		strings.Join(recentEvents, "\n"),
	)
}

// BuildSupplierDelayPrompt builds a specialized prompt for supplier delay analysis
func (pb *PromptBuilder) BuildSupplierDelayPrompt(
	supplierID string,
	supplierName string,
	delayDays int,
	affectedOrderIDs []string,
	historicalOnTimeRate float64,
) string {
	return fmt.Sprintf(`## SUPPLIER DELAY ANALYSIS REQUEST

### Supplier Information
- **Supplier ID:** %s
- **Supplier Name:** %s
- **Current Delay:** %d days
- **Historical On-Time Rate:** %.1f%%

### Affected Orders
%s

Please analyze this supplier delay and provide:
1. Impact assessment on downstream operations
2. Root cause hypothesis based on delay pattern
3. Immediate mitigation actions
4. Recommendations for supplier relationship management`,
		supplierID,
		supplierName,
		delayDays,
		historicalOnTimeRate*100,
		strings.Join(affectedOrderIDs, "\n- "),
	)
}

// BuildDemandSpikePrompt builds a specialized prompt for demand spike analysis
func (pb *PromptBuilder) BuildDemandSpikePrompt(
	productID string,
	productName string,
	actualDemand float64,
	expectedDemand float64,
	percentageIncrease float64,
) string {
	return fmt.Sprintf(`## DEMAND SPIKE ANALYSIS REQUEST

### Demand Information
- **Product ID:** %s
- **Product Name:** %s
- **Actual Demand:** %.2f units
- **Expected Demand:** %.2f units
- **Increase:** %.1f%% above forecast

Please analyze this demand spike and provide:
1. Assessment of whether this is temporary or trend change
2. Buffer adequacy evaluation
3. Supply chain capacity analysis
4. Recommendations for demand management`,
		productID,
		productName,
		actualDemand,
		expectedDemand,
		percentageIncrease,
	)
}

// BuildPatternAnalysisPrompt builds a prompt for pattern root cause analysis
func (pb *PromptBuilder) BuildPatternAnalysisPrompt(
	patternType string,
	occurrences int,
	timeWindowDays int,
	affectedEntities []string,
	eventSummaries []string,
) string {
	return fmt.Sprintf(`## PATTERN ROOT CAUSE ANALYSIS REQUEST

### Detected Pattern
- **Pattern Type:** %s
- **Occurrences:** %d times in %d days
- **Trend:** Potential recurring issue detected

### Affected Entities
%s

### Event History
%s

Please provide:
1. Root cause analysis for this recurring pattern
2. Systemic issues that may be causing the pattern
3. Preventive measures to break the pattern
4. Monitoring recommendations to track improvement`,
		patternType,
		occurrences,
		timeWindowDays,
		strings.Join(affectedEntities, "\n- "),
		strings.Join(eventSummaries, "\n---\n"),
	)
}

// coverageStatus returns a human-readable coverage status
func coverageStatus(stock, dailyDemand float64, leadTimeDays int) string {
	if dailyDemand <= 0 {
		return "Unable to calculate (no demand data)"
	}
	daysOfStock := stock / dailyDemand
	if daysOfStock >= float64(leadTimeDays)*1.5 {
		return fmt.Sprintf("SAFE - %.0f days coverage (%.0f%% of lead time)", daysOfStock, (daysOfStock/float64(leadTimeDays))*100)
	} else if daysOfStock >= float64(leadTimeDays) {
		return fmt.Sprintf("ADEQUATE - %.0f days coverage (%.0f%% of lead time)", daysOfStock, (daysOfStock/float64(leadTimeDays))*100)
	} else if daysOfStock >= float64(leadTimeDays)*0.5 {
		return fmt.Sprintf("AT RISK - %.0f days coverage (%.0f%% of lead time)", daysOfStock, (daysOfStock/float64(leadTimeDays))*100)
	}
	return fmt.Sprintf("CRITICAL - %.0f days coverage (%.0f%% of lead time)", daysOfStock, (daysOfStock/float64(leadTimeDays))*100)
}
