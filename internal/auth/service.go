package auth

import (
	"context"
	"crypto/subtle"
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
	email := normalizeEmail(input.Email)
	if err := validateEmail(email); err != nil {
		return RegisterResult{}, ErrInvalidEmail
	}

	if err := validatePassword(input.Password); err != nil {
		return RegisterResult{}, ErrInvalidPassword
	}

	existingUser, err := s.queries.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{
		Email: email,
	})
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
		Email: email,
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
		VerificationCodeExpiresAt: now.Add(s.verificationCodeTTL),
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, input VerifyEmailInput) error {
	email := normalizeEmail(input.Email)
	if err := validateEmail(email); err != nil {
		return ErrInvalidEmail
	}

	code := normalizeVerificationCode(input.Code)
	if err := validateVerificationCode(code); err != nil {
		return ErrInvalidVerificationCode
	}

	user, err := s.queries.GetUserByEmail(ctx, sqlc.GetUserByEmailParams{
		Email: email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return fmt.Errorf("get user by email: %w", err)
	}

	if user.DeletedAt.Valid {
		return ErrUserDeleted
	}

	if user.EmailVerified {
		return nil
	}

	storedCode, err := s.verificationCodeStore.GetVerificationCode(ctx, email)
	if err != nil {
		if errors.Is(err, otp.ErrVerificationCodeNotFound) {
			return ErrVerificationCodeExpired
		}

		return fmt.Errorf("get verification code: %w", err)
	}

	if subtle.ConstantTimeCompare([]byte(storedCode.Code), []byte(code)) != 1 {
		return ErrInvalidVerificationCode
	}

	_, err = s.queries.MarkEmailVerified(ctx, sqlc.MarkEmailVerifiedParams{
		ID: user.ID,
	})
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}

	if err = s.verificationCodeStore.DeleteVerificationCode(ctx, email); err != nil {
		return fmt.Errorf("delete verification code: %w", err)
	}

	return nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	email := normalizeEmail(input.Email)
	if err := validateEmail(email); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	if err := validatePassword(input.Password); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	user, err := s.queries.GetActiveUserByEmail(ctx, sqlc.GetActiveUserByEmailParams{
		Email: email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, ErrInvalidCredentials
		}

		return TokenPair{}, fmt.Errorf("get active user by email: %w", err)
	}

	if err = s.passwordHasher.Compare(input.Password, user.PasswordHash); err != nil {
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

	return s.buildTokenPair(user, session.ID, refreshToken, refreshExpiresAt)
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	if input.RefreshToken == "" {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	oldRefreshHash := s.tokenManager.HashRefreshToken(input.RefreshToken)

	session, err := s.queries.GetActiveSessionByRefreshHash(ctx, sqlc.GetActiveSessionByRefreshHashParams{
		RefreshHash: oldRefreshHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, ErrInvalidRefreshToken
		}

		return TokenPair{}, fmt.Errorf("get active session by refresh hash: %w", err)
	}

	now := time.Now().UTC()
	if !session.ExpiresAt.Valid || !session.ExpiresAt.Time.After(now) {
		return TokenPair{}, ErrRefreshTokenExpired
	}

	user, err := s.queries.GetActiveUserByID(ctx, sqlc.GetActiveUserByIDParams{
		ID: session.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, ErrInvalidRefreshToken
		}

		return TokenPair{}, fmt.Errorf("get active user by id: %w", err)
	}

	newRefreshToken, newRefreshExpiresAt, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate new refresh token: %w", err)
	}

	newRefreshHash := s.tokenManager.HashRefreshToken(newRefreshToken)

	rotatedSession, err := s.queries.RotateSessionRefreshHash(ctx, sqlc.RotateSessionRefreshHashParams{
		ID:          session.ID,
		RefreshHash: newRefreshHash,
		ExpiresAt:   timestamptzValue(newRefreshExpiresAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, ErrInvalidRefreshToken
		}

		return TokenPair{}, fmt.Errorf("rotate session refresh hash: %w", err)
	}

	return s.buildTokenPair(user, rotatedSession.ID, newRefreshToken, newRefreshExpiresAt)
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken == "" {
		return ErrInvalidRefreshToken
	}

	refreshHash := s.tokenManager.HashRefreshToken(input.RefreshToken)

	session, err := s.queries.GetSessionByRefreshHash(ctx, sqlc.GetSessionByRefreshHashParams{
		RefreshHash: refreshHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidRefreshToken
		}

		return fmt.Errorf("get session by refresh hash: %w", err)
	}

	if session.RevokedAt.Valid {
		return nil
	}

	_, err = s.queries.RevokeSession(ctx, sqlc.RevokeSessionParams{
		ID: session.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidRefreshToken
		}

		return fmt.Errorf("revoke session: %w", err)
	}

	return nil
}

func (s *Service) buildTokenPair(
	user sqlc.User,
	sessionID pgtype.UUID,
	refreshToken string,
	refreshExpiresAt time.Time,
) (TokenPair, error) {
	userID, err := uuidFromPG(user.ID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("convert user id: %w", err)
	}

	sid, err := uuidFromPG(sessionID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("convert session id: %w", err)
	}

	accessToken, accessExpiresAt, err := s.tokenManager.NewAccessToken(AccessTokenInput{
		UserID:        userID.String(),
		Email:         user.Email,
		Roles:         []string{"user"},
		SessionID:     sid.String(),
		EmailVerified: user.EmailVerified,
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
