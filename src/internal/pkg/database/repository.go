package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository provides common CRUD operations for any entity type
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new BaseRepository
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// Insert inserts a new entity into the database
func (r *BaseRepository[T]) Insert(ctx context.Context, entity *T) error {
	if err := r.db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("insert entity: %w", err)
	}
	return nil
}

// InsertBatch inserts multiple entities into the database
func (r *BaseRepository[T]) InsertBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(entities).Error; err != nil {
		return fmt.Errorf("insert batch entities: %w", err)
	}
	return nil
}

// UpdateByID updates an entity by its ID
func (r *BaseRepository[T]) UpdateByID(ctx context.Context, id uint, entity *T) error {
	if err := r.db.WithContext(ctx).Model(entity).Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("update entity by id %d: %w", id, err)
	}
	return nil
}

// UpdateWhere updates entities matching conditions with the provided updates
func (r *BaseRepository[T]) UpdateWhere(ctx context.Context, conditions map[string]interface{}, updates map[string]interface{}) error {
	query := r.db.WithContext(ctx).Model(new(T))
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	if err := query.Updates(updates).Error; err != nil {
		return fmt.Errorf("update entities where conditions: %w", err)
	}
	return nil
}

// GetByID retrieves an entity by its ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity with id %d not found", id)
		}
		return nil, fmt.Errorf("get entity by id %d: %w", id, err)
	}
	return &entity, nil
}

// GetAll retrieves all entities with optional limit and offset
func (r *BaseRepository[T]) GetAll(ctx context.Context, limit, offset int) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("get all entities: %w", err)
	}
	return entities, nil
}

// GetWhere retrieves entities matching the provided conditions
func (r *BaseRepository[T]) GetWhere(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	var entities []*T
	query := r.db.WithContext(ctx)
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("get entities where conditions: %w", err)
	}
	return entities, nil
}

// DeleteByID deletes an entity by its ID
func (r *BaseRepository[T]) DeleteByID(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return fmt.Errorf("delete entity by id %d: %w", id, err)
	}
	return nil
}

// DeleteWhere deletes entities matching the provided conditions
func (r *BaseRepository[T]) DeleteWhere(ctx context.Context, conditions map[string]interface{}) error {
	query := r.db.WithContext(ctx).Model(new(T))
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Delete(new(T)).Error; err != nil {
		return fmt.Errorf("delete entities where conditions: %w", err)
	}
	return nil
}

// Count counts entities matching the provided conditions
func (r *BaseRepository[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(new(T))
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count entities: %w", err)
	}
	return count, nil
}

// Exists checks if any entities match the provided conditions
func (r *BaseRepository[T]) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	count, err := r.Count(ctx, conditions)
	if err != nil {
		return false, fmt.Errorf("check entity exists: %w", err)
	}
	return count > 0, nil
}

// WithTx returns a new BaseRepository with the provided transaction
func (r *BaseRepository[T]) WithTx(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: tx}
}

// GetDB returns the underlying database connection
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}

// MasterRepo provides repository operations for entities stored in the master database
// Used for: tenant metadata, user authentication, system configuration, cross-tenant data
type MasterRepo[T any] struct {
	*BaseRepository[T]
}

// NewMasterRepo creates a new repository connected to the master database
func NewMasterRepo[T any](dbManager *DatabaseManager) *MasterRepo[T] {
	return &MasterRepo[T]{
		BaseRepository: NewBaseRepository[T](dbManager.MasterDB),
	}
}

// TenantRepo provides repository operations for entities stored in tenant databases
// Used for: products, orders, customers, tenant-specific business data
// Dynamically connects to the appropriate tenant database based on context
type TenantRepo[T any] struct {
	connManager *TenantConnectionManager
}

// NewTenantRepo creates a new repository with dynamic tenant database connection
func NewTenantRepo[T any](connManager *TenantConnectionManager) *TenantRepo[T] {
	return &TenantRepo[T]{
		connManager: connManager,
	}
}

// getTenantDB is a helper that extracts tenant ID from context and gets the database
func (r *TenantRepo[T]) getTenantDB(ctx context.Context) (*gorm.DB, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	return r.connManager.GetTenantDB(ctx, tenantID)
}

