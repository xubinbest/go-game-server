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

// Lock 获取分布式锁（内置看门狗），会重试直到获取成功或超时
// 看门狗会在 ttl/3 间隔进行续约，直到调用 Unlock 或 ctx 取消
func (r *RedisCache) Lock(ctx context.Context, key string, ttl time.Duration, timeout time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("invalid ttl: %v", ttl)
	}
	if timeout <= 0 {
		return fmt.Errorf("invalid timeout: %v", timeout)
	}
	if err := r.acquireLockWithRetry(ctx, key, ttl, timeout); err != nil {
		return err
	}
	// 启动看门狗（若已存在旧的看门狗，先取消）
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

	return nil
}

// StartWatchdog 在已持有锁的前提下启动看门狗（供 TryLock 成功后使用）
func (r *RedisCache) StartWatchdog(ctx context.Context, key string, ttl time.Duration) (context.CancelFunc, error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("invalid ttl: %v", ttl)
	}
	// 校验是否当前进程持有该锁
	r.mu.Lock()
	if _, ok := r.lockTokens[key]; !ok {
		r.mu.Unlock()
		return nil, fmt.Errorf("watchdog start failed: not lock owner for key=%s", key)
	}
	// 若已有看门狗，先取消
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

	return func() {
		r.mu.Lock()
		if cur, ok := r.watchdogs[key]; ok && cur.id == rid {
			delete(r.watchdogs, key)
		}
		r.mu.Unlock()
		cancel()
	}, nil
}

// acquireLockWithRetry 带重试的获取锁
func (r *RedisCache) acquireLockWithRetry(ctx context.Context, key string, ttl, timeout time.Duration) error {
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
