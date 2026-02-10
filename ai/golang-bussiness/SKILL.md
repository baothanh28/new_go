---
name: golang-business
description: Generate complete business service layers (model, repository, service, handler, router, module, app) from a model definition. Use when creating new business entities, adding CRUD operations, or when the user mentions generating service components, creating new models, or scaffolding business logic.
---

# Go Business Service Generator

Automatically generate complete service layers following clean architecture patterns with dependency injection (fx), proper error handling, and idiomatic Go code.

## 🎯 Key Principles: Clean Architecture & Separation of Concerns

### One Entity = One File

**IMPORTANT**: Each business entity must have its own separate file in model/, dto/, repository/, service/, handler/, and router/ directories.

**File Naming**: 
- Model: `{entity}.go` (e.g., `product.go`, `category.go`)
- DTO: `{entity}_dto.go` (e.g., `product_dto.go`, `category_dto.go`)
- Others: `{entity}_repository.go`, `{entity}_service.go`, `{entity}_handler.go`, `{entity}_router.go`

**Struct Naming**: Use entity name as prefix: `{Entity}`, `{Entity}Repository`, `{Entity}Service`, `{Entity}Handler`

### Separation: Models vs DTOs

**CRITICAL**: Always separate domain models from Data Transfer Objects (DTOs):

- **`model/`** folder - Database entities ONLY (GORM tags)
- **`dto/`** folder - API contracts (Request/Response DTOs with validation tags)

**Why?**
- ✅ Models stay internal to service layer
- ✅ DTOs control API contracts independently 
- ✅ Database schema changes don't break API
- ✅ Security: exclude sensitive fields from responses
- ✅ Flexibility: different API versions can use same model

## 🚀 Model Creation Workflow

**When user says "create model {EntityName}":**

1. **Automatically create** `src/internal/service/{service}/model/{entity}.go` with:
   - Entity struct with base fields (ID, timestamps, soft delete)
   - Custom fields based on user requirements
   - GORM tags for database constraints
   - JSON tags for API serialization (optional, but DTO is preferred)
   - TableName() method
   - ❌ **DO NOT** include DTOs in model file

2. **Automatically create** `src/internal/service/{service}/dto/{entity}_dto.go` with:
   - CreateRequest struct with validation tags
   - UpdateRequest struct with optional pointer fields
   - Response struct for API responses
   - Converter functions: `To{Entity}Response(entity)`, `To{Entity}ResponseList(entities)`, `ToEntity()` method

3. **Then automatically proceed** to generate complete service layers:
   - Repository (Step 3)
   - Service (Step 4)
   - Handler (Step 5)
   - Router (Step 6)
   - Update Module (Step 7)
   - Update Migration (Step 8)
   - Update App (Step 9)

**The user should only need to say "create model Product" and the entire service layer is generated automatically!**

## When to Use

Trigger this skill when:
- User says "create model {entity_name}" or "generate model for {entity_name}"
- User creates a new model file in `src/internal/service/{service_name}/model/`
- User asks to "generate service from model"
- User wants CRUD operations for a business entity
- User mentions scaffolding, generating handlers, or creating a new service

## Generation Workflow

### Step 0: Create Model (If Requested)

**When**: User says "create model {entity_name}" or provides entity specification

**Location**: `src/internal/service/{service}/model/{entity_lowercase}.go`

**File Naming**: Use lowercase entity name (e.g., `product.go`, `category.go`, `order.go`)

**Template Pattern**:

```go
package model

import (
	"time"
	"gorm.io/gorm"
)

// {Entity} represents a {entity} in the system
type {Entity} struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Add entity-specific fields here based on user requirements
	// Example for Product:
	// Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	// Description string  `gorm:"type:text" json:"description"`
	// Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	// SKU         string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
	// Stock       int     `gorm:"default:0" json:"stock"`
	// CategoryID  uint    `gorm:"index" json:"category_id"`
	// IsActive    bool    `gorm:"default:true" json:"is_active"`
}

// TableName sets the table name for {Entity}
func (e *{Entity}) TableName() string {
	return "{table_name}"
}
```

