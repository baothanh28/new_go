package service

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"myapp/internal/service/product/dto"
	"myapp/internal/service/product/repository"
)

var (
	ErrProductTestOnlyNotFound = errors.New("product test only not found")
	ErrCodeExists              = errors.New("code already exists")
)

// ProductTestOnlyService handles product test only business logic
type ProductTestOnlyService struct {
	repo *repository.ProductTestOnlyRepository
}

// NewProductTestOnlyService creates a new product test only service
func NewProductTestOnlyService(repo *repository.ProductTestOnlyRepository) *ProductTestOnlyService {
	return &ProductTestOnlyService{repo: repo}
}

// CreateProductTestOnly creates a new product test only
func (s *ProductTestOnlyService) CreateProductTestOnly(ctx context.Context, req *dto.CreateProductTestOnlyRequest) (*dto.ProductTestOnlyResponse, error) {
	// Check if code already exists
	exists, err := s.repo.CodeExists(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("check code exists: %w", err)
	}
	if exists {
		return nil, ErrCodeExists
	}

	// Create entity from request DTO
	entity := req.ToProductTestOnlyEntity()

	// Save to database
	if err := s.repo.Insert(ctx, entity); err != nil {
		return nil, fmt.Errorf("create product test only: %w", err)
	}

	// Convert to response DTO before returning
	return dto.ToProductTestOnlyResponse(entity), nil
}

// GetProductTestOnlyByID retrieves product test only by ID
func (s *ProductTestOnlyService) GetProductTestOnlyByID(ctx context.Context, id uint) (*dto.ProductTestOnlyResponse, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductTestOnlyNotFound
		}
		return nil, fmt.Errorf("get product test only by ID: %w", err)
	}
	// Convert to response DTO before returning
	return dto.ToProductTestOnlyResponse(entity), nil
}

// GetProductTestOnlyByCode retrieves product test only by code
func (s *ProductTestOnlyService) GetProductTestOnlyByCode(ctx context.Context, code string) (*dto.ProductTestOnlyResponse, error) {
	entity, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductTestOnlyNotFound
		}
		return nil, fmt.Errorf("get product test only by code: %w", err)
	}
	// Convert to response DTO before returning
	return dto.ToProductTestOnlyResponse(entity), nil
}

// GetAllProductTestOnly retrieves all product test only records with pagination
func (s *ProductTestOnlyService) GetAllProductTestOnly(ctx context.Context, limit, offset int) ([]*dto.ProductTestOnlyResponse, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	entities, err := s.repo.GetAllPaginated(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all product test only: %w", err)
	}
	
	// Convert entities to response DTOs
	return dto.ToProductTestOnlyResponseList(entities), nil
}

// GetProductTestOnlyByType retrieves product test only records by type
func (s *ProductTestOnlyService) GetProductTestOnlyByType(ctx context.Context, entityType string, limit, offset int) ([]*dto.ProductTestOnlyResponse, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	entities, err := s.repo.GetByType(ctx, entityType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get product test only by type: %w", err)
	}
	
	// Convert entities to response DTOs
	return dto.ToProductTestOnlyResponseList(entities), nil
}

// SearchProductTestOnly searches product test only records by name
func (s *ProductTestOnlyService) SearchProductTestOnly(ctx context.Context, name string, limit, offset int) ([]*dto.ProductTestOnlyResponse, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	entities, err := s.repo.SearchByName(ctx, name, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search product test only: %w", err)
	}
	
	// Convert entities to response DTOs
	return dto.ToProductTestOnlyResponseList(entities), nil
}

// UpdateProductTestOnly updates a product test only
func (s *ProductTestOnlyService) UpdateProductTestOnly(ctx context.Context, id uint, req *dto.UpdateProductTestOnlyRequest) (*dto.ProductTestOnlyResponse, error) {
	// Check if entity exists and get current data
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductTestOnlyNotFound
		}
		return nil, fmt.Errorf("get product test only by ID: %w", err)
	}

	// If code is being updated, check if new code already exists
	if req.Code != nil && *req.Code != entity.Code {
		exists, err := s.repo.CodeExists(ctx, *req.Code)
		if err != nil {
			return nil, fmt.Errorf("check code exists: %w", err)
		}
		if exists {
			return nil, ErrCodeExists
		}
		entity.Code = *req.Code
	}

	// Update fields if provided
	if req.Name != nil {
		entity.Name = *req.Name
	}
	if req.Type != nil {
		entity.Type = *req.Type
	}

	// Save updates
	if err := s.repo.UpdateByID(ctx, id, entity); err != nil {
		return nil, fmt.Errorf("update product test only: %w", err)
	}

	// Convert to response DTO before returning
	return dto.ToProductTestOnlyResponse(entity), nil
}

// DeleteProductTestOnly deletes a product test only (soft delete)
func (s *ProductTestOnlyService) DeleteProductTestOnly(ctx context.Context, id uint) error {
	// Check if entity exists
	_, err := s.GetProductTestOnlyByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete entity
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("delete product test only: %w", err)
	}

	return nil
}
