package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Close() error
	SetToken(ctx context.Context, userID int64, token string, expiration time.Duration) error
	GetToken(ctx context.Context, userID int64) (string, error)
	DeleteToken(ctx context.Context, userID int64) error

	// 通用缓存方法
	SPop(ctx context.Context, key string) *redis.StringCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) *redis.StringCmd
	Delete(ctx context.Context, key string) error
	Expire(ctx context.Context, key string, expiration time.Duration) error
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd
	ZRevRank(ctx context.Context, key, member string) *redis.IntCmd
	ZScore(ctx context.Context, key, member string) *redis.FloatCmd

	// 分布式锁方法
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Lock(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) error
	Unlock(ctx context.Context, key string) error

	// 订阅/发布
	Subscribe(ctx context.Context, channel string) (<-chan interface{}, error)
	Unsubscribe(ctx context.Context, channel string) error
	Publish(ctx context.Context, channel string, message interface{}) error
}
