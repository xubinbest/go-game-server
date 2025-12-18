package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// State 熔断器状态
type State int32

const (
	StateClosed   State = iota // 关闭状态：正常处理请求
	StateOpen                  // 开启状态：拒绝请求，直接返回错误
	StateHalfOpen              // 半开状态：允许少量请求通过，用于探测服务是否恢复
)

// String 返回状态字符串
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config 熔断器配置
type Config struct {
	// 失败阈值：连续失败多少次后开启熔断
	FailureThreshold int
	// 成功阈值：半开状态下成功多少次后关闭熔断
	SuccessThreshold int
	// 超时时间：开启状态持续多久后进入半开状态
	Timeout time.Duration
	// 半开状态最大请求数：半开状态下最多允许多少个请求通过
	HalfOpenMaxRequests int
	// 是否启用熔断器
	Enabled bool
}

// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		Timeout:             60 * time.Second,
		HalfOpenMaxRequests: 3,
		Enabled:             true,
	}
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	config Config
	logger *zap.Logger

	// 状态
	state int32 // State

	// 统计信息
	failureCount    int64 // 当前失败计数
	successCount    int64 // 半开状态下的成功计数
	halfOpenRequest int64 // 半开状态下的请求计数
	lastFailureTime int64 // 最后失败时间（UnixNano）

	// 互斥锁
	mu sync.RWMutex
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config Config, logger *zap.Logger) *CircuitBreaker {
	if logger == nil {
		logger = utils.GetLogger()
	}
	return &CircuitBreaker{
		config: config,
		logger: logger,
		state:  int32(StateClosed),
	}
}

// Execute 执行函数，带熔断保护
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.config.Enabled {
		return fn()
	}

	// 检查状态
	state := cb.getState()
	if state == StateOpen {
		// 检查是否超时，可以进入半开状态
		if cb.shouldAttemptReset() {
			cb.setState(StateHalfOpen)
			atomic.StoreInt64(&cb.halfOpenRequest, 0)
			atomic.StoreInt64(&cb.successCount, 0)
			state = StateHalfOpen
		} else {
			cb.logger.Debug("Circuit breaker is OPEN, request rejected",
				zap.String("state", state.String()))
			return errors.New("circuit breaker is open")
		}
	}

	// 半开状态下检查请求数限制
	if state == StateHalfOpen {
		requests := atomic.AddInt64(&cb.halfOpenRequest, 1)
		if requests > int64(cb.config.HalfOpenMaxRequests) {
			atomic.AddInt64(&cb.halfOpenRequest, -1)
			cb.logger.Debug("Half-open max requests exceeded",
				zap.Int64("requests", requests))
			return errors.New("circuit breaker half-open max requests exceeded")
		}
	}

	// 执行函数
	err := fn()

	// 根据结果更新状态
	cb.onResult(err, state)
	return err
}

// getState 获取当前状态
func (cb *CircuitBreaker) getState() State {
	return State(atomic.LoadInt32(&cb.state))
}

// setState 设置状态
func (cb *CircuitBreaker) setState(newState State) {
	oldState := State(atomic.SwapInt32(&cb.state, int32(newState)))
	if oldState != newState {
		cb.logger.Info("Circuit breaker state changed",
			zap.String("old", oldState.String()),
			zap.String("new", newState.String()))
	}
}

// onResult 处理执行结果
func (cb *CircuitBreaker) onResult(err error, currentState State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// 执行失败
		cb.onFailure(currentState)
	} else {
		// 执行成功
		cb.onSuccess(currentState)
	}
}

// onFailure 处理失败
func (cb *CircuitBreaker) onFailure(currentState State) {
	atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())

	switch currentState {
	case StateClosed:
		// 关闭状态：增加失败计数
		failures := atomic.AddInt64(&cb.failureCount, 1)
		if failures >= int64(cb.config.FailureThreshold) {
			cb.setState(StateOpen)
			atomic.StoreInt64(&cb.failureCount, 0)
			cb.logger.Warn("Circuit breaker opened due to failures",
				zap.Int64("failures", failures))
		}

	case StateHalfOpen:
		// 半开状态：失败则重新开启熔断
		cb.setState(StateOpen)
		atomic.StoreInt64(&cb.halfOpenRequest, 0)
		atomic.StoreInt64(&cb.successCount, 0)
		cb.logger.Warn("Circuit breaker reopened due to failure in half-open state")

	default:
		// 开启状态：不需要处理
	}
}

// onSuccess 处理成功
func (cb *CircuitBreaker) onSuccess(currentState State) {
	switch currentState {
	case StateClosed:
		// 关闭状态：重置失败计数
		atomic.StoreInt64(&cb.failureCount, 0)

	case StateHalfOpen:
		// 半开状态：增加成功计数
		successes := atomic.AddInt64(&cb.successCount, 1)
		if successes >= int64(cb.config.SuccessThreshold) {
			cb.setState(StateClosed)
			atomic.StoreInt64(&cb.halfOpenRequest, 0)
			atomic.StoreInt64(&cb.successCount, 0)
			atomic.StoreInt64(&cb.failureCount, 0)
			cb.logger.Info("Circuit breaker closed after successful recovery",
				zap.Int64("successes", successes))
		}

	default:
		// 开启状态：不应该执行到这里
	}
}

// shouldAttemptReset 检查是否应该尝试重置（从开启状态进入半开状态）
func (cb *CircuitBreaker) shouldAttemptReset() bool {
	lastFailure := atomic.LoadInt64(&cb.lastFailureTime)
	if lastFailure == 0 {
		return true
	}
	elapsed := time.Since(time.Unix(0, lastFailure))
	return elapsed >= cb.config.Timeout
}

// GetState 获取当前状态（用于监控）
func (cb *CircuitBreaker) GetState() State {
	return cb.getState()
}

// GetStats 获取统计信息（用于监控）
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.getState().String(),
		"failure_count":     atomic.LoadInt64(&cb.failureCount),
		"success_count":     atomic.LoadInt64(&cb.successCount),
		"half_open_request": atomic.LoadInt64(&cb.halfOpenRequest),
	}
}
