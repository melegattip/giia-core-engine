package sales_order

import (
	"context"
	"time"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/google/uuid"
)

type CreateSOUseCase struct {
	soRepo         providers.SalesOrderRepository
	catalogClient  providers.CatalogServiceClient
	eventPublisher providers.EventPublisher
}

func NewCreateSOUseCase(
	soRepo providers.SalesOrderRepository,
	catalogClient providers.CatalogServiceClient,
	publisher providers.EventPublisher,
) *CreateSOUseCase {
	return &CreateSOUseCase{
		soRepo:         soRepo,
		catalogClient:  catalogClient,
		eventPublisher: publisher,
	}
}

type CreateSOInput struct {
	OrganizationID uuid.UUID
	SONumber       string
	CustomerID     uuid.UUID
	OrderDate      time.Time
	DueDate        time.Time
	LineItems      []domain.SOLineItem
}

func (uc *CreateSOUseCase) Execute(ctx context.Context, input *CreateSOInput) (*domain.SalesOrder, error) {
	if input == nil {
		return nil, domain.NewValidationError("input cannot be nil")
	}

	existingSO, _ := uc.soRepo.GetBySONumber(ctx, input.SONumber, input.OrganizationID)
	if existingSO != nil {
		return nil, domain.NewConflictError("sales order with this SO number already exists")
	}

	for _, item := range input.LineItems {
		_, err := uc.catalogClient.GetProduct(ctx, item.ProductID)
		if err != nil {
			return nil, domain.NewValidationError("invalid product_id: " + item.ProductID.String())
		}
	}

	so, err := domain.NewSalesOrder(
		input.OrganizationID,
		input.CustomerID,
		input.SONumber,
		input.OrderDate,
		input.DueDate,
		input.LineItems,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.soRepo.Create(ctx, so); err != nil {
		return nil, err
	}

	uc.eventPublisher.PublishSOCreated(ctx, so)

	return so, nil
}