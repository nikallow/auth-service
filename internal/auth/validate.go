package auth

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeVerificationCode(code string) string {
	return strings.TrimSpace(code)
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}

	if addr.Address != email {
		return ErrInvalidEmail
	}

	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrInvalidPassword
	}

	if utf8.RuneCountInString(password) < minPasswordLength {
		return ErrInvalidPassword
	}

	if len(password) > maxPasswordBytes {
		return ErrInvalidPassword
	}

	return nil
}

func validateVerificationCode(code string) error {
	if len(code) != verificationCodeLength {
		return ErrInvalidVerificationCode
	}

	for i := 0; i < len(code); i++ {
		if code[i] < '0' || code[i] > '9' {
			return ErrInvalidVerificationCode
		}
	}

	return nil
}