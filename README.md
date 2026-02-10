# MyApp - Golang RESTful API Project

A clean architecture RESTful API built with Go, featuring dependency injection, JWT authentication, and tenant/master database pattern.

## 📁 Project Structure

```
base/
├── src/                        # Source code
│   ├── cmd/                    # Application entry points
│   ├── internal/               # Internal packages
│   ├── config/                 # Configuration files
│   └── Makefile               # Build commands
├── docs/                       # Documentation
│   ├── README.md              # Full API documentation
│   ├── DEV_GUIDE.md           # Development guide with Air setup
│   ├── QUICKSTART.md          # Quick start guide
│   └── SERVICE_ARCHITECTURE.md # Architecture documentation
├── deployment/                 # Deployment configurations
│   ├── docker-compose.yml     # Docker Compose for local dev
│   └── README.md              # Deployment documentation
└── README.md                   # This file
```

## 🚀 Quick Start

### 1. Start Databases

```bash
cd src
make docker-up
```

### 2. Run Migrations

```bash
make migrate
```

### 3. Start Development Server

**With hot reload (recommended):**
```bash
make dev
```

**Without hot reload:**
```bash
make run
```

The API server will start at `http://localhost:8080`

## 📖 Documentation

- **[Full Documentation](docs/README.md)** - Complete API documentation and features
- **[Development Guide](docs/DEV_GUIDE.md)** - Testing with Air hot reload
- **[Quick Start Guide](docs/QUICKSTART.md)** - Step-by-step setup
- **[Architecture](docs/SERVICE_ARCHITECTURE.md)** - Service architecture patterns
- **[Deployment](deployment/README.md)** - Docker and deployment configuration

## ✨ Features

- **Clean Architecture** - Clear separation of concerns
- **Dependency Injection** - Using Uber FX
- **JWT Authentication** - Secure token-based auth
- **Tenant/Master Pattern** - Separate databases
- **Hot Reload** - Air for development
- **PostgreSQL** - With GORM ORM
- **Structured Logging** - Using Zap
- **Comprehensive Testing** - Unit and integration tests

## 🛠️ Technology Stack

- **Framework**: Echo v4
- **Dependency Injection**: Uber FX
- **CLI**: Cobra
- **Logger**: Zap
- **ORM**: GORM
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Hot Reload**: Air

## 📝 Available Commands

```bash
make help          # Display all available commands
make build         # Build the application binary
make run           # Run the application
make dev           # Run with hot reload (Air)
make test          # Run tests with coverage
make migrate       # Run database migrations
make docker-up     # Start PostgreSQL with Docker
make docker-down   # Stop Docker containers
make clean         # Clean build artifacts
make deps          # Download dependencies
```

## 🧪 Test the API

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123", "role": "user"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

## 📦 API Endpoints

### Public Endpoints
- `GET /health` - Health check
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check
- `POST /api/auth/register` - Register user
- `POST /api/auth/login` - User login

### Protected Endpoints
- `GET /api/auth/me` - Get current user info

## 🏗️ Architecture

The project follows clean architecture principles with three main layers:

1. **Infrastructure Layer** (`internal/pkg/`) - Config, logging, database, server
2. **Service Layer** (`internal/service/`) - Business logic, models, repositories
3. **Application Layer** (`internal/module/`, `cmd/`) - Dependency injection, bootstrap

## 🔧 Configuration

Configuration is managed through:
1. Configuration file (`config/config.yaml`)
2. Environment variables (prefix: `MYAPP_`)
3. CLI flags

Priority: CLI flags > Environment variables > Config file > Defaults

## 🤝 Contributing

1. Create a feature branch
2. Make your changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Submit a pull request

## 📄 License

[Your License Here]

## 📞 Support

For detailed information, check the documentation in the `docs/` folder.

For issues and questions, please open an issue on the repository.
