# Quick Start Guide

This guide will help you get the MyApp API server up and running in minutes.

## Prerequisites

- Go 1.21 or higher installed
- PostgreSQL 14 or higher (or Docker to run PostgreSQL)

## Step-by-Step Setup

### 1. Start PostgreSQL Databases

**Option A: Using Docker (Recommended)**

```bash
# Start PostgreSQL containers
docker-compose up -d

# Verify containers are running
docker ps
```

This will start two PostgreSQL databases:
- `myapp-master-db` on port 5432 (master_db)
- `myapp-tenant-db` on port 5433 (tenant_db)

**Option B: Use Existing PostgreSQL**

If you have PostgreSQL installed, create two databases:

```sql
CREATE DATABASE master_db;
CREATE DATABASE tenant_db;
```

Then update `config/config.yaml` with your PostgreSQL credentials.

### 2. Configure the Application

Edit `config/config.yaml` if needed. The default configuration works with the Docker setup:

```yaml
master_database:
  host: "localhost"
  port: 5432
  name: "master_db"
  user: "postgres"
  password: "password"

tenant_database:
  host: "localhost"
  port: 5432  # Change to 5433 if using docker-compose setup
  name: "tenant_db"
  user: "postgres"
  password: "password"
```

**Important**: Change the JWT secret in production!

### 3. Run Database Migrations

```bash
go run cmd/api/main.go migrate
```

Expected output:
```
Running database migrations...
Running auth service migrations
Auth service migrations completed successfully
All migrations completed successfully!
Migrations completed successfully!
```

### 4. Start the Server

```bash
go run cmd/api/main.go serve
```

Expected output:
```
{"level":"info","msg":"Starting HTTP server","addr":"0.0.0.0:8080"}
```

The server is now running at `http://localhost:8080`

## Test the API

### 1. Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "service": "myapp",
  "time": "2024-01-01T00:00:00Z"
}
```

### 2. Register a User

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "role": "user"
  }'
```

Response:
```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "test@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Save the token** - you'll need it for authenticated requests!

### 4. Get Current User Info (Protected Route)

```bash
# Replace YOUR_TOKEN with the actual token from login response
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response:
```json
{
  "id": 1,
  "email": "test@example.com",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z"
}
```

## Available Endpoints

### Public Endpoints (No Authentication Required)

- `GET /health` - Health check
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login

### Protected Endpoints (Requires JWT Token)

- `GET /api/auth/me` - Get current user information

## Tenant/Master Database Pattern

The application supports separate databases for system-wide and tenant-specific data.

### Master Database Request (Default)

```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Request-Type: master"
```

### Tenant Database Request

```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: tenant-123"
```

## Using Make Commands

The project includes a Makefile for common tasks:

```bash
make help          # Show all available commands
make build         # Build the binary
make run           # Run the server
make test          # Run tests
make migrate       # Run migrations
make docker-up     # Start Docker containers
make docker-down   # Stop Docker containers
make clean         # Clean build artifacts
```

## Building for Production

1. Build the binary:
```bash
make build
```

2. Run the binary:
```bash
./bin/myapp serve
```

3. Or specify a custom config file:
```bash
./bin/myapp serve --config /path/to/config.yaml
```

## Environment Variables

You can override configuration using environment variables:

```bash
export MYAPP_SERVER_PORT=9000
export MYAPP_JWT_SECRET=your-super-secret-key-at-least-32-chars
export MYAPP_LOGGER_LEVEL=debug
./bin/myapp serve
```

## Troubleshooting

### Database Connection Issues

1. Check if PostgreSQL is running:
```bash
docker ps  # if using Docker
```

2. Test database connection:
```bash
psql -h localhost -p 5432 -U postgres -d master_db
```

3. Check configuration in `config/config.yaml`

### Port Already in Use

If port 8080 is already in use, change it in `config/config.yaml`:
```yaml
server:
  port: 9000  # Use a different port
```

### Migration Errors

If migrations fail, check:
1. Database is running and accessible
2. Credentials in config are correct
3. Database exists

### JWT Token Issues

If you get "invalid token" errors:
1. Make sure you're including the full token with "Bearer " prefix
2. Check if the token has expired (default: 24 hours)
3. Verify JWT secret hasn't changed

## Next Steps

1. **Add More Services**: Follow the pattern in `internal/service/` to add new features
2. **Add Tests**: Write unit and integration tests for your services
3. **API Documentation**: Generate API docs using Swagger/OpenAPI
4. **Deployment**: Deploy to your cloud provider (AWS, GCP, Azure)
5. **Monitoring**: Add metrics and monitoring (Prometheus, Grafana)

## Need Help?

- Check the full README.md for detailed documentation
- Review the code in `internal/service/` for examples
- Check logs for error messages (logger outputs to console)

Happy coding! 🚀
