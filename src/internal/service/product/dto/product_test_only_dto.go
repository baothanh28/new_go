package dto

import (
	"time"
	"myapp/internal/service/product/model"
)

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

// ToProductTestOnlyResponse converts model.ProductTestOnly to ProductTestOnlyResponse
func ToProductTestOnlyResponse(entity *model.ProductTestOnly) *ProductTestOnlyResponse {
	if entity == nil {
		return nil
	}
	return &ProductTestOnlyResponse{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Name:      entity.Name,
		Type:      entity.Type,
		Code:      entity.Code,
	}
}

// ToProductTestOnlyResponseList converts a slice of entities to a slice of responses
func ToProductTestOnlyResponseList(entities []*model.ProductTestOnly) []*ProductTestOnlyResponse {
	responses := make([]*ProductTestOnlyResponse, len(entities))
	for i, entity := range entities {
		responses[i] = ToProductTestOnlyResponse(entity)
	}
	return responses
}

// ToProductTestOnlyEntity converts CreateProductTestOnlyRequest to model.ProductTestOnly
func (req *CreateProductTestOnlyRequest) ToProductTestOnlyEntity() *model.ProductTestOnly {
	return &model.ProductTestOnly{
		Name: req.Name,
		Type: req.Type,
		Code: req.Code,
	}
}
