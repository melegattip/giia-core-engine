package event_processing

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/events"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type BufferEventHandlerImpl struct {
	stockoutAnalyzer StockoutRiskAnalyzer
	logger           logger.Logger
}

type StockoutRiskAnalyzer interface {
	Execute(ctx context.Context, event *events.Event) error
}

func NewBufferEventHandler(
	stockoutAnalyzer StockoutRiskAnalyzer,
	logger logger.Logger,
) providers.BufferEventHandler {
	return &BufferEventHandlerImpl{
		stockoutAnalyzer: stockoutAnalyzer,
		logger:           logger,
	}
}

func (h *BufferEventHandlerImpl) Handle(ctx context.Context, event *events.Event) error {
	h.logger.Info(ctx, "Processing buffer event", logger.Tags{
		"event_type": event.Type,
		"event_id":   event.ID,
	})

	if event.Type == "buffer.below_minimum" {
		return h.stockoutAnalyzer.Execute(ctx, event)
	}

	return nil
}
