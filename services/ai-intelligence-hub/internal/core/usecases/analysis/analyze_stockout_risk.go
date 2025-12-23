package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type AnalyzeStockoutRiskUseCase struct {
	aiAnalyzer       providers.AIAnalyzer
	ragKnowledge     providers.RAGKnowledge
	notificationRepo providers.NotificationRepository
	logger           logger.Logger
}

func NewAnalyzeStockoutRiskUseCase(
	aiAnalyzer providers.AIAnalyzer,
	ragKnowledge providers.RAGKnowledge,
	notificationRepo providers.NotificationRepository,
	logger logger.Logger,
) *AnalyzeStockoutRiskUseCase {
	return &AnalyzeStockoutRiskUseCase{
		aiAnalyzer:       aiAnalyzer,
		ragKnowledge:     ragKnowledge,
		notificationRepo: notificationRepo,
		logger:           logger,
	}
}

func (uc *AnalyzeStockoutRiskUseCase) Execute(ctx context.Context, event *events.Event) error {
	if event == nil {
		return errors.NewBadRequest("event cannot be nil")
	}

	uc.logger.Info(ctx, "Analyzing stockout risk", logger.Tags{
		"event_type": event.Type,
		"event_id":   event.ID,
	})

	productID, ok := event.Data["product_id"].(string)
	if !ok {
		return errors.NewBadRequest("missing product_id in event data")
	}

	currentStock, _ := event.Data["current_stock"].(float64)
	minBuffer, _ := event.Data["min_buffer"].(float64)
	dailyConsumption := 10.0

	if dc, ok := event.Data["daily_consumption"].(float64); ok {
		dailyConsumption = dc
	}

	daysUntilStockout := currentStock / dailyConsumption

	knowledge, err := uc.ragKnowledge.Retrieve(ctx, "DDMRP buffer below minimum stockout prevention emergency order", 3)
	if err != nil {
		uc.logger.Warn(ctx, "Failed to retrieve knowledge, continuing without RAG", logger.Tags{
			"error": err.Error(),
		})
		knowledge = []string{}
	}

	analysisContext := map[string]interface{}{
		"product_id":        productID,
		"current_stock":     currentStock,
		"min_buffer":        minBuffer,
		"daily_consumption": dailyConsumption,
		"days_to_stockout":  daysUntilStockout,
	}

	prompt := fmt.Sprintf(`
Analyze this buffer below minimum situation for product "%s".

Current State:
- On-hand stock: %.2f units
- Daily consumption: %.2f units/day
- Days until stockout: %.1f days
- Min buffer: %.2f units

Provide:
1. Risk assessment (is stockout imminent?)
2. Revenue impact estimation
3. Root cause analysis (why below minimum?)
4. Immediate action recommendations (emergency order? supplier switch?)
5. Long-term prevention (buffer recalculation?)

Use DDMRP best practices from the provided knowledge.
	`, productID, currentStock, dailyConsumption, daysUntilStockout, minBuffer)

	aiRequest := &providers.AIAnalysisRequest{
		Event:          event,
		Context:        analysisContext,
		DDMRPKnowledge: knowledge,
		Prompt:         prompt,
	}

	aiResponse, err := uc.aiAnalyzer.Analyze(ctx, aiRequest)
	if err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("AI analysis failed: %v", err))
	}

	orgID, err := uuid.Parse(event.OrganizationID)
	if err != nil {
		return errors.NewBadRequest("invalid organization_id")
	}

	userID := uuid.New()

	priority := uc.determinePriority(daysUntilStockout)

	notification := domain.NewNotification(
		orgID,
		userID,
		domain.NotificationTypeAlert,
		priority,
		fmt.Sprintf("Stockout Risk: %s", productID),
		aiResponse.Summary,
	)

	notification.FullAnalysis = aiResponse.FullAnalysis
	notification.Reasoning = aiResponse.Reasoning

	var timeToImpact *time.Duration
	if aiResponse.ImpactAssessment.TimeToImpactHours > 0 {
		duration := time.Duration(aiResponse.ImpactAssessment.TimeToImpactHours) * time.Hour
		timeToImpact = &duration
	}

	notification.Impact = domain.ImpactAssessment{
		RiskLevel:        aiResponse.ImpactAssessment.RiskLevel,
		RevenueImpact:    aiResponse.ImpactAssessment.RevenueImpact,
		CostImpact:       aiResponse.ImpactAssessment.CostImpact,
		TimeToImpact:     timeToImpact,
		AffectedOrders:   aiResponse.ImpactAssessment.AffectedOrders,
		AffectedProducts: aiResponse.ImpactAssessment.AffectedProducts,
	}

	for i, aiRec := range aiResponse.Recommendations {
		notification.Recommendations = append(notification.Recommendations, domain.Recommendation{
			Action:          aiRec.Action,
			Reasoning:       aiRec.Reasoning,
			ExpectedOutcome: aiRec.ExpectedOutcome,
			Effort:          aiRec.Effort,
			Impact:          aiRec.Impact,
			PriorityOrder:   i + 1,
		})
	}

	notification.SourceEvents = []string{event.ID}
	notification.RelatedEntities = map[string][]string{
		"product_ids": {productID},
	}

	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return errors.NewInternalServerError(fmt.Sprintf("failed to create notification: %v", err))
	}

	uc.logger.Info(ctx, "Stockout risk notification created", logger.Tags{
		"notification_id": notification.ID.String(),
		"product_id":      productID,
		"priority":        string(priority),
		"risk_level":      aiResponse.ImpactAssessment.RiskLevel,
	})

	return nil
}

func (uc *AnalyzeStockoutRiskUseCase) determinePriority(daysUntilStockout float64) domain.NotificationPriority {
	switch {
	case daysUntilStockout < 2:
		return domain.NotificationPriorityCritical
	case daysUntilStockout < 5:
		return domain.NotificationPriorityHigh
	case daysUntilStockout < 10:
		return domain.NotificationPriorityMedium
	default:
		return domain.NotificationPriorityLow
	}
}
