package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"gorm.io/gorm"
)

// PurchaseOrderModel represents the database model for purchase orders
type PurchaseOrderModel struct {
	ID                  uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID      uuid.UUID         `gorm:"type:uuid;not null;index:idx_po_org"`
	PONumber            string            `gorm:"type:varchar(50);not null;index:idx_po_number"`
	SupplierID          uuid.UUID         `gorm:"type:uuid;not null;index:idx_po_supplier"`
	Status              string            `gorm:"type:varchar(20);not null;default:'draft'"`
	OrderDate           time.Time         `gorm:"not null"`
	ExpectedArrivalDate time.Time         `gorm:"not null"`
	ActualArrivalDate   *time.Time        `gorm:""`
	DelayDays           int               `gorm:"default:0"`
	IsDelayed           bool              `gorm:"default:false;index:idx_po_delayed"`
	TotalAmount         float64           `gorm:"type:decimal(15,2);default:0"`
	CreatedBy           uuid.UUID         `gorm:"type:uuid;not null"`
	CreatedAt           time.Time         `gorm:"not null;default:now()"`
	UpdatedAt           time.Time         `gorm:"not null;default:now()"`
	LineItems           []POLineItemModel `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE"`
}

func (PurchaseOrderModel) TableName() string {
	return "purchase_orders"
}

// POLineItemModel represents the database model for purchase order line items
type POLineItemModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index:idx_poli_po"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null;index:idx_poli_product"`
	Quantity        float64   `gorm:"type:decimal(15,4);not null"`
	ReceivedQty     float64   `gorm:"type:decimal(15,4);default:0"`
	UnitCost        float64   `gorm:"type:decimal(15,4);default:0"`
	LineTotal       float64   `gorm:"type:decimal(15,2);default:0"`
}

func (POLineItemModel) TableName() string {
	return "purchase_order_lines"
}

type purchaseOrderRepository struct {
	db *gorm.DB
}

// NewPurchaseOrderRepository creates a new purchase order repository
func NewPurchaseOrderRepository(db *gorm.DB) providers.PurchaseOrderRepository {
	return &purchaseOrderRepository{db: db}
}

func (r *purchaseOrderRepository) scopeByOrg(orgID uuid.UUID) *gorm.DB {
	return r.db.Where("organization_id = ?", orgID)
}

func (r *purchaseOrderRepository) Create(ctx context.Context, po *domain.PurchaseOrder) error {
	model := r.toModel(po)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *purchaseOrderRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	var model PurchaseOrderModel
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Preload("LineItems").
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *purchaseOrderRepository) GetByPONumber(ctx context.Context, poNumber string, organizationID uuid.UUID) (*domain.PurchaseOrder, error) {
	var model PurchaseOrderModel
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Preload("LineItems").
		Where("po_number = ?", poNumber).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *purchaseOrderRepository) Update(ctx context.Context, po *domain.PurchaseOrder) error {
	model := r.toModel(po)

	// Use transaction to update order and line items
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the purchase order
		if err := tx.Where("id = ? AND organization_id = ?", model.ID, model.OrganizationID).
			Updates(model).Error; err != nil {
			return err
		}

		// Delete existing line items and recreate
		if err := tx.Where("purchase_order_id = ?", model.ID).Delete(&POLineItemModel{}).Error; err != nil {
			return err
		}

		// Create new line items
		if len(model.LineItems) > 0 {
			for i := range model.LineItems {
				model.LineItems[i].PurchaseOrderID = model.ID
			}
			if err := tx.Create(&model.LineItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *purchaseOrderRepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	result := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("id = ?", id).
		Delete(&PurchaseOrderModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *purchaseOrderRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.PurchaseOrder, int64, error) {
	var models []PurchaseOrderModel
	var total int64

	query := r.scopeByOrg(organizationID).WithContext(ctx).Model(&PurchaseOrderModel{})

	// Apply filters
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "status":
				query = query.Where("status = ?", value)
			case "supplier_id":
				query = query.Where("supplier_id = ?", value)
			case "is_delayed":
				query = query.Where("is_delayed = ?", value)
			case "from_date":
				query = query.Where("order_date >= ?", value)
			case "to_date":
				query = query.Where("order_date <= ?", value)
			}
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate
	offset := (page - 1) * pageSize
	if err := query.
		Preload("LineItems").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	result := make([]*domain.PurchaseOrder, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, total, nil
}

func (r *purchaseOrderRepository) GetDelayedOrders(ctx context.Context, organizationID uuid.UUID) ([]*domain.PurchaseOrder, error) {
	var models []PurchaseOrderModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Preload("LineItems").
		Where("is_delayed = ?", true).
		Where("status NOT IN ?", []string{"received", "closed", "cancelled"}).
		Order("delay_days DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.PurchaseOrder, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, nil
}

// toModel converts domain entity to database model
func (r *purchaseOrderRepository) toModel(po *domain.PurchaseOrder) *PurchaseOrderModel {
	lineItems := make([]POLineItemModel, len(po.LineItems))
	for i, item := range po.LineItems {
		lineItems[i] = POLineItemModel{
			ID:              item.ID,
			PurchaseOrderID: po.ID,
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			ReceivedQty:     item.ReceivedQty,
			UnitCost:        item.UnitCost,
			LineTotal:       item.LineTotal,
		}
	}

	return &PurchaseOrderModel{
		ID:                  po.ID,
		OrganizationID:      po.OrganizationID,
		PONumber:            po.PONumber,
		SupplierID:          po.SupplierID,
		Status:              string(po.Status),
		OrderDate:           po.OrderDate,
		ExpectedArrivalDate: po.ExpectedArrivalDate,
		ActualArrivalDate:   po.ActualArrivalDate,
		DelayDays:           po.DelayDays,
		IsDelayed:           po.IsDelayed,
		TotalAmount:         po.TotalAmount,
		CreatedBy:           po.CreatedBy,
		CreatedAt:           po.CreatedAt,
		UpdatedAt:           po.UpdatedAt,
		LineItems:           lineItems,
	}
}

// toDomain converts database model to domain entity
func (r *purchaseOrderRepository) toDomain(model *PurchaseOrderModel) *domain.PurchaseOrder {
	lineItems := make([]domain.POLineItem, len(model.LineItems))
	for i, item := range model.LineItems {
		lineItems[i] = domain.POLineItem{
			ID:              item.ID,
			PurchaseOrderID: item.PurchaseOrderID,
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			ReceivedQty:     item.ReceivedQty,
			UnitCost:        item.UnitCost,
			LineTotal:       item.LineTotal,
		}
	}

	return &domain.PurchaseOrder{
		ID:                  model.ID,
		OrganizationID:      model.OrganizationID,
		PONumber:            model.PONumber,
		SupplierID:          model.SupplierID,
		Status:              domain.POStatus(model.Status),
		OrderDate:           model.OrderDate,
		ExpectedArrivalDate: model.ExpectedArrivalDate,
		ActualArrivalDate:   model.ActualArrivalDate,
		DelayDays:           model.DelayDays,
		IsDelayed:           model.IsDelayed,
		TotalAmount:         model.TotalAmount,
		LineItems:           lineItems,
		CreatedBy:           model.CreatedBy,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}
