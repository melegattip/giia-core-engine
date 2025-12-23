package catalog

import (
	"context"

	"github.com/giia/giia-core-engine/services/ddmrp-engine-service/internal/core/providers"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type CatalogGRPCClient struct {
	conn   *grpc.ClientConn
	client interface{}
}

func NewCatalogGRPCClient(catalogURL string) (*CatalogGRPCClient, error) {
	conn, err := grpc.Dial(catalogURL, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &CatalogGRPCClient{
		conn: conn,
	}, nil
}

func (c *CatalogGRPCClient) GetProduct(ctx context.Context, productID uuid.UUID) (*providers.Product, error) {
	return &providers.Product{
		ID:              productID,
		OrganizationID:  uuid.New(),
		SKU:             "MOCK-SKU",
		Name:            "Mock Product",
		BufferProfileID: func() *uuid.UUID { id := uuid.New(); return &id }(),
		Status:          "active",
	}, nil
}

func (c *CatalogGRPCClient) GetBufferProfile(ctx context.Context, bufferProfileID uuid.UUID) (*providers.BufferProfile, error) {
	return &providers.BufferProfile{
		ID:                bufferProfileID,
		Name:              "Mock Profile",
		LeadTimeFactor:    0.5,
		VariabilityFactor: 0.5,
		OrderFrequency:    7,
		Status:            "active",
	}, nil
}

func (c *CatalogGRPCClient) GetSupplier(ctx context.Context, supplierID uuid.UUID) (*providers.Supplier, error) {
	return &providers.Supplier{
		ID:     supplierID,
		Code:   "MOCK-SUPPLIER",
		Name:   "Mock Supplier",
		Status: "active",
	}, nil
}

func (c *CatalogGRPCClient) GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]providers.ProductSupplier, error) {
	return []providers.ProductSupplier{
		{
			ProductID:    productID,
			SupplierID:   uuid.New(),
			LeadTimeDays: 30,
			IsPrimary:    true,
			MOQ:          100,
		},
	}, nil
}

func (c *CatalogGRPCClient) GetPrimarySupplier(ctx context.Context, productID uuid.UUID) (*providers.ProductSupplier, error) {
	return &providers.ProductSupplier{
		ProductID:    productID,
		SupplierID:   uuid.New(),
		LeadTimeDays: 30,
		IsPrimary:    true,
		MOQ:          100,
	}, nil
}

func (c *CatalogGRPCClient) Close() error {
	return c.conn.Close()
}
