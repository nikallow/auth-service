package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nikallow/auth-service/internal/otp"
	goredis "github.com/redis/go-redis/v9"
)

const verifyEmailKeyPrefix = "verify_email"

type VerificationCodeStore struct {
	client *goredis.Client
}

func NewVerificationCodeStore(client *goredis.Client) *VerificationCodeStore {
	return &VerificationCodeStore{
		client: client,
	}
}

func (s *VerificationCodeStore) SetVerificationCode(
	ctx context.Context,
	email string,
	value otp.VerificationCode,
	ttl time.Duration,
) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal verification code: %w", err)
	}

	if err := s.client.Set(ctx, verificationCodeKey(email), payload, ttl).Err(); err != nil {
		return fmt.Errorf("set verification code: %w", err)
	}

	return nil
}

func (s *VerificationCodeStore) GetVerificationCode(ctx context.Context, email string) (otp.VerificationCode, error) {
	payload, err := s.client.Get(ctx, verificationCodeKey(email)).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return otp.VerificationCode{}, otp.ErrVerificationCodeNotFound
		}

		return otp.VerificationCode{}, fmt.Errorf("get verification code: %w", err)
	}

	var value otp.VerificationCode
	if err := json.Unmarshal(payload, &value); err != nil {
		return otp.VerificationCode{}, fmt.Errorf("unmarshall verification code: %w", err)
	}

	return value, err
}

func (s *VerificationCodeStore) DeleteVerificationCode(ctx context.Context, email string) error {
	if err := s.client.Del(ctx, verificationCodeKey(email)).Err(); err != nil {
		return fmt.Errorf("delete verification code: %w", err)
	}

	return nil
}

func verificationCodeKey(email string) string {
	return verifyEmailKeyPrefix + ":" + email
}
