package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nikallow/auth-service/internal/otp"
	sqlc "github.com/nikallow/auth-service/internal/storage/postgres/gen"
)

const defaultVerificationCodeTTL = 5 * time.Minute

type Service struct {
	queries               sqlc.Querier
	passwordHasher        *PasswordHasher
	tokenManager          *TokenManager
	verificationCodeStore otp.VerificationCodeStore
	verificationCodeTTL   time.Duration
}

func NewService(
	queries sqlc.Querier,
	passwordHasher *PasswordHasher,
	tokenManager *TokenManager,
	verificationCodeStore otp.VerificationCodeStore,
	verificationCodeTTL time.Duration,
) *Service {
	if verificationCodeTTL == 0 {
		verificationCodeTTL = defaultVerificationCodeTTL
	}

	return &Service{
		queries:               queries,
		passwordHasher:        passwordHasher,
		tokenManager:          tokenManager,
		verificationCodeStore: verificationCodeStore,
		verificationCodeTTL:   verificationCodeTTL,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (RegisterResult, error) {
	existingUser, err := s.queries.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{Email: input.Email})
	if err == nil {
		if existingUser.DeletedAt.Valid {
			return RegisterResult{}, ErrUserDeleted
		}

		return RegisterResult{}, ErrUserAlreadyExists
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return RegisterResult{}, fmt.Errorf("get user by email: %w", err)
	}

	passwordHash, err := s.passwordHasher.Hash(input.Password)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        input.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return RegisterResult{}, fmt.Errorf("create user: %w", err)
	}

	verificationCode, err := GenerateVerificationCode()
	if err != nil {
		return RegisterResult{}, fmt.Errorf("generate verification code: %w", err)
	}

	now := time.Now().UTC()
	if err := s.verificationCodeStore.SetVerificationCode(
		ctx,
		user.Email,
		otp.VerificationCode{
			Code:      verificationCode,
			CreatedAt: now,
		},
		s.verificationCodeTTL); err != nil {
		return RegisterResult{}, fmt.Errorf("set verification code: %w", err)
	}

	userID, err := uuidFromPG(user.ID)
	if err != nil {
		return RegisterResult{}, fmt.Errorf("convert user id: %w", err)
	}

	return RegisterResult{
		UserID:                    userID,
		Email:                     user.Email,
		EmailVerified:             user.EmailVerified,
		VerificationCodeExpiresAt: now.Add(s.verificationCodeTTL),
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	user, err := s.queries.GetActiveUserByEmail(ctx, sqlc.GetActiveUserByEmailParams{
		Email: input.Email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, ErrInvalidCredentials
		}

		return TokenPair{}, fmt.Errorf("get active user by email: %w", err)
	}

	if err := s.passwordHasher.Compare(input.Password, user.PasswordHash); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	if !user.EmailVerified {
		return TokenPair{}, ErrEmailNotVerified
	}

	refreshToken, refreshExpiresAt, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	refreshHash := s.tokenManager.HashRefreshToken(refreshToken)

	session, err := s.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:      user.ID,
		RefreshHash: refreshHash,
		ExpiresAt:   timestamptzValue(refreshExpiresAt),
		UserAgent:   textValue(input.UserAgent),
		Ip:          textValue(input.IP),
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("create session: %w", err)
	}

	userID, err := uuidFromPG(user.ID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("convert user id: %w", err)
	}

	sessionID, err := uuidFromPG(session.ID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("convert session id: %w", err)
	}

	accessToken, accessExpiresAt, err := s.tokenManager.NewAccessToken(AccessTokenInput{
		UserID:    userID.String(),
		Email:     user.Email,
		Roles:     []string{"user"},
		SessionID: sessionID.String(),
	})
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate access token: %w", err)
	}

	return TokenPair{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessExpiresAt,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}, nil
}

func textValue(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}

	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func timestamptzValue(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func uuidFromPG(id pgtype.UUID) (uuid.UUID, error) {
	if !id.Valid {
		return uuid.Nil, fmt.Errorf("uuid is not valid")
	}

	return id.Bytes, nil
}
