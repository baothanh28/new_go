# Golang Web API Development Requirements

## Project Overview
I need help designing and implementing a RESTful Web API in Golang with clean architecture principles and dependency injection.

## Technology Stack Requirements

### Core Framework & Libraries
- **Dependency Injection**: Uber FX (`go.uber.org/fx`)
- **HTTP Server**: Echo framework (`github.com/labstack/echo/v4`)
- **CLI Framework**: Cobra (`github.com/spf13/cobra`)
- **Logger**: Structured logging (prefer `go.uber.org/zap` or `github.com/sirupsen/logrus`)


## Go Idioms and Best Practices (CRITICAL)

**These principles must be followed throughout the entire codebase:**

1. **Error Handling**: Always wrap errors with context using `fmt.Errorf` with `%w`
   - Never ignore errors (no blank identifier `_`)
   - Return errors early, keep happy path unindented

2. **Context Usage**: Pass `context.Context` as first parameter in all functions
   - Never store context in structs
   - Use `context.Background()` for top-level operations
   - Pass context through the entire call chain

3. **Interface Design**: Accept interfaces, return structs
   - Define interfaces where they're used (consumer side)
   - Keep interfaces small and focused

4. **Zero Value Usefulness**: Design types to work with zero values
   - Example: `sync.Mutex`, `bytes.Buffer` work without initialization

5. **Dependency Injection**: No global mutable state
   - Use FX for dependency injection
   - Pass dependencies through constructors

6. **Naming Conventions**:
   - Package names: short, lowercase, single word
   - Constructors: Use `New` prefix
   - No stuttering: `user.Service` not `user.UserService`

7. **Code Organization**:
   - Simple is better than clever
   - Clear is better than concise
   - Explicit is better than implicit

## Key Requirements

### Service Architecture Flexibility
**IMPORTANT**: Services have flexible architecture patterns:
- **Not all services require database access** - Only include repository layer when needed
- **Database-backed services**: Repository → Service → Handler (3 layers)
- **External API services**: Service → Handler (2 layers, no repository)
- **Stateless services**: Handler only (1 layer, no service/repository)
- FX module should only provide the layers that are actually needed
- Examples of services without database: health checks, email, SMS, payment gateways, file uploads, validators

### Database Provider Conditional Infrastructure
**CRITICAL**: Infrastructure code generation MUST be conditional based on the database provider configuration:
- **PostgreSQL provider (`"postgres"`)** → Generate PostgreSQL-specific infrastructure
  - Import: `gorm.io/driver/postgres`
  - DSN format: PostgreSQL connection string
  - Optimizations: PostgreSQL-specific settings
  
- **MySQL provider (`"mysql"`)** → Generate MySQL-specific infrastructure
  - Import: `gorm.io/driver/mysql`
  - DSN format: MySQL connection string
  - Optimizations: MySQL-specific settings

- Master and tenant databases can each specify their own provider
- Only import the driver packages that are actually being used
- Generate the appropriate infrastructure code based on the provider(s) specified in configuration

## Functional Requirements

### 1. Dependency Injection with FX

**Main Application Module** (`internal/pkg/module/app.go`):
```go
package module

import (
    "go.uber.org/fx"
    "myapp/internal/pkg/config"
    "myapp/internal/pkg/logger"
    "myapp/internal/pkg/database"
    "myapp/internal/pkg/server"
    "myapp/internal/router"
    "myapp/internal/service/product"
    "myapp/internal/service/auth"
)

var AppModule = fx.Options(
    // Infrastructure modules
    config.Module,
    logger.Module,
    database.Module,  // Optional: only if services need database
    server.Module,
    
    // Service modules (mix of database-backed and stateless services)
    product.Module,   // Has database: Repository → Service → Handler
    auth.Module,      // Has database: Repository → Service → Handler
    email.Module,     // No database: Service → Handler (external API)
    health.Module,    // No database: Handler only (stateless)
    
    // Router module (registers all routes)
    router.Module,
)
```