### Step 1: Create DTOs

**When**: Immediately after creating model

**Location**: `src/internal/service/{service}/dto/{entity_lowercase}_dto.go`

**File Naming**: Use lowercase entity name with `_dto` suffix (e.g., `product_dto.go`, `category_dto.go`)

**Template Pattern**:

```go
package dto

import (
	"time"
	"myapp/internal/service/{service}/model"
)

// Create{Entity}Request defines the request structure for creating a {entity}
type Create{Entity}Request struct {
	// Add required fields with validation tags
	// Example:
	// Name        string  `json:"name" validate:"required,min=1,max=255"`
	// Description string  `json:"description"`
	// Price       float64 `json:"price" validate:"required,gt=0"`
	// SKU         string  `json:"sku" validate:"required,min=1,max=100"`
	// Stock       int     `json:"stock" validate:"gte=0"`
	// CategoryID  uint    `json:"category_id" validate:"required"`
}

// Update{Entity}Request defines the request structure for updating a {entity}
type Update{Entity}Request struct {
	// Add optional fields for update (use pointers for optional fields)
	// Example:
	// Name        *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	// Description *string  `json:"description,omitempty"`
	// Price       *float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	// Stock       *int     `json:"stock,omitempty" validate:"omitempty,gte=0"`
	// IsActive    *bool    `json:"is_active,omitempty"`
}

// {Entity}Response defines the response structure for {entity}
type {Entity}Response struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Add all entity fields to include in response
	// Example:
	// Name        string  `json:"name"`
	// Description string  `json:"description"`
	// Price       float64 `json:"price"`
	// SKU         string  `json:"sku"`
	// Stock       int     `json:"stock"`
	// IsActive    bool    `json:"is_active"`
}

// To{Entity}Response converts model.{Entity} to {Entity}Response
func To{Entity}Response(entity *model.{Entity}) *{Entity}Response {
	if entity == nil {
		return nil
	}
	return &{Entity}Response{
		ID:        entity.ID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		// Map all fields from entity to response
	}
}

// To{Entity}ResponseList converts a slice of entities to a slice of responses
func To{Entity}ResponseList(entities []*model.{Entity}) []*{Entity}Response {
	responses := make([]*{Entity}Response, len(entities))
	for i, entity := range entities {
		responses[i] = To{Entity}Response(entity)
	}
	return responses
}

// ToEntity converts Create{Entity}Request to model.{Entity}
func (req *Create{Entity}Request) ToEntity() *model.{Entity} {
	return &model.{Entity}{
		// Map request fields to entity
		// Example:
		// Name:        req.Name,
		// Description: req.Description,
		// Price:       req.Price,
		// SKU:         req.SKU,
		// Stock:       req.Stock,
	}
}
```

**Model Creation Rules**:
1. **File per Entity**: Create separate file for each entity in the `model/` directory
2. **Base Fields**: Always include ID, CreatedAt, UpdatedAt, DeletedAt (soft delete)
3. **GORM Tags**: Use appropriate GORM tags for database constraints:
   - `primarykey` for ID
   - `type:` for specific column types
   - `not null` for required fields
   - `uniqueIndex` for unique fields
   - `index` for searchable fields
   - `default:` for default values
4. **JSON Tags**: Include JSON tags (optional, but DTO is preferred for API)
5. **TableName Method**: Implement `TableName()` to explicitly set table name (usually plural)
6. **Keep Model Clean**: ❌ DO NOT include DTOs, validation, or API concerns in model file

**DTO Creation Rules**:
1. **Separate DTO File**: Create `dto/{entity}_dto.go` for each entity
2. **Request DTOs**: Create separate `Create` and `Update` request structs
   - Use pointers in Update requests for optional fields
   - Add validation tags (`validate:`)
