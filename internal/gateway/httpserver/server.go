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

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/gateway/grpcpool"
	"github.xubinbest.com/go-game-server/internal/gateway/messagerouter"
	"github.xubinbest.com/go-game-server/internal/middleware"
	"github.xubinbest.com/go-game-server/internal/registry"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type HTTPServer struct {
	reg    registry.Registry
	logger *zap.Logger
	cfg    *config.Config
	router *mux.Router
	server *http.Server
	Addr   string
}

func New(port int, reg registry.Registry, logger *zap.Logger, cfg *config.Config) *HTTPServer {
	s := &HTTPServer{
		reg:    reg,
		logger: logger,
		cfg:    cfg,
		router: mux.NewRouter(),
		Addr:   fmt.Sprintf(":%d", port),
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

	// Protected routes (with auth)
	protectedHandler := s.handleAPIRequest
	protectedHandler = middleware.NewLoggingMiddleware(s.logger)(protectedHandler)
	protectedHandler = middleware.NewRateLimitMiddleware(s.cfg.RateLimit)(protectedHandler)
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

	// Handle message and get response
	resp, err := messagerouter.HandleMessage(r, pc.ClientConn)
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

	return err
}
