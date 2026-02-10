package database

import (
	"context"
	"fmt"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// TenantIDKey is the context key for tenant ID
	TenantIDKey contextKey = "tenantID"
)

// WithTenantID adds tenant ID to context
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, TenantIDKey, tenantID)
}

// GetTenantID retrieves tenant ID from context
func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}
