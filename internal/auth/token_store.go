package auth

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
)

// TokenStore 定义令牌存储接口，便于替换实现（如内存/Redis/Mongo）
type TokenStore interface {
	// SetToken 存储用户令牌，设置过期时间
	SetToken(ctx context.Context, userID int64, token string, expiration time.Duration) error
	// GetToken 获取用户令牌，未命中返回错误
	GetToken(ctx context.Context, userID int64) (string, error)
	// DeleteToken 删除用户令牌
	DeleteToken(ctx context.Context, userID int64) error
}

// RedisTokenStore 基于通用缓存的 Redis 实现
type RedisTokenStore struct {
	cacheClient cache.Cache
}

// NewRedisTokenStore 创建 RedisTokenStore
func NewRedisTokenStore(c cache.Cache) *RedisTokenStore {
	return &RedisTokenStore{cacheClient: c}
}

func (s *RedisTokenStore) SetToken(ctx context.Context, userID int64, token string, expiration time.Duration) error {
	return s.cacheClient.Set(ctx, tokenKey(userID), token, expiration)
}

func (s *RedisTokenStore) GetToken(ctx context.Context, userID int64) (string, error) {
	return s.cacheClient.Get(ctx, tokenKey(userID)).Result()
}

func (s *RedisTokenStore) DeleteToken(ctx context.Context, userID int64) error {
	return s.cacheClient.Delete(ctx, tokenKey(userID))
}

// tokenKey 生成令牌键
func tokenKey(userId int64) string {
	return fmt.Sprintf("token:%d", userId)
}
