package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"room-booking/internal/interface/http/handler"
	"room-booking/internal/service"
)

func AuthMiddleware(jwtService *service.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization header format"})
				return
			}

			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				status := http.StatusUnauthorized
				msg := "invalid or expired token"
				switch {
				case errors.Is(err, service.ErrExpiredToken):
					msg = "token has expired"
				case errors.Is(err, service.ErrInvalidToken):
					msg = "invalid token signature"
				case errors.Is(err, service.ErrInvalidRole):
					msg = "insufficient permissions"
				}

				respondJSON(w, status, map[string]string{"error": msg})
				return
			}

			ctx := context.WithValue(r.Context(), handler.CtxUserID, claims.UserID)
			ctx = context.WithValue(ctx, handler.CtxUserRole, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(expectedRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(handler.CtxUserRole).(string)
			if !ok || role == "" {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}
			if expectedRole != "" && role != expectedRole {
				respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied: insufficient permissions"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAnyRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(handler.CtxUserRole).(string)
			if !ok || role == "" {
				respondJSON(w, http.StatusUnauthorized, map[string]string{"error": "authentication required"})
				return
			}

			for _, allowed := range allowedRoles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			respondJSON(w, http.StatusForbidden, map[string]string{"error": "access denied: insufficient permissions"})
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
