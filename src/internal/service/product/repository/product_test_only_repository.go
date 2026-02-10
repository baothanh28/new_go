package repository

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
	"myapp/internal/service/product/model"
)

// ProductTestOnlyRepository handles product test only data access
type ProductTestOnlyRepository struct {
	*database.TenantRepo[model.ProductTestOnly]
	db *gorm.DB
}

// NewProductTestOnlyRepository creates a new product test only repository using tenant database
func NewProductTestOnlyRepository(dbManager *database.DatabaseManager) *ProductTestOnlyRepository {
	return &ProductTestOnlyRepository{
		TenantRepo: database.NewTenantRepo[model.ProductTestOnly](dbManager.TenantConnManager),
		db:         dbManager.TenantDB, // For custom queries
	}
}

// GetByCode retrieves a product test only by code
func (r *ProductTestOnlyRepository) GetByCode(ctx context.Context, code string) (*model.ProductTestOnly, error) {
	var entity model.ProductTestOnly
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&entity).Error
	if err != nil {
		return nil, fmt.Errorf("get product test only by code: %w", err)
	}
	return &entity, nil
}

// CodeExists checks if a code already exists
func (r *ProductTestOnlyRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProductTestOnly{}).Where("code = ?", code).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check code exists: %w", err)
	}
	return count > 0, nil
}

// GetByType retrieves product test only records by type
func (r *ProductTestOnlyRepository) GetByType(ctx context.Context, entityType string, limit, offset int) ([]*model.ProductTestOnly, error) {
	var entities []*model.ProductTestOnly
	err := r.db.WithContext(ctx).
		Where("type = ?", entityType).
		Limit(limit).
		Offset(offset).
		Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("get product test only by type: %w", err)
	}
	return entities, nil
}

// SearchByName searches product test only records by name
func (r *ProductTestOnlyRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]*model.ProductTestOnly, error) {
	var entities []*model.ProductTestOnly
	err := r.db.WithContext(ctx).
		Where("name LIKE ?", "%"+name+"%").
		Limit(limit).
		Offset(offset).
		Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("search product test only by name: %w", err)
	}
	return entities, nil
}

// GetAllPaginated retrieves all product test only records with pagination
func (r *ProductTestOnlyRepository) GetAllPaginated(ctx context.Context, limit, offset int) ([]*model.ProductTestOnly, error) {
	var entities []*model.ProductTestOnly
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&entities).Error
	if err != nil {
		return nil, fmt.Errorf("get all product test only paginated: %w", err)
	}
	return entities, nil
}
