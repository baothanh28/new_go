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
