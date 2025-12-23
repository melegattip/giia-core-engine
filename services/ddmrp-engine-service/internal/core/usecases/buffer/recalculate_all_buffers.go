package buffer

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/google/uuid"
)

type RecalculateAllBuffersUseCase struct {
	bufferRepo          *CalculateBufferUseCase
	listBuffersUseCase  *ListBuffersUseCase
}

func NewRecalculateAllBuffersUseCase(
	calculateUseCase *CalculateBufferUseCase,
	listUseCase *ListBuffersUseCase,
) *RecalculateAllBuffersUseCase {
	return &RecalculateAllBuffersUseCase{
		bufferRepo:          calculateUseCase,
		listBuffersUseCase:  listUseCase,
	}
}

type RecalculateAllBuffersInput struct {
	OrganizationID uuid.UUID
}

func (uc *RecalculateAllBuffersUseCase) Execute(ctx context.Context, input RecalculateAllBuffersInput) error {
	if input.OrganizationID == uuid.Nil {
		return errors.NewBadRequest("organization_id is required")
	}

	buffers, err := uc.listBuffersUseCase.Execute(ctx, ListBuffersInput{
		OrganizationID: input.OrganizationID,
	})
	if err != nil {
		return errors.NewInternalServerError("failed to list buffers for recalculation")
	}

	for _, buffer := range buffers {
		_, err := uc.bufferRepo.Execute(ctx, CalculateBufferInput{
			ProductID:      buffer.ProductID,
			OrganizationID: buffer.OrganizationID,
		})
		if err != nil {
			continue
		}
	}

	return nil
}
