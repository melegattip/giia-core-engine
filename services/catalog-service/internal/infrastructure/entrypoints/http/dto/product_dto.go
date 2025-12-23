package dto

import (
	"time"

	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/google/uuid"
)

type CreateProductRequest struct {
	SKU             string     `json:"sku" validate:"required,max=100"`
	Name            string     `json:"name" validate:"required,max=255"`
	Description     string     `json:"description"`
	Category        string     `json:"category,omitempty" validate:"max=100"`
	UnitOfMeasure   string     `json:"unit_of_measure" validate:"required,max=50"`
	BufferProfileID *uuid.UUID `json:"buffer_profile_id,omitempty"`
}

type UpdateProductRequest struct {
	Name            string     `json:"name,omitempty" validate:"max=255"`
	Description     string     `json:"description"`
	Category        string     `json:"category,omitempty" validate:"max=100"`
	UnitOfMeasure   string     `json:"unit_of_measure,omitempty" validate:"max=50"`
	Status          string     `json:"status,omitempty"`
	BufferProfileID *uuid.UUID `json:"buffer_profile_id,omitempty"`
}

type ProductResponse struct {
	ID              uuid.UUID         `json:"id"`
	SKU             string            `json:"sku"`
	Name            string            `json:"name"`
	Description     string            `json:"description,omitempty"`
	Category        string            `json:"category,omitempty"`
	UnitOfMeasure   string            `json:"unit_of_measure"`
	Status          string            `json:"status"`
	OrganizationID  uuid.UUID         `json:"organization_id"`
	BufferProfileID *uuid.UUID        `json:"buffer_profile_id,omitempty"`
	Suppliers       []SupplierSummary `json:"suppliers,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type SupplierSummary struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	LeadTimeDays int       `json:"lead_time_days"`
	UnitCost     *float64  `json:"unit_cost,omitempty"`
	IsPrimary    bool      `json:"is_primary"`
}

type PaginatedProductsResponse struct {
	Products   []ProductResponse `json:"products"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalCount int64             `json:"total_count"`
	TotalPages int               `json:"total_pages"`
}

func ToProductResponse(product *domain.Product) ProductResponse {
	response := ProductResponse{
		ID:              product.ID,
		SKU:             product.SKU,
		Name:            product.Name,
		Description:     product.Description,
		Category:        product.Category,
		UnitOfMeasure:   product.UnitOfMeasure,
		Status:          string(product.Status),
		OrganizationID:  product.OrganizationID,
		BufferProfileID: product.BufferProfileID,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}

	if len(product.Suppliers) > 0 {
		response.Suppliers = make([]SupplierSummary, len(product.Suppliers))
		for i, ps := range product.Suppliers {
			response.Suppliers[i] = SupplierSummary{
				ID:           ps.SupplierID,
				Code:         ps.Supplier.Code,
				Name:         ps.Supplier.Name,
				LeadTimeDays: ps.LeadTimeDays,
				UnitCost:     ps.UnitCost,
				IsPrimary:    ps.IsPrimarySupplier,
			}
		}
	}

	return response
}

func ToProductListResponse(products []*domain.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = ToProductResponse(product)
	}
	return responses
}
