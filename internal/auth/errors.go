package auth

import "errors"

var (
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrUserDeleted             = errors.New("user is deleted")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrEmailNotVerified        = errors.New("email is not verified")
	ErrInvalidRefreshToken     = errors.New("invalid refresh token")
	ErrRefreshTokenExpired     = errors.New("refresh token expired")
	ErrInvalidEmail            = errors.New("invalid email")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code expired")
)
