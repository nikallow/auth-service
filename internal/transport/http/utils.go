package httptransport

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

func expiresInSeconds(expiresAt time.Time) int {
	seconds := int(time.Until(expiresAt).Seconds())
	if seconds <= 0 {
		return 0
	}

	return seconds
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}

func bearerTokenFromRequest(r *http.Request) (string, error) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if authorization == "" {
		return "", errors.New("authorization header is empty")
	}

	parts := strings.Fields(authorization)
	if len(parts) != 2 {
		return "", errors.New("authorization header format is invalid")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization scheme is invalid")
	}

	if parts[1] == "" {
		return "", errors.New("bearer token is empty")
	}

	return parts[1], nil
}
