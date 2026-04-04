package auth

import (
	"time"

	"github.com/google/uuid"
)

type RegisterInput struct {
	Email    string
	Password string
}

type RegisterResult struct {
	UserID                    uuid.UUID
	Email                     string
	VerificationCodeExpiresAt time.Time
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
}

type RefreshInput struct {
	RefreshToken string
}

type LogoutInput struct {
	RefreshToken string
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}
