package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"gorm.io/gorm"
)

// InventoryTransactionModel represents the database model for inventory transactions
type InventoryTransactionModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_txn_org"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_txn_product"`
	LocationID      uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_txn_location"`
	Type            string    `gorm:"type:varchar(20);not null"`
	Quantity        float64   `gorm:"type:decimal(15,4);not null"`
	UnitCost        float64   `gorm:"type:decimal(15,4);default:0"`
	ReferenceType   string    `gorm:"type:varchar(50)"`
	ReferenceID     uuid.UUID `gorm:"type:uuid;index:idx_inv_txn_ref"`
	Reason          string    `gorm:"type:text"`
	TransactionDate time.Time `gorm:"not null;index:idx_inv_txn_date"`
	CreatedBy       uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
}

func (InventoryTransactionModel) TableName() string {
	return "inventory_transactions"
}

type inventoryTransactionRepository struct {
	db *gorm.DB
}

// NewInventoryTransactionRepository creates a new inventory transaction repository
func NewInventoryTransactionRepository(db *gorm.DB) providers.InventoryTransactionRepository {
	return &inventoryTransactionRepository{db: db}
}

func (r *inventoryTransactionRepository) scopeByOrg(orgID uuid.UUID) *gorm.DB {
	return r.db.Where("organization_id = ?", orgID)
}

func (r *inventoryTransactionRepository) Create(ctx context.Context, txn *domain.InventoryTransaction) error {
	model := r.toModel(txn)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *inventoryTransactionRepository) GetByID(ctx context.Context, id, organizationID uuid.UUID) (*domain.InventoryTransaction, error) {
	var model InventoryTransactionModel
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
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

func (r *inventoryTransactionRepository) List(ctx context.Context, organizationID, productID uuid.UUID, filters map[string]interface{}, page, pageSize int) ([]*domain.InventoryTransaction, int64, error) {
	var models []InventoryTransactionModel
	var total int64

	query := r.scopeByOrg(organizationID).WithContext(ctx).Model(&InventoryTransactionModel{})

	// Filter by product if provided
	if productID != uuid.Nil {
		query = query.Where("product_id = ?", productID)
	}

	// Apply additional filters
	for key, value := range filters {
		if value != nil && value != "" {
			switch key {
			case "type":
				query = query.Where("type = ?", value)
			case "location_id":
				query = query.Where("location_id = ?", value)
			case "reference_type":
				query = query.Where("reference_type = ?", value)
			case "from_date":
				query = query.Where("transaction_date >= ?", value)
			case "to_date":
				query = query.Where("transaction_date <= ?", value)
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
		Offset(offset).
		Limit(pageSize).
		Order("transaction_date DESC, created_at DESC").
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	// Convert to domain
	result := make([]*domain.InventoryTransaction, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, total, nil
}

func (r *inventoryTransactionRepository) GetByReferenceID(ctx context.Context, referenceType string, referenceID, organizationID uuid.UUID) ([]*domain.InventoryTransaction, error) {
	var models []InventoryTransactionModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("reference_type = ?", referenceType).
		Where("reference_id = ?", referenceID).
		Order("created_at ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.InventoryTransaction, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, nil
}

// toModel converts domain entity to database model
func (r *inventoryTransactionRepository) toModel(txn *domain.InventoryTransaction) *InventoryTransactionModel {
	return &InventoryTransactionModel{
		ID:              txn.ID,
		OrganizationID:  txn.OrganizationID,
		ProductID:       txn.ProductID,
		LocationID:      txn.LocationID,
		Type:            string(txn.Type),
		Quantity:        txn.Quantity,
		UnitCost:        txn.UnitCost,
		ReferenceType:   txn.ReferenceType,
		ReferenceID:     txn.ReferenceID,
		Reason:          txn.Reason,
		TransactionDate: txn.TransactionDate,
		CreatedBy:       txn.CreatedBy,
		CreatedAt:       txn.CreatedAt,
	}
}

// toDomain converts database model to domain entity
func (r *inventoryTransactionRepository) toDomain(model *InventoryTransactionModel) *domain.InventoryTransaction {
	return &domain.InventoryTransaction{
		ID:              model.ID,
		OrganizationID:  model.OrganizationID,
		ProductID:       model.ProductID,
		LocationID:      model.LocationID,
		Type:            domain.TransactionType(model.Type),
		Quantity:        model.Quantity,
		UnitCost:        model.UnitCost,
		ReferenceType:   model.ReferenceType,
		ReferenceID:     model.ReferenceID,
		Reason:          model.Reason,
		TransactionDate: model.TransactionDate,
		CreatedBy:       model.CreatedBy,
		CreatedAt:       model.CreatedAt,
	}
}
