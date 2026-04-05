package httptransport

import (
	"github.com/nikallow/auth-service/internal/auth"
)

type Handler struct {
	authService         *auth.Service
	refreshCookieSecure bool
}

func NewHandler(authService *auth.Service, refreshCookieSecure bool) *Handler {
	return &Handler{
		authService:         authService,
		refreshCookieSecure: refreshCookieSecure,
	}
}
