package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/game_service"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/snowflake"
	"github.xubinbest.com/go-game-server/internal/telemetry"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Initialize logger first
	logger, err := utils.NewLogger()
	if err != nil {
		// Fallback to zap default logger if initialization fails
		logger, _ = zap.NewProduction()
	}
	utils.SetLogger(logger)
	defer utils.Sync()

	reg, err := registry.NewRegistry()
	if err != nil {
		utils.Fatal("Failed to create registry", zap.Error(err))
	}
	defer reg.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cfg *config.Config
	if configLocation := config.GetConfigLocation(); configLocation == "local" {
		cfg = config.LoadConfig("config.yaml")
	} else {
		content, err := reg.LoadConfig(ctx, "game.yaml", "DEFAULT_GROUP")
		if err != nil {
			utils.Fatal("failed to load config from registry", zap.Error(err))
		}
		cfg, err = config.ParseConfig([]byte(content))
		if err != nil {
			utils.Fatal("failed to parse config", zap.Error(err))
		}
	}

	cacheClient, err := cache.NewRedisCache(cfg)
	if err != nil {
		utils.Fatal("Failed to initialize cache", zap.Error(err))
	}
	defer cacheClient.Close()

	sf, err := snowflake.NewSnowflakeWithRedis(cacheClient)
	if err != nil {
		utils.Fatal("Failed to initialize snowflake", zap.Error(err))
	}

	dbClient, err := db.NewDatabaseClient(sf, cfg)
	if err != nil {
		utils.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer dbClient.Close()

	// 初始化OpenTelemetry
	var metricsServer *http.Server
	if cfg.Telemetry.Enabled {
		if err := telemetry.InitTracer(&cfg.Telemetry, logger); err != nil {
			logger.Error("Failed to initialize OpenTelemetry tracer", zap.Error(err))
		}
		metricsPort := 9090
		if metricsPortStr := os.Getenv("METRICS_PORT"); metricsPortStr != "" {
			if p, err := strconv.Atoi(metricsPortStr); err == nil {
				metricsPort = p
			}
		}
		metricsServer = telemetry.StartMetricsServer(metricsPort, logger)
	}

	// Create gRPC server with interceptors
	var opts []grpc.ServerOption
	if cfg.Telemetry.Enabled {
		opts = append(opts,
			grpc.UnaryInterceptor(telemetry.ServerTracingAndMetricsInterceptor()),
			grpc.StreamInterceptor(telemetry.StreamServerTracingInterceptor()),
		)
	}
	grpcServer := grpc.NewServer(opts...)

	// kubernetes health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	gameService := game_service.NewGameGRPCService(ctx, cacheClient, cfg)
	if err != nil {
		utils.Fatal("Failed to create game server", zap.Error(err))
	}
	pb.RegisterGameServiceServer(grpcServer, gameService)

	// Get pod IP from environment variable (set by Kubernetes Downward API)
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		podIP = "localhost" // fallback for local development
	}

	// Get port from environment variable with fallback to config
	portStr := os.Getenv("GRPC_PORT")
	if portStr == "" {
		utils.Fatal("GRPC_PORT is not set")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		utils.Fatal("GRPC_PORT is not a valid integer", zap.Error(err))
	}

	utils.Info("Starting game service on port", zap.Int("port", port))
	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		utils.Fatal("failed to listen", zap.Error(err))
	}

	// Initialize service instance
	instance := createServiceInstance(podIP, port)

	// Register service
	if err := reg.Register(ctx, instance); err != nil {
		utils.Fatal("Failed to register service", zap.Error(err))
	}

	// Setup graceful shutdown
	setupShutdownHandler(grpcServer, reg, instance, port, lis, metricsServer, cfg, logger)

	utils.Info("Server exited properly")
}

func createServiceInstance(podIP string, port int) *registry.ServiceInstance {
	// Get service name from environment variable
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "game-service"
	}

	return &registry.ServiceInstance{
		ID:          os.Getenv("POD_NAME"), // Use pod name as instance ID
		Name:        serviceName,
		Version:     "1.0.0",
		Metadata:    map[string]string{"protocol": "grpc"},
		Ip:          podIP,
		ServiceHost: serviceName + ".game-server.svc.cluster.local",
		Port:        port,
	}
}

func setupShutdownHandler(
	grpcServer *grpc.Server,
	reg registry.Registry,
	instance *registry.ServiceInstance,
	port int,
	lis net.Listener,
	metricsServer *http.Server,
	cfg *config.Config,
	logger *zap.Logger,
) {
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		utils.Info("Game gRPC service started", zap.Int("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			utils.Fatal("failed to serve", zap.Error(err))
		}
	}()

	<-quit
	utils.Info("Shutting down server...")

	// 关闭metrics服务器
	if metricsServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.ShutdownMetricsServer(shutdownCtx, metricsServer, logger); err != nil {
			logger.Error("Failed to shutdown metrics server", zap.Error(err))
		}
	}

	grpcServer.GracefulStop()

	// 关闭OpenTelemetry tracer
	if cfg.Telemetry.Enabled {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown tracer", zap.Error(err))
		}
	}

	// Deregister service
	if err := reg.Deregister(context.Background(), instance); err != nil {
		utils.Error("Failed to deregister service", zap.Error(err))
	}

	// Sync logs before exit
	defer utils.Sync()
}
