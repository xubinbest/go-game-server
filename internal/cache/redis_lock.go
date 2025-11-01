package cache

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

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
