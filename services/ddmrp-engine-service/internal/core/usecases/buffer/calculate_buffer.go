package buffer

import (
	"context"
	"math"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type CalculateBufferUseCase struct {
	bufferRepo           providers.BufferRepository
	demandAdjustmentRepo providers.DemandAdjustmentRepository
	bufferAdjustmentRepo providers.BufferAdjustmentRepository
	bufferHistoryRepo    providers.BufferHistoryRepository
	catalogClient        providers.CatalogServiceClient
	aduRepo              providers.ADURepository
	eventPublisher       providers.EventPublisher
}

func NewCalculateBufferUseCase(
	bufferRepo providers.BufferRepository,
	demandAdjustmentRepo providers.DemandAdjustmentRepository,
	bufferAdjustmentRepo providers.BufferAdjustmentRepository,
	bufferHistoryRepo providers.BufferHistoryRepository,
	catalogClient providers.CatalogServiceClient,
	aduRepo providers.ADURepository,
	eventPublisher providers.EventPublisher,
) *CalculateBufferUseCase {
	return &CalculateBufferUseCase{
		bufferRepo:           bufferRepo,
		demandAdjustmentRepo: demandAdjustmentRepo,
		bufferAdjustmentRepo: bufferAdjustmentRepo,
		bufferHistoryRepo:    bufferHistoryRepo,
		catalogClient:        catalogClient,
		aduRepo:              aduRepo,
		eventPublisher:       eventPublisher,
	}
}

type CalculateBufferInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
}

func (uc *CalculateBufferUseCase) Execute(ctx context.Context, input CalculateBufferInput) (*domain.Buffer, error) {
	if input.ProductID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}

	product, err := uc.catalogClient.GetProduct(ctx, input.ProductID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get product from catalog")
	}

	if product.BufferProfileID == nil {
		return nil, errors.NewBadRequest("product has no buffer profile assigned")
	}

	bufferProfile, err := uc.catalogClient.GetBufferProfile(ctx, *product.BufferProfileID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get buffer profile")
	}

	adu, err := uc.aduRepo.GetLatest(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get latest ADU")
	}

	baseCPD := math.Ceil(adu.ADUValue)

	today := time.Now()
	activeFADs, err := uc.demandAdjustmentRepo.GetActiveForDate(ctx, input.ProductID, input.OrganizationID, today)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get active FADs")
	}

	adjustedCPD := domain.ApplyAdjustedCPD(baseCPD, activeFADs)

	primarySupplier, err := uc.catalogClient.GetPrimarySupplier(ctx, input.ProductID)
	if err != nil {
		return nil, errors.NewInternalServerError("failed to get primary supplier")
	}

	moq := primarySupplier.MOQ
	orderFrequency := bufferProfile.OrderFrequency
	ltd := primarySupplier.LeadTimeDays

	redBase, redSafe, redZone, yellowZone, greenZone := domain.CalculateBufferZones(
		adjustedCPD,
		ltd,
		bufferProfile.LeadTimeFactor,
		bufferProfile.VariabilityFactor,
		moq,
		orderFrequency,
	)

	buffer, err := uc.bufferRepo.GetByProduct(ctx, input.ProductID, input.OrganizationID)
	if err != nil {
		buffer = &domain.Buffer{
			ID:              uuid.New(),
			ProductID:       input.ProductID,
			OrganizationID:  input.OrganizationID,
			BufferProfileID: *product.BufferProfileID,
			CreatedAt:       time.Now(),
		}
	}

	oldZone := buffer.Zone

	buffer.CPD = adjustedCPD
	buffer.LTD = ltd
	buffer.RedBase = redBase
	buffer.RedSafe = redSafe
	buffer.RedZone = redZone
	buffer.YellowZone = yellowZone
	buffer.GreenZone = greenZone
	buffer.LastRecalculatedAt = time.Now()
	buffer.UpdatedAt = time.Now()

	hasAdjustments := len(activeFADs) > 0

	activeBufferAdjs, err := uc.bufferAdjustmentRepo.GetActiveForDate(ctx, buffer.ID, today)
	if err == nil && len(activeBufferAdjs) > 0 {
		hasAdjustments = true
		for _, adj := range activeBufferAdjs {
			switch adj.TargetZone {
			case domain.ZoneRed:
				buffer.RedZone *= adj.Factor
			case domain.ZoneYellow:
				buffer.YellowZone *= adj.Factor
			case domain.ZoneGreen:
				buffer.GreenZone *= adj.Factor
			case domain.ZoneAll:
				buffer.RedZone *= adj.Factor
				buffer.YellowZone *= adj.Factor
				buffer.GreenZone *= adj.Factor
			}
		}
	}

	buffer.DetermineZone()

	if err := uc.bufferRepo.Save(ctx, buffer); err != nil {
		return nil, errors.NewInternalServerError("failed to save buffer")
	}

	moqPtr := &moq
	orderFreqPtr := &orderFrequency
	history := domain.NewBufferHistory(
		buffer,
		bufferProfile.LeadTimeFactor,
		bufferProfile.VariabilityFactor,
		moqPtr,
		orderFreqPtr,
		hasAdjustments,
	)

	if err := uc.bufferHistoryRepo.Create(ctx, history); err != nil {
		return nil, errors.NewInternalServerError("failed to create buffer history")
	}

	if err := uc.eventPublisher.PublishBufferCalculated(ctx, buffer); err != nil {
		return nil, errors.NewInternalServerError("failed to publish buffer calculated event")
	}

	if oldZone != buffer.Zone {
		if err := uc.eventPublisher.PublishBufferStatusChanged(ctx, buffer, oldZone); err != nil {
			return nil, errors.NewInternalServerError("failed to publish status changed event")
		}
	}

	if buffer.AlertLevel == domain.AlertReplenish || buffer.AlertLevel == domain.AlertCritical {
		if err := uc.eventPublisher.PublishBufferAlertTriggered(ctx, buffer); err != nil {
			return nil, errors.NewInternalServerError("failed to publish alert event")
		}
	}

	return buffer, nil
}
