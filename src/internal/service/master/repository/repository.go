package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/pkg/database"
	"myapp/internal/service/master/model"
)

// Repository handles master data access using master database
type Repository struct {
	*database.MasterRepo[model.Master]
	db *gorm.DB
}

// NewRepository creates a new master repository using master database
func NewRepository(dbManager *database.DatabaseManager) *Repository {
	return &Repository{
		MasterRepo: database.NewMasterRepo[model.Master](dbManager),
		db:         dbManager.MasterDB, // For custom queries
	}
}

// GetByCode retrieves a master record by code
func (r *Repository) GetByCode(ctx context.Context, code string) (*model.Master, error) {
	var master model.Master
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&master).Error
	if err != nil {
		return nil, err
	}
	return &master, nil
}

// CodeExists checks if a code already exists
func (r *Repository) CodeExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Master{}).Where("code = ?", code).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("check code exists: %w", err)
	}
	return count > 0, nil
}

// GetByType retrieves master records by type
func (r *Repository) GetByType(ctx context.Context, masterType string, limit, offset int) ([]*model.Master, error) {
	var masters []*model.Master
	query := r.db.WithContext(ctx).Where("type = ?", masterType)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Find(&masters).Error
	if err != nil {
		return nil, err
	}
	return masters, nil
}

// GetActiveMasters retrieves all active master records
func (r *Repository) GetActiveMasters(ctx context.Context, limit, offset int) ([]*model.Master, error) {
	var masters []*model.Master
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	err := query.Order("created_at DESC").Find(&masters).Error
	if err != nil {
		return nil, err
	}
	return masters, nil
}

// SearchMasters searches master records by name or description
func (r *Repository) SearchMasters(ctx context.Context, query string, limit, offset int) ([]*model.Master, error) {
	var masters []*model.Master
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
	
	err := dbQuery.Order("created_at DESC").Find(&masters).Error
	if err != nil {
		return nil, err
	}
	return masters, nil
}
