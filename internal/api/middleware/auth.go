package middleware

import (
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/kuda/internal/api/handlers"
)

func AuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("Authorization")
			if key == "" || key != apiKey {
			slog.Error("Auth failed", "received", key, "expected", apiKey)
				handlers.WriteJson(w, http.StatusUnauthorized, map[string]string{
					"message": "invalid or missing api key",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
