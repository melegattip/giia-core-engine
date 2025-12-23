package supplier

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
)

type ListSuppliersRequest struct {
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page" validate:"gte=1"`
	PageSize int    `json:"page_size" validate:"gte=1,lte=100"`
}

type ListSuppliersResponse struct {
	Suppliers  []*domain.Supplier `json:"suppliers"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type ListSuppliersUseCase struct {
	supplierRepo providers.SupplierRepository
	logger       logger.Logger
}

func NewListSuppliersUseCase(
	supplierRepo providers.SupplierRepository,
	logger logger.Logger,
) *ListSuppliersUseCase {
	return &ListSuppliersUseCase{
		supplierRepo: supplierRepo,
		logger:       logger,
	}
}

func (uc *ListSuppliersUseCase) Execute(ctx context.Context, req *ListSuppliersRequest) (*ListSuppliersResponse, error) {
	if req == nil {
		return nil, errors.NewBadRequest("request cannot be nil")
	}

	orgID, ok := ctx.Value("organization_id").(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return nil, errors.NewBadRequest("organization ID is required in context")
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filters := make(map[string]interface{})
	if req.Status != "" {
		filters["status"] = req.Status
	}

	suppliers, total, err := uc.supplierRepo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to list suppliers", nil)
		return nil, err
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	return &ListSuppliersResponse{
		Suppliers:  suppliers,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
