package adu

import (
	"context"
	"math"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
)

type CalculateADUUseCase struct {
	aduRepo providers.ADURepository
}

func NewCalculateADUUseCase(aduRepo providers.ADURepository) *CalculateADUUseCase {
	return &CalculateADUUseCase{
		aduRepo: aduRepo,
	}
}

type CalculateADUInput struct {
	ProductID      uuid.UUID
	OrganizationID uuid.UUID
	DemandData     []float64
	Method         domain.ADUMethod
	PeriodDays     int
	Alpha          float64
}

func (uc *CalculateADUUseCase) Execute(ctx context.Context, input CalculateADUInput) (*domain.ADUCalculation, error) {
	if input.ProductID == uuid.Nil {
		return nil, errors.NewBadRequest("product_id is required")
	}
	if input.OrganizationID == uuid.Nil {
		return nil, errors.NewBadRequest("organization_id is required")
	}
	if len(input.DemandData) == 0 {
		return nil, errors.NewBadRequest("demand data is required")
	}
	if !input.Method.IsValid() {
		return nil, errors.NewBadRequest("invalid ADU calculation method")
	}

	var aduValue float64
	switch input.Method {
	case domain.ADUMethodAverage:
		aduValue = uc.calculateSimpleAverage(input.DemandData)
	case domain.ADUMethodExponential:
		aduValue = uc.calculateExponentialSmoothing(input.DemandData, input.Alpha)
	case domain.ADUMethodWeighted:
		aduValue = uc.calculateWeightedMovingAverage(input.DemandData)
	default:
		return nil, errors.NewBadRequest("unsupported ADU calculation method")
	}

	adu := &domain.ADUCalculation{
		ID:              uuid.New(),
		ProductID:       input.ProductID,
		OrganizationID:  input.OrganizationID,
		CalculationDate: time.Now(),
		ADUValue:        aduValue,
		Method:          input.Method,
		PeriodDays:      input.PeriodDays,
		CreatedAt:       time.Now(),
	}

	if err := uc.aduRepo.Create(ctx, adu); err != nil {
		return nil, errors.NewInternalServerError("failed to save ADU calculation")
	}

	return adu, nil
}

func (uc *CalculateADUUseCase) calculateSimpleAverage(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range data {
		sum += value
	}

	return sum / float64(len(data))
}

func (uc *CalculateADUUseCase) calculateExponentialSmoothing(data []float64, alpha float64) float64 {
	if len(data) == 0 {
		return 0
	}

	if alpha <= 0 || alpha > 1 {
		alpha = 0.3
	}

	smoothed := data[0]
	for i := 1; i < len(data); i++ {
		smoothed = alpha*data[i] + (1-alpha)*smoothed
	}

	return smoothed
}

func (uc *CalculateADUUseCase) calculateWeightedMovingAverage(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	n := len(data)
	totalWeight := float64(n * (n + 1) / 2)
	weightedSum := 0.0

	for i := 0; i < n; i++ {
		weight := float64(i + 1)
		weightedSum += data[i] * weight
	}

	return math.Round(weightedSum/totalWeight*100) / 100
}
