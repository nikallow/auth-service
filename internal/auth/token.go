package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nikallow/auth-service/internal/config"
)

const refreshTokenSize = 32

type TokenManager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

type AccessTokenInput struct {
	UserID    string
	Email     string
	Roles     []string
	SessionID string
}

type AccessTokenClaims struct {
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	SessionID string   `json:"sid"`

	jwt.RegisteredClaims
}

func NewTokenManager(cfg config.JWT, issuer string) *TokenManager {
	return &TokenManager{
		secret:     []byte(cfg.Secret),
		issuer:     issuer,
		accessTTL:  time.Duration(cfg.AccessTTLMin) * time.Minute,
		refreshTTL: time.Duration(cfg.RefreshTTLHour) * time.Hour,
	}
}

func (m *TokenManager) NewAccessToken(input AccessTokenInput) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessTTL)

	claims := AccessTokenClaims{
		Email:     input.Email,
		Roles:     input.Roles,
		SessionID: input.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   input.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signedToken, expiresAt, nil
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}

		return m.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, fmt.Errorf("parse access token: invalid claims type")
	}

	return claims, nil
}

func (m *TokenManager) NewRefreshToken() (string, time.Time, error) {
	tokenBytes := make([]byte, refreshTokenSize)

	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	expiresAt := time.Now().UTC().Add(m.refreshTTL)
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	return token, expiresAt, nil
}

func (m *TokenManager) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}

func (m *TokenManager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *TokenManager) RefreshTTL() time.Duration {
	return m.refreshTTL
}