**Module Structure**:
- Each infrastructure component (config, logger, database, server) has its own `module.go` in `internal/pkg/{component}/`
- Each service has its own `module.go` that provides repositories, services, and handlers
- **Not all services require database access** - Some services may:
  - Be stateless (e.g., utility services, formatters, validators)
  - Use external APIs only (e.g., payment gateways, email services)
  - Work with in-memory data or caching only
  - Only depend on other services without direct database access
  - In these cases, the service module should NOT include repository layer
  - Service layer can be injected directly without repository dependency
- Router module registers all service routes onto the Echo server
- FX manages the entire application lifecycle and dependency graph
- Support graceful shutdown

**FX Constructor Best Practices**:
```go
// Good: Accept interfaces, return concrete types
func NewUserService(repo UserRepository, logger *zap.Logger) *UserService {
    return &UserService{
        repo:   repo,
        logger: logger,
    }
}

// Good: Use functional options for complex configuration
type ServerOption func(*Server)

func NewServer(cfg *Config, opts ...ServerOption) *Server {
    s := &Server{
        addr:    cfg.Addr,
        timeout: 30 * time.Second,
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}
```

**Graceful Shutdown with FX**:
```go
func RegisterHooks(lc fx.Lifecycle, server *echo.Echo, logger *zap.Logger) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                logger.Info("Starting HTTP server", zap.String("addr", server.Server.Addr))
                if err := server.Start(":8080"); err != nil && err != http.ErrServerClosed {
                    logger.Fatal("Server startup failed", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            logger.Info("Shutting down server")
            return server.Shutdown(ctx)
        },
    })
}
```

**Middleware Chain Order** (Important):
1. Logger middleware (request/response logging)
2. Context middleware (tenant/master detection) - **MUST be early in chain**
3. JWT authentication middleware (for protected routes)
4. Other middleware (CORS, rate limiting, etc.)

**Database Module Requirements**:
- Database module provides `DatabaseManager` with both `MasterDB` and `TenantDB`
- Configuration supports separate master and tenant database configs
- Each database connection can use different drivers (postgres or mysql)
- Context middleware selects appropriate database based on request headers

### 2. HTTP Server with Echo
- Setup Echo server in `internal/pkg/server/` with proper middleware stack
- Echo instance is provided by FX and injected into router and handlers
- Implement RESTful endpoints
- Add request validation using struct tags
- Include error handling middleware with proper error wrapping
- Support JSON request/response

**Error Handling in Handlers**:
- Always wrap errors with context: `fmt.Errorf("create user: %w", err)`
- Return appropriate HTTP status codes
- Use custom error types for domain-specific errors
- Example:
```go
func (h *UserHandler) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("bind request: %w", err))
    }
    
    user, err := h.service.CreateUser(c.Request().Context(), &req)
    if err != nil {
        return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("create user: %w", err))
    }
    
    return c.JSON(http.StatusCreated, user)
}
```

**Middleware Stack Setup**:
```go
// Global middleware (applied to all routes)
e.Use(middleware.Logger())          // 1. Request logging
e.Use(ContextMiddleware(dbManager)) // 2. Tenant/Master context
e.Use(middleware.Recover())         // 3. Panic recovery
e.Use(middleware.CORS())            // 4. CORS handling

// Protected routes group
protected := e.Group("/api")
protected.Use(JWTMiddleware(jwtConfig)) // JWT auth for protected routes
```

- Context middleware must be registered early to ensure database context is available
- JWT middleware is applied only to protected route groups
- All handlers can access request context via `c.Get("requestContext")`

### 3. Tenant/Master Context Middleware

**Purpose**: Determine request context type (tenant or master) based on HTTP headers before processing request

**Implementation** (`internal/pkg/middleware/context_middleware.go`):
- Read custom headers to identify request type:
  - `X-Tenant-ID`: Identifies tenant-specific requests
  - `X-Request-Type`: Explicitly marks "tenant" or "master" requests
  
