package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP请求metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "path"},
	)

	// WebSocket metrics
	WebSocketConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Current number of WebSocket connections",
		},
		[]string{"status"},
	)

	WebSocketMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"direction"},
	)

	// gRPC客户端metrics
	GRPCClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total number of gRPC client requests",
		},
		[]string{"service", "method", "status"},
	)

	GRPCClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_client_request_duration_seconds",
			Help:    "gRPC client request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	// gRPC服务端metrics
	GRPCServerRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_requests_total",
			Help: "Total number of gRPC server requests",
		},
		[]string{"service", "method", "status"},
	)

	GRPCServerRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_request_duration_seconds",
			Help:    "gRPC server request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)

	GRPCServerMsgSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_msg_sent_total",
			Help: "Total number of gRPC server messages sent",
		},
		[]string{"service", "method"},
	)

	GRPCServerMsgReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_msg_received_total",
			Help: "Total number of gRPC server messages received",
		},
		[]string{"service", "method"},
	)

	// 缓存metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type"},
	)

	CacheErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_errors_total",
			Help: "Total number of cache errors",
		},
		[]string{"cache_type"},
	)

	CacheHitRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate (0.0-1.0)",
		},
		[]string{"cache_type"},
	)
)
