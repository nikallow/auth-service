package httptransport

import (
	"errors"
	"net/http"

	"github.com/nikallow/auth-service/internal/auth"
)

func (h *Handler) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidEmail),
		errors.Is(err, auth.ErrInvalidPassword),
		errors.Is(err, auth.ErrInvalidVerificationCode),
		errors.Is(err, auth.ErrVerificationCodeExpired):
		writeError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, auth.ErrUserAlreadyExists),
		errors.Is(err, auth.ErrUserDeleted):
		writeError(w, http.StatusConflict, err.Error())

	case errors.Is(err, auth.ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, auth.ErrInvalidCredentials),
		errors.Is(err, auth.ErrInvalidRefreshToken),
		errors.Is(err, auth.ErrRefreshTokenExpired):
		writeError(w, http.StatusUnauthorized, err.Error())

	case errors.Is(err, auth.ErrEmailNotVerified):
		writeError(w, http.StatusForbidden, err.Error())

	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