- Create request context with database selection:
```go
type RequestContext struct {
    Type     string    // "tenant" or "master"
    TenantID string    // Empty for master requests
    Database *gorm.DB  // Selected database connection
}
```

- Middleware logic:
```go
func ContextMiddleware(dbManager *DatabaseManager) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            tenantID := c.Request().Header.Get("X-Tenant-ID")
            requestType := c.Request().Header.Get("X-Request-Type")
            
            ctx := &RequestContext{}
            
            if requestType == "master" || tenantID == "" {
                ctx.Type = "master"
                ctx.Database = dbManager.MasterDB
            } else {
                ctx.Type = "tenant"
                ctx.TenantID = tenantID
                ctx.Database = dbManager.TenantDB
            }
            
            c.Set("requestContext", ctx)
            return next(c)
        }
    }
}
```

- Apply middleware globally in server setup
- Handlers access context via `c.Get("requestContext").(*RequestContext)`
- Repositories use the appropriate database from request context

### 4. JWT Authentication Middleware
- **CRITICAL**: Implement JWT middleware in `internal/service/auth/jwt_middleware.go`
- Middleware must parse and validate JWT tokens from Authorization header
- After successful validation, inject user information into Echo context
- User info should be accessible in handlers via `c.Get("user")` or similar pattern

**Example Implementation**:
```go
type UserContext struct {
    UserID   uint   `json:"user_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

func JWTMiddleware(jwtSecret string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            token := extractToken(c.Request())
            if token == "" {
                return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization token")
            }
            
            claims, err := validateToken(token, jwtSecret)
            if err != nil {
                return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("invalid token: %w", err))
            }
            
            userCtx := &UserContext{
                UserID: claims.UserID,
                Email:  claims.Email,
                Role:   claims.Role,
            }
            
            c.Set("user", userCtx)
            return next(c)
        }
    }
}

// Helper function to extract user from context
func GetUserFromContext(c echo.Context) (*UserContext, error) {
    user, ok := c.Get("user").(*UserContext)
    if !ok || user == nil {
        return nil, errors.New("user not found in context")
    }
    return user, nil
}
```

- Protected routes should use this middleware to ensure authentication
- Always use the helper function to safely extract user from context
- Wrap errors properly when token validation fails

### 5. Database Layer with GORM

**Database Provider Infrastructure**:

**IMPORTANT**: Infrastructure generation must be conditional based on the database provider:
- **If provider is `postgres`**: Generate PostgreSQL-specific infrastructure:
  - Import `gorm.io/driver/postgres`
  - Use PostgreSQL DSN format: `host=%s port=%d user=%s password=%s dbname=%s sslmode=disable`
  - Include PostgreSQL-specific configuration and optimizations
  - Migration files should use PostgreSQL syntax
  
- **If provider is `mysql`**: Generate MySQL-specific infrastructure:
  - Import `gorm.io/driver/mysql`
  - Use MySQL DSN format: `%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local`
  - Include MySQL-specific configuration and optimizations
  - Migration files should use MySQL syntax

- Both master and tenant databases can use different providers independently
- Configuration must specify the driver type for each database
- Code generation should adapt to the specified provider(s)

**Database Connection Manager** (`internal/pkg/database/manager.go`):

Support multiple database types (PostgreSQL, MySQL) with tenant/master pattern:

```go
type DatabaseConfig struct {
    Driver   string // "postgres" or "mysql"
    Host     string
    Port     int
    Name     string
    User     string
    Password string
    // Connection pool settings
    MaxOpenConns int
    MaxIdleConns int
}

