---
name: golang-business
description: Generate complete business service layers (model, repository, service, handler, router, module, app) from a model definition. Use when creating new business entities, adding CRUD operations, or when the user mentions generating service components, creating new models, or scaffolding business logic.
---

# Go Business Service Generator

Automatically generate complete service layers following clean architecture patterns with dependency injection (fx), proper error handling, and idiomatic Go code.

## When to Use

Trigger this skill when:
- User creates a new model file in `src/internal/service/{service_name}/model/`
- User asks to "generate service from model"
- User wants CRUD operations for a business entity
- User mentions scaffolding, generating handlers, or creating a new service

## Generation Workflow

### Step 1: Read and Analyze Model

Read the model file to extract:
- Entity name (e.g., `Product`, `Order`, `User`)
- Table name from `TableName()` method
- Field definitions and types
- Validation tags from request structs
- Unique fields (for existence checks)
- Business-specific fields (for custom queries)

### Step 2: Generate Repository

**Location**: `src/internal/service/{service}/repository/repository.go`

**Template Pattern**:

```go
package repository

import (
	"context"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
	"myapp/internal/service/{service}/model"
)

// Repository handles {entity} data access
type Repository struct {
	*database.BaseRepository[model.{Entity}]
	db *gorm.DB
}

// NewRepository creates a new {entity} repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		BaseRepository: database.NewBaseRepository[model.{Entity}](db),
		db:             db,
	}
}

// Custom query methods based on model fields
// Example: GetBySKU, GetByEmail, GetByCategory, SearchByName, etc.
```

**Repository Rules**:
1. Embed `BaseRepository[T]` for standard CRUD
2. Add custom methods for unique fields (email, SKU, username)
3. Add query methods for searchable fields (category, status)
4. Add pagination support (`limit, offset int` parameters)
5. Add specialized updates (UpdateStock, UpdateStatus)
6. Use `ctx context.Context` as first parameter
7. Return `(*model.Entity, error)` or `([]*model.Entity, error)`
8. Wrap errors with `fmt.Errorf` for context

### Step 3: Generate Service

**Location**: `src/internal/service/{service}/service/service.go`

**Template Pattern**:

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"myapp/internal/service/{service}/model"
	"myapp/internal/service/{service}/repository"
)

var (
	Err{Entity}NotFound = errors.New("{entity} not found")
	// Add domain-specific errors based on model constraints
	// Example: ErrSKUExists, ErrEmailExists, ErrInsufficientStock
)

// Service handles {entity} business logic
type Service struct {
	repo *repository.Repository
}

// NewService creates a new {entity} service
func NewService(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// Create{Entity} creates a new {entity}
func (s *Service) Create{Entity}(ctx context.Context, req *model.Create{Entity}Request) (*model.{Entity}, error) {
	// Add business validations here
	// Check unique constraints
	// Transform request to entity
	// Call repository
	// Wrap errors with context
}

// Get{Entity}ByID retrieves {entity} by ID
func (s *Service) Get{Entity}ByID(ctx context.Context, id uint) (*model.{Entity}, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Err{Entity}NotFound
		}
		return nil, fmt.Errorf("get {entity} by ID: %w", err)
	}
	return entity, nil
}

// Standard methods: GetAll, Update, Delete
// Custom methods based on model: Search, GetByCategory, UpdateStock
```

**Service Rules**:
1. Define sentinel errors at package level
2. Accept context as first parameter
3. Accept request DTOs, not entities directly
4. Perform business validations before repository calls
5. Transform `gorm.ErrRecordNotFound` to domain errors
6. Wrap all errors with `fmt.Errorf("operation: %w", err)`
7. Return domain models, not request/response DTOs

### Step 4: Generate Handler

**Location**: `src/internal/service/{service}/handler/handler.go`

**Template Pattern**:

```go
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"myapp/internal/service/{service}/model"
	"myapp/internal/service/{service}/service"
)

// Handler handles {entity} HTTP requests
type Handler struct {
	service *service.Service
}

// NewHandler creates a new {entity} handler
func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

