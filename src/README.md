# MyApp - Golang RESTful API

A clean architecture RESTful API built with Go, featuring dependency injection, JWT authentication, and tenant/master database pattern.

## Features

- **Clean Architecture** - Clear separation of concerns with layered architecture
- **Dependency Injection** - Using Uber FX for managing dependencies
- **JWT Authentication** - Secure token-based authentication
- **Tenant/Master Pattern** - Separate databases for system and tenant data
- **PostgreSQL** - Robust database with GORM ORM
- **Structured Logging** - Using Zap for high-performance logging
- **Graceful Shutdown** - Proper lifecycle management
- **Generic Repository** - Reusable CRUD operations for all entities

## Technology Stack

- **Framework**: Echo v4
- **Dependency Injection**: Uber FX
- **CLI**: Cobra
- **Logger**: Zap
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT (golang-jwt)

## Project Structure

```
src/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   ├── pkg/                           # Shared infrastructure
│   │   ├── config/                    # Configuration management
│   │   ├── logger/                    # Structured logging
│   │   ├── database/                  # Database connections
│   │   ├── server/                    # Echo HTTP server
│   │   └── middleware/                # Shared middleware
│   ├── service/                       # Business services
│   │   ├── auth/                      # Authentication service
│   │   └── health/                    # Health check service
│   ├── router/                        # Route registration
│   └── module/                        # Main app module
├── config/                            # Configuration files
│   └── config.yaml                    # Default configuration
├── go.mod                             # Go module definition
├── Makefile                           # Build commands
└── README.md                          # This file
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 14 or higher
- Docker and Docker Compose (optional, for local development)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd src
```

2. Install dependencies:
```bash
make deps
```

3. Set up PostgreSQL databases:

**Option A: Using Docker (recommended)**
```bash
docker-compose up -d
```

**Option B: Manual setup**
Create two PostgreSQL databases:
- `master_db` - for system-wide data (users, authentication)
- `tenant_db` - for tenant-specific data

4. Configure the application:

Edit `config/config.yaml` or set environment variables:
```bash
export MYAPP_SERVER_PORT=8080
export MYAPP_MASTER_DATABASE_HOST=localhost
export MYAPP_MASTER_DATABASE_NAME=master_db
export MYAPP_MASTER_DATABASE_USER=postgres
export MYAPP_MASTER_DATABASE_PASSWORD=password
# ... and so on
```

5. Run database migrations:
```bash
make migrate
```

6. Start the application:
```bash
make run
```

The API server will start at `http://localhost:8080`

## Configuration

Configuration can be provided through:
1. Configuration file (`config/config.yaml`)
2. Environment variables (prefix: `MYAPP_`, e.g., `MYAPP_SERVER_PORT`)
3. CLI flags

Priority: CLI flags > Environment variables > Config file > Defaults

### Configuration Structure

```yaml
server:
  host: "0.0.0.0"
  port: 8080

master_database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  name: "master_db"
  user: "postgres"
  password: "password"
  max_open_conns: 25
  max_idle_conns: 5

tenant_database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  name: "tenant_db"
  user: "postgres"
  password: "password"
  max_open_conns: 25
  max_idle_conns: 5

jwt:
  secret: "your-secret-key-change-in-production-must-be-at-least-32-characters"
  expiration_hours: 24

logger:
  level: "info"  # debug, info, warn, error
  format: "json" # json, console
```

## API Endpoints

### Health Check Endpoints (Public)

```bash
# Basic health check
GET /health

# Readiness check
GET /health/ready

# Liveness check
GET /health/live
```

### Authentication Endpoints

#### Register a new user (Public)
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "role": "user"
}
```

#### Login (Public)
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "message": "Login successful",
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Get current user info (Protected)
```bash
GET /api/auth/me
Authorization: Bearer <token>

Response:
{
  "id": 1,
  "email": "user@example.com",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z"
}
```

## Tenant/Master Context

The application supports tenant/master database pattern through HTTP headers:

```bash
# Master database request
GET /api/resource
X-Request-Type: master

# Tenant database request
GET /api/resource
X-Tenant-ID: tenant-123
```

If no headers are provided, the request defaults to master database.

## Development

### Available Make Commands

```bash
make help          # Display all available commands
make build         # Build the application binary
make run           # Run the application
make test          # Run tests with coverage
make migrate       # Run database migrations
make clean         # Clean build artifacts
make deps          # Download dependencies
make docker-up     # Start PostgreSQL with Docker
make docker-down   # Stop Docker containers
make lint          # Run linter
make fmt           # Format code
make version       # Display version info
```

### Running Tests

```bash
make test
```

Coverage report will be generated at `coverage.html`

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

## Docker Support

Create a `docker-compose.yml` file for local development:

```yaml
version: '3.8'

services:
  master-db:
    image: postgres:15-alpine
    container_name: myapp-master-db
    environment:
      POSTGRES_DB: master_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - master-data:/var/lib/postgresql/data

  tenant-db:
    image: postgres:15-alpine
    container_name: myapp-tenant-db
    environment:
      POSTGRES_DB: tenant_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5433:5432"
    volumes:
      - tenant-data:/var/lib/postgresql/data

volumes:
  master-data:
  tenant-data:
```

Start databases:
```bash
make docker-up
```

## Architecture

### Clean Architecture Layers

1. **Infrastructure Layer** (`internal/pkg/`)
   - Configuration management
   - Logging
   - Database connections
   - HTTP server
   - Middleware

2. **Service Layer** (`internal/service/`)
   - Business logic
   - Domain models
   - Repositories
   - HTTP handlers

3. **Application Layer** (`internal/module/`, `cmd/`)
   - Dependency injection
   - Application bootstrap
   - CLI commands

### Service Patterns

The application supports flexible service architectures:

- **Full Stack** (Repository → Service → Handler): Auth service
- **Service Layer** (Service → Handler): Email, SMS services
- **Handler Only** (Handler): Health check service

## Go Idioms

This project follows Go best practices:

- Context as first parameter in all functions
- Error wrapping with `fmt.Errorf` and `%w`
- No global mutable state
- Accept interfaces, return structs
- Proper naming conventions
- Early return for errors
- Clear error messages with context

## Adding New Services

To add a new service:

1. Create service directory: `internal/service/myservice/`
2. Implement layers as needed:
   - `model.go` - Domain models (if using database)
   - `repository.go` - Database operations (if using database)
   - `service.go` - Business logic
   - `handler.go` - HTTP handlers
   - `module.go` - FX module
3. Add module to `internal/module/app.go`
4. Register routes in `internal/router/router.go`

## License

[Your License Here]

## Contributing

[Your Contributing Guidelines Here]

## Support

For issues and questions, please open an issue on the repository.
