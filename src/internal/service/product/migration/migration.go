package migration

import (
	"fmt"

	"gorm.io/gorm"
	"myapp/internal/service/product/model"
)

// RunMigrations runs database migrations for product service
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		return fmt.Errorf("failed to migrate product table: %w", err)
	}

	// Add any additional migrations here
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes creates additional database indexes
func createIndexes(db *gorm.DB) error {
	// Index on category for faster category queries
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)").Error; err != nil {
		return err
	}

	// Index on is_active for filtering active products
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_products_is_active ON products(is_active)").Error; err != nil {
		return err
	}

	// Composite index on category and is_active
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_products_category_active ON products(category, is_active)").Error; err != nil {
		return err
	}

	// Full-text search index on name and description (PostgreSQL example)
	// Uncomment if using PostgreSQL
	// if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_products_search ON products USING gin(to_tsvector('english', name || ' ' || description))").Error; err != nil {
	// 	return err
	// }

	return nil
}

// Seed adds initial product data (optional)
func Seed(db *gorm.DB) error {
	// Check if products already exist
	var count int64
	if err := db.Model(&model.Product{}).Count(&count).Error; err != nil {
		return err
	}

	// Only seed if no products exist
	if count > 0 {
		return nil
	}

	seedProducts := []*model.Product{
		{
			Name:        "Sample Product 1",
			Description: "This is a sample product for testing",
			Price:       29.99,
			Stock:       100,
			SKU:         "SAMPLE-001",
			Category:    "Electronics",
			IsActive:    true,
		},
		{
			Name:        "Sample Product 2",
			Description: "Another sample product",
			Price:       49.99,
			Stock:       50,
			SKU:         "SAMPLE-002",
			Category:    "Books",
			IsActive:    true,
		},
	}

	for _, product := range seedProducts {
		if err := db.Create(product).Error; err != nil {
			return fmt.Errorf("failed to seed product %s: %w", product.Name, err)
		}
	}

	return nil
}
