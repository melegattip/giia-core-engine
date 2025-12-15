package product

import (
	"context"

	"github.com/giia/giia-core-engine/pkg/logger"
	"github.com/giia/giia-core-engine/services/catalog-service/internal/core/providers"
)

type SearchProductsRequest struct {
	Query    string `json:"query"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Category string `json:"category,omitempty"`
	Status   string `json:"status,omitempty"`
}

type SearchProductsUseCase struct {
	productRepo providers.ProductRepository
	logger      logger.Logger
}

func NewSearchProductsUseCase(
	productRepo providers.ProductRepository,
	logger logger.Logger,
) *SearchProductsUseCase {
	return &SearchProductsUseCase{
		productRepo: productRepo,
		logger:      logger,
	}
}

func (uc *SearchProductsUseCase) Execute(ctx context.Context, req *SearchProductsRequest) (*PaginatedProductsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	filters := make(map[string]interface{})
	if req.Category != "" {
		filters["category"] = req.Category
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	products, total, err := uc.productRepo.Search(ctx, req.Query, filters, req.Page, req.PageSize)
	if err != nil {
		uc.logger.Error(ctx, err, "Failed to search products", logger.Tags{
			"query":     req.Query,
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
