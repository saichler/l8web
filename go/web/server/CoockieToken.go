package server

import (
	"net/http"
	"strings"
)

var BearerCookieName = "bToken"

func extractToken(r *http.Request) string {
	// 1. Try cookie first (primary method for browser requests)
	cookie, err := r.Cookie(BearerCookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// 2. Fallback to Authorization header (for API clients)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// 3. Fallback to query parameter (for initial page load redirect)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	return ""
}
