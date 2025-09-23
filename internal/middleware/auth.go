package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.xubinbest.com/go-game-server/internal/auth"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

type contextKey string

const userContextKey contextKey = "user"

func NewAuthMiddleware(cfg config.AuthConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查请求路径是否在白名单中
			for _, path := range cfg.WhitelistPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			tokenStr := r.Header.Get("Authorization")
			if tokenStr == "" {
				sendError(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			utils.Info("Http Token", zap.String("token", tokenStr))
			claims, err := auth.ParseToken(tokenStr, cfg.SecretKey)

			if err != nil {
				sendError(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// 将用户信息添加到请求上下文中
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		}
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
