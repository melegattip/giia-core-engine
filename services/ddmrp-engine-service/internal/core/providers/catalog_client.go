package providers

import (
	"context"

	"github.com/google/uuid"
)

type Product struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	SKU             string
	Name            string
	Description     string
	Category        string
	UnitOfMeasure   string
	BufferProfileID *uuid.UUID
	Status          string
}

type BufferProfile struct {
	ID                uuid.UUID
	OrganizationID    uuid.UUID
	Name              string
	Description       string
	ADUMethod         string
	LeadTimeFactor    float64
	VariabilityFactor float64
	OrderFrequency    int
	Status            string
}

type Supplier struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Code           string
	Name           string
	ContactName    string
	ContactEmail   string
	Status         string
}

type ProductSupplier struct {
	ProductID      uuid.UUID
	SupplierID     uuid.UUID
	LeadTimeDays   int
	IsPrimary      bool
	MOQ            int
}

type CatalogServiceClient interface {
	GetProduct(ctx context.Context, productID uuid.UUID) (*Product, error)
	GetBufferProfile(ctx context.Context, bufferProfileID uuid.UUID) (*BufferProfile, error)
	GetSupplier(ctx context.Context, supplierID uuid.UUID) (*Supplier, error)
	GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]ProductSupplier, error)
	GetPrimarySupplier(ctx context.Context, productID uuid.UUID) (*ProductSupplier, error)
}
