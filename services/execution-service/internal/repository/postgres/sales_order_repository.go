package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"gorm.io/gorm"
)

// SalesOrderModel represents the database model for sales orders
type SalesOrderModel struct {
	ID                 uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID     uuid.UUID         `gorm:"type:uuid;not null;index:idx_so_org"`
	SONumber           string            `gorm:"type:varchar(50);not null;index:idx_so_number"`
	CustomerID         uuid.UUID         `gorm:"type:uuid;not null;index:idx_so_customer"`
	Status             string            `gorm:"type:varchar(20);not null;default:'pending'"`
	OrderDate          time.Time         `gorm:"not null"`
	DueDate            time.Time         `gorm:"not null"`
	ShipDate           *time.Time        `gorm:""`
	DeliveryNoteIssued bool              `gorm:"default:false"`
	DeliveryNoteNumber string            `gorm:"type:varchar(50)"`
	DeliveryNoteDate   *time.Time        `gorm:""`
	TotalAmount        float64           `gorm:"type:decimal(15,2);default:0"`
	CreatedAt          time.Time         `gorm:"not null;default:now()"`
	UpdatedAt          time.Time         `gorm:"not null;default:now()"`
	LineItems          []SOLineItemModel `gorm:"foreignKey:SalesOrderID;constraint:OnDelete:CASCADE"`
}

func (SalesOrderModel) TableName() string {
	return "sales_orders"
}

// SOLineItemModel represents the database model for sales order line items
type SOLineItemModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SalesOrderID uuid.UUID `gorm:"type:uuid;not null;index:idx_soli_so"`
	ProductID    uuid.UUID `gorm:"type:uuid;not null;index:idx_soli_product"`
	Quantity     float64   `gorm:"type:decimal(15,4);not null"`
	UnitPrice    float64   `gorm:"type:decimal(15,4);default:0"`
	LineTotal    float64   `gorm:"type:decimal(15,2);default:0"`
}

func (SOLineItemModel) TableName() string {
	return "sales_order_lines"
}

type salesOrderRepository struct {
	db *gorm.DB
}

// NewSalesOrderRepository creates a new sales order repository
func NewSalesOrderRepository(db *gorm.DB) providers.SalesOrderRepository {
	return &salesOrderRepository{db: db}
}

func (r *salesOrderRepository) scopeByOrg(orgID uuid.UUID) *gorm.DB {
	return r.db.Where("organization_id = ?", orgID)
}

func (r *salesOrderRepository) Create(ctx context.Context, so *domain.SalesOrder) error {
	model := r.toModel(so)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *salesOrderRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	var model SalesOrderModel
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

func (r *salesOrderRepository) GetBySONumber(ctx context.Context, soNumber string, organizationID uuid.UUID) (*domain.SalesOrder, error) {
	var model SalesOrderModel
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Preload("LineItems").
		Where("so_number = ?", soNumber).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *salesOrderRepository) Update(ctx context.Context, so *domain.SalesOrder) error {
	model := r.toModel(so)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the sales order
		if err := tx.Where("id = ? AND organization_id = ?", model.ID, model.OrganizationID).
			Updates(model).Error; err != nil {
			return err
		}

		// Delete existing line items and recreate
		if err := tx.Where("sales_order_id = ?", model.ID).Delete(&SOLineItemModel{}).Error; err != nil {
			return err
		}

		// Create new line items
		if len(model.LineItems) > 0 {
			for i := range model.LineItems {
				model.LineItems[i].SalesOrderID = model.ID
			}
			if err := tx.Create(&model.LineItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *salesOrderRepository) Delete(ctx context.Context, id, organizationID uuid.UUID) error {
	result := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("id = ?", id).
		Delete(&SalesOrderModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *salesOrderRepository) List(ctx context.Context, organizationID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.SalesOrder, int64, error) {
	var models []SalesOrderModel
	var total int64

	query := r.scopeByOrg(organizationID).WithContext(ctx).Model(&SalesOrderModel{})

	// Apply filters
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "status":
				query = query.Where("status = ?", value)
			case "customer_id":
				query = query.Where("customer_id = ?", value)
			case "delivery_note_issued":
				query = query.Where("delivery_note_issued = ?", value)
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
	result := make([]*domain.SalesOrder, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, total, nil
}

func (r *salesOrderRepository) GetQualifiedDemand(ctx context.Context, organizationID, productID uuid.UUID) (float64, error) {
	var totalDemand float64

	// Qualified demand = confirmed orders without delivery note issued
	err := r.db.WithContext(ctx).
		Model(&SOLineItemModel{}).
		Joins("JOIN sales_orders ON sales_orders.id = sales_order_lines.sales_order_id").
		Where("sales_orders.organization_id = ?", organizationID).
		Where("sales_orders.status = ?", "confirmed").
		Where("sales_orders.delivery_note_issued = ?", false).
		Where("sales_order_lines.product_id = ?", productID).
		Select("COALESCE(SUM(sales_order_lines.quantity), 0)").
		Scan(&totalDemand).Error

	if err != nil {
		return 0, err
	}

	return totalDemand, nil
}

// toModel converts domain entity to database model
func (r *salesOrderRepository) toModel(so *domain.SalesOrder) *SalesOrderModel {
	lineItems := make([]SOLineItemModel, len(so.LineItems))
	for i, item := range so.LineItems {
		lineItems[i] = SOLineItemModel{
			ID:           item.ID,
			SalesOrderID: so.ID,
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			LineTotal:    item.LineTotal,
		}
	}

	return &SalesOrderModel{
		ID:                 so.ID,
		OrganizationID:     so.OrganizationID,
		SONumber:           so.SONumber,
		CustomerID:         so.CustomerID,
		Status:             string(so.Status),
		OrderDate:          so.OrderDate,
		DueDate:            so.DueDate,
		ShipDate:           so.ShipDate,
		DeliveryNoteIssued: so.DeliveryNoteIssued,
		DeliveryNoteNumber: so.DeliveryNoteNumber,
		DeliveryNoteDate:   so.DeliveryNoteDate,
		TotalAmount:        so.TotalAmount,
		CreatedAt:          so.CreatedAt,
		UpdatedAt:          so.UpdatedAt,
		LineItems:          lineItems,
	}
}

// toDomain converts database model to domain entity
func (r *salesOrderRepository) toDomain(model *SalesOrderModel) *domain.SalesOrder {
	lineItems := make([]domain.SOLineItem, len(model.LineItems))
	for i, item := range model.LineItems {
		lineItems[i] = domain.SOLineItem{
			ID:           item.ID,
			SalesOrderID: item.SalesOrderID,
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			LineTotal:    item.LineTotal,
		}
	}

	return &domain.SalesOrder{
		ID:                 model.ID,
		OrganizationID:     model.OrganizationID,
		SONumber:           model.SONumber,
		CustomerID:         model.CustomerID,
		Status:             domain.SOStatus(model.Status),
		OrderDate:          model.OrderDate,
		DueDate:            model.DueDate,
		ShipDate:           model.ShipDate,
		DeliveryNoteIssued: model.DeliveryNoteIssued,
		DeliveryNoteNumber: model.DeliveryNoteNumber,
		DeliveryNoteDate:   model.DeliveryNoteDate,
		TotalAmount:        model.TotalAmount,
		LineItems:          lineItems,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}
