package domain

import (
	"time"

	"github.com/giia/giia-core-engine/pkg/errors"
	"github.com/google/uuid"
)

type ProductSupplier struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID         uuid.UUID `json:"product_id" gorm:"type:uuid;not null;index:idx_product_suppliers_product_id"`
	SupplierID        uuid.UUID `json:"supplier_id" gorm:"type:uuid;not null;index:idx_product_suppliers_supplier_id"`
	LeadTimeDays      int       `json:"lead_time_days" gorm:"not null"`
	UnitCost          *float64  `json:"unit_cost,omitempty" gorm:"type:decimal(12,2)"`
	IsPrimarySupplier bool      `json:"is_primary_supplier" gorm:"not null;default:false"`
	Supplier          *Supplier `json:"supplier,omitempty" gorm:"foreignKey:SupplierID"`
	Product           *Product  `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	CreatedAt         time.Time `json:"created_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ProductSupplier) TableName() string {
	return "product_suppliers"
}

func (ps *ProductSupplier) Validate() error {
	if ps.ProductID == uuid.Nil {
		return errors.NewBadRequest("product ID is required")
	}
	if ps.SupplierID == uuid.Nil {
		return errors.NewBadRequest("supplier ID is required")
	}
	if ps.LeadTimeDays < 0 {
		return errors.NewBadRequest("lead time days cannot be negative")
	}
	if ps.UnitCost != nil && *ps.UnitCost < 0 {
		return errors.NewBadRequest("unit cost cannot be negative")
	}
	return nil
}