// Insert inserts a new entity into the tenant database
func (r *TenantRepo[T]) Insert(ctx context.Context, entity *T) error {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	if err := db.WithContext(ctx).Create(entity).Error; err != nil {
		return fmt.Errorf("insert entity: %w", err)
	}
	return nil
}

// InsertBatch inserts multiple entities into the tenant database
func (r *TenantRepo[T]) InsertBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	if err := db.WithContext(ctx).Create(entities).Error; err != nil {
		return fmt.Errorf("insert batch entities: %w", err)
	}
	return nil
}

// UpdateByID updates an entity by its ID in the tenant database
func (r *TenantRepo[T]) UpdateByID(ctx context.Context, id uint, entity *T) error {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	if err := db.WithContext(ctx).Model(entity).Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("update entity by id %d: %w", id, err)
	}
	return nil
}

// UpdateWhere updates entities matching conditions with the provided updates
func (r *TenantRepo[T]) UpdateWhere(ctx context.Context, conditions map[string]interface{}, updates map[string]interface{}) error {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	query := db.WithContext(ctx).Model(new(T))
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	if err := query.Updates(updates).Error; err != nil {
		return fmt.Errorf("update entities where conditions: %w", err)
	}
	return nil
}

// GetByID retrieves an entity by its ID from the tenant database
func (r *TenantRepo[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant database: %w", err)
	}
	var entity T
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("entity with id %d not found", id)
		}
		return nil, fmt.Errorf("get entity by id %d: %w", id, err)
	}
	return &entity, nil
}

// GetAll retrieves all entities with optional limit and offset from the tenant database
func (r *TenantRepo[T]) GetAll(ctx context.Context, limit, offset int) ([]*T, error) {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant database: %w", err)
	}
	var entities []*T
	query := db.WithContext(ctx)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("get all entities: %w", err)
	}
	return entities, nil
}

// GetWhere retrieves entities matching the provided conditions from the tenant database
func (r *TenantRepo[T]) GetWhere(ctx context.Context, conditions map[string]interface{}) ([]*T, error) {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("get tenant database: %w", err)
	}
	var entities []*T
	query := db.WithContext(ctx)
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("get entities where conditions: %w", err)
	}
	return entities, nil
}

// DeleteByID deletes an entity by its ID from the tenant database
func (r *TenantRepo[T]) DeleteByID(ctx context.Context, id uint) error {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	if err := db.WithContext(ctx).Delete(new(T), id).Error; err != nil {
		return fmt.Errorf("delete entity by id %d: %w", id, err)
	}
	return nil
}

// DeleteWhere deletes entities matching the provided conditions from the tenant database
func (r *TenantRepo[T]) DeleteWhere(ctx context.Context, conditions map[string]interface{}) error {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return fmt.Errorf("get tenant database: %w", err)
	}
	query := db.WithContext(ctx).Model(new(T))
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Delete(new(T)).Error; err != nil {
		return fmt.Errorf("delete entities where conditions: %w", err)
	}
	return nil
}

// Count counts entities matching the provided conditions in the tenant database
func (r *TenantRepo[T]) Count(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	db, err := r.getTenantDB(ctx)
	if err != nil {
		return 0, fmt.Errorf("get tenant database: %w", err)
	}
	var count int64
	query := db.WithContext(ctx).Model(new(T))
	
	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}
	
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count entities: %w", err)
	}
	return count, nil
}

// Exists checks if any entities match the provided conditions in the tenant database
func (r *TenantRepo[T]) Exists(ctx context.Context, conditions map[string]interface{}) (bool, error) {
	count, err := r.Count(ctx, conditions)
	if err != nil {
		return false, fmt.Errorf("check entity exists: %w", err)
	}
	return count > 0, nil
}

// WithTx returns a new BaseRepository with the provided transaction
// Note: This requires the transaction to be created from the correct tenant database
func (r *TenantRepo[T]) WithTx(tx *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: tx}
}

// GetDB returns the underlying database connection for the current tenant
func (r *TenantRepo[T]) GetDB(ctx context.Context) (*gorm.DB, error) {
	return r.getTenantDB(ctx)
}
