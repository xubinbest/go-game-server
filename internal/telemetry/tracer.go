package telemetry

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
)

// InitTracer 初始化OpenTelemetry Tracer
func InitTracer(cfg *config.TelemetryConfig, logger *zap.Logger) error {
	if !cfg.Enabled {
		logger.Info("Telemetry is disabled")
		return nil
	}

	// 创建OTLP HTTP导出器（Jaeger支持OTLP协议）
	// 端点格式: host:port (例如: jaeger:4318)
	// 如果端点包含http://或https://，需要先解析
	endpoint := cfg.OTLP.Endpoint
	if endpoint == "" {
		return fmt.Errorf("OTLP endpoint is required")
	}

	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // 默认使用HTTP，生产环境建议使用HTTPS
	)
	if err != nil {
		return fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 配置采样策略
	var sampler sdktrace.Sampler
	switch cfg.Sampling.Type {
	case "always_on":
		sampler = sdktrace.AlwaysSample()
	case "always_off":
		sampler = sdktrace.NeverSample()
	case "traceidratio":
		ratio := cfg.Sampling.Ratio
		if ratio <= 0 {
			ratio = 0.1 // 默认10%
		}
		if ratio > 1.0 {
			ratio = 1.0
		}
		sampler = sdktrace.TraceIDRatioBased(ratio)
	default:
		// 默认使用10%采样率
		sampler = sdktrace.TraceIDRatioBased(0.1)
	}

	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建TracerProvider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置为全局TracerProvider
	otel.SetTracerProvider(tracerProvider)

	// 创建Tracer
	tracer = otel.Tracer(cfg.ServiceName)

	logger.Info("OpenTelemetry tracer initialized",
		zap.String("service", cfg.ServiceName),
		zap.String("otlp_endpoint", cfg.OTLP.Endpoint),
		zap.String("sampling_type", cfg.Sampling.Type),
		zap.Float64("sampling_ratio", cfg.Sampling.Ratio),
	)

	return nil
}

// GetTracer 获取Tracer实例
func GetTracer() trace.Tracer {
	if tracer == nil {
		return otel.Tracer("default")
	}
	return tracer
}

// Shutdown 关闭TracerProvider
func Shutdown(ctx context.Context) error {
	if tracerProvider != nil {
		return tracerProvider.Shutdown(ctx)
	}
	return nil
}
