package httptransport

import (
	"context"

	"github.com/nikallow/auth-service/internal/auth"
)

type AuthService interface {
	Register(ctx context.Context, input auth.RegisterInput) (auth.RegisterResult, error)
	VerifyEmail(ctx context.Context, input auth.VerifyEmailInput) error
	Login(ctx context.Context, input auth.LoginInput) (auth.TokenPair, error)
	Refresh(ctx context.Context, input auth.RefreshInput) (auth.TokenPair, error)
	Logout(ctx context.Context, input auth.LogoutInput) error
}

type AccessTokenParser interface {
	ParseAccessToken(tokenString string) (*auth.AccessTokenClaims, error)
}
