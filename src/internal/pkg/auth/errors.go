package auth

import "fmt"

// Custom error types for authentication operations

// ErrInvalidCredentials is returned when email/password combination is invalid
type ErrInvalidCredentials struct {
	Message string
}

func (e *ErrInvalidCredentials) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid credentials"
}

// ErrEmailExists is returned when trying to register with an existing email
type ErrEmailExists struct {
	Email string
}

func (e *ErrEmailExists) Error() string {
	return fmt.Sprintf("email %s already exists", e.Email)
}

// ErrTokenExpired is returned when a token has expired
type ErrTokenExpired struct {
	Message string
}

func (e *ErrTokenExpired) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "token has expired"
}

// ErrTokenInvalid is returned when a token is invalid or malformed
type ErrTokenInvalid struct {
	Message string
}

func (e *ErrTokenInvalid) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "invalid token"
}

// ErrTokenRevoked is returned when a token has been revoked
type ErrTokenRevoked struct {
	Message string
}

func (e *ErrTokenRevoked) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "token has been revoked"
}

// ErrRefreshTokenNotFound is returned when refresh token is not found in database
type ErrRefreshTokenNotFound struct {
	Message string
}

func (e *ErrRefreshTokenNotFound) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "refresh token not found"
}

// ErrUserNotFound is returned when a user is not found
type ErrUserNotFound struct {
	ID    uint
	Email string
}

func (e *ErrUserNotFound) Error() string {
	if e.Email != "" {
		return fmt.Sprintf("user with email %s not found", e.Email)
	}
	return fmt.Sprintf("user with id %d not found", e.ID)
}