3. **Response DTO**: Create response struct to control what data is exposed
4. **Converter Functions**: 
   - `To{Entity}Response(entity)` - Convert single entity to DTO
   - `To{Entity}ResponseList(entities)` - Convert slice of entities
   - `(req *Create{Entity}Request) ToEntity()` - Convert request DTO to entity
5. **Nil Safety**: Always check for nil entities in converter functions

**Common Field Types by Entity**:

**Product/Item**:
```go
Name        string  `gorm:"type:varchar(255);not null" json:"name"`
Description string  `gorm:"type:text" json:"description"`
Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`
SKU         string  `gorm:"type:varchar(100);uniqueIndex;not null" json:"sku"`
Stock       int     `gorm:"default:0" json:"stock"`
CategoryID  uint    `gorm:"index" json:"category_id"`
IsActive    bool    `gorm:"default:true" json:"is_active"`
```

**User/Account**:
```go
Email     string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
Username  string `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
Password  string `gorm:"type:varchar(255);not null" json:"-"`
FirstName string `gorm:"type:varchar(100)" json:"first_name"`
LastName  string `gorm:"type:varchar(100)" json:"last_name"`
Role      string `gorm:"type:varchar(50);default:'user'" json:"role"`
IsActive  bool   `gorm:"default:true" json:"is_active"`
```

**Category/Type**:
```go
Name        string `gorm:"type:varchar(255);not null" json:"name"`
Code        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
Description string `gorm:"type:text" json:"description"`
Type        string `gorm:"type:varchar(50);index" json:"type"`
IsActive    bool   `gorm:"default:true" json:"is_active"`
```

**Order/Transaction**:
```go
OrderNumber string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_number"`
UserID      uint      `gorm:"index;not null" json:"user_id"`
TotalAmount float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
Status      string    `gorm:"type:varchar(50);index;default:'pending'" json:"status"`
OrderDate   time.Time `gorm:"not null" json:"order_date"`
Notes       string    `gorm:"type:text" json:"notes"`
```

**Model Creation Workflow**:
1. Ask user for entity name if not provided
2. Ask user for key fields and their types if not specified
3. Create model file in `src/internal/service/{service}/model/{entity}.go`
4. Generate base entity struct with GORM tags
5. Generate TableName() method
6. Create DTO file in `src/internal/service/{service}/dto/{entity}_dto.go`
7. Generate Create/Update request structs with validation tags
8. Generate Response struct
9. Generate converter functions (To{Entity}Response, ToEntity, etc.)
10. Confirm creation and proceed to generate service layers (Step 2+)

### Step 2: Read and Analyze Model and DTOs

Read the model file to extract:
- Entity name (e.g., `Product`, `Order`, `User`)
- Table name from `TableName()` method
- Field definitions and types
- Validation tags from request structs
- Unique fields (for existence checks)
- Business-specific fields (for custom queries)

### Step 3: Generate Repository

**Location**: `src/internal/service/{service}/repository/{entity_lowercase}_repository.go`

**File Naming**: Use lowercase entity name with underscore (e.g., `product_repository.go`, `category_repository.go`)

**Template Pattern**:

```go
package repository

import (
	"context"
	"gorm.io/gorm"
	"myapp/internal/pkg/database"
	"myapp/internal/service/{service}/model"
)

// {Entity}Repository handles {entity} data access
type {Entity}Repository struct {
	*database.BaseRepository[model.{Entity}]
	db *gorm.DB
}

