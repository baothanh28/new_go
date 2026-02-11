# Authentication Module

Enterprise-grade JWT authentication module with RS256 asymmetric encryption, refresh token rotation, and OAuth 2.0 security patterns.

## Features

- **RS256 JWT Tokens**: Asymmetric encryption using RSA key pairs (more secure than HS256)
- **Access Token**: Short-lived tokens (15 minutes) for API access
- **Refresh Token**: Long-lived tokens (7 days) with automatic rotation
- **Token Rotation**: Each refresh generates new access + refresh tokens
- **Token Revocation**: Blacklist-based token revocation on logout
- **Password Security**: bcrypt hashing with configurable cost
- **PKCE Support**: Proof Key for Code Exchange for enhanced security
- **Role-Based Access Control**: Middleware for role-based authorization
- **Automatic Cleanup**: Background worker for expired token cleanup

## Architecture

```
src/internal/pkg/auth/
├── models.go              # Database models (User, RefreshToken, TokenBlacklist)
├── dto.go                 # Request/Response DTOs
├── repository.go          # User repository operations
├── token_repository.go    # Refresh token & blacklist operations
├── service.go             # Business logic (Register, Login, Refresh, Logout)
├── token_manager.go       # RS256 token generation & validation
├── pkce.go                # PKCE challenge/verifier validation
├── middleware.go          # JWT authentication middleware
├── handler.go             # HTTP handlers for auth endpoints
├── router.go              # Route registration
├── module.go              # fx dependency injection module
├── errors.go              # Custom error types
├── utils.go               # Helper functions (password hashing)
└── keys/                  # RSA key pair storage
    ├── private.pem        # RSA private key (gitignored)
    ├── public.pem         # RSA public key
    └── keygen.go          # Key generation utility
```

## Setup

### 1. Generate RSA Key Pair

If keys don't exist, generate them:

```bash
cd src/internal/pkg/auth/keys
go run generate_keys.go .
```

This creates:
- `private.pem` - Private key for signing tokens (gitignored)
- `public.pem` - Public key for verifying tokens (committed to git)

### 2. Configuration

Update `src/config/config.yaml`:

```yaml
auth:
  access_token_duration: "15m"      # Access token lifetime
  refresh_token_duration: "168h"    # Refresh token lifetime (7 days)
  rsa_private_key_path: "src/internal/pkg/auth/keys/private.pem"
  rsa_public_key_path: "src/internal/pkg/auth/keys/public.pem"
  issuer: "myapp-auth-service"      # JWT issuer claim
  bcrypt_cost: 12                   # bcrypt hashing cost (4-31)
```

### 3. Database Migration

The module automatically migrates tables on startup:
- `users` - User accounts
- `refresh_tokens` - Refresh token storage
- `token_blacklist` - Revoked access tokens

## Usage

### Integration

Add the auth module to your fx application:

```go
import "myapp/internal/pkg/auth"

app := fx.New(
    // ... other modules
    auth.Module,
    // ... other modules
)
```

### API Endpoints

#### Register
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123",
  "role": "user"
}
```

#### Login
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123"
}

Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "a1b2c3...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": 1,
    "email": "user@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Refresh Token
```http
POST /api/auth/refresh
Content-Type: application/json

{
  "refresh_token": "a1b2c3..."
}

Response:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "x9y8z7...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

#### Logout
```http
POST /api/auth/logout
Authorization: Bearer eyJhbGc...

Response:
{
  "message": "logged out successfully"
}
```

#### Get Current User
```http
GET /api/auth/me
Authorization: Bearer eyJhbGc...

Response:
{
  "id": 1,
  "email": "user@example.com",
  "role": "user",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Protecting Routes

Use the JWT middleware to protect routes:

```go
import "myapp/internal/pkg/auth"

// Get middleware from service
middleware := auth.JWTMiddleware(authService, logger)

// Apply to routes
e.GET("/protected", handler, middleware)
```

### Role-Based Access Control

Use role-based middleware:

```go
import "myapp/internal/pkg/auth"

// Require specific roles
adminOnly := auth.RequireRole("admin")
e.GET("/admin", adminHandler, middleware, adminOnly)
```

### Accessing User Context

In handlers, extract user information:

```go
import "myapp/internal/pkg/auth"

func MyHandler(c echo.Context) error {
    user, err := auth.GetUserFromContext(c)
    if err != nil {
        return err
    }
    
    // Use user.UserID, user.Email, user.Role
    return c.JSON(200, map[string]interface{}{
        "user_id": user.UserID,
        "email": user.Email,
    })
}
```

## Security Features

### Token Security
- **RS256**: Asymmetric encryption prevents secret key exposure
- **Short-lived Access Tokens**: 15-minute expiration reduces attack window
- **Token Rotation**: Refresh tokens rotate on each use
- **JTI-based Revocation**: Access tokens revoked via JWT ID (JTI)

### Password Security
- **bcrypt Hashing**: One-way password hashing with configurable cost
- **Minimum Length**: 8 characters required
- **Never Exposed**: Passwords never returned in API responses

### PKCE Support
Optional PKCE validation for enhanced security:

```go
import "myapp/internal/pkg/auth"

// Validate code verifier against challenge
valid := auth.ValidateCodeVerifier(verifier, challenge, "S256")
```

## Background Workers

### Token Cleanup

A background worker automatically cleans up expired tokens every hour:
- Removes expired entries from `token_blacklist`
- Removes expired `refresh_tokens`

No manual intervention required.

## Error Handling

The module provides custom error types:

- `ErrInvalidCredentials` - Invalid email/password
- `ErrEmailExists` - Email already registered
- `ErrTokenExpired` - Token has expired
- `ErrTokenInvalid` - Invalid or malformed token
- `ErrTokenRevoked` - Token has been revoked
- `ErrRefreshTokenNotFound` - Refresh token not found
- `ErrUserNotFound` - User not found

## Testing

Example test structure:

```go
func TestLogin(t *testing.T) {
    // Setup test database
    // Create test user
    // Call service.Login()
    // Verify tokens returned
    // Verify refresh token stored
}
```

## Production Considerations

1. **RSA Key Security**: Keep `private.pem` secure, never commit to git
2. **Key Rotation**: Rotate RSA keys periodically (update config)
3. **Token Expiration**: Adjust durations based on security requirements
4. **Rate Limiting**: Add rate limiting to login/register endpoints
5. **Monitoring**: Monitor failed login attempts and token usage
6. **HTTPS Only**: Always use HTTPS in production
7. **CORS**: Configure CORS appropriately for your frontend

## Migration from Existing Auth

If migrating from `internal/service/auth`:

1. Update imports to use `myapp/internal/pkg/auth`
2. Replace HS256 tokens with RS256 tokens
3. Update token handling to use access + refresh tokens
4. Update middleware usage
5. Migrate existing users (passwords remain compatible)

## License

Part of the myapp project.
