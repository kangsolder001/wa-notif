package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"wa-notif/internal/usecase/dto"
)

type AuthMiddleware struct {
	bearerToken string
}

func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{
		bearerToken: strings.TrimSpace(token),
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no token is configured, allow requests (open local access)
		if m.bearerToken == "" {
			next(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		var token string

		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		} else if queryToken := r.URL.Query().Get("token"); queryToken != "" {
			// Fallback: allow query param for browser/image requests
			token = strings.TrimSpace(queryToken)
		}

		if token == "" || token != m.bearerToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(dto.APIResponse{
				Success: false,
				Error: &dto.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Missing or invalid Bearer token in Authorization header",
				},
			})
			return
		}

		next(w, r)
	}
}
