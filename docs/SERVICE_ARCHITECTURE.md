# Service Architecture

## Overview

Each service now has its own `app.go` file that defines a complete, self-contained application module. This architecture supports both:

1. **Monolithic deployment** - All services running together in a single process
2. **Microservices deployment** - Each service running independently

## Directory Structure

```
src/
├── cmd/
│   ├── api/                    # Monolithic API server (all services)
│   │   └── main.go
│   ├── auth-service/           # Standalone auth service
│   │   └── main.go
│   ├── product-service/        # Standalone product service
│   │   └── main.go
│   └── health-service/         # Standalone health service
│       └── main.go
├── internal/
│   ├── module/
│   │   └── app.go             # Combines all services for monolithic deployment
│   └── service/
│       ├── auth/
│       │   ├── app.go         # Auth service app module
│       │   ├── router.go      # Auth route registration
│       │   ├── module.go      # Auth dependencies
│       │   ├── handler.go
│       │   ├── service.go
│       │   ├── repository.go
│       │   └── migration.go
│       ├── product/
│       │   ├── app.go         # Product service app module
│       │   ├── router.go      # Product route registration
│       │   ├── module/
│       │   │   └── module.go  # Product dependencies
│       │   ├── handler/
│       │   ├── service/
│       │   ├── repository/
│       │   └── migration/
│       └── health/
│           ├── app.go         # Health service app module
│           ├── router.go      # Health route registration
│           ├── module.go      # Health dependencies
│           └── handler.go
```

## Service App Modules

Each service has an `app.go` file that includes:

### 1. **Infrastructure Modules**
- Configuration (config)
- Logging (logger)
- Database (if needed)
- HTTP Server (server)

### 2. **Service Module**
- Service-specific dependencies (repositories, services, handlers)

### 3. **Route Registration**
- Service-specific route registration function

### Example: Auth Service (`service/auth/app.go`)

```go
package auth

import (
	"go.uber.org/fx"
	"myapp/internal/pkg/config"
	"myapp/internal/pkg/database"
	"myapp/internal/pkg/logger"
	"myapp/internal/pkg/server"
)

var AppModule = fx.Options(
	// Infrastructure
	config.Module,
	logger.Module,
	database.Module,
	server.Module,
	
	// Auth service
	Module,
	
	// Routes
	fx.Invoke(RegisterAuthRoutes),
)
```

## Running Services

### Monolithic Mode (All Services Together)

Run all services in a single process:

```bash
# Build
go build -o bin/api ./src/cmd/api

# Run
./bin/api serve

# Run migrations for all services
./bin/api migrate
```

### Microservices Mode (Services Independent)

Run each service independently:

```bash
# Auth Service
go build -o bin/auth-service ./src/cmd/auth-service
./bin/auth-service serve

# Product Service
go build -o bin/product-service ./src/cmd/product-service
./bin/product-service serve

# Health Service
go build -o bin/health-service ./src/cmd/health-service
./bin/health-service serve
```

## Benefits

### 1. **Flexibility**
- Start with monolithic deployment for simplicity
- Migrate to microservices as needed
- No code changes required to switch deployment modes

### 2. **Development**
- Run only the service you're working on
- Faster build and restart times
- Easier debugging and testing

### 3. **Testing**
- Each service can be tested in isolation
- Easier integration testing
- Better test coverage

### 4. **Deployment**
- Deploy services independently
- Scale services based on load
- Independent versioning and releases

## Adding a New Service

To add a new service (e.g., `order`):

1. **Create service directory structure:**
   ```
   internal/service/order/
   ├── app.go           # Service app module
   ├── router.go        # Route registration
   ├── module.go        # Service dependencies
   ├── handler.go
   ├── service.go
   ├── repository.go    # If database needed
   └── migration.go     # If database needed
   ```

2. **Create `app.go`:**
   ```go
   package order

   import (
       "go.uber.org/fx"
       "myapp/internal/pkg/config"
       "myapp/internal/pkg/database"
       "myapp/internal/pkg/logger"
       "myapp/internal/pkg/server"
   )

   var AppModule = fx.Options(
       config.Module,
       logger.Module,
       database.Module,
       server.Module,
       Module,
       fx.Invoke(RegisterOrderRoutes),
   )
   ```

3. **Create `router.go`:**
   ```go
   package order

   import (
       "github.com/labstack/echo/v4"
       "go.uber.org/zap"
   )

   func RegisterOrderRoutes(
       e *echo.Echo,
       handler *Handler,
       logger *zap.Logger,
   ) {
       logger.Info("Registering order routes")
       api := e.Group("/api/orders")
       api.GET("", handler.ListOrders)
       api.POST("", handler.CreateOrder)
       logger.Info("Order routes registered")
   }
   ```

4. **Create `module.go`:**
   ```go
   package order

   import "go.uber.org/fx"

   var Module = fx.Options(
       fx.Provide(
           NewRepository,
           NewService,
           NewHandler,
       ),
   )
   ```

5. **Add to monolithic app** (`internal/module/app.go`):
   ```go
   import "myapp/internal/service/order"

   var AppModule = fx.Options(
       // ... existing modules
       order.Module,
       fx.Invoke(order.RegisterOrderRoutes),
   )
   ```

6. **Create standalone command** (`cmd/order-service/main.go`):
   ```go
   package main

   import (
       "go.uber.org/fx"
       "myapp/internal/service/order"
   )

   func runServe(cmd *cobra.Command, args []string) error {
       app := fx.New(
           order.AppModule,
           fx.NopLogger,
       )
       // ... start/stop logic
   }
   ```

## Migration Strategy

### Phase 1: Monolithic (Current)
- All services in `cmd/api`
- Single deployment
- Shared infrastructure

### Phase 2: Hybrid
- Keep monolith for most services
- Extract high-load services (e.g., product)
- Gradual migration

### Phase 3: Microservices
- All services independent
- Service mesh for communication
- Distributed deployment

## Best Practices

1. **Keep services independent** - Each service should manage its own domain
2. **Avoid cross-service imports** - Services shouldn't depend on each other directly
3. **Use interfaces** - Define clear contracts between services
4. **Consistent structure** - All services follow the same directory layout
5. **Database per service** - Each service has its own database/schema (if needed)

## Configuration

Each service can have its own configuration while sharing common infrastructure config:

```yaml
# config/config.yaml
server:
  port: 8080
  
database:
  master:
    host: localhost
    
services:
  auth:
    jwt_secret: "secret"
  product:
    cache_ttl: 3600
  health:
    check_interval: 30
```

## Conclusion

This architecture provides maximum flexibility - start simple with a monolith, and evolve to microservices as your application grows. Each service is self-contained and can be deployed independently without code changes.
