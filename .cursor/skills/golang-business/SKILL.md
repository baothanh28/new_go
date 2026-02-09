---
name: golang-business
description: Generate complete business service layers (model, repository, service, handler, router, module, app) from a model definition. Use when creating new business entities, adding CRUD operations, or when the user mentions generating service components, creating new models, or scaffolding business logic.
---

# Go Business Service Generator

Automatically generate complete service layers following clean architecture patterns with dependency injection (fx), proper error handling, and idiomatic Go code.

## 🎯 Key Principle: One Entity = One File

**IMPORTANT**: Each business entity must have its own separate file in repository/, service/, handler/, and router/ directories.

**File Naming**: Use lowercase with underscores: `{entity}_repository.go`, `{entity}_service.go`, `{entity}_handler.go`, `{entity}_router.go`

**Struct Naming**: Use entity name as prefix: `{Entity}Repository`, `{Entity}Service`, `{Entity}Handler`

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

### Step 3: Generate Service

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
	"myapp/internal/service/{service}/model"
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
func (s *{Entity}Service) Create{Entity}(ctx context.Context, req *model.Create{Entity}Request) (*model.{Entity}, error) {
	// Add business validations here
	// Check unique constraints
	// Transform request to entity
	// Call repository
	// Wrap errors with context
}

// Get{Entity}ByID retrieves {entity} by ID
func (s *{Entity}Service) Get{Entity}ByID(ctx context.Context, id uint) (*model.{Entity}, error) {
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
	"myapp/internal/service/{service}/model"
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

### Step 7: Update Migration

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

### Step 8: Update or Create App

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

- [ ] Model file exists and is readable
- [ ] Target directories exist (`handler/`, `service/`, etc.)
- [ ] Extract entity name, table name, and all fields
- [ ] Identify unique constraints for existence checks
- [ ] Identify searchable/filterable fields
- [ ] Check if service already exists (update vs create)

## Post-Generation Checklist

After generating code, verify:

### File Organization
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

### Dependency Injection
- [ ] Repository constructor added to module (`repository.New{Entity}Repository`)
- [ ] Service constructor added to module (`service.New{Entity}Service`)
- [ ] Handler constructor added to module (`handler.New{Entity}Handler`)
- [ ] Router registration added to app (`fx.Invoke({service}router.Register{Entity}Routes)`)

### Code Quality
- [ ] All imports are correct
- [ ] Error handling follows patterns (wrap with context, use sentinel errors)
- [ ] All methods have context as first parameter
- [ ] Response methods use `ToResponse()`
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

1. **Create/Update Model** (`model/model.go`):
   - Add `Category` struct
   - Add `CreateCategoryRequest`, `UpdateCategoryRequest`, `CategoryResponse`
   - Add `ToResponse()` method

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
│   └── model.go
├── repository/
│   ├── product_repository.go
│   └── category_repository.go      ← NEW FILE
├── service/
│   ├── product_service.go
│   └── category_service.go         ← NEW FILE
├── handler/
│   ├── product_handler.go
│   └── category_handler.go         ← NEW FILE
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

- Model: `model.go` (all models in one file per service, OR separate files per entity)
- Repository: `{entity_lowercase}_repository.go` (e.g., `product_repository.go`, `category_repository.go`)
- Service: `{entity_lowercase}_service.go` (e.g., `product_service.go`, `category_service.go`)
- Handler: `{entity_lowercase}_handler.go` (e.g., `product_handler.go`, `category_handler.go`)
- Router: `{entity_lowercase}_router.go` (e.g., `product_router.go`, `category_router.go`)
- Router Middleware: `middleware.go` (shared middleware, create only once)
- Module: `module.go` (single file, registers all entities)
- App: `app.go` (at service root, single file)

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
| Model | `model.go` | `{Entity}` | - | Data structures and DTOs |
| Repository | `{entity}_repository.go` | `{Entity}Repository` | `New{Entity}Repository()` | Data access layer |
| Service | `{entity}_service.go` | `{Entity}Service` | `New{Entity}Service()` | Business logic layer |
| Handler | `{entity}_handler.go` | `{Entity}Handler` | `New{Entity}Handler()` | HTTP request handlers |
| Router | `{entity}_router.go` | - | `Register{Entity}Routes()` | Route registration |
| Middleware | `middleware.go` | - | Various | Shared middleware (create once) |
| Module | `module.go` | - | - | fx.Provide all constructors |
| App | `app.go` | - | - | fx.Invoke all route registrations |
| Migration | `migration.go` | - | - | AutoMigrate all entities |

## File Organization Strategy

### ❌ OLD APPROACH (Don't Use)
```
product/
├── repository/
│   └── repository.go          ← All repositories in one file
├── service/
│   └── service.go             ← All services in one file
├── handler/
│   └── handler.go             ← All handlers in one file
└── router/
    └── router.go              ← All routes in one file
```

### ✅ NEW APPROACH (Use This)
```
product/
├── repository/
│   ├── product_repository.go      ← One file per entity
│   └── category_repository.go     ← One file per entity
├── service/
│   ├── product_service.go         ← One file per entity
│   └── category_service.go        ← One file per entity
├── handler/
│   ├── product_handler.go         ← One file per entity
│   └── category_handler.go        ← One file per entity
└── router/
    ├── product_router.go          ← One file per entity
    ├── category_router.go         ← One file per entity
    └── middleware.go              ← Shared middleware (one file)
```

### When to Create New Files
- **Always** create separate files for each entity within repository/, service/, handler/, and router/ directories
- Example: When adding a `Category` entity to the `product` service:
  - Create `repository/category_repository.go`
  - Create `service/category_service.go`
  - Create `handler/category_handler.go`
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
- **ALWAYS create separate files for each entity** (repository, service, handler, router)
- Generate code following existing project conventions
- Use the same error handling patterns as existing services
- Maintain consistency with existing middleware patterns
- Include comprehensive comments for generated code
- Generate complete CRUD operations by default
- Add custom methods based on model field analysis
- Keep file names lowercase with underscores for Go conventions
