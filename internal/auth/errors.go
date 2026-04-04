package auth

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserDeleted         = errors.New("user is deleted")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailNotVerified    = errors.New("email is not verified")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)
