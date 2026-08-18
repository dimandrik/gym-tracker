package middleware

import (
	"context"
	"gym_tracker/internal/auth"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

// кладёт userID в контекст запроса под UserIDKey для последующих хендлеров
func AuthMiddleware(secret string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := auth.ValidateToken(tokenString, secret)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, token.UserID)
			next(w, r.WithContext(ctx))
		}
	}
}
