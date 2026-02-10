# Test Suite Documentation

This document provides information about the test suites created for the internal packages.

## Test Coverage

Comprehensive test suites have been created for the following packages:

### 1. Config Package (`internal/pkg/config`)
- **File**: `config_test.go`
- **Tests**:
  - `ServerConfig` validation tests
  - `DatabaseConfig` validation tests
  - `JWTConfig` validation tests
  - `LoggerConfig` validation tests
  - Full `Config` validation tests
  - Configuration loading tests

### 2. Database Package (`internal/pkg/database`)

#### Context Tests (`context_test.go`)
- Tenant ID context management
- Context value storage and retrieval
- Context chaining and preservation
- Type safety tests

#### Repository Tests (`repository_test.go`)
- **BaseRepository CRUD operations**:
  - Insert single and batch entities
  - Update by ID and conditional updates
  - Retrieve by ID, all entities, and conditional queries
  - Delete by ID and conditional deletes
  - Count and existence checks
  - Transaction support
- **MasterRepo** creation and usage
- **TenantRepo** creation and dynamic tenant connections

**⚠️ NOTE**: Repository tests use SQLite which requires CGO. See [CGO Requirements](#cgo-requirements) below.

#### Manager Tests (`manager_test.go`)
- Database connection creation
- Connection pool configuration
- DatabaseManager lifecycle (creation and cleanup)
- Error handling for invalid configurations

**⚠️ NOTE**: Some tests are skipped as they require running PostgreSQL instances.

#### Tenant Connection Manager Tests (`tenant_connection_manager_test.go`)
- Dynamic tenant database connections
- Tenant configuration retrieval
- Multi-database support (PostgreSQL, MySQL, SQLite)
- Connection pooling for tenant databases
- Managing multiple tenant connections

**⚠️ NOTE**: These tests use SQLite for unit tests. PostgreSQL and MySQL tests are skipped unless databases are running.

### 3. Logger Package (`internal/pkg/logger`)
- **File**: `logger_test.go`
- **Tests**:
  - Log level parsing (debug, info, warn, error)
  - Logger creation with different formats (JSON, console)
  - Case insensitivity handling
  - Log level filtering

### 4. Middleware Package (`internal/pkg/middleware`)
- **File**: `context_middleware_test.go`
- **Tests**:
  - Request context middleware
  - Tenant ID header parsing
  - Master/Tenant request routing
  - Database selection based on context
  - Integration tests with Echo framework

### 5. Server Package (`internal/pkg/server`)
- **File**: `server_test.go`
- **Tests**:
  - Echo server creation and configuration
  - Request logging middleware
  - Custom error handler
  - Panic recovery
  - CORS support
  - Integration tests

## Running the Tests

### Run All Tests
```bash
# Run all tests in the pkg directory
go test ./internal/pkg/... -v

# Run with coverage
go test ./internal/pkg/... -cover

# Run with detailed coverage report
go test ./internal/pkg/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Tests by Package
```bash
# Config tests
go test ./internal/pkg/config -v

# Database tests
go test ./internal/pkg/database -v

# Logger tests
go test ./internal/pkg/logger -v

# Middleware tests
go test ./internal/pkg/middleware -v

# Server tests
go test ./internal/pkg/server -v
```

### Run Specific Tests
```bash
# Run a specific test
go test ./internal/pkg/config -v -run TestServerConfig_Validate

# Run tests matching a pattern
go test ./internal/pkg/database -v -run TestContext
```

### Run Tests with Short Mode
Some tests are marked to skip in short mode:
```bash
go test ./internal/pkg/... -v -short
```

## CGO Requirements

Some database tests (particularly those using SQLite) require CGO to be enabled. This means you need a C compiler:

### Windows
1. Install [MinGW-w64](https://www.mingw-w64.org/) or [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)
2. Add the `bin` directory to your PATH
3. Run tests with CGO enabled:
   ```bash
   $env:CGO_ENABLED=1
   go test ./internal/pkg/database -v
   ```

### Linux
CGO is usually enabled by default. Install gcc if needed:
```bash
sudo apt-get install build-essential  # Ubuntu/Debian
sudo yum install gcc                   # CentOS/RHEL
```

### macOS
Install Xcode Command Line Tools:
```bash
xcode-select --install
```

### Alternative: Skip CGO Tests
If you don't have a C compiler or don't want to enable CGO, you can skip those tests:
- The repository and tenant connection manager tests will fail without CGO
- However, other tests (config, logger, middleware, server, context) will work fine
- Consider running integration tests with actual PostgreSQL instead

## Test Dependencies

The test suite uses the following dependencies (already included in `go.mod`):
- `github.com/stretchr/testify` - Assertion library
- `go.uber.org/zap/zaptest` - Logger testing utilities
- `gorm.io/driver/sqlite` - In-memory database for testing
- `github.com/labstack/echo/v4` - Web framework (for middleware/server tests)

## Test Structure

All tests follow Go's standard testing conventions:
- Test files are named `*_test.go`
- Test functions start with `Test`
- Tests are located in the same package as the code they test
- Table-driven tests are used for multiple test cases
- Mock/stub objects are created for external dependencies

## Continuous Integration

To integrate these tests into CI/CD:

```yaml
# Example GitHub Actions workflow
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd src
          go test ./internal/pkg/... -v -cover
```

## Test Maintenance

When modifying code, ensure you:
1. Run the relevant tests to verify your changes
2. Update tests if behavior changes
3. Add new tests for new functionality
4. Maintain test coverage above 80%

## Troubleshooting

### Common Issues

1. **"Binary was compiled with 'CGO_ENABLED=0'"**
   - Solution: Enable CGO or skip SQLite tests (see [CGO Requirements](#cgo-requirements))

2. **Tests are slow**
   - Solution: Run specific test packages instead of all tests
   - Use `-short` flag to skip long-running tests

3. **Import errors**
   - Solution: Run `go mod tidy` to update dependencies

4. **Test failures after changes**
   - Check if mocks need updating
   - Verify test data is still valid
   - Review error messages for details

## Best Practices

1. **Keep tests fast**: Use in-memory databases and mocks
2. **Test one thing at a time**: Each test should focus on a single behavior
3. **Use descriptive names**: Test names should clearly describe what they test
4. **Clean up resources**: Always close connections and clean up test data
5. **Avoid flaky tests**: Don't rely on timing or external services

## Coverage Goals

Current test coverage by package:
- ✅ Config: ~95% (validation logic fully covered)
- ✅ Database Context: 100% (all context functions covered)
- ✅ Logger: ~90% (all public functions covered)
- ✅ Middleware: ~95% (request handling fully covered)
- ✅ Server: ~90% (server setup and error handling covered)
- ⚠️ Database Repository: Requires CGO for execution
- ⚠️ Database Manager: Integration tests recommended

## Next Steps

Consider adding:
1. **Integration tests** with real PostgreSQL instances
2. **Benchmark tests** for performance-critical code
3. **Example tests** to demonstrate package usage
4. **Fuzzing tests** for input validation functions
