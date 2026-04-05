package httptransport

import (
	"errors"
	"strings"
)

var errInvalidRequest = errors.New("invalid request body")

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r registerRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errInvalidRequest
	}

	if r.Password == "" {
		return errInvalidRequest
	}

	return nil
}

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (r verifyEmailRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errInvalidRequest
	}

	if strings.TrimSpace(r.Code) == "" {
		return errInvalidRequest
	}

	return nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r loginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errInvalidRequest
	}

	if r.Password == "" {
		return errInvalidRequest
	}

	return nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type meResponse struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	Roles         []string `json:"roles"`
	EmailVerified bool     `json:"email_verified"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}
