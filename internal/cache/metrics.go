package cache

import (
	"github.xubinbest.com/go-game-server/internal/telemetry"
)

// CacheMetrics 缓存统计（使用Prometheus metrics）
type CacheMetrics struct {
	cacheType string
}

func NewCacheMetrics(cacheType string) *CacheMetrics {
	return &CacheMetrics{
		cacheType: cacheType,
	}
}

func (cm *CacheMetrics) IncrementHits() {
	telemetry.CacheHitsTotal.WithLabelValues(cm.cacheType).Inc()
	cm.updateHitRate()
}

func (cm *CacheMetrics) IncrementMisses() {
	telemetry.CacheMissesTotal.WithLabelValues(cm.cacheType).Inc()
	cm.updateHitRate()
}

func (cm *CacheMetrics) IncrementErrors() {
	telemetry.CacheErrorsTotal.WithLabelValues(cm.cacheType).Inc()
}

func (cm *CacheMetrics) updateHitRate() {
	// 命中率应该通过Prometheus查询计算
	// 在Grafana中使用: rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))
	// 这里不需要手动计算，Prometheus会自动处理
}

func (cm *CacheMetrics) HitRate() float64 {
	// 命中率应该通过Prometheus查询计算，这里返回0作为占位符
	// 实际命中率应该在Grafana中使用 rate(cache_hits_total) / (rate(cache_hits_total) + rate(cache_misses_total))
	return 0
}

func (cm *CacheMetrics) GetStats() map[string]int64 {
	// 返回空map，实际统计应该从Prometheus查询
	return map[string]int64{}
}
