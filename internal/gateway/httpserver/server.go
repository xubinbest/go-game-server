package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/circuitbreaker"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/gateway/grpcpool"
	"github.xubinbest.com/go-game-server/internal/gateway/messagerouter"
	"github.xubinbest.com/go-game-server/internal/middleware"
	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/telemetry"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type HTTPServer struct {
	reg         registry.Registry
	logger      *zap.Logger
	cfg         *config.Config
	router      *mux.Router
	server      *http.Server
	Addr        string
	cacheClient cache.Cache
	cbManager   *circuitbreaker.Manager
}

func New(port int, reg registry.Registry, logger *zap.Logger, cfg *config.Config, cacheClient cache.Cache) *HTTPServer {
	s := &HTTPServer{
		reg:         reg,
		logger:      logger,
		cfg:         cfg,
		router:      mux.NewRouter(),
		Addr:        fmt.Sprintf(":%d", port),
		cacheClient: cacheClient,
	}

	// 初始化OpenTelemetry
	if cfg.Telemetry.Enabled {
		if err := telemetry.InitTracer(&cfg.Telemetry, logger); err != nil {
			logger.Error("Failed to initialize OpenTelemetry tracer", zap.Error(err))
		}
	}

	// 初始化熔断器管理器
	if cfg.CircuitBreaker.Enabled {
		s.cbManager = circuitbreaker.NewManager(cfg.CircuitBreaker, logger)
	}

	s.registerRoutes()
	return s
}

var (
	rrIndex   uint64       // for round-robin
	grpcPools = sync.Map{} // map[serviceName]*grpcpool.GRPCPool
)

func (s *HTTPServer) registerRoutes() {
	// Health check
	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Metrics endpoint for Prometheus
	s.router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Protected routes (with auth)
	protectedHandler := s.handleAPIRequest
	protectedHandler = middleware.NewLoggingMiddleware(s.logger)(protectedHandler)

	// 使用分布式限流（如果启用）
	if s.cfg.DistributedRateLimit.Enabled && s.cacheClient != nil {
		protectedHandler = middleware.NewDistributedRateLimitMiddleware(
			s.cacheClient,
			s.cfg.DistributedRateLimit,
			s.logger,
		)(protectedHandler)
	}

	protectedHandler = middleware.NewAuthMiddleware(s.cfg.Auth)(protectedHandler)
	s.router.HandleFunc("/api/{service}/{path:.*}", protectedHandler)

}

func (s *HTTPServer) selectInstance(instances []*registry.ServiceInstance) string {
	if len(instances) == 0 {
		return ""
	}

	switch s.cfg.LoadBalancer.Strategy {
	case "random":
		rand.Seed(time.Now().UnixNano())
		ins := instances[rand.Intn(len(instances))]
		return ins.Ip + ":" + fmt.Sprint(ins.Port)
	case "leastconn":
		// TODO: implement connection tracking
		ins := instances[0]
		return ins.Ip + ":" + fmt.Sprint(ins.Port)
	default:
		// Round-robin by default
		staticIndex := atomic.AddUint64(&rrIndex, 1) % uint64(len(instances))
		ins := instances[staticIndex]
		return ins.Ip + ":" + fmt.Sprint(ins.Port)
	}
}

func (s *HTTPServer) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	// Discover service instances
	instances, err := s.reg.Discover(r.Context(), serviceName)
	if err != nil {
		s.logger.Error("Service discovery failed",
			zap.String("service", serviceName),
			zap.Error(err))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	if len(instances) == 0 {
		s.logger.Error("No available instances",
			zap.String("service", serviceName))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Select instance based on load balancing strategy
	target := s.selectInstance(instances)

	// Get or create gRPC pool for this service
	pool, _ := grpcPools.LoadOrStore(serviceName, grpcpool.New(100, 30*time.Minute))
	grpcPool := pool.(*grpcpool.GRPCPool)

	// Get gRPC connection
	pc, err := grpcPool.GetConn(target, s.cfg)
	if err != nil {
		s.logger.Error("Failed to get gRPC connection",
			zap.String("service", serviceName),
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer pc.Close()

	// Handle message and get response (with circuit breaker if enabled)
	var resp proto.Message
	if s.cbManager != nil {
		resp, err = messagerouter.HandleMessage(r, pc.ClientConn, s.cbManager)
	} else {
		resp, err = messagerouter.HandleMessage(r, pc.ClientConn, nil)
	}
	if err != nil {
		s.logger.Error("Failed to handle message",
			zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert protobuf response to JSON
	jsonData, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("Failed to marshal response",
			zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set response headers and write JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *HTTPServer) Start() error {
	s.server = &http.Server{
		Addr:         s.Addr,
		Handler:      s,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	var err error

	if s.server != nil {
		if shutdownErr := s.server.Shutdown(ctx); shutdownErr != nil {
			err = shutdownErr
		}
	}

	// 关闭OpenTelemetry tracer
	if s.cfg.Telemetry.Enabled {
		if shutdownErr := telemetry.Shutdown(ctx); shutdownErr != nil {
			if err == nil {
				err = shutdownErr
			}
		}
	}

	return err
}
