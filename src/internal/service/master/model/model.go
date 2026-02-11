package model

import (
	"time"

	"gorm.io/gorm"
)

// Master represents a master data record in the system
type Master struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Code        string `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"`
	Type        string `gorm:"type:varchar(50);index" json:"type"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
}

// TableName sets the table name for Master
func (m *Master) TableName() string {
	return "masters"
}

// CreateMasterRequest defines the request structure for creating a master record
type CreateMasterRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Code        string `json:"code" validate:"required,min=1,max=100"`
	Type        string `json:"type" validate:"required,min=1,max=50"`
}

// UpdateMasterRequest defines the request structure for updating a master record
type UpdateMasterRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Code        *string `json:"code,omitempty" validate:"omitempty,min=1,max=100"`
	Type        *string `json:"type,omitempty" validate:"omitempty,min=1,max=50"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// MasterResponse defines the response structure for master record
type MasterResponse struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Code        string    `json:"code"`
	Type        string    `json:"type"`
	IsActive    bool      `json:"is_active"`
}

// ToResponse converts Master to MasterResponse
func (m *Master) ToResponse() *MasterResponse {
	return &MasterResponse{
		ID:          m.ID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		Name:        m.Name,
		Description: m.Description,
		Code:        m.Code,
		Type:        m.Type,
		IsActive:    m.IsActive,
	}
}
