package repositories

import (
	"context"

	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/domain"
	"github.com/melegattip/giia-core-engine/services/catalog-service/internal/core/providers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) providers.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) tenantScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
			return db.Where("organization_id = ?", orgID)
		}
		return db
	}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	if orgID, ok := ctx.Value("organization_id").(uuid.UUID); ok && orgID != uuid.Nil {
		product.OrganizationID = orgID
	}

	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		if errors.IsDuplicateKeyError(err) {
			return errors.NewConflict("product with this SKU already exists for this organization")
		}
		return errors.NewInternalServerError("failed to create product")
	}

	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", id).
		First(&product).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("product not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve product")
	}

	return &product, nil
}

func (r *productRepository) GetByIDWithSuppliers(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Preload("Suppliers").
		Preload("Suppliers.Supplier").
		Where("id = ?", id).
		First(&product).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("product not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve product")
	}

	return &product, nil
}

func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	var product domain.Product
	err := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("sku = ?", sku).
		First(&product).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFound("product not found")
		}
		return nil, errors.NewInternalServerError("failed to retrieve product")
	}

	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Where("id = ?", product.ID).
		Updates(product)

	if result.Error != nil {
		if errors.IsDuplicateKeyError(result.Error) {
			return errors.NewConflict("product with this SKU already exists for this organization")
		}
		return errors.NewInternalServerError("failed to update product")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("product not found")
	}

	return nil
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.Product{}).
		Where("id = ?", id).
		Update("status", domain.ProductStatusInactive)

	if result.Error != nil {
		return errors.NewInternalServerError("failed to delete product")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("product not found")
	}

	return nil
}

func (r *productRepository) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	query := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.Product{})

	for key, value := range filters {
		if value != nil && value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count products")
	}

	offset := (page - 1) * pageSize
	err := query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&products).Error

	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to list products")
	}

	return products, total, nil
}

func (r *productRepository) Search(ctx context.Context, query string, filters map[string]interface{}, page, pageSize int) ([]*domain.Product, int64, error) {
	var products []*domain.Product
	var total int64

	db := r.db.WithContext(ctx).
		Scopes(r.tenantScope(ctx)).
		Model(&domain.Product{})

	if query != "" {
		db = db.Where("sku ILIKE ? OR name ILIKE ?", "%"+query+"%", "%"+query+"%")
	}

	for key, value := range filters {
		if value != nil && value != "" {
			db = db.Where(key+" = ?", value)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.NewInternalServerError("failed to count products")
	}

	offset := (page - 1) * pageSize
	err := db.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&products).Error

	if err != nil {
		return nil, 0, errors.NewInternalServerError("failed to search products")
	}

	return products, total, nil
}

func (r *productRepository) AssociateSupplier(ctx context.Context, productSupplier *domain.ProductSupplier) error {
	if err := r.db.WithContext(ctx).Create(productSupplier).Error; err != nil {
		if errors.IsDuplicateKeyError(err) {
			return errors.NewConflict("supplier is already associated with this product")
		}
		return errors.NewInternalServerError("failed to associate supplier")
	}

	return nil
}

func (r *productRepository) RemoveSupplier(ctx context.Context, productID, supplierID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("product_id = ? AND supplier_id = ?", productID, supplierID).
		Delete(&domain.ProductSupplier{})

	if result.Error != nil {
		return errors.NewInternalServerError("failed to remove supplier association")
	}

	if result.RowsAffected == 0 {
		return errors.NewNotFound("supplier association not found")
	}

	return nil
}

func (r *productRepository) GetProductSuppliers(ctx context.Context, productID uuid.UUID) ([]*domain.ProductSupplier, error) {
	var productSuppliers []*domain.ProductSupplier

	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Where("product_id = ?", productID).
		Find(&productSuppliers).Error

	if err != nil {
		return nil, errors.NewInternalServerError("failed to retrieve product suppliers")
	}

	return productSuppliers, nil
}
