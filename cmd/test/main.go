package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/snowflake"
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

	// Load configuration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cfg *config.Config
	if configLocation := config.GetConfigLocation(); configLocation == "local" {
		cfg = config.LoadConfig("config.yaml")
	} else {
		content, err := reg.LoadConfig(ctx, "test.yaml", "DEFAULT_GROUP")
		if err != nil {
			utils.Fatal("failed to load config from registry", zap.Error(err))
		}
		cfg, err = config.ParseConfig([]byte(content))
		if err != nil {
			utils.Fatal("failed to parse config", zap.Error(err))
		}
	}

	utils.Info("cfg", zap.Any("config", cfg))

	// Initialize cache
	cacheClient, err := cache.NewRedisCache(cfg)
	if err != nil {
		utils.Fatal("Failed to initialize cache", zap.Error(err))
	}
	utils.Info("Cache client initialized successfully")
	defer cacheClient.Close()

	sf, err := snowflake.NewSnowflakeWithRedis(cacheClient)

	if err != nil {
		utils.Fatal("Failed to initialize snowflake", zap.Error(err))
	}

	// Initialize database
	dbClient, err := db.NewDatabaseClient(sf, cfg)
	if err != nil {
		utils.Fatal("Failed to initialize database", zap.Error(err))
	}
	utils.Info("Database client initialized successfully")
	defer dbClient.Close()

	// Create gRPC server with increased message size limit
	grpcServer := grpc.NewServer()

	// kubernetes health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

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

	utils.Info("Starting test service on port", zap.Int("port", port))
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

	setupShutdownHandler(grpcServer, reg, instance, port, lis)

	utils.Info("Server exited properly")
}

func createServiceInstance(podIP string, port int) *registry.ServiceInstance {
	// Get service name from environment variable
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "test-service"
	}

	podName := os.Getenv("POD_NAME")
	if podName == "" {
		podName = "default-pod-name" // Fallback for local development
	}

	version := os.Getenv("POD_VERSION")
	if version == "" {
		version = "1.0.0" // Default version if not set
	}

	metadata := map[string]string{
		"pod_name": podName,
		"protocol": "grpc",
	}

	return &registry.ServiceInstance{
		ID:          podName,
		Name:        serviceName,
		Version:     version,
		Metadata:    metadata,
		Ip:          podIP,
		Port:        port,
		ClusterName: "test-cluster",
	}
}

func setupShutdownHandler(
	grpcServer *grpc.Server,
	reg registry.Registry,
	instance *registry.ServiceInstance,
	port int,
	lis net.Listener,
) {
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		utils.Info("test gRPC service started", zap.Int("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			utils.Fatal("failed to serve", zap.Error(err))
		}
	}()

	<-quit
	utils.Info("Shutting down server...")

	grpcServer.GracefulStop()

	// Deregister service
	if err := reg.Deregister(context.Background(), instance); err != nil {
		utils.Error("Failed to deregister service", zap.Error(err))
	}

	// Sync logs before exit
	defer utils.Sync()
}