type DatabaseManager struct {
    MasterDB *gorm.DB
    TenantDB *gorm.DB
}
```

**Database Factory with Driver Support**:

**When generating infrastructure, adapt imports and code based on provider**:

If provider is **PostgreSQL only**:
```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        config.Host, config.Port, config.User, config.Password, config.Name)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("open postgres database %s: %w", config.Name, err)
    }
    
    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("get underlying database connection: %w", err)
    }
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    
    return db, nil
}
```

If provider is **MySQL only**:
```go
import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.User, config.Password, config.Host, config.Port, config.Name)
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("open mysql database %s: %w", config.Name, err)
    }
    
    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("get underlying database connection: %w", err)
    }
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    
    return db, nil
}
```

If **supporting both providers** (when master and tenant use different drivers):
```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
    var dialector gorm.Dialector
    
    switch config.Driver {
    case "postgres":
        dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
            config.Host, config.Port, config.User, config.Password, config.Name)
        dialector = postgres.Open(dsn)
        
    case "mysql":
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            config.User, config.Password, config.Host, config.Port, config.Name)
        dialector = mysql.Open(dsn)
        
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
    }
    
    db, err := gorm.Open(dialector, &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("open %s database %s: %w", config.Driver, config.Name, err)
    }
    
    // Configure connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, fmt.Errorf("get underlying database connection: %w", err)
    }
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    
    return db, nil
}
```

**Tenant/Master Database Support**:
- Maintain two database connection pools:
  - **Master DB**: For system-wide operations, admin functions, and non-tenant data
  - **Tenant DB**: For tenant-specific data isolation
- FX module provides both connections:
```go
fx.Provide(
    NewMasterDatabase,
    NewTenantDatabase,
    NewDatabaseManager,
)
```

**Base Repository** (`internal/pkg/database/repository.go`):
Implement a generic base repository with common CRUD operations that all service repositories can embed:

```go
type BaseRepository[T any] struct {
    db *gorm.DB
}

// Common operations to implement (context.Context as first parameter):
// Insert operations
- Insert(ctx context.Context, entity *T) error
- InsertBatch(ctx context.Context, entities []*T) error

// Update operations
- UpdateByID(ctx context.Context, id uint, entity *T) error
- UpdateWhere(ctx context.Context, id uint, updates map[string]interface{}) error

// Get operations
- GetByID(ctx context.Context, id uint) (*T, error)
- GetAll(ctx context.Context, limit, offset int) ([]*T, error)
- GetWhere(ctx context.Context, conditions map[string]interface{}) ([]*T, error)

// Delete operations
- DeleteByID(ctx context.Context, id uint) error
- DeleteWhere(ctx context.Context, conditions map[string]interface{}) error

// Query operations
- Count(ctx context.Context, conditions map[string]interface{}) (int64, error)
- Exists(ctx context.Context, conditions map[string]interface{}) (bool, error)

