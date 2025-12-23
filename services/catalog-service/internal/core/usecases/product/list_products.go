package product

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
)

type ListProductsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Category string `json:"category,omitempty"`
	Status   string `json:"status,omitempty"`
}

type PaginatedProductsResponse struct {
	Products   []*domain.Product `json:"products"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalCount int64             `json:"total_count"`
	TotalPages int               `json:"total_pages"`
}

type ListProductsUseCase struct {
	productRepo providers.ProductRepository
	logger      logger.Logger
}

func NewListProductsUseCase(
	productRepo providers.ProductRepository,
	logger logger.Logger,
) *ListProductsUseCase {
	return &ListProductsUseCase{
		productRepo: productRepo,
		logger:      logger,
	}
}

func (uc *ListProductsUseCase) Execute(ctx context.Context, req *ListProductsRequest) (*PaginatedProductsResponse, error) {
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
	if req.Category != "" {
		filters["category"] = req.Category
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	products, total, err := uc.productRepo.List(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to list products", logger.Tags{
			"page":      req.Page,
			"page_size": req.PageSize,
		})
		return nil, err
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &PaginatedProductsResponse{
		Products:   products,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
	}, nil
}
