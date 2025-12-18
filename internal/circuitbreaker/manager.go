package circuitbreaker

import (
	"context"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// Manager 熔断器管理器
type Manager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
	config   config.CircuitBreakerConfig
	logger   *zap.Logger
}

// NewManager 创建熔断器管理器
func NewManager(cfg config.CircuitBreakerConfig, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = utils.GetLogger()
	}
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
		config:   cfg,
		logger:   logger,
	}
}

// GetBreaker 获取或创建熔断器
func (m *Manager) GetBreaker(key string) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[key]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	// 创建新的熔断器
	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if breaker, exists := m.breakers[key]; exists {
		return breaker
	}

	// 使用配置创建熔断器，设置默认值
	cbConfig := Config{
		FailureThreshold:    m.config.FailureThreshold,
		SuccessThreshold:    m.config.SuccessThreshold,
		Timeout:             m.config.Timeout,
		HalfOpenMaxRequests: m.config.HalfOpenMaxRequests,
		Enabled:             m.config.Enabled,
	}

	// 设置默认值
	if cbConfig.FailureThreshold == 0 {
		cbConfig.FailureThreshold = 5
	}
	if cbConfig.SuccessThreshold == 0 {
		cbConfig.SuccessThreshold = 2
	}
	if cbConfig.Timeout == 0 {
		cbConfig.Timeout = 60 * time.Second
	}
	if cbConfig.HalfOpenMaxRequests == 0 {
		cbConfig.HalfOpenMaxRequests = 3
	}

	breaker = NewCircuitBreaker(cbConfig, m.logger)
	m.breakers[key] = breaker
	return breaker
}

// Execute 执行函数，使用指定key的熔断器
func (m *Manager) Execute(ctx context.Context, key string, fn func() error) error {
	breaker := m.GetBreaker(key)
	return breaker.Execute(ctx, fn)
}

// GetBreakerStats 获取所有熔断器的统计信息
func (m *Manager) GetBreakerStats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for key, breaker := range m.breakers {
		stats[key] = breaker.GetStats()
	}
	return stats
}