// Create{Entity} handles {entity} creation
// POST /api/{entities}
func (h *Handler) Create{Entity}(c echo.Context) error {
	var req model.Create{Entity}Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	entity, err := h.service.Create{Entity}(c.Request().Context(), &req)
	if err != nil {
		// Map service errors to HTTP status codes
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create {entity}",
		})
	}

	return c.JSON(http.StatusCreated, entity.ToResponse())
}

// Standard CRUD handlers: Get, GetAll, Update, Delete
// Custom handlers based on business logic
```

**Handler Rules**:
1. Extract request with `c.Bind(&req)`
2. Validate with `c.Validate(&req)`
3. Pass `c.Request().Context()` to service
4. Map domain errors to HTTP status codes:
   - 400: Invalid input, validation errors
   - 404: Entity not found
   - 409: Conflict (duplicate key, etc.)
   - 500: Unexpected errors
5. Return `entity.ToResponse()` for responses
6. Use consistent error response format
7. Include route documentation in comments

### Step 5: Generate Router

**Location**: `src/internal/service/{service}/router/router.go`

**Template Pattern**:

```go
package router

import (
	"myapp/internal/service/{service}/handler"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Register{Entity}Routes registers all {entity}-related routes
func Register{Entity}Routes(
	e *echo.Echo,
	{entity}Handler *handler.Handler,
	logger *zap.Logger,
) {
	logger.Info("Registering {entity} routes")

	api := e.Group("/api")

	// Public routes - basic rate limiting
	public{Entities} := api.Group("/{entities}")
	public{Entities}.GET("", {entity}Handler.Get{Entities})
	public{Entities}.GET("/:id", {entity}Handler.Get{Entity})

	// Protected routes - authentication required
	protected{Entities} := api.Group("/{entities}", authMiddleware())
	protected{Entities}.POST("", {entity}Handler.Create{Entity})
	protected{Entities}.PUT("/:id", {entity}Handler.Update{Entity})
	protected{Entities}.DELETE("/:id", {entity}Handler.Delete{Entity})

	logger.Info("{Entity} routes registered successfully")
}

// Include middleware stubs for reference
```

**Router Rules**:
1. Group routes by access level (public, protected, admin)
2. Apply middleware at group level, not per-route
3. Use RESTful naming conventions
4. Include logger for registration confirmation
5. Document middleware requirements
6. Include example middleware implementations as comments

### Step 6: Generate Module

**Location**: `src/internal/service/{service}/module/module.go`

**Template Pattern**:

```go
package module

import (
	"go.uber.org/fx"
	"myapp/internal/service/{service}/handler"
	"myapp/internal/service/{service}/repository"
	"myapp/internal/service/{service}/service"
)

// Module exports {service} service dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewRepository,
		service.NewService,
		handler.NewHandler,
	),
)
```

**Module Rules**:
1. Use `fx.Provide` for constructor functions
2. Export as package-level `Module` variable
3. Keep in order: repository → service → handler
4. One module per service

### Step 7: Update or Create App

**Location**: `src/internal/service/{service}/app.go`

**Template Pattern**:

```go
package {service}

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
	{service}module "myapp/internal/service/{service}/module"
	{service}router "myapp/internal/service/{service}/router"
)

