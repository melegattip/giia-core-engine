// Package grpc provides gRPC handlers for the Execution Service.
package grpc

import (
	"context"

	"github.com/google/uuid"

	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/usecases/inventory"
)

// ExecutionService implements the gRPC ExecutionService interface.
type ExecutionService struct {
	inventoryBalanceRepo providers.InventoryBalanceRepository
	inventoryTxnRepo     providers.InventoryTransactionRepository
	purchaseOrderRepo    providers.PurchaseOrderRepository
	salesOrderRepo       providers.SalesOrderRepository
	recordTxnUseCase     *inventory.RecordTransactionUseCase
}

// NewExecutionService creates a new ExecutionService.
func NewExecutionService(
	balanceRepo providers.InventoryBalanceRepository,
	txnRepo providers.InventoryTransactionRepository,
	poRepo providers.PurchaseOrderRepository,
	soRepo providers.SalesOrderRepository,
	recordTxnUC *inventory.RecordTransactionUseCase,
) *ExecutionService {
	return &ExecutionService{
		inventoryBalanceRepo: balanceRepo,
		inventoryTxnRepo:     txnRepo,
		purchaseOrderRepo:    poRepo,
		salesOrderRepo:       soRepo,
		recordTxnUseCase:     recordTxnUC,
	}
}

// BalanceResponse represents the response for GetInventoryBalance.
type BalanceResponse struct {
	Balances       []*InventoryBalanceData
	TotalOnHand    float64
	TotalAvailable float64
}

// InventoryBalanceData represents inventory balance data in gRPC response.
type InventoryBalanceData struct {
	ID             string
	OrganizationID string
	ProductID      string
	LocationID     string
	OnHand         float64
	Reserved       float64
	Available      float64
	UpdatedAt      string
}

