package model

import (
	"time"
)

// Product represents a product entity
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Stock       int       `gorm:"type:int;default:0" json:"stock"`
	SKU         string    `gorm:"type:varchar(100);uniqueIndex" json:"sku"`
	Category    string    `gorm:"type:varchar(100)" json:"category"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=3,max=255"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"gte=0"`
	SKU         string  `json:"sku" validate:"required,min=3,max=100"`
	Category    string  `json:"category"`
}

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	Name        *string  `json:"name" validate:"omitempty,min=3,max=255"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price" validate:"omitempty,gt=0"`
	Stock       *int     `json:"stock" validate:"omitempty,gte=0"`
	Category    *string  `json:"category"`
	IsActive    *bool    `json:"is_active"`
}

// ProductResponse represents product response
type ProductResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	SKU         string    `json:"sku"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Product to ProductResponse
func (p *Product) ToResponse() *ProductResponse {
	return &ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		SKU:         p.SKU,
		Category:    p.Category,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
