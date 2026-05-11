package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/luqmanshaban/kuda/repository"
	"github.com/luqmanshaban/kuda/utils"
)

type AuthMiddleware struct {
	Repo *repository.UserRepository
}

func NewAuthMiddleware(repo *repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{Repo: repo}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := r.Header.Get("Authorization")
		if k == "" {
			utils.WriteJson(w, http.StatusUnauthorized, map[string]string{"message": "api key missing"})
			return 
		}

		user, err := m.Repo.GetUserWithApi(k)
		if err != nil {
			slog.Error("authorizing user", "repository", "middleware", "op", "auth_user", "err", err)
			utils.WriteJson(w, http.StatusInternalServerError, map[string]string{"message": "invalid api key"})
			return 
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}