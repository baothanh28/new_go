package database

import (
	"time"
)

// Tenant represents a tenant entity stored in the master database
// Contains database connection information for the tenant's dedicated database
type Tenant struct {
	ID         string    `gorm:"primaryKey;type:varchar(100)" json:"id"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	DBType     string    `gorm:"type:varchar(50);not null;default:'mysql';column:db_type" json:"db_type"` // Database type: mysql, postgresql, etc.
	DBHost     string    `gorm:"type:varchar(255);not null;column:db_host" json:"db_host"`
	DBPort     int       `gorm:"not null;column:db_port" json:"db_port"`
	DBName     string    `gorm:"type:varchar(100);not null;column:db_name" json:"db_name"`
	DBUser     string    `gorm:"type:varchar(100);not null;column:db_user" json:"db_user"`
	DBPassword string    `gorm:"type:varchar(255);not null;column:db_password" json:"-"` // Don't expose in JSON
	IsActive   bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName specifies the table name for Tenant
func (Tenant) TableName() string {
	return "tenants"
}
