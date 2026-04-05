package httptransport

import (
	"context"
	"net/http"

	"github.com/nikallow/auth-service/internal/auth"
)

type contextKey string

const accessTokenClaimsContextKey contextKey = "access_token_claims"

func (h *Handler) RequireAccessToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := bearerTokenFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or missing access token")
			return
		}

		claims, err := h.tokenParser.ParseAccessToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired access token")
			return
		}

		ctx := context.WithValue(r.Context(), accessTokenClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func accessTokenClaimsFromContext(ctx context.Context) (*auth.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(accessTokenClaimsContextKey).(*auth.AccessTokenClaims)
	return claims, ok
}
