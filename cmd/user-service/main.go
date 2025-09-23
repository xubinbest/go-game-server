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
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/snowflake"
	"github.xubinbest.com/go-game-server/internal/user"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// Initialize service registry
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
		content, err := reg.LoadConfig(ctx, "user.yaml", "DEFAULT_GROUP")
		if err != nil {
			utils.Fatal("failed to load config from registry", zap.Error(err))
		}
		cfg, err = config.ParseConfig([]byte(content))
		if err != nil {
			utils.Fatal("failed to parse config", zap.Error(err))
		}
	}

	// 初始化 redis缓存
	cacheClient, err := cache.NewRedisCache(cfg)
	if err != nil {
		utils.Fatal("Failed to initialize cache", zap.Error(err))
	}
	defer cacheClient.Close()

	// 初始化雪花算发的ID生成器
	sf, err := snowflake.NewSnowflakeWithRedis(cacheClient)
	if err != nil {
		utils.Fatal("Failed to initialize snowflake", zap.Error(err))
	}

	// 初始化数据库
	dbClient, err := db.NewDatabaseClient(sf, cfg)
	if err != nil {
		utils.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer dbClient.Close()

	// 初始化设计配置管理器
	designConfigManager := designconfig.NewDesignConfigManager(reg, user.Tables)
	if err := designConfigManager.Start(); err != nil {
		utils.Fatal("Failed to initialize design config manager", zap.Error(err))
	}

	// Create gRPC server with increased message size limit
	grpcServer := grpc.NewServer()

	// kubernetes health check
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	userGRPCServer := user.NewUserGRPCServer(dbClient, cacheClient, sf, cfg, designConfigManager)
	pb.RegisterUserServiceServer(grpcServer, userGRPCServer)

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

	utils.Info("Starting user service", zap.Int("port", port))
	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		utils.Fatal("failed to listen", zap.Error(err))
	}

	// Initialize service instance
	instance := createServiceInstance(cfg, podIP, port)

	// Register service
	if err := reg.Register(ctx, instance); err != nil {
		utils.Fatal("Failed to register service", zap.Error(err))
	}

	// Setup graceful shutdown
	setupShutdownHandler(grpcServer, reg, instance, port, lis)

	utils.Info("Server exited properly")
}

func createServiceInstance(cfg *config.Config, podIP string, port int) *registry.ServiceInstance {
	// Get service name from environment variable
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "user-service"
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
) {
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		utils.Info("User gRPC service started", zap.Int("port", port))
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
