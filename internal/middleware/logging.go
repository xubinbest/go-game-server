package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.xubinbest.com/go-game-server/internal/telemetry"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func NewLoggingMiddleware(logger *zap.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 提取trace context
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			// 创建span
			tracer := otel.Tracer("gateway")
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.path", r.URL.Path),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.host", r.Host),
					attribute.String("http.remote_addr", r.RemoteAddr),
				),
			)
			defer span.End()

			// 将trace context注入到请求中
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			ww := &responseWriterWrapper{w: w}

			// Process request
			next.ServeHTTP(ww, r)

			// 记录span属性
			duration := time.Since(start)
			span.SetAttributes(
				attribute.Int("http.status_code", ww.status),
				attribute.Int64("http.duration_ms", duration.Milliseconds()),
			)

			// 设置span状态
			if ww.status >= 400 {
				span.SetStatus(codes.Error, "HTTP error")
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// 记录metrics
			statusStr := strconv.Itoa(ww.status)
			telemetry.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusStr).Inc()
			telemetry.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path, statusStr).Observe(duration.Seconds())

			// Log request details with trace ID
			spanCtx := span.SpanContext()
			logger.Info("Request processed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.status),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("trace_id", spanCtx.TraceID().String()),
				zap.String("span_id", spanCtx.SpanID().String()),
			)
		}
	}
}

// NewTracingMiddleware 创建OpenTelemetry追踪中间件（使用otelhttp）
func NewTracingMiddleware(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

type responseWriterWrapper struct {
	w      http.ResponseWriter
	status int
}

func (w *responseWriterWrapper) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.status = statusCode
	w.w.WriteHeader(statusCode)
}