// New{Entity}Repository creates a new {entity} repository
func New{Entity}Repository(db *gorm.DB) *{Entity}Repository {
	return &{Entity}Repository{
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

### Step 4: Generate Service

**Location**: `src/internal/service/{service}/service/{entity_lowercase}_service.go`

**File Naming**: Use lowercase entity name with underscore (e.g., `product_service.go`, `category_service.go`)

**Template Pattern**:

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"myapp/internal/service/{service}/dto"
	"myapp/internal/service/{service}/repository"
)

var (
	Err{Entity}NotFound = errors.New("{entity} not found")
	// Add domain-specific errors based on model constraints
	// Example: ErrSKUExists, ErrEmailExists, ErrInsufficientStock
)

// {Entity}Service handles {entity} business logic
type {Entity}Service struct {
	repo *repository.{Entity}Repository
}

// New{Entity}Service creates a new {entity} service
func New{Entity}Service(repo *repository.{Entity}Repository) *{Entity}Service {
	return &{Entity}Service{repo: repo}
}

// Create{Entity} creates a new {entity}
func (s *{Entity}Service) Create{Entity}(ctx context.Context, req *dto.Create{Entity}Request) (*dto.{Entity}Response, error) {
	// Add business validations here
	// Check unique constraints
	
	// Transform request DTO to entity
	entity := req.ToEntity()
	
	// Save to database
	if err := s.repo.Insert(ctx, entity); err != nil {
		return nil, fmt.Errorf("create {entity}: %w", err)
	}
	
	// Convert entity to response DTO before returning
	return dto.To{Entity}Response(entity), nil
}

// Get{Entity}ByID retrieves {entity} by ID
func (s *{Entity}Service) Get{Entity}ByID(ctx context.Context, id uint) (*dto.{Entity}Response, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Err{Entity}NotFound
		}
		return nil, fmt.Errorf("get {entity} by ID: %w", err)
	}
	// Convert entity to response DTO before returning
	return dto.To{Entity}Response(entity), nil
}

// GetAll{Entity} retrieves all {entity} records with pagination
func (s *{Entity}Service) GetAll{Entity}(ctx context.Context, limit, offset int) ([]*dto.{Entity}Response, error) {
	entities, err := s.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get all {entity}: %w", err)
	}
	// Convert entities to response DTOs
	return dto.To{Entity}ResponseList(entities), nil
}

// Standard methods: Update, Delete
// Custom methods based on model: Search, GetByCategory, UpdateStock
```

**Service Rules**:
1. Define sentinel errors at package level
2. Accept context as first parameter
3. **Import `dto` package, NOT `model` package**
4. Accept request DTOs (`dto.Create{Entity}Request`), not entities directly
5. Return response DTOs (`dto.{Entity}Response`), never raw entities
6. Use `req.ToEntity()` to convert request DTO to entity
7. Use `dto.To{Entity}Response(entity)` to convert entity to response DTO
8. Use `dto.To{Entity}ResponseList(entities)` for slice conversions
9. Perform business validations before repository calls
10. Transform `gorm.ErrRecordNotFound` to domain errors
11. Wrap all errors with `fmt.Errorf("operation: %w", err)`

### Step 5: Generate Handler

**Location**: `src/internal/service/{service}/handler/{entity_lowercase}_handler.go`

**File Naming**: Use lowercase entity name with underscore (e.g., `product_handler.go`, `category_handler.go`)

**Template Pattern**:

```go
package handler

import (
	"errors"
	"net/http"
	"strconv"
	"github.com/labstack/echo/v4"
	"myapp/internal/service/{service}/dto"
	"myapp/internal/service/{service}/service"
)

// {Entity}Handler handles {entity} HTTP requests
type {Entity}Handler struct {
	service *service.{Entity}Service
}

// New{Entity}Handler creates a new {entity} handler
func New{Entity}Handler(service *service.{Entity}Service) *{Entity}Handler {
	return &{Entity}Handler{service: service}
}

// Create{Entity} handles {entity} creation
// POST /api/{entities}
func (h *{Entity}Handler) Create{Entity}(c echo.Context) error {
	var req dto.Create{Entity}Request
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

	response, err := h.service.Create{Entity}(c.Request().Context(), &req)
	if err != nil {
		// Map service errors to HTTP status codes
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create {entity}",
		})
	}

	return c.JSON(http.StatusCreated, response)
}

