package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/grpc/status"
)

// NewServerStatsHandler 创建gRPC服务端StatsHandler（用于追踪）
func NewServerStatsHandler() stats.Handler {
	return otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	)
}

// NewClientStatsHandler 创建gRPC客户端StatsHandler（用于追踪）
func NewClientStatsHandler() stats.Handler {
	return otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(otel.GetTracerProvider()),
		otelgrpc.WithPropagators(otel.GetTextMapPropagator()),
	)
}

// UnaryServerTracingInterceptor 创建gRPC服务端追踪拦截器（兼容旧API）
// 注意：新版本otelgrpc使用StatsHandler，此函数返回空拦截器，实际应使用StatsHandler
func UnaryServerTracingInterceptor() grpc.UnaryServerInterceptor {
	// 在新版本中，追踪通过StatsHandler处理，这里返回一个空拦截器
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

// UnaryClientTracingInterceptor 创建gRPC客户端追踪拦截器（兼容旧API）
// 注意：新版本otelgrpc使用StatsHandler，此函数返回空拦截器，实际应使用StatsHandler
func UnaryClientTracingInterceptor() grpc.UnaryClientInterceptor {
	// 在新版本中，追踪通过StatsHandler处理，这里返回一个空拦截器
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// UnaryServerMetricsInterceptor 创建gRPC服务端metrics拦截器
func UnaryServerMetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// 提取服务名和方法名
		service, method := parseMethod(info.FullMethod)

		// 调用处理函数
		resp, err := handler(ctx, req)

		// 计算延迟
		duration := time.Since(start)

		// 获取状态码
		statusCode := "OK"
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code().String()
			} else {
				statusCode = "Unknown"
			}
		}

		// 记录metrics
		GRPCServerRequestsTotal.WithLabelValues(service, method, statusCode).Inc()
		GRPCServerRequestDuration.WithLabelValues(service, method, statusCode).Observe(duration.Seconds())

		return resp, err
	}
}

// UnaryClientMetricsInterceptor 创建gRPC客户端metrics拦截器
func UnaryClientMetricsInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		// 提取服务名和方法名
		service, methodName := parseMethod(method)

		// 调用invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// 计算延迟
		duration := time.Since(start)

		// 获取状态码
		statusCode := "OK"
		if err != nil {
			if s, ok := status.FromError(err); ok {
				statusCode = s.Code().String()
			} else {
				statusCode = "Unknown"
			}
		}

		// 记录metrics
		GRPCClientRequestsTotal.WithLabelValues(service, methodName, statusCode).Inc()
		GRPCClientRequestDuration.WithLabelValues(service, methodName, statusCode).Observe(duration.Seconds())

		return err
	}
}

// StreamServerTracingInterceptor 创建gRPC流式服务端追踪拦截器（兼容旧API）
// 注意：新版本otelgrpc使用StatsHandler，此函数返回空拦截器，实际应使用StatsHandler
func StreamServerTracingInterceptor() grpc.StreamServerInterceptor {
	// 在新版本中，追踪通过StatsHandler处理，这里返回一个空拦截器
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
}

// StreamClientTracingInterceptor 创建gRPC流式客户端追踪拦截器（兼容旧API）
// 注意：新版本otelgrpc使用StatsHandler，此函数返回空拦截器，实际应使用StatsHandler
func StreamClientTracingInterceptor() grpc.StreamClientInterceptor {
	// 在新版本中，追踪通过StatsHandler处理，这里返回一个空拦截器
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// parseMethod 解析gRPC方法名，返回服务名和方法名
func parseMethod(fullMethod string) (service, method string) {
	// fullMethod格式: /package.Service/Method
	parts := splitMethod(fullMethod)
	if len(parts) >= 2 {
		service = parts[0]
		method = parts[1]
	} else if len(parts) == 1 {
		service = "unknown"
		method = parts[0]
	} else {
		service = "unknown"
		method = "unknown"
	}
	return
}

// splitMethod 分割方法名
func splitMethod(fullMethod string) []string {
	// 移除开头的/
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}

	// 按/分割
	parts := make([]string, 0, 2)
	lastIndex := 0
	for i, char := range fullMethod {
		if char == '/' {
			if i > lastIndex {
				parts = append(parts, fullMethod[lastIndex:i])
			}
			lastIndex = i + 1
		}
	}
	if lastIndex < len(fullMethod) {
		parts = append(parts, fullMethod[lastIndex:])
	}

	return parts
}

// ServerTracingAndMetricsInterceptor 组合追踪和metrics的服务器拦截器
func ServerTracingAndMetricsInterceptor() grpc.UnaryServerInterceptor {
	tracingInterceptor := UnaryServerTracingInterceptor()
	metricsInterceptor := UnaryServerMetricsInterceptor()

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 先应用metrics拦截器
		return metricsInterceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
			// 再应用追踪拦截器
			return tracingInterceptor(ctx, req, info, handler)
		})
	}
}

// ClientTracingAndMetricsInterceptor 组合追踪和metrics的客户端拦截器
func ClientTracingAndMetricsInterceptor() grpc.UnaryClientInterceptor {
	tracingInterceptor := UnaryClientTracingInterceptor()
	metricsInterceptor := UnaryClientMetricsInterceptor()

	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// 先应用追踪拦截器（创建span），再应用metrics拦截器（记录metrics）
		return metricsInterceptor(ctx, method, req, reply, cc, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return tracingInterceptor(ctx, method, req, reply, cc, invoker, opts...)
		}, opts...)
	}
}