// AppModule combines infrastructure and {service} service modules
var AppModule = fx.Options(
	// Infrastructure modules
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// {Service} service module
	{service}module.Module,
	
	// Router registration
	fx.Invoke({service}router.Register{Entity}Routes),
)
```

**App Rules**:
1. Import infrastructure modules (config, logger, database, server)
2. Import service module with alias: `{service}module`
3. Import router with alias: `{service}router`
4. Use `fx.Invoke` for router registration
5. Create as package-level `AppModule` variable

## Model Analysis Checklist

When analyzing a model, extract:

- [ ] Entity name (struct name)
- [ ] Table name from `TableName()` method
- [ ] Unique fields (email, SKU, username) → Generate existence checks
- [ ] Searchable fields (name, description) → Generate search methods
- [ ] Categorical fields (category, status, type) → Generate filter methods
- [ ] Numeric fields for operations (stock, quantity) → Generate update methods
- [ ] Timestamp fields (created_at, updated_at) → Include in queries
- [ ] Request struct fields → Map to validation rules
- [ ] Boolean flags (is_active, is_deleted) → Generate filter methods

## Applying golang-patterns Skill

Always apply principles from the `golang-patterns` skill:

1. **Error Handling**: Wrap errors with context, use sentinel errors
2. **Interfaces**: Define repository interface at consumer side (service)
3. **Context**: Always accept `context.Context` as first parameter
4. **Naming**: Short, lowercase package names; clear function names
5. **Zero Values**: Design structs to work without initialization
6. **Return Values**: Return concrete types, accept interfaces
7. **Concurrency**: Use context for cancellation in long operations

## Pre-Generation Checklist

Before generating code, verify:

- [ ] Model file exists and is readable
- [ ] Target directories exist (`handler/`, `service/`, etc.)
- [ ] Extract entity name, table name, and all fields
- [ ] Identify unique constraints for existence checks
- [ ] Identify searchable/filterable fields
- [ ] Check if service already exists (update vs create)

## Post-Generation Checklist

After generating code, verify:

- [ ] All imports are correct
- [ ] Package names match directory structure
- [ ] Entity name is consistently PascalCase
- [ ] Service name is consistently lowercase
- [ ] fx dependency injection is properly configured
- [ ] Router is registered in AppModule
- [ ] Error handling follows patterns
- [ ] All methods have context parameter
- [ ] Response methods use `ToResponse()`

## Common Patterns by Model Type

### E-commerce Product
- Repository: `GetBySKU`, `GetByCategory`, `SearchProducts`, `GetActiveProducts`, `UpdateStock`
- Service: Stock validation, price validation, SKU uniqueness
- Handlers: Inventory management, search, filtering

### User/Account
- Repository: `GetByEmail`, `GetByUsername`, `GetActiveUsers`
- Service: Email uniqueness, password hashing, role validation
- Handlers: Authentication, profile management

### Order/Transaction
- Repository: `GetByUserID`, `GetByStatus`, `GetByDateRange`
- Service: Status transitions, payment validation, inventory checks
- Handlers: Order lifecycle, status updates

### Content (Blog, Post, Article)
- Repository: `GetBySlug`, `GetByAuthor`, `GetPublished`, `SearchByTitle`
- Service: Slug generation, publish/unpublish, author verification
- Handlers: Content CRUD, publishing workflow

## Example Command Flow

```bash
# User creates a new model
# File: src/internal/service/order/model/model.go

# AI should automatically:
1. Read and analyze the Order model
2. Generate repository with Order-specific queries
3. Generate service with order business logic
4. Generate handler with order HTTP endpoints
5. Generate router with order routes
6. Generate module for dependency injection
7. Update/create app.go with module registration
```

## File Naming Conventions

- Model: `model.go` (all models in one file per service)
- Repository: `repository.go`
- Service: `service.go`
- Handler: `handler.go`
- Router: `router.go`
- Module: `module.go`
- App: `app.go` (at service root)

## Import Path Pattern

```go
import (
	"myapp/internal/service/{service}/model"
	"myapp/internal/service/{service}/repository"
	"myapp/internal/service/{service}/service"
	"myapp/internal/service/{service}/handler"
	{service}module "myapp/internal/service/{service}/module"
	{service}router "myapp/internal/service/{service}/router"
)
```

## Error Response Format

Consistent JSON error responses:

```go
map[string]string{"error": "descriptive message"}
```

For lists with metadata:

```go
map[string]interface{}{
	"items":  responses,
	"limit":  limit,
	"offset": offset,
}
```

## Quick Reference

| Component | Responsibility | Key Pattern |
|-----------|---------------|-------------|
| Model | Data structure | Entity, Request, Response DTOs |
| Repository | Data access | Embed BaseRepository, add custom queries |
| Service | Business logic | Validate, transform, coordinate |
| Handler | HTTP layer | Bind, validate, map errors to status codes |
| Router | Route registration | Group by access level, apply middleware |
| Module | DI configuration | fx.Provide constructors |
| App | Module composition | Combine infrastructure + service modules |

## Additional Notes

- Always read the golang-patterns skill before generating code
- Generate code following existing project conventions
- Use the same error handling patterns as existing services
- Maintain consistency with existing middleware patterns
- Include comprehensive comments for generated code
- Generate complete CRUD operations by default
- Add custom methods based on model field analysis
