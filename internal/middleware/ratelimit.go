package middleware

import (
	"net/http"
	"sync"

	"github.xubinbest.com/go-game-server/internal/config"

	"go.uber.org/ratelimit"
)

type rateLimiter struct {
	limiters map[string]ratelimit.Limiter
	mu       sync.Mutex
	cfg      config.RateLimitConfig
}

func NewRateLimitMiddleware(cfg config.RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	rl := &rateLimiter{
		limiters: make(map[string]ratelimit.Limiter),
		cfg:      cfg,
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 获取客户端IP作为限流key
			ip := r.RemoteAddr

			// 获取或创建该IP的限流器
			rl.mu.Lock()
			limiter, exists := rl.limiters[ip]
			if !exists {
				limiter = ratelimit.New(rl.cfg.RequestsPerSecond)
				rl.limiters[ip] = limiter
			}
			rl.mu.Unlock()

			// Take()会阻塞直到允许请求通过
			limiter.Take()

			next.ServeHTTP(w, r)
		}
	}
}
