package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// CacheManager 缓存管理器
type CacheManager struct {
	cache Cache
}

func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// GetOrSet 获取或设置缓存（防止击穿）
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, strategy CacheStrategy,
	loader func() (interface{}, error)) (interface{}, error) {

	// 1. 尝试从缓存获取
	result, err := cm.cache.Get(ctx, key).Result()
	if err == nil {
		// 缓存命中
		if result == "null" {
			return nil, nil // 空值缓存
		}
		// 直接返回JSON字符串，让调用方自己反序列化
		return []byte(result), nil
	}

	// 2. 缓存未命中，使用分布式锁防止击穿
	if strategy.UseLock {
		lockKey := fmt.Sprintf("lock:%s", key)
		locked, err := cm.cache.TryLock(ctx, lockKey, strategy.LockTimeout)
		if err != nil {
			utils.Error("try lock failed", zap.Error(err))
		} else if !locked {
			// 等待其他进程加载完成
			return cm.waitForCache(ctx, key, strategy.LockTimeout)
		}
		defer cm.cache.Unlock(ctx, lockKey)
	}

	// 3. 从数据源加载
	data, err := loader()
	if err != nil {
		return nil, err
	}

	// 4. 设置缓存
	ttl := strategy.TTL
	if data == nil && strategy.CacheEmpty {
		ttl = strategy.EmptyTTL
	}

	if data != nil {
		if jsonData, err := json.Marshal(data); err == nil {
			cm.cache.Set(ctx, key, jsonData, ttl)
		}
	} else if strategy.CacheEmpty {
		// 缓存空值防止穿透
		cm.cache.Set(ctx, key, "null", ttl)
	}

	return data, nil
}

// Invalidate 失效缓存
func (cm *CacheManager) Invalidate(ctx context.Context, key string) error {
	return cm.cache.Delete(ctx, key)
}

// InvalidatePattern 批量失效缓存
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// 这里需要实现模式删除，可以使用Lua脚本
	// 暂时返回nil，后续可以扩展
	return nil
}

// waitForCache 等待缓存加载完成
func (cm *CacheManager) waitForCache(ctx context.Context, key string, timeout time.Duration) (interface{}, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := cm.cache.Get(ctx, key).Result()
		if err == nil {
			if result == "null" {
				return nil, nil
			}
			// 直接返回JSON字符串，让调用方自己反序列化
			return []byte(result), nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil, fmt.Errorf("cache load timeout")
}
