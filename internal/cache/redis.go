package cache

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client redis.Cmdable
}

func NewRedisCache(cfg *config.Config) (Cache, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Redis.Cluster,
		Password:     cfg.Redis.Password,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			utils.Info("Connected to Redis cluster", zap.String("connection", cn.String()))
			return nil
		},
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		utils.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	cache := &RedisCache{client: client}

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

func (r *RedisCache) SetToken(ctx context.Context, userID int64, token string, expiration time.Duration) error {
	return r.client.Set(ctx, tokenKey(userID), token, expiration).Err()
}

func (r *RedisCache) GetToken(ctx context.Context, userID int64) (string, error) {
	return r.client.Get(ctx, tokenKey(userID)).Result()
}

func (r *RedisCache) DeleteToken(ctx context.Context, userID int64) error {
	return r.client.Del(ctx, tokenKey(userID)).Err()
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

// TryLock 尝试获取分布式锁
func (r *RedisCache) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, lockKey(key), "1", ttl).Result()
}

// Lock 获取分布式锁，会重试直到获取成功或超时
func (r *RedisCache) Lock(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) error {
	start := time.Now()
	for {
		ok, err := r.TryLock(ctx, key, ttl)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("acquire lock timeout")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Unlock 释放分布式锁
func (r *RedisCache) Unlock(ctx context.Context, key string) error {
	return r.client.Del(ctx, lockKey(key)).Err()
}

func tokenKey(userId int64) string {
	return fmt.Sprintf("toke:%d", userId)
}

func lockKey(key string) string {
	return fmt.Sprintf("lock:%s", key)
}

// Subscribe 订阅频道
func (r *RedisCache) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	var pubsub *redis.PubSub
	switch c := r.client.(type) {
	case *redis.Client:
		pubsub = c.Subscribe(ctx, channel)
	case *redis.ClusterClient:
		pubsub = c.Subscribe(ctx, channel)
	default:
		return nil, fmt.Errorf("unsupported client type for Subscribe")
	}
	ch := make(chan interface{})

	go func() {
		for msg := range pubsub.Channel() {
			ch <- msg.Payload
		}
	}()

	return ch, nil
}

// Unsubscribe 取消订阅频道
func (r *RedisCache) Unsubscribe(ctx context.Context, channel string) error {

	var pubsub *redis.PubSub
	switch c := r.client.(type) {
	case *redis.Client:
		pubsub = c.Subscribe(ctx, channel)
	case *redis.ClusterClient:
		pubsub = c.Subscribe(ctx, channel)
	default:
		return fmt.Errorf("unsupported client type for Subscribe")
	}
	return pubsub.Unsubscribe(ctx, channel)
}

// Publish 发布消息到频道
func (r *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
