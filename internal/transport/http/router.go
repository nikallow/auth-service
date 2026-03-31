package httptransport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	handler := NewHandler()

	router := chi.NewRouter()
	router.Get("/health", handler.Health)

	return router
}
