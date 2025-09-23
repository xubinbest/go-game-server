package cache

import (
	"sync/atomic"
)

// CacheMetrics 缓存统计
type CacheMetrics struct {
	hits   int64
	misses int64
	errors int64
}

func NewCacheMetrics() *CacheMetrics {
	return &CacheMetrics{}
}

func (cm *CacheMetrics) IncrementHits() {
	atomic.AddInt64(&cm.hits, 1)
}

func (cm *CacheMetrics) IncrementMisses() {
	atomic.AddInt64(&cm.misses, 1)
}

func (cm *CacheMetrics) IncrementErrors() {
	atomic.AddInt64(&cm.errors, 1)
}

func (cm *CacheMetrics) HitRate() float64 {
	total := atomic.LoadInt64(&cm.hits) + atomic.LoadInt64(&cm.misses)
	if total == 0 {
		return 0
	}
	return float64(atomic.LoadInt64(&cm.hits)) / float64(total)
}

func (cm *CacheMetrics) GetStats() map[string]int64 {
	return map[string]int64{
		"hits":   atomic.LoadInt64(&cm.hits),
		"misses": atomic.LoadInt64(&cm.misses),
		"errors": atomic.LoadInt64(&cm.errors),
	}
}
