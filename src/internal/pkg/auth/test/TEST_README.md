# Authentication Module Tests

Comprehensive test suite for the authentication module covering all components.

## Test Files

### 1. `utils_test.go` - Password Hashing Tests
- `TestHashPassword` - Tests password hashing with various inputs
- `TestVerifyPassword` - Tests password verification
- `TestHashPassword_DifferentPasswords` - Ensures different passwords produce different hashes
- `TestHashPassword_SamePasswordDifferentHashes` - Verifies salt randomization

### 2. `pkce_test.go` - PKCE Tests
- `TestGenerateCodeChallenge_Plain` - Tests plain code challenge generation
- `TestGenerateCodeChallenge_S256` - Tests S256 code challenge generation
- `TestValidateCodeVerifier_Plain` - Tests plain verifier validation
- `TestValidateCodeVerifier_S256` - Tests S256 verifier validation
- `TestGenerateCodeVerifier` - Tests code verifier generation
- `TestPKCE_EndToEnd` - End-to-end PKCE flow test

### 3. `token_manager_test.go` - Token Manager Tests
- `TestNewTokenManager` - Tests token manager creation
- `TestTokenManager_GenerateAccessToken` - Tests access token generation
- `TestTokenManager_GenerateAccessToken_Claims` - Verifies token claims
- `TestTokenManager_GenerateRefreshToken` - Tests refresh token generation
- `TestTokenManager_ValidateAccessToken` - Tests token validation
- `TestTokenManager_ValidateAccessToken_Expired` - Tests expired token handling
- `TestTokenManager_ExtractClaims` - Tests claims extraction
- `TestTokenManager_DifferentUsers` - Tests token uniqueness per user

### 4. `repository_test.go` - User Repository Tests
- `TestRepository_GetByEmail` - Tests user retrieval by email
- `TestRepository_EmailExists` - Tests email existence check
- `TestRepository_GetByID` - Tests user retrieval by ID
- `TestRepository_Create` - Tests user creation

### 5. `token_repository_test.go` - Token Repository Tests
- `TestTokenRepository_SaveRefreshToken` - Tests refresh token storage
- `TestTokenRepository_GetRefreshToken` - Tests refresh token retrieval
- `TestTokenRepository_RevokeRefreshToken` - Tests token revocation
- `TestTokenRepository_RevokeAllUserTokens` - Tests bulk token revocation
- `TestTokenRepository_AddToBlacklist` - Tests blacklist addition
- `TestTokenRepository_IsBlacklisted` - Tests blacklist checking
- `TestTokenRepository_CleanupExpiredTokens` - Tests expired token cleanup

### 6. `service_test.go` - Service Layer Tests
- `TestService_Register` - Tests user registration
- `TestService_Login` - Tests user login
- `TestService_RefreshToken` - Tests token refresh with rotation
- `TestService_Logout` - Tests user logout
- `TestService_ValidateToken` - Tests token validation
- `TestService_GetUserByID` - Tests user retrieval

### 7. `middleware_test.go` - Middleware Tests
- `TestJWTMiddleware_ValidToken` - Tests middleware with valid token
- `TestJWTMiddleware_MissingToken` - Tests middleware without token
- `TestJWTMiddleware_InvalidToken` - Tests middleware with invalid token
- `TestRequireRole` - Tests role-based access control
- `TestRequireRole_InsufficientPermissions` - Tests permission denial
- `TestGetUserFromContext` - Tests user context extraction
- `TestGetUserIDFromContext` - Tests user ID extraction

### 8. `handler_test.go` - HTTP Handler Tests
- `TestHandler_Register` - Tests registration endpoint
- `TestHandler_Login` - Tests login endpoint
- `TestHandler_RefreshToken` - Tests refresh endpoint
- `TestHandler_Logout` - Tests logout endpoint
- `TestHandler_GetCurrentUser` - Tests current user endpoint

## Running Tests

### Run All Tests
```bash
cd src/internal/pkg/auth/test
go test -v
```

Or from project root:
```bash
go test ./internal/pkg/auth/test -v
```

### Run with Coverage
```bash
go test -v -cover
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Test File
```bash
go test -v -run TestService
go test -v -run TestHandler
```

### Run Specific Test
```bash
go test -v -run TestService_Register
```

## Test Coverage Goals

- **Unit Tests**: All utility functions, repositories, and service methods
- **Integration Tests**: Complete authentication flows
- **Edge Cases**: Invalid inputs, expired tokens, revoked tokens
- **Security Tests**: Token validation, password hashing, PKCE

## Test Database

Tests use SQLite in-memory database for fast execution:
- No external database required
- Tests run in isolation
- Fast test execution
- Easy cleanup

**Note**: Repository, service, middleware, and handler tests require CGO to be enabled (SQLite dependency).

### Running Tests with CGO

**Windows**:
```bash
# Install MinGW-w64 or TDM-GCC
# Add to PATH, then run:
go test -tags=cgo ./internal/pkg/auth -v
```

**Linux/Mac**:
```bash
# CGO usually enabled by default
go test ./internal/pkg/auth -v
```

### Tests That Don't Require CGO

These tests can run without CGO:
- `utils_test.go` - Password hashing tests
- `pkce_test.go` - PKCE function tests  
- `token_manager_test.go` - Token manager tests (uses temp files for keys)

### Tests That Require CGO

These tests require CGO (SQLite):
- `repository_test.go` - User repository tests
- `token_repository_test.go` - Token repository tests
- `service_test.go` - Service layer tests
- `middleware_test.go` - Middleware tests
- `handler_test.go` - HTTP handler tests

## Mocking

- **Token Manager**: Uses real RSA keys generated in temp directory
- **Database**: Uses in-memory SQLite
- **Logger**: Uses zap.NewNop() for no-op logging

## Test Patterns

### Table-Driven Tests
Most tests use table-driven approach for comprehensive coverage:
```go
tests := []struct {
    name     string
    input    string
    wantErr  bool
}{
    // test cases
}
```

### Setup/Teardown
Each test file includes setup functions:
- `setupTestService()` - Creates complete service with dependencies
- `setupTestRepository()` - Creates repository with test database
- `setupTestTokenManager()` - Creates token manager with temp keys

## Notes

- Tests use lower bcrypt cost (10) for faster execution
- RSA keys are generated in temporary directories
- All tests are isolated and can run in parallel
- No external dependencies required
