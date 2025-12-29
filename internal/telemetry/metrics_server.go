package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// StartMetricsServer 启动Prometheus metrics HTTP服务器
func StartMetricsServer(port int, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		logger.Info("Starting metrics server", zap.Int("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	return server
}

// ShutdownMetricsServer 关闭metrics服务器
func ShutdownMetricsServer(ctx context.Context, server *http.Server, logger *zap.Logger) error {
	if server == nil {
		return nil
	}
	logger.Info("Shutting down metrics server")
	return server.Shutdown(ctx)
}