// Get{Entity} handles retrieving a single {entity} by ID
// GET /api/{entities}/:id
func (h *{Entity}Handler) Get{Entity}(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid ID format",
		})
	}

	response, err := h.service.Get{Entity}ByID(c.Request().Context(), uint(id))
	if err != nil {
		if errors.Is(err, service.Err{Entity}NotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "{Entity} not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get {entity}",
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Standard CRUD handlers: GetAll, Update, Delete
// Custom handlers based on business logic
```

**Handler Rules**:
1. **Import `dto` package, NOT `model` package**
2. Extract request with `c.Bind(&req)` using DTO types
3. Validate with `c.Validate(&req)`
4. Pass `c.Request().Context()` to service
5. **Service returns DTOs, pass them directly to JSON response**
6. Map domain errors to HTTP status codes:
   - 400: Invalid input, validation errors
   - 404: Entity not found
   - 409: Conflict (duplicate key, etc.)
   - 500: Unexpected errors
7. Use consistent error response format
8. Include route documentation in comments
9. **DO NOT** call `.ToResponse()` - service already returns DTOs

### Step 6: Generate Router

**Location**: `src/internal/service/{service}/router/{entity_lowercase}_router.go`

**File Naming**: Use lowercase entity name with underscore (e.g., `product_router.go`, `category_router.go`)

**Router Function Naming**: Each entity gets its own registration function

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
	{entity}Handler *handler.{Entity}Handler,
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
1. Each entity has its own router file and registration function
2. Group routes by access level (public, protected, admin)
3. Apply middleware at group level, not per-route
4. Use RESTful naming conventions
5. Include logger for registration confirmation
6. Document middleware requirements
7. Create a shared `middleware.go` file for common middleware implementations (only if it doesn't exist)

### Step 7: Generate Module

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
		// Repositories
		repository.New{Entity}Repository,
		// Add more: repository.New{OtherEntity}Repository,
		
		// Services
		service.New{Entity}Service,
		// Add more: service.New{OtherEntity}Service,
		
		// Handlers
		handler.New{Entity}Handler,
		// Add more: handler.New{OtherEntity}Handler,
	),
)
```

**Module Rules**:
1. Use `fx.Provide` for constructor functions
2. Export as package-level `Module` variable
3. Keep in order: repository → service → handler
4. Add all entity constructors in the Module:
   - `repository.New{Entity}Repository`
   - `service.New{Entity}Service`
   - `handler.New{Entity}Handler`
5. One module file per service (contains all entities)

### Step 8: Update Migration

**Location**: `src/internal/service/{service}/migration/migration.go`

**Update Pattern**:

```go
// RunMigrations runs database migrations for {service} service
func RunMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.{Entity}{}); err != nil {
		return fmt.Errorf("failed to migrate {entity} table: %w", err)
	}
	
	// Add other entities...
	
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	return nil
}

// createIndexes - Add entity-specific indexes
// - Index on unique fields (code, email, sku)
// - Index on searchable fields (name, type, status)
// - Composite indexes for common queries
```

**Migration Rules**:
1. Add `db.AutoMigrate(&model.{Entity}{})` for new entity
2. Add relevant indexes in `createIndexes()` function
3. Update `Seed()` function if sample data is needed
4. Add comments for PostgreSQL-specific features (full-text search, etc.)

### Step 9: Update or Create App

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
	// Add more: fx.Invoke({service}router.Register{OtherEntity}Routes),
)
```

**App Rules**:
1. Import infrastructure modules (config, logger, database, server)
2. Import service module with alias: `{service}module`
3. Import router with alias: `{service}router`
4. Use `fx.Invoke` for each entity's router registration function
5. Create as package-level `AppModule` variable
6. Add one `fx.Invoke` per entity route registration

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

- [ ] If user says "create model", generate model file first (Step 0)
- [ ] Create DTO file immediately after model (Step 1)
- [ ] Model file exists and is readable (or just created)
- [ ] DTO file exists and is readable (or just created)
- [ ] Target directories exist (`model/`, `dto/`, `handler/`, `service/`, etc.)
- [ ] Extract entity name, table name, and all fields
- [ ] Identify unique constraints for existence checks
- [ ] Identify searchable/filterable fields
- [ ] Check if service already exists (update vs create)

## Post-Generation Checklist

After generating code, verify:

### File Organization
- [ ] Each entity has its own model file (`{entity}.go`)
- [ ] Each entity has its own DTO file (`{entity}_dto.go`)
- [ ] Each entity has its own repository file (`{entity}_repository.go`)
- [ ] Each entity has its own service file (`{entity}_service.go`)
- [ ] Each entity has its own handler file (`{entity}_handler.go`)
- [ ] Each entity has its own router file (`{entity}_router.go`)
- [ ] File names use lowercase with underscores

### Naming Conventions
- [ ] Struct names include entity prefix (`{Entity}Repository`, `{Entity}Service`, `{Entity}Handler`)
- [ ] Constructor names match struct names (`New{Entity}Repository`, `New{Entity}Service`)
- [ ] Entity name is consistently PascalCase throughout
- [ ] Package names match directory structure

### DTO Separation
- [ ] Model file contains ONLY entity struct and TableName() method
- [ ] DTO file contains request/response structs and converters
- [ ] Service imports `dto` package, NOT `model` package directly for DTOs
- [ ] Handler imports `dto` package, NOT `model` package
- [ ] Service returns DTOs (`dto.{Entity}Response`), not entities
- [ ] Handler receives DTOs from service and passes directly to JSON response

### Dependency Injection
- [ ] Repository constructor added to module (`repository.New{Entity}Repository`)
- [ ] Service constructor added to module (`service.New{Entity}Service`)
- [ ] Handler constructor added to module (`handler.New{Entity}Handler`)
- [ ] Router registration added to app (`fx.Invoke({service}router.Register{Entity}Routes)`)

### Code Quality
- [ ] All imports are correct
- [ ] Error handling follows patterns (wrap with context, use sentinel errors)
- [ ] All methods have context as first parameter
- [ ] DTO converter functions handle nil entities safely
- [ ] Service uses `dto.To{Entity}Response()` and `dto.To{Entity}ResponseList()`
- [ ] Handler does NOT call `.ToResponse()` (service already returns DTOs)
- [ ] Migration file updated with new entity
- [ ] Indexes added for unique and searchable fields

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

### Scenario: Adding a Category entity to the product service

**User Request**: "Create Category model (id, name, type, code) and generate complete service"

**AI Actions**:

1. **Create Model** (`model/category.go`) - **NEW SEPARATE FILE**:
   - Create `Category` struct with fields:
     - ID, CreatedAt, UpdatedAt, DeletedAt (base fields)
     - Name (varchar 255, not null, searchable)
     - Code (varchar 50, unique index, not null)
     - Type (varchar 50, index, for filtering)
     - Description (text, optional)
     - IsActive (bool, default true)
   - Implement `TableName()` method returning "categories"
   - Create `CreateCategoryRequest` with validation tags
   - Create `UpdateCategoryRequest` with optional pointer fields
   - Create `CategoryResponse` struct
   - Implement `ToResponse()` method

2. **Create Repository** (`repository/category_repository.go`):
   ```
   - CategoryRepository struct
   - NewCategoryRepository() constructor
   - GetByCode() - unique field query
   - GetByType() - filter query
   - CodeExists() - uniqueness check
   - SearchCategories() - search functionality
   ```

3. **Create Service** (`service/category_service.go`):
   ```
   - ErrCategoryNotFound, ErrCodeExists - domain errors
   - CategoryService struct
   - NewCategoryService() constructor
   - CreateCategory() - with validation
   - GetCategoryByID(), GetCategoryByCode()
   - GetAllCategories(), GetCategoriesByType()
   - UpdateCategory(), DeleteCategory()
   ```

4. **Create Handler** (`handler/category_handler.go`):
   ```
   - CategoryHandler struct
   - NewCategoryHandler() constructor
   - CreateCategory(), GetCategory(), GetCategories()
   - UpdateCategory(), DeleteCategory()
   ```

5. **Create Router** (`router/category_router.go`):
   ```
   - RegisterCategoryRoutes() function
   - Public, protected, admin route groups
   ```

6. **Update Module** (`module/module.go`):
   ```go
   fx.Provide(
       repository.NewProductRepository,
       repository.NewCategoryRepository,  // ADD THIS
       service.NewProductService,
       service.NewCategoryService,        // ADD THIS
       handler.NewProductHandler,
       handler.NewCategoryHandler,        // ADD THIS
   )
   ```

7. **Update App** (`app.go`):
   ```go
   fx.Invoke(productrouter.RegisterProductRoutes),
   fx.Invoke(productrouter.RegisterCategoryRoutes),  // ADD THIS
   ```

**Result**: Clean, maintainable file structure:
```
product/
├── model/
│   ├── product.go
│   └── category.go                 ← NEW MODEL FILE (entity only)
├── dto/
│   ├── product_dto.go
│   └── category_dto.go             ← NEW DTO FILE (request/response DTOs)
├── repository/
│   ├── product_repository.go
│   └── category_repository.go      ← NEW FILE
├── service/
│   ├── product_service.go
│   └── category_service.go         ← NEW FILE (imports dto, returns DTOs)
├── handler/
│   ├── product_handler.go
│   └── category_handler.go         ← NEW FILE (imports dto)
├── router/
│   ├── product_router.go
│   ├── category_router.go          ← NEW FILE
│   └── middleware.go
├── module/
│   └── module.go                   ← UPDATED
├── migration/
│   └── migration.go                ← UPDATED (add Category to AutoMigrate)
└── app.go                          ← UPDATED
```

## File Naming Conventions

**IMPORTANT: Each entity gets its own file for better separation of concerns**

- **Model**: `{entity_lowercase}.go` (e.g., `product.go`, `category.go`) - **ONE FILE PER ENTITY, ENTITY ONLY**
- **DTO**: `{entity_lowercase}_dto.go` (e.g., `product_dto.go`, `category_dto.go`) - **ONE FILE PER ENTITY**
- **Repository**: `{entity_lowercase}_repository.go` (e.g., `product_repository.go`, `category_repository.go`)
- **Service**: `{entity_lowercase}_service.go` (e.g., `product_service.go`, `category_service.go`)
- **Handler**: `{entity_lowercase}_handler.go` (e.g., `product_handler.go`, `category_handler.go`)
- **Router**: `{entity_lowercase}_router.go` (e.g., `product_router.go`, `category_router.go`)
- **Router Middleware**: `middleware.go` (shared middleware, create only once)
- **Module**: `module.go` (single file, registers all entities)
- **App**: `app.go` (at service root, single file)

**Entity Naming Convention**:
- Struct names: Use entity name with suffix (e.g., `ProductRepository`, `CategoryService`, `OrderHandler`)
- Constructor names: Use New prefix with full struct name (e.g., `NewProductRepository`, `NewCategoryService`)
- File names: Use lowercase with underscores (e.g., `product_repository.go`, `user_service.go`)

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

| Component | File Name | Struct Name | Constructor | Responsibility |
|-----------|-----------|-------------|-------------|----------------|
| Model | `{entity}.go` | `{Entity}` | - | Database entity ONLY (GORM tags) |
| DTO | `{entity}_dto.go` | `Create{Entity}Request`, `Update{Entity}Request`, `{Entity}Response` | - | API contracts (request/response) |
| Repository | `{entity}_repository.go` | `{Entity}Repository` | `New{Entity}Repository()` | Data access layer |
| Service | `{entity}_service.go` | `{Entity}Service` | `New{Entity}Service()` | Business logic (accepts & returns DTOs) |
| Handler | `{entity}_handler.go` | `{Entity}Handler` | `New{Entity}Handler()` | HTTP handlers (uses DTOs) |
| Router | `{entity}_router.go` | - | `Register{Entity}Routes()` | Route registration |
| Middleware | `middleware.go` | - | Various | Shared middleware (create once) |
| Module | `module.go` | - | - | fx.Provide all constructors |
| App | `app.go` | - | - | fx.Invoke all route registrations |
| Migration | `migration.go` | - | - | AutoMigrate all entities |

## File Organization Strategy

### ❌ OLD APPROACH (Don't Use)
```
product/
├── model/
│   └── product.go             ← Entity + DTOs mixed together
├── repository/
│   └── repository.go          ← All repositories in one file
├── service/
│   └── service.go             ← All services in one file, returns entities
├── handler/
│   └── handler.go             ← All handlers in one file
└── router/
    └── router.go              ← All routes in one file
```

### ✅ NEW APPROACH (Use This)
```
product/
├── model/
│   ├── product.go                 ← Entity ONLY (GORM tags)
│   └── category.go                ← Entity ONLY (GORM tags)
├── dto/
│   ├── product_dto.go             ← Request/Response DTOs
│   └── category_dto.go            ← Request/Response DTOs
├── repository/
│   ├── product_repository.go      ← One file per entity
│   └── category_repository.go     ← One file per entity
├── service/
│   ├── product_service.go         ← Imports dto, returns DTOs
│   └── category_service.go        ← Imports dto, returns DTOs
├── handler/
│   ├── product_handler.go         ← Imports dto
│   └── category_handler.go        ← Imports dto
└── router/
    ├── product_router.go          ← One file per entity
    ├── category_router.go         ← One file per entity
    └── middleware.go              ← Shared middleware (one file)
```

### When to Create New Files
- **Always** create separate files for each entity within model/, dto/, repository/, service/, handler/, and router/ directories
- Example: When user says "create model Category" or adding a `Category` entity to the `product` service:
  - Create `model/category.go` (entity struct only)
  - Create `dto/category_dto.go` (request/response DTOs)
  - Create `repository/category_repository.go`
  - Create `service/category_service.go` (imports dto package)
  - Create `handler/category_handler.go` (imports dto package)
  - Create `router/category_router.go`

### When to Update Existing Files
- **Module** (`module/module.go`): Update to add new constructor functions
- **App** (`app.go`): Update to add new router registration calls
- **Middleware** (`router/middleware.go`): Only create once; reuse for all entities

### Entity Struct and Function Naming
- Repository: `{Entity}Repository`, `New{Entity}Repository()`
- Service: `{Entity}Service`, `New{Entity}Service()`
- Handler: `{Entity}Handler`, `New{Entity}Handler()`
- Router: `Register{Entity}Routes()`

### Benefits of Separate Files
1. **Maintainability**: Easier to locate and modify entity-specific code
2. **Readability**: Smaller, focused files are easier to understand
3. **Collaboration**: Reduces merge conflicts when multiple developers work on different entities
4. **Testing**: Easier to write and organize entity-specific tests
5. **Scalability**: Clean separation as the service grows

## Additional Notes

- Always read the golang-patterns skill before generating code
- **ALWAYS create separate files for each entity** (model, dto, repository, service, handler, router)
- **When user says "create model", automatically create BOTH model and DTO files**:
  - `model/{entity}.go` - Entity struct only
  - `dto/{entity}_dto.go` - Request/Response DTOs
- **Service layer MUST import dto package and return DTOs, NOT entities**
- **Handler layer MUST import dto package and work with DTOs only**
- **Models stay internal to repository and service layers**
- Generate code following existing project conventions
- Use the same error handling patterns as existing services
- Maintain consistency with existing middleware patterns
- Include comprehensive comments for generated code
- Generate complete CRUD operations by default
- Add custom methods based on model field analysis
- Keep file names lowercase with underscores for Go conventions
- After creating model and DTO, automatically proceed to generate complete service layers (Steps 2-9)
- **Clean Architecture**: Model → Repository → Service (converts to DTO) → Handler → Client
