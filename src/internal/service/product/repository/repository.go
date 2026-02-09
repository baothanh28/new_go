package repository

import (
	"context"

	"github.com/base/src/internal/pkg/database"
	"github.com/base/src/internal/service/product/model"
	"gorm.io/gorm"
)

// Repository handles product data access
type Repository struct {
	*database.BaseRepository[model.Product]
	db *gorm.DB
}

// NewRepository creates a new product repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository[model.Product](db),
		db:             db,
	}
}

// GetBySKU retrieves a product by SKU
func (r *Repository) GetBySKU(ctx context.Context, sku string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByCategory retrieves products by category
func (r *Repository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*model.Product, error) {
	var products []*model.Product
	query := r.db.WithContext(ctx).Where("category = ?", category)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// GetActiveProducts retrieves all active products
func (r *Repository) GetActiveProducts(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	var products []*model.Product
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("created_at DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// SKUExists checks if a SKU already exists
func (r *Repository) SKUExists(ctx context.Context, sku string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Product{}).Where("sku = ?", sku).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SearchProducts searches products by name or description
func (r *Repository) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*model.Product, error) {
	var products []*model.Product
	searchQuery := "%" + query + "%"
	
	dbQuery := r.db.WithContext(ctx).
		Where("name LIKE ? OR description LIKE ?", searchQuery, searchQuery).
		Where("is_active = ?", true)
	
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	}
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}
	
	err := dbQuery.Order("created_at DESC").Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

// UpdateStock updates product stock
func (r *Repository) UpdateStock(ctx context.Context, id uint, quantity int) error {
	return r.db.WithContext(ctx).
		Model(&model.Product{}).
		Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", quantity)).
		Error
}
