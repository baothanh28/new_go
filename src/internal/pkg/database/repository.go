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
