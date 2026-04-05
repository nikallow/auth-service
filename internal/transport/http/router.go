package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	apiBasePath  = "/api/v1"
	authBasePath = apiBasePath + "/auth"
)

func NewRouter(handler *Handler) http.Handler {
	router := chi.NewRouter()
	router.Get("/health", handler.Health)

	router.Route(apiBasePath, func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.Register)
			r.Post("/verify", handler.VerifyEmail)
			r.Post("/login", handler.Login)
			r.Post("/refresh", handler.Refresh)
			r.Post("/logout", handler.Logout)
			r.With(handler.RequireAccessToken).Get("/me", handler.Me)
		})
	})

	return router
}
