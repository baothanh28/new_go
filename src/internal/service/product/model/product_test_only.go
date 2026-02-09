package model

import (
	"time"
	"gorm.io/gorm"
)

// ProductTestOnly represents a product test only entity in the system
type ProductTestOnly struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Name string `gorm:"type:varchar(255);not null" json:"name"`
	Type string `gorm:"type:varchar(100);index;not null" json:"type"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
}

// TableName sets the table name for ProductTestOnly
func (p *ProductTestOnly) TableName() string {
	return "product_test_only"
}

// CreateProductTestOnlyRequest defines the request structure for creating a product test only
type CreateProductTestOnlyRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Type string `json:"type" validate:"required,min=1,max=100"`
	Code string `json:"code" validate:"required,min=1,max=50"`
}

// UpdateProductTestOnlyRequest defines the request structure for updating a product test only
type UpdateProductTestOnlyRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Type *string `json:"type,omitempty" validate:"omitempty,min=1,max=100"`
	Code *string `json:"code,omitempty" validate:"omitempty,min=1,max=50"`
}

// ProductTestOnlyResponse defines the response structure for product test only
type ProductTestOnlyResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Code      string    `json:"code"`
}

// ToResponse converts ProductTestOnly to ProductTestOnlyResponse
func (p *ProductTestOnly) ToResponse() *ProductTestOnlyResponse {
	return &ProductTestOnlyResponse{
		ID:        p.ID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		Name:      p.Name,
		Type:      p.Type,
		Code:      p.Code,
	}
}