// GetInventoryBalance returns inventory balances for a product.
func (s *ExecutionService) GetInventoryBalance(ctx context.Context, orgID, productID uuid.UUID, locationID *uuid.UUID) (*BalanceResponse, error) {
	balances, err := s.inventoryBalanceRepo.GetByProduct(ctx, orgID, productID)
	if err != nil {
		return nil, err
	}

	var totalOnHand, totalAvailable float64
	balanceData := make([]*InventoryBalanceData, 0, len(balances))

	for _, b := range balances {
		if locationID != nil && b.LocationID != *locationID {
			continue
		}
		totalOnHand += b.OnHand
		totalAvailable += b.Available
		balanceData = append(balanceData, &InventoryBalanceData{
			ID:             b.ID.String(),
			OrganizationID: b.OrganizationID.String(),
			ProductID:      b.ProductID.String(),
			LocationID:     b.LocationID.String(),
			OnHand:         b.OnHand,
			Reserved:       b.Reserved,
			Available:      b.Available,
			UpdatedAt:      b.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &BalanceResponse{
		Balances:       balanceData,
		TotalOnHand:    totalOnHand,
		TotalAvailable: totalAvailable,
	}, nil
}

// PendingPOResponse represents the response for GetPendingPurchaseOrders.
type PendingPOResponse struct {
	Orders               []*PendingPurchaseOrder
	TotalPendingQuantity float64
}

// PendingPurchaseOrder represents a pending PO with relevant details.
type PendingPurchaseOrder struct {
	ID                  string
	PONumber            string
	Status              string
	ExpectedArrivalDate string
	PendingQuantity     float64
}

// GetPendingPurchaseOrders returns pending POs for a product.
func (s *ExecutionService) GetPendingPurchaseOrders(ctx context.Context, orgID, productID uuid.UUID) (*PendingPOResponse, error) {
	// Get all confirmed or partial POs
	filters := map[string]interface{}{
		"status": []domain.POStatus{domain.POStatusConfirmed, domain.POStatusPartial},
	}

	orders, _, err := s.purchaseOrderRepo.List(ctx, orgID, filters, 1, 1000)
	if err != nil {
		return nil, err
	}

	var totalPending float64
	pendingOrders := make([]*PendingPurchaseOrder, 0)

	for _, po := range orders {
		for _, lineItem := range po.LineItems {
			if lineItem.ProductID == productID {
				pending := lineItem.Quantity - lineItem.ReceivedQty
				if pending > 0 {
					totalPending += pending
					pendingOrders = append(pendingOrders, &PendingPurchaseOrder{
						ID:                  po.ID.String(),
						PONumber:            po.PONumber,
						Status:              string(po.Status),
						ExpectedArrivalDate: po.ExpectedArrivalDate.Format("2006-01-02T15:04:05Z"),
						PendingQuantity:     pending,
					})
				}
			}
		}
	}

	return &PendingPOResponse{
		Orders:               pendingOrders,
		TotalPendingQuantity: totalPending,
	}, nil
}

// OnOrderResponse represents the response for GetPendingOnOrder.
type OnOrderResponse struct {
	OnOrder        float64
	PendingPOCount int32
}

// GetPendingOnOrder returns total quantity on order for a product.
func (s *ExecutionService) GetPendingOnOrder(ctx context.Context, orgID, productID uuid.UUID) (*OnOrderResponse, error) {
	result, err := s.GetPendingPurchaseOrders(ctx, orgID, productID)
	if err != nil {
		return nil, err
	}

	return &OnOrderResponse{
		OnOrder:        result.TotalPendingQuantity,
		PendingPOCount: int32(len(result.Orders)),
	}, nil
}

// DemandResponse represents the response for GetQualifiedDemand.
type DemandResponse struct {
	QualifiedDemand  float64
	ConfirmedSOCount int32
}

// GetQualifiedDemand returns qualified demand for a product.
func (s *ExecutionService) GetQualifiedDemand(ctx context.Context, orgID, productID uuid.UUID) (*DemandResponse, error) {
	demand, err := s.salesOrderRepo.GetQualifiedDemand(ctx, orgID, productID)
	if err != nil {
		return nil, err
	}

	// Get count of confirmed SOs
	filters := map[string]interface{}{
		"status": domain.SOStatusConfirmed,
	}
	orders, _, err := s.salesOrderRepo.List(ctx, orgID, filters, 1, 1000)
	if err != nil {
		return nil, err
	}

	var confirmedCount int32
	for _, so := range orders {
		for _, item := range so.LineItems {
			if item.ProductID == productID {
				confirmedCount++
				break
			}
		}
	}

	return &DemandResponse{
		QualifiedDemand:  demand,
		ConfirmedSOCount: confirmedCount,
	}, nil
}

// TransactionResponse represents the response for RecordTransaction.
type TransactionResponse struct {
	Transaction    *domain.InventoryTransaction
	UpdatedBalance *domain.InventoryBalance
}

// RecordTransactionInput represents input for recording a transaction.
type RecordTransactionInput struct {
	OrganizationID uuid.UUID
	ProductID      uuid.UUID
	LocationID     uuid.UUID
	Type           domain.TransactionType
	Quantity       float64
	UnitCost       float64
	ReferenceType  string
	ReferenceID    uuid.UUID
	Reason         string
	CreatedBy      uuid.UUID
}

// RecordTransaction records an inventory transaction.
func (s *ExecutionService) RecordTransaction(ctx context.Context, input *RecordTransactionInput) (*TransactionResponse, error) {
	ucInput := &inventory.RecordTransactionInput{
		OrganizationID: input.OrganizationID,
		ProductID:      input.ProductID,
		LocationID:     input.LocationID,
		Type:           input.Type,
		Quantity:       input.Quantity,
		UnitCost:       input.UnitCost,
		ReferenceType:  input.ReferenceType,
		ReferenceID:    input.ReferenceID,
		Reason:         input.Reason,
		CreatedBy:      input.CreatedBy,
	}

	txn, err := s.recordTxnUseCase.Execute(ctx, ucInput)
	if err != nil {
		return nil, err
	}

	balance, err := s.inventoryBalanceRepo.GetOrCreate(ctx, input.OrganizationID, input.ProductID, input.LocationID)
	if err != nil {
		return nil, err
	}

	return &TransactionResponse{
		Transaction:    txn,
		UpdatedBalance: balance,
	}, nil
}
