package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWithTenantID tests adding tenant ID to context
func TestWithTenantID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
	}{
		{
			name:     "add valid tenant ID",
			tenantID: "tenant-123",
		},
		{
			name:     "add UUID tenant ID",
			tenantID: "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "add numeric tenant ID",
			tenantID: "12345",
		},
		{
			name:     "add empty tenant ID",
			tenantID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := WithTenantID(ctx, tt.tenantID)
			
			assert.NotNil(t, newCtx)
			
			// Verify the value is stored in context
			value := newCtx.Value(TenantIDKey)
			assert.Equal(t, tt.tenantID, value)
		})
	}
}

// TestGetTenantID tests retrieving tenant ID from context
func TestGetTenantID(t *testing.T) {
	tests := []struct {
		name      string
		setupCtx  func() context.Context
		wantID    string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "get valid tenant ID",
			setupCtx: func() context.Context {
				return WithTenantID(context.Background(), "tenant-123")
			},
			wantID:  "tenant-123",
			wantErr: false,
		},
		{
			name: "get UUID tenant ID",
			setupCtx: func() context.Context {
				return WithTenantID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")
			},
			wantID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name: "context without tenant ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantID:  "",
			wantErr: true,
			errMsg:  "tenant ID not found in context",
		},
		{
			name: "context with empty tenant ID",
			setupCtx: func() context.Context {
				return WithTenantID(context.Background(), "")
			},
			wantID:  "",
			wantErr: true,
			errMsg:  "tenant ID not found in context",
		},
		{
			name: "context with wrong type value",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), TenantIDKey, 123)
			},
			wantID:  "",
			wantErr: true,
			errMsg:  "tenant ID not found in context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			tenantID, err := GetTenantID(ctx)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, tenantID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, tenantID)
			}
		})
	}
}

// TestContextChaining tests that context values are preserved through chaining
func TestContextChaining(t *testing.T) {
	t.Run("tenant ID preserved through context chain", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTenantID(ctx, "tenant-456")
		ctx = context.WithValue(ctx, "other-key", "other-value")
		
		// Verify tenant ID is still accessible
		tenantID, err := GetTenantID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "tenant-456", tenantID)
		
		// Verify other value is also accessible
		otherValue := ctx.Value("other-key")
		assert.Equal(t, "other-value", otherValue)
	})

	t.Run("overwrite tenant ID in context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTenantID(ctx, "tenant-old")
		ctx = WithTenantID(ctx, "tenant-new")
		
		tenantID, err := GetTenantID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "tenant-new", tenantID)
	})
}

// TestTenantIDKey tests the context key constant
func TestTenantIDKey(t *testing.T) {
	t.Run("TenantIDKey is unique", func(t *testing.T) {
		// Verify that our custom type prevents key collisions
		ctx := context.Background()
		ctx = context.WithValue(ctx, "tenantID", "wrong-value") // string key
		ctx = WithTenantID(ctx, "correct-value")                // contextKey type
		
		// String key should not interfere with our contextKey
		tenantID, err := GetTenantID(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "correct-value", tenantID)
		
		// String key value should still be accessible
		stringValue := ctx.Value("tenantID")
		assert.Equal(t, "wrong-value", stringValue)
	})
}
