package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client     redis.Cmdable
	mu         sync.Mutex
	lockTokens map[string]string // 进程内记录：每个锁键对应的持有者token
	pubsubs    map[string]*redis.PubSub
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

	cache := &RedisCache{
		client:     client,
		lockTokens: make(map[string]string),
		pubsubs:    make(map[string]*redis.PubSub),
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

// Renew 尝试续约锁TTL，仅当当前进程仍为持有者时生效
func (r *RedisCache) Renew(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, fmt.Errorf("invalid ttl: %v", ttl)
	}
	r.mu.Lock()
	token, ok := r.lockTokens[key]
	r.mu.Unlock()
	if !ok || token == "" {
		return false, nil
	}
	res, err := renewLua.Run(ctx, r.client, []string{key}, token, ttl.Milliseconds()).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// TryLock 尝试获取分布式锁
func (r *RedisCache) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return false, fmt.Errorf("invalid ttl: %v", ttl)
	}
	token, err := generateToken()
	if err != nil {
		return false, err
	}
	// 注意：此处直接使用传入的 key，调用方需传完整锁键
	ok, err := r.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil || !ok {
		return ok, err
	}
	r.setTokenForKey(key, token)
	return true, nil
}

// Lock 获取分布式锁，会重试直到获取成功或超时
func (r *RedisCache) Lock(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("invalid ttl: %v", ttl)
	}
	if timeout <= 0 {
		return fmt.Errorf("invalid timeout: %v", timeout)
	}
	start := time.Now()
	backoff := 25 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
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
		time.Sleep(backoff + randomJitter(20))
		if backoff < 500*time.Millisecond {
			backoff *= 2
		}
	}
}

// LockWithWatchdog 获取分布式锁并启动自动续约，看门狗会在 ttl/3 间隔续约，直到取消
// 返回的 cancel 可手动停止看门狗；调用 Unlock 也会自动停止
func (r *RedisCache) LockWithWatchdog(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) (context.CancelFunc, error) {
	if err := r.Lock(ctx, key, ttl, timeout); err != nil {
		return nil, err
	}
	// 若已有看门狗，先取消（防重入场景调用）
	r.mu.Lock()
	if old, ok := r.watchdogs[key]; ok {
		delete(r.watchdogs, key)
		r.mu.Unlock()
		old.cancel()
	} else {
		r.mu.Unlock()
	}

	cctx, cancel := context.WithCancel(context.Background())
	rid, err := generateToken()
	if err != nil {
		rid = fmt.Sprintf("wd-%d", time.Now().UnixNano())
	}
	r.mu.Lock()
	r.watchdogs[key] = watchdogEntry{id: rid, cancel: cancel}
	r.mu.Unlock()

	interval := ttl / 3
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-cctx.Done():
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				ok, err := r.Renew(cctx, key, ttl)
				if err != nil {
					utils.Error("redis lock renew failed", zap.Error(err))
				}
				if !ok {
					// 不是持有者或锁已失效，停止看门狗
					return
				}
			}
		}
	}()

	myId := rid
	return func() {
		r.mu.Lock()
		if cur, ok := r.watchdogs[key]; ok && cur.id == myId {
			delete(r.watchdogs, key)
		}
		r.mu.Unlock()
		cancel()
	}, nil
}

// Unlock 释放分布式锁
func (r *RedisCache) Unlock(ctx context.Context, key string) error {
	// 优先停止看门狗
	r.mu.Lock()
	if entry, ok := r.watchdogs[key]; ok {
		delete(r.watchdogs, key)
		r.mu.Unlock()
		entry.cancel()
	} else {
		r.mu.Unlock()
	}

	// 仅当当前进程持有 token 时才尝试释放
	if token, ok := r.popTokenForKey(key); ok {
		_, err := unlockLua.Run(ctx, r.client, []string{key}, token).Result()
		return err
	}
	// 未找到 token：可能是未持有或进程重启后遗失，交给 TTL 过期
	return nil
}

func tokenKey(userId int64) string {
	return fmt.Sprintf("toke:%d", userId)
}

// Subscribe 订阅频道
func (r *RedisCache) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	// 复用或创建该频道的 PubSub
	r.mu.Lock()
	pubsub, ok := r.pubsubs[channel]
	if !ok {
		switch c := r.client.(type) {
		case *redis.Client:
			pubsub = c.Subscribe(ctx, channel)
		case *redis.ClusterClient:
			pubsub = c.Subscribe(ctx, channel)
		default:
			r.mu.Unlock()
			return nil, fmt.Errorf("unsupported client type for Subscribe")
		}
		r.pubsubs[channel] = pubsub
	}
	r.mu.Unlock()

	ch := make(chan interface{})

	go func(ps *redis.PubSub, out chan interface{}) {
		for msg := range ps.Channel() {
			out <- msg.Payload
		}
		close(out)
	}(pubsub, ch)

	return ch, nil
}

// Unsubscribe 取消订阅频道
func (r *RedisCache) Unsubscribe(ctx context.Context, channel string) error {
	r.mu.Lock()
	ps, ok := r.pubsubs[channel]
	if ok {
		delete(r.pubsubs, channel)
	}
	r.mu.Unlock()
	if !ok || ps == nil {
		// 没有找到对应订阅，视为已退订
		return nil
	}
	// 退订该频道并关闭 PubSub（单通道场景可直接关闭）
	if err := ps.Unsubscribe(ctx, channel); err != nil {
		return err
	}
	return ps.Close()
}

// Publish 发布消息到频道
func (r *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
