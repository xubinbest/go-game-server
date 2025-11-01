package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheManager 缓存管理器
type CacheManager struct {
	cache Cache
	// 缓存指标（命中/未命中/错误），用于监控与调优
	metrics *CacheMetrics
}

func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache:   cache,
		metrics: NewCacheMetrics(),
	}
}

// MetricsStats 返回当前缓存指标统计
func (cm *CacheManager) MetricsStats() map[string]int64 {
	if cm.metrics == nil {
		return map[string]int64{"hits": 0, "misses": 0, "errors": 0}
	}
	return cm.metrics.GetStats()
}

// GetOrSet 获取或设置缓存（防止击穿）
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, strategy CacheStrategy,
	loader func() (interface{}, error)) (interface{}, error) {
	// 1) 首次读取缓存
	hit, cached, _ := cm.getCached(ctx, key)
	if hit {
		return cached, nil
	}
	if cm.metrics != nil {
		cm.metrics.IncrementMisses()
	}

	// 2) 获取锁或等待他人填充
	gotLock, unlock, waitedData, waitedNull, err := cm.acquireOrWait(ctx, key, strategy)
	if err == nil && (waitedData != nil || waitedNull) {
		if waitedNull {
			return nil, nil
		}
		return waitedData, nil
	}
	if err != nil {
		// 获取锁/等待失败（已在内部记录错误），继续尝试自行加载
	}
	if gotLock && unlock != nil {
		defer unlock()
	}

	// 3) 二次校验，避免竞态下的无谓回源
	hit2, cached2, _ := cm.getCached(ctx, key)
	if hit2 {
		return cached2, nil
	}

	// 4) 回源加载
	data, loadErr := loader()
	if loadErr != nil {
		if cm.metrics != nil {
			cm.metrics.IncrementErrors()
		}
		return nil, loadErr
	}

	// 5) 写回缓存（忽略写回错误，但计入指标）
	_ = cm.writeBack(ctx, key, data, strategy)
	return data, nil
}

// Invalidate 失效缓存
func (cm *CacheManager) Invalidate(ctx context.Context, key string) error {
	return cm.cache.Delete(ctx, key)
}

// InvalidatePattern 批量失效缓存
func (cm *CacheManager) InvalidatePattern(ctx context.Context, pattern string) error {
	// 集群模式需要遍历所有分片；单机直接 SCAN
	rc, ok := cm.cache.(*RedisCache)
	if !ok {
		return fmt.Errorf("unsupported cache implementation for pattern invalidation")
	}

	deleteByScan := func(c redis.Cmdable) error {
		var cursor uint64
		for {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			scan := c.Scan(ctx, cursor, pattern, 200)
			keys, next, err := scan.Result()
			if err != nil {
				return err
			}
			cursor = next
			if len(keys) > 0 {
				if err := c.Del(ctx, keys...).Err(); err != nil {
					return err
				}
			}
			if cursor == 0 {
				break
			}
		}
		return nil
	}

	switch c := rc.client.(type) {
	case *redis.Client:
		return deleteByScan(c)
	case *redis.ClusterClient:
		return c.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			return deleteByScan(shard)
		})
	default:
		return fmt.Errorf("unsupported redis client type for pattern invalidation")
	}
}

// waitForCache 等待缓存加载完成
func (cm *CacheManager) waitForCache(ctx context.Context, key string, timeout time.Duration) (interface{}, error) {
	deadline := time.Now().Add(timeout)
	backoff := 10 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		result, err := cm.cache.Get(ctx, key).Result()
		if err == nil {
			if result == "null" {
				if cm.metrics != nil {
					cm.metrics.IncrementHits()
				}
				return nil, nil
			}
			// 直接返回JSON字符串，让调用方自己反序列化
			if cm.metrics != nil {
				cm.metrics.IncrementHits()
			}
			return []byte(result), nil
		}
		if time.Now().After(deadline) {
			if cm.metrics != nil {
				cm.metrics.IncrementErrors()
			}
			return nil, fmt.Errorf("cache load timeout")
		}
		// 指数退避 + 抖动，减小热键竞争
		time.Sleep(backoff + randomJitter(10))
		if backoff < 200*time.Millisecond {
			backoff *= 2
		}
	}
}

// getCached 读取缓存，返回是否命中/空值/数据
func (cm *CacheManager) getCached(ctx context.Context, key string) (bool, []byte, error) {
	result, err := cm.cache.Get(ctx, key).Result()
	if err != nil || result == "null" {
		return false, nil, err
	}
	if cm.metrics != nil {
		cm.metrics.IncrementHits()
	}
	return true, []byte(result), nil
}

// acquireOrWait 获取锁或等待他人填充；返回是否持有锁、解锁函数、等待到的数据/空值
func (cm *CacheManager) acquireOrWait(ctx context.Context, key string, strategy CacheStrategy) (bool, func(), []byte, bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	locked, err := cm.cache.TryLock(ctx, lockKey, strategy.LockTimeout)
	if err != nil {
		utils.Error("尝试获取分布式锁失败", zap.Error(err))
		if cm.metrics != nil {
			cm.metrics.IncrementErrors()
		}
		if v, werr := cm.waitForCache(ctx, key, strategy.LockTimeout); werr == nil {
			if v == nil {
				return false, nil, nil, true, nil
			}
			if b, ok := v.([]byte); ok {
				return false, nil, b, false, nil
			}
			return false, nil, nil, false, nil
		}
		return false, nil, nil, false, err
	}
	if !locked {
		v, werr := cm.waitForCache(ctx, key, strategy.LockTimeout)
		if werr != nil {
			return false, nil, nil, false, werr
		}
		if v == nil {
			return false, nil, nil, true, nil
		}
		if b, ok := v.([]byte); ok {
			return false, nil, b, false, nil
		}
		return false, nil, nil, false, nil
	}
	// 成功加锁：启动看门狗并返回解锁函数
	if _, werr := cm.cache.StartWatchdog(ctx, lockKey, strategy.LockTimeout); werr != nil {
		utils.Error("start watchdog failed", zap.Error(werr))
		if cm.metrics != nil {
			cm.metrics.IncrementErrors()
		}
	}
	unlock := func() { _ = cm.cache.Unlock(ctx, lockKey) }
	return true, unlock, nil, false, nil
}

// writeBack 写回缓存（失败仅计入错误指标，不返回错误给调用者）
func (cm *CacheManager) writeBack(ctx context.Context, key string, data interface{}, strategy CacheStrategy) error {
	ttl := strategy.TTL
	if data == nil {
		ttl = strategy.EmptyTTL
	}
	if data != nil {
		if jsonData, mErr := json.Marshal(data); mErr == nil {
			if sErr := cm.cache.Set(ctx, key, jsonData, ttl); sErr != nil {
				if cm.metrics != nil {
					cm.metrics.IncrementErrors()
				}
				return sErr
			}
		} else {
			if cm.metrics != nil {
				cm.metrics.IncrementErrors()
			}
			return mErr
		}
	} else {
		if sErr := cm.cache.Set(ctx, key, "null", ttl); sErr != nil {
			if cm.metrics != nil {
				cm.metrics.IncrementErrors()
			}
			return sErr
		}
	}
	return nil
}