// Transaction support
- WithTx(tx *gorm.DB) *BaseRepository[T]
```

**Error Handling in Repository**:
- All repository methods should wrap errors with context using fmt.Errorf with %w
- Example: `return nil, fmt.Errorf("get user by id %d: %w", id, err)`
- Use GORM's context-aware methods: `db.WithContext(ctx).Find()`

**Service Repositories**:
- Each service repository embeds `BaseRepository` and adds domain-specific queries
- Example: `ProductRepository` embeds `BaseRepository[Product]`
- Repositories receive database connection from request context (via context middleware)
- Connection pooling configuration for both master and tenant databases
- Each service has its own `migration.go` file that defines migrations for its tables

**Services Without Database**:
- **IMPORTANT**: Not all services require database access
- Services may be stateless or depend on external resources only

**Service Module Patterns**:

*Pattern 1: Service with Database (Repository → Service → Handler)*
```go
// internal/service/product/module.go
var Module = fx.Options(
    fx.Provide(
        NewProductRepository,  // Repository layer
        NewProductService,     // Service layer (depends on repository)
        NewProductHandler,     // Handler layer (depends on service)
    ),
)
```

*Pattern 2: Service without Database (Service → Handler)*
```go
// internal/service/email/module.go
var Module = fx.Options(
    fx.Provide(
        // No repository layer needed
        NewEmailService,       // Service layer (uses external email API)
        NewEmailHandler,       // Handler layer (depends on service)
    ),
)
```

*Pattern 3: Stateless Service (Handler only)*
```go
// internal/service/health/module.go
var Module = fx.Options(
    fx.Provide(
        NewHealthHandler,      // Handler only (no service/repository needed)
    ),
)
```

**Examples of Services Without Database**:
- **Health Check Service**: Returns system status, no data persistence
- **Email Service**: Uses external SMTP/email API
- **SMS Service**: Uses external SMS gateway
- **Payment Service**: Integrates with payment providers (Stripe, PayPal)
- **File Upload Service**: Handles file storage to cloud (S3, GCS)
- **Validation Service**: Stateless validation logic
- **Transform Service**: Data transformation without persistence
- **Cache Service**: In-memory or Redis-based caching
- **Notification Service**: Pushes to external notification systems

### 6. Configuration Management
- Config module in `internal/pkg/config/` provides configuration to all components
- Support configuration via file, environment variables, and CLI flags
- Use priority: CLI flags > Environment variables > Config file > Defaults
- Validate configuration on startup with proper error wrapping

**Configuration Structure**:
```go
type Config struct {
    Server         ServerConfig     `mapstructure:"server"`
    MasterDatabase DatabaseConfig   `mapstructure:"master_database"`
    TenantDatabase DatabaseConfig   `mapstructure:"tenant_database"`
    JWT            JWTConfig        `mapstructure:"jwt"`
    Logger         LoggerConfig     `mapstructure:"logger"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
    if err := c.Server.Validate(); err != nil {
        return fmt.Errorf("validate server config: %w", err)
    }
    if err := c.MasterDatabase.Validate(); err != nil {
        return fmt.Errorf("validate master database config: %w", err)
    }
    if err := c.TenantDatabase.Validate(); err != nil {
        return fmt.Errorf("validate tenant database config: %w", err)
    }
    if err := c.JWT.Validate(); err != nil {
        return fmt.Errorf("validate jwt config: %w", err)
    }
    return nil
}

// LoadConfig loads and validates configuration
func LoadConfig(configPath string) (*Config, error) {
    cfg := &Config{}
    
    if err := readConfigFile(configPath, cfg); err != nil {
        return nil, fmt.Errorf("read config file %s: %w", configPath, err)
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("validate config: %w", err)
    }
    
    return cfg, nil
}
```

- **MasterDatabase**: Configuration for master/system database
  - **Driver** (postgres/mysql) - **This determines which infrastructure to generate**
  - Connection details
  - Pool settings
  
- **TenantDatabase**: Configuration for tenant database
  - **Driver** (postgres/mysql) - **This determines which infrastructure to generate**
  - Connection details
  - Pool settings
  
- **Infrastructure Generation Rules**:
  - If **MasterDatabase.Driver = "postgres"**: Generate PostgreSQL infrastructure for master DB
  - If **MasterDatabase.Driver = "mysql"**: Generate MySQL infrastructure for master DB
  - If **TenantDatabase.Driver = "postgres"**: Generate PostgreSQL infrastructure for tenant DB
  - If **TenantDatabase.Driver = "mysql"**: Generate MySQL infrastructure for tenant DB
  - If both use the same driver: Only import one driver package
  - If they use different drivers: Import both driver packages and use switch/case logic
  
- Both databases can use different drivers if needed
- Config validation ensures required fields are present

### 7. Logging
- Logger module in `internal/pkg/logger/` provides structured logger to all components
- Use `go.uber.org/zap` or `github.com/sirupsen/logrus`
- Log levels: DEBUG, INFO, WARN, ERROR
- Request/response logging middleware (should be first in middleware chain)
- Correlation ID support for request tracing
- Logger is injected via FX into all services and handlers
- Log tenant/master context type for better debugging
