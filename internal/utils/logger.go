package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
)

// NewLogger 创建新的logger实例（推荐使用）
func NewLogger() (*zap.Logger, error) {
	// 配置zap
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建logger
	log, err := config.Build()
	if err != nil {
		return nil, err
	}

	// 替换zap包中全局的logger实例
	zap.ReplaceGlobals(log)

	return log, nil
}

// SetLogger 设置全局logger实例（用于依赖注入）
func SetLogger(log *zap.Logger) {
	logger = log
	zap.ReplaceGlobals(log)
}

// GetLogger 获取logger实例，如果未设置则创建默认实例
func GetLogger() *zap.Logger {
	if logger == nil {
		// 如果未设置，创建默认logger（向后兼容）
		var err error
		logger, err = NewLogger()
		if err != nil {
			// 如果创建失败，使用zap的全局logger
			return zap.L()
		}
	}
	return logger
}

// Info 输出信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Debug 输出调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Fatal 输出致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *zap.Logger {
	return GetLogger().With(zap.Any(key, value))
}

// WithFields 添加多个字段
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}
