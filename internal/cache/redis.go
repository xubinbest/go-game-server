package cache

import (
	"context"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client     redis.Cmdable
	mu         sync.Mutex
	lockTokens map[string]string // 进程内记录：每个锁键对应的持有者token
	pubsubs    map[string][]*redis.PubSub
	watchdogs  map[string]watchdogEntry
}

type watchdogEntry struct {
	id     string
	cancel context.CancelFunc
}

func NewRedisCache(cfg *config.Config) (Cache, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Redis.Cluster,
		Password:     cfg.Redis.Password,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		OnConnect:    func(ctx context.Context, cn *redis.Conn) error { return nil },
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	cache := &RedisCache{
		client:     client,
		lockTokens: make(map[string]string),
		pubsubs:    make(map[string][]*redis.PubSub),
		watchdogs:  make(map[string]watchdogEntry),
	}

	return cache, nil
}

func (r *RedisCache) Close() error {
	switch c := r.client.(type) {
	case *redis.Client:
		return c.Close()
	case *redis.ClusterClient:
		return c.Close()
	default:
		return nil
	}
}

func (r *RedisCache) SPop(ctx context.Context, key string) *redis.StringCmd {
	return r.client.SPop(ctx, key)
}

func (r *RedisCache) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return r.client.SAdd(ctx, key, members...)
}

func (r *RedisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.client.SetNX(ctx, key, value, expiration)
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) *redis.StringCmd {
	return r.client.Get(ctx, key)
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisCache) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return r.client.ZAdd(ctx, key, members...)
}

func (r *RedisCache) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return r.client.ZRevRangeWithScores(ctx, key, start, stop)
}

func (r *RedisCache) ZRevRank(ctx context.Context, key, member string) *redis.IntCmd {
	return r.client.ZRevRank(ctx, key, member)
}

func (r *RedisCache) ZScore(ctx context.Context, key, member string) *redis.FloatCmd {
	return r.client.ZScore(ctx, key, member)
}
