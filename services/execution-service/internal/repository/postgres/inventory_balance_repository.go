package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/execution-service/internal/core/providers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InventoryBalanceModel represents the database model for inventory balances
type InventoryBalanceModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_bal_org;uniqueIndex:uq_inv_bal_org_product_location,priority:1"`
	ProductID      uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_bal_product;uniqueIndex:uq_inv_bal_org_product_location,priority:2"`
	LocationID     uuid.UUID `gorm:"type:uuid;not null;index:idx_inv_bal_location;uniqueIndex:uq_inv_bal_org_product_location,priority:3"`
	OnHand         float64   `gorm:"type:decimal(15,4);default:0"`
	Reserved       float64   `gorm:"type:decimal(15,4);default:0"`
	Available      float64   `gorm:"type:decimal(15,4);default:0"`
	UpdatedAt      time.Time `gorm:"not null;default:now()"`
}

func (InventoryBalanceModel) TableName() string {
	return "inventory_balances"
}

type inventoryBalanceRepository struct {
	db *gorm.DB
}

// NewInventoryBalanceRepository creates a new inventory balance repository
func NewInventoryBalanceRepository(db *gorm.DB) providers.InventoryBalanceRepository {
	return &inventoryBalanceRepository{db: db}
}

func (r *inventoryBalanceRepository) scopeByOrg(orgID uuid.UUID) *gorm.DB {
	return r.db.Where("organization_id = ?", orgID)
}

func (r *inventoryBalanceRepository) GetOrCreate(ctx context.Context, organizationID, productID, locationID uuid.UUID) (*domain.InventoryBalance, error) {
	var model InventoryBalanceModel

	// Try to find existing balance
	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("product_id = ?", productID).
		Where("location_id = ?", locationID).
		First(&model).Error

	if err == nil {
		return r.toDomain(&model), nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new balance if not exists
	model = InventoryBalanceModel{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		ProductID:      productID,
		LocationID:     locationID,
		OnHand:         0,
		Reserved:       0,
		Available:      0,
		UpdatedAt:      time.Now(),
	}

	// Use upsert to handle race conditions
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "organization_id"}, {Name: "product_id"}, {Name: "location_id"}},
			DoNothing: true,
		}).
		Create(&model).Error; err != nil {
		return nil, err
	}

	// Re-fetch to get the actual record (in case of race condition)
	err = r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("product_id = ?", productID).
		Where("location_id = ?", locationID).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *inventoryBalanceRepository) UpdateOnHand(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&InventoryBalanceModel{}).
		Where("organization_id = ?", organizationID).
		Where("product_id = ?", productID).
		Where("location_id = ?", locationID).
		Updates(map[string]interface{}{
			"on_hand":    gorm.Expr("on_hand + ?", quantity),
			"available":  gorm.Expr("GREATEST(on_hand + ? - reserved, 0)", quantity),
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Balance doesn't exist, create it first
		_, err := r.GetOrCreate(ctx, organizationID, productID, locationID)
		if err != nil {
			return err
		}
		// Try update again
		return r.UpdateOnHand(ctx, organizationID, productID, locationID, quantity)
	}

	return nil
}

func (r *inventoryBalanceRepository) UpdateReserved(ctx context.Context, organizationID, productID, locationID uuid.UUID, quantity float64) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&InventoryBalanceModel{}).
		Where("organization_id = ?", organizationID).
		Where("product_id = ?", productID).
		Where("location_id = ?", locationID).
		Updates(map[string]interface{}{
			"reserved":   gorm.Expr("reserved + ?", quantity),
			"available":  gorm.Expr("GREATEST(on_hand - (reserved + ?), 0)", quantity),
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		// Balance doesn't exist, create it first
		_, err := r.GetOrCreate(ctx, organizationID, productID, locationID)
		if err != nil {
			return err
		}
		// Try update again
		return r.UpdateReserved(ctx, organizationID, productID, locationID, quantity)
	}

	return nil
}

func (r *inventoryBalanceRepository) GetByProduct(ctx context.Context, organizationID, productID uuid.UUID) ([]*domain.InventoryBalance, error) {
	var models []InventoryBalanceModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("product_id = ?", productID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.InventoryBalance, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, nil
}

func (r *inventoryBalanceRepository) GetByLocation(ctx context.Context, organizationID, locationID uuid.UUID) ([]*domain.InventoryBalance, error) {
	var models []InventoryBalanceModel

	err := r.scopeByOrg(organizationID).
		WithContext(ctx).
		Where("location_id = ?", locationID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.InventoryBalance, len(models))
	for i := range models {
		result[i] = r.toDomain(&models[i])
	}

	return result, nil
}

// toDomain converts database model to domain entity
func (r *inventoryBalanceRepository) toDomain(model *InventoryBalanceModel) *domain.InventoryBalance {
	return &domain.InventoryBalance{
		ID:             model.ID,
		OrganizationID: model.OrganizationID,
		ProductID:      model.ProductID,
		LocationID:     model.LocationID,
		OnHand:         model.OnHand,
		Reserved:       model.Reserved,
		Available:      model.Available,
		UpdatedAt:      model.UpdatedAt,
	}
}
