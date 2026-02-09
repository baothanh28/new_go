package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/service/product/model"
	"myapp/internal/service/product/repository"
)

var (
	// ErrProductNotFound is returned when product is not found
	ErrProductNotFound = errors.New("product not found")
	// ErrSKUExists is returned when SKU already exists
	ErrSKUExists = errors.New("product with this SKU already exists")
	// ErrInsufficientStock is returned when stock is insufficient
	ErrInsufficientStock = errors.New("insufficient stock")
)

// Service handles product business logic
type Service struct {
	repo *repository.Repository
}

// NewService creates a new product service
func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, req *model.CreateProductRequest) (*model.Product, error) {
	// Check if SKU already exists
	exists, err := s.repo.SKUExists(ctx, req.SKU)
	if err != nil {
		return nil, fmt.Errorf("check SKU existence: %w", err)
	}
	if exists {
		return nil, ErrSKUExists
	}

	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		SKU:         req.SKU,
		Category:    req.Category,
		IsActive:    true,
	}

	if err := s.repo.Insert(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}

	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *Service) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by ID: %w", err)
	}
	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (s *Service) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	product, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("get product by SKU: %w", err)
	}
	return product, nil
}

// GetAllProducts retrieves all products with pagination
func (s *Service) GetAllProducts(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	products, err := s.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all products: %w", err)
	}
	return products, nil
}

// GetActiveProducts retrieves all active products with pagination
func (s *Service) GetActiveProducts(ctx context.Context, limit, offset int) ([]*model.Product, error) {
	products, err := s.repo.GetActiveProducts(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get active products: %w", err)
	}
	return products, nil
}

// GetProductsByCategory retrieves products by category
func (s *Service) GetProductsByCategory(ctx context.Context, category string, limit, offset int) ([]*model.Product, error) {
	products, err := s.repo.GetByCategory(ctx, category, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get products by category: %w", err)
	}
	return products, nil
}

// SearchProducts searches products by name or description
func (s *Service) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*model.Product, error) {
	products, err := s.repo.SearchProducts(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search products: %w", err)
	}
	return products, nil
}

// UpdateProduct updates a product
func (s *Service) UpdateProduct(ctx context.Context, id uint, req *model.UpdateProductRequest) (*model.Product, error) {
	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateByID(ctx, id, product); err != nil {
		return nil, fmt.Errorf("update product: %w", err)
	}

	return product, nil
}

// DeleteProduct deletes a product
func (s *Service) DeleteProduct(ctx context.Context, id uint) error {
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("delete product: %w", err)
	}
	return nil
}

// UpdateStock updates product stock
func (s *Service) UpdateStock(ctx context.Context, id uint, quantity int) error {
	product, err := s.GetProductByID(ctx, id)
	if err != nil {
		return err
	}

	newStock := product.Stock + quantity
	if newStock < 0 {
		return ErrInsufficientStock
	}

	if err := s.repo.UpdateStock(ctx, id, quantity); err != nil {
		return fmt.Errorf("update stock: %w", err)
	}

	return nil
}
