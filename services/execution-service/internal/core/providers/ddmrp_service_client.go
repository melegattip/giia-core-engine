package providers

import (
	"context"

	"github.com/google/uuid"
)

type BufferStatus struct {
	ProductID       uuid.UUID
	Zone            string
	NetFlowPosition float64
	TopOfGreen      float64
	AlertLevel      string
}

type DDMRPServiceClient interface {
	GetBufferStatus(ctx context.Context, organizationID, productID uuid.UUID) (*BufferStatus, error)
	UpdateNetFlowPosition(ctx context.Context, organizationID, productID uuid.UUID) error
	GetProductsInRedZone(ctx context.Context, organizationID uuid.UUID) ([]*BufferStatus, error)
}