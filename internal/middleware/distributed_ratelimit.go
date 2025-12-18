package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// DistributedRateLimiter 基于Redis的分布式限流器
type DistributedRateLimiter struct {
	cache  cache.Cache
	cfg    config.DistributedRateLimitConfig
	logger *zap.Logger
}

// NewDistributedRateLimitMiddleware 创建分布式限流中间件
func NewDistributedRateLimitMiddleware(
	cacheClient cache.Cache,
	cfg config.DistributedRateLimitConfig,
	logger *zap.Logger,
) func(http.HandlerFunc) http.HandlerFunc {
	// 设置默认窗口时间
	if cfg.Window == 0 {
		cfg.Window = time.Second
	}

	limiter := &DistributedRateLimiter{
		cache:  cacheClient,
		cfg:    cfg,
		logger: logger,
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 获取限流key（IP或用户ID）
			key := limiter.getRateLimitKey(r)

			// 检查是否允许通过
			allowed, remaining, err := limiter.allow(r.Context(), key)
			if err != nil {
				limiter.logger.Error("Rate limit check failed",
					zap.String("key", key),
					zap.Error(err))
				// Redis错误时，允许通过（fail-open策略）
				next.ServeHTTP(w, r)
				return
			}

			// 设置限流响应头
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.cfg.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.cfg.Window).Unix()))

			if !allowed {
				limiter.logger.Warn("Rate limit exceeded",
					zap.String("key", key),
					zap.Int("limit", limiter.cfg.RequestsPerSecond))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// getRateLimitKey 获取限流key
func (rl *DistributedRateLimiter) getRateLimitKey(r *http.Request) string {
	// 优先使用用户ID（如果已认证）
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return fmt.Sprintf("ratelimit:user:%s", userID)
	}

	// 使用IP地址
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return fmt.Sprintf("ratelimit:ip:%s", ip)
}

// allow 检查是否允许请求（滑动窗口算法）
func (rl *DistributedRateLimiter) allow(ctx context.Context, key string) (bool, int, error) {
	now := time.Now()
	windowStart := now.Add(-rl.cfg.Window)

	// 使用Redis ZSet实现滑动窗口
	// Key: ratelimit:{key}
	// Score: 时间戳（毫秒）
	// Member: 请求ID（UUID或时间戳+随机数）

	// 1. 移除窗口外的旧记录
	zsetKey := fmt.Sprintf("ratelimit:zset:%s", key)
	minScore := float64(windowStart.UnixMilli())

	// 2. 统计窗口内的请求数
	count, err := rl.cache.ZCount(ctx, zsetKey, minScore, float64(now.UnixMilli())).Result()
	if err != nil {
		// Redis错误时返回允许通过（fail-open策略）
		rl.logger.Warn("Failed to count rate limit", zap.Error(err))
		return true, rl.cfg.RequestsPerSecond, nil
	}

	// 3. 检查是否超过限制
	if int(count) >= rl.cfg.RequestsPerSecond {
		remaining := 0
		return false, remaining, nil
	}

	// 4. 添加当前请求记录
	member := fmt.Sprintf("%d:%d", now.UnixMilli(), now.Nanosecond())
	score := float64(now.UnixMilli())
	if err := rl.cache.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: member,
	}).Err(); err != nil {
		rl.logger.Warn("Failed to add rate limit record", zap.Error(err))
		return true, rl.cfg.RequestsPerSecond, nil
	}

	// 5. 设置过期时间（窗口大小 + 1秒缓冲）
	if err := rl.cache.Expire(ctx, zsetKey, rl.cfg.Window+time.Second); err != nil {
		rl.logger.Warn("Failed to set expire for rate limit key", zap.Error(err))
	}

	remaining := rl.cfg.RequestsPerSecond - int(count) - 1
	return true, remaining, nil
}
