package middleware

import (
	"context"
	auth2 "github.com/NRKA/backend-bootcamp-assignment-2024/internal/service/auth"
	"net/http"
	"strings"
)

const (
	client    = "client"
	moderator = "moderator"
	claimsKey = "claims"
)

func TokenAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := auth2.ParseToken(tokenString)
		if err != nil {
			switch err {
			case auth2.ErrTokenExpired:
				http.Error(w, "Token expired", http.StatusUnauthorized)
			case auth2.ErrTokenNotValidYet:
				http.Error(w, "Token not valid yet", http.StatusUnauthorized)
			case auth2.ErrTokenInvalid:
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			default:
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AuthOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(claimsKey).(*auth2.Claims)
		if claims.Role != client && claims.Role != moderator {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ModerationOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value(claimsKey).(*auth2.Claims)
		if claims.Role != moderator {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
