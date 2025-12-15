package domain

import (
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

type Product struct {
	ID              uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SKU             string            `json:"sku" gorm:"type:varchar(100);not null;index:idx_products_sku"`
	Name            string            `json:"name" gorm:"type:varchar(255);not null"`
	Description     string            `json:"description" gorm:"type:text"`
	Category        string            `json:"category" gorm:"type:varchar(100);index:idx_products_category"`
	UnitOfMeasure   string            `json:"unit_of_measure" gorm:"type:varchar(50);not null"`
	Status          ProductStatus     `json:"status" gorm:"type:varchar(20);not null;default:'active';index:idx_products_status"`
	OrganizationID  uuid.UUID         `json:"organization_id" gorm:"type:uuid;not null;index:idx_products_organization_id"`
	BufferProfileID *uuid.UUID        `json:"buffer_profile_id,omitempty" gorm:"type:uuid"`
	Suppliers       []ProductSupplier `json:"suppliers,omitempty" gorm:"foreignKey:ProductID"`
	CreatedAt       time.Time         `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time         `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (Product) TableName() string {
	return "products"
}

func (p *Product) Validate() error {
	if p.SKU == "" {
		return errors.NewBadRequest("SKU is required")
	}
	if len(p.SKU) > 100 {
		return errors.NewBadRequest("SKU must be 100 characters or less")
	}
	if p.Name == "" {
		return errors.NewBadRequest("product name is required")
	}
	if len(p.Name) > 255 {
		return errors.NewBadRequest("product name must be 255 characters or less")
	}
	if p.UnitOfMeasure == "" {
		return errors.NewBadRequest("unit of measure is required")
	}
	if len(p.UnitOfMeasure) > 50 {
		return errors.NewBadRequest("unit of measure must be 50 characters or less")
	}
	if p.OrganizationID == uuid.Nil {
		return errors.NewBadRequest("organization ID is required")
	}
	if !p.IsValidStatus() {
		return errors.NewBadRequest("invalid product status")
	}
	return nil
}

func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

func (p *Product) IsValidStatus() bool {
	switch p.Status {
	case ProductStatusActive, ProductStatusInactive, ProductStatusDiscontinued:
		return true
	default:
		return false
	}
}

func (p *Product) Deactivate() {
	p.Status = ProductStatusInactive
}

func (p *Product) Activate() {
	p.Status = ProductStatusActive
}

func (p *Product) Discontinue() {
	p.Status = ProductStatusDiscontinued
}
