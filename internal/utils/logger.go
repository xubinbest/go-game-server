package utils

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// InitLogger 初始化日志
func InitLogger() {
	once.Do(func() {
		// 配置zap
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// 创建logger
		var err error
		logger, err = config.Build()
		if err != nil {
			panic("Failed to initialize logger: " + err.Error())
		}

		// 替换zap包中全局的logger实例
		zap.ReplaceGlobals(logger)
	})
}

// Info 输出信息日志
func Info(msg string, fields ...zap.Field) {
	InitLogger()
	logger.Info(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...zap.Field) {
	InitLogger()
	logger.Error(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...zap.Field) {
	InitLogger()
	logger.Warn(msg, fields...)
}

// Debug 输出调试日志
func Debug(msg string, fields ...zap.Field) {
	InitLogger()
	logger.Debug(msg, fields...)
}

// Fatal 输出致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
	InitLogger()
	logger.Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *zap.Logger {
	InitLogger()
	return logger.With(zap.Any(key, value))
}

// WithFields 添加多个字段
func WithFields(fields ...zap.Field) *zap.Logger {
	InitLogger()
	return logger.With(fields...)
}

// GetLogger 获取logger实例
func GetLogger() *zap.Logger {
	InitLogger()
	return logger
}
