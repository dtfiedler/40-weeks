package middleware

import (
	"context"
	"net/http"
	"strings"

	"simple-go/api/config"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID  int    `json:"user_id"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// ClaimsKey is the context key for JWT claims
type contextKey string
const ClaimsKey contextKey = "claims"

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Get claims from context
		claims, ok := r.Context().Value(ClaimsKey).(*Claims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Check if user is admin from JWT claims
		if !claims.IsAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next(w, r)
	})
}