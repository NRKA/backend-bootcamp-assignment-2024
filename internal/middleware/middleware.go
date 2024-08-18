package middleware

import (
	"context"
	"net/http"
	"strings"
)

// TokenAuthenticator проверяет наличие и валидность токена в заголовке запроса.
func TokenAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Удаляем префикс "Bearer "
		token = strings.TrimPrefix(token, "Bearer ")

		// Проверка валидности токена и его типа
		// Здесь ваша логика проверки токенов и извлечения информации из них
		if !isValidToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Установите токен в контекст
		ctx := context.WithValue(r.Context(), "token", token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Проверка валидности токена (реализуйте по вашему усмотрению)
func isValidToken(token string) bool {
	// Ваш код для проверки токенов
	// Например, запрос к базе данных для проверки токена
	return true
}

// AuthOnlyMiddleware проверяет, что токен имеет права авторизованного пользователя.
func AuthOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token").(string)
		if !isAuthToken(token) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ModerationOnlyMiddleware проверяет, что токен имеет права модератора.
func ModerationOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token").(string)
		if !isModerationToken(token) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Проверка прав токена (реализуйте по вашему усмотрению)
func isAuthToken(token string) bool {
	// Ваша логика проверки, например, токен должен быть `auth_token_client` или `auth_token_moderator`
	return token == "auth_token_client" || token == "auth_token_moderator"
}

func isModerationToken(token string) bool {
	// Ваша логика проверки, например, токен должен быть `auth_token_moderator`
	return token == "auth_token_moderator" || isModeratorTokenFromDB(token)
}

func isModeratorTokenFromDB(token string) bool {
	// Запрос к базе данных для проверки, что токен принадлежит модератору
	return false
}
