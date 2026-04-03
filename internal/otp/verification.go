package otp

import (
	"context"
	"errors"
	"time"
)

var ErrVerificationCodeNotFound = errors.New("verification code not found")

type VerificationCode struct {
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type VerificationCodeStore interface {
	SetVerificationCode(ctx context.Context, email string, value VerificationCode, ttl time.Duration) error
	GetVerificationCode(ctx context.Context, email string) (VerificationCode, error)
	DeleteVerificationCode(ctx context.Context, email string) error
}
