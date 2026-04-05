package httptransport

import (
	"errors"
	"net/http"

	"github.com/nikallow/auth-service/internal/auth"
)

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, err := h.authService.Register(r.Context(), auth.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, messageResponse{
		Message: "user registered, verification required",
	})
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.authService.VerifyEmail(r.Context(), auth.VerifyEmailInput{
		Email: req.Email,
		Code:  req.Code,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messageResponse{
		Message: "email verified",
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.authService.Login(r.Context(), auth.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: r.UserAgent(),
		IP:        clientIP(r),
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken, result.RefreshTokenExpiresAt, h.refreshCookieSecure)

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresInSeconds(result.AccessTokenExpiresAt),
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshTokenCookie(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	result, err := h.authService.Refresh(r.Context(), auth.RefreshInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			clearRefreshTokenCookie(w, h.refreshCookieSecure)
		}

		h.writeAuthError(w, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken, result.RefreshTokenExpiresAt, h.refreshCookieSecure)

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresInSeconds(result.AccessTokenExpiresAt),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := readRefreshTokenCookie(r)
	if err != nil {
		clearRefreshTokenCookie(w, h.refreshCookieSecure)
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	err = h.authService.Logout(r.Context(), auth.LogoutInput{
		RefreshToken: refreshToken,
	})
	if err != nil {
		if errors.Is(err, auth.ErrInvalidRefreshToken) {
			clearRefreshTokenCookie(w, h.refreshCookieSecure)
		}

		h.writeAuthError(w, err)
		return
	}

	clearRefreshTokenCookie(w, h.refreshCookieSecure)

	writeJSON(w, http.StatusOK, messageResponse{
		Message: "session revoked",
	})
}
