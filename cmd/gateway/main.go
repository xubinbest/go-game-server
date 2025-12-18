package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/gateway/chatmessage"
	"github.xubinbest.com/go-game-server/internal/gateway/httpserver"
	"github.xubinbest.com/go-game-server/internal/gateway/wsserver"
	"github.xubinbest.com/go-game-server/internal/registry"
	"github.xubinbest.com/go-game-server/internal/utils"
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

	ctx := context.Background()

	// Initialize components
	reg := initRegistry(logger)
	defer reg.Close()

	cfg := loadConfig(ctx, logger, reg)

	// Get pod IP from environment variable (set by Kubernetes Downward API)
	podIP := os.Getenv("POD_IP")
	if podIP == "" {
		logger.Fatal("POD_IP is not set")
	}

	httpPortStr := os.Getenv("HTTP_PORT")
	if httpPortStr == "" {
		logger.Fatal("HTTP_PORT is not set")
	}
	httpPort, err := strconv.Atoi(httpPortStr)
	if err != nil {
		logger.Fatal("HTTP_PORT is not a valid integer", zap.Error(err))
	}

	wsPortStr := os.Getenv("WEBSOCKET_PORT")
	if wsPortStr == "" {
		logger.Fatal("WEBSOCKET_PORT is not set")
	}
	wsPort, err := strconv.Atoi(wsPortStr)
	if err != nil {
		logger.Fatal("WEBSOCKET_PORT is not a valid integer", zap.Error(err))
	}

	// Initialize service instance
	instance := createServiceInstance(podIP, httpPort, wsPort, logger)

	// Register service
	if err := reg.Register(ctx, instance); err != nil {
		logger.Fatal("Failed to register service", zap.Error(err))
	}

	// Initialize Redis cache
	cacheClient, err := cache.NewRedisCache(cfg)
	if err != nil {
		logger.Fatal("Failed to create Redis cache", zap.Error(err))
	}

	// Initialize servers
	httpServer := httpserver.New(httpPort, reg, logger, cfg, cacheClient)
	wsServer := wsserver.New(wsPort, reg, logger, cfg, cacheClient)

	// Setup graceful shutdown
	setupShutdownHandler(ctx, logger, httpServer, wsServer, reg, instance)

	// Initialize chat message broadcast
	initChatBroadcast(logger, cfg, cacheClient)

	// Start servers
	startServers(logger, httpServer, wsServer, httpPort, wsPort)
}

func initRegistry(logger *zap.Logger) registry.Registry {
	reg, err := registry.NewRegistry()
	if err != nil {
		logger.Fatal("Failed to create registry client", zap.Error(err))
	}
	return reg
}

func loadConfig(ctx context.Context, logger *zap.Logger, reg registry.Registry) *config.Config {
	content, err := reg.LoadConfig(ctx, "gateway.yaml", "DEFAULT_GROUP")
	if err != nil {
		logger.Fatal("failed to load config from registry", zap.Error(err))
	}
	cfg, err := config.ParseConfig([]byte(content))
	if err != nil {
		logger.Fatal("failed to parse config", zap.Error(err))
	}
	return cfg
}

func createServiceInstance(podIP string, httpPort int, wsPort int, logger *zap.Logger) *registry.ServiceInstance {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "gateway-service"
	}

	return &registry.ServiceInstance{
		ID:          os.Getenv("POD_NAME"),
		Name:        serviceName,
		Version:     "1.0.0",
		Ip:          podIP,
		ServiceHost: serviceName + ".game-server.svc.cluster.local",
		Port:        httpPort,
		Metadata: map[string]string{
			"protocol":  "http/websocket",
			"http_port": strconv.Itoa(httpPort),
			"ws_port":   strconv.Itoa(wsPort),
		},
	}
}

func setupShutdownHandler(
	ctx context.Context,
	logger *zap.Logger,
	httpServer *httpserver.HTTPServer,
	wsServer *wsserver.WSServer,
	reg registry.Registry,
	instance *registry.ServiceInstance,
) {
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownServers(shutdownCtx, logger, httpServer, wsServer)
		deregisterService(ctx, logger, reg, instance)
	}()
}

func shutdownServers(
	ctx context.Context,
	logger *zap.Logger,
	httpServer *httpserver.HTTPServer,
	wsServer *wsserver.WSServer,
) {
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	if err := wsServer.Shutdown(ctx); err != nil {
		logger.Error("WebSocket server shutdown error", zap.Error(err))
	}
}

func deregisterService(
	ctx context.Context,
	logger *zap.Logger,
	reg registry.Registry,
	instance *registry.ServiceInstance,
) {
	if reg != nil && instance != nil {
		if err := reg.Deregister(ctx, instance); err != nil {
			logger.Error("Failed to deregister service", zap.Error(err))
		}
	}
}

func initChatBroadcast(logger *zap.Logger, cfg *config.Config, cacheClient cache.Cache) {
	chatMessage, err := chatmessage.NewChatMessageBroadcast(cacheClient)
	if err != nil {
		logger.Fatal("Failed to create chat message broadcast", zap.Error(err))
	}

	if err := chatMessage.SubscribeChatMessageBroadcast(context.Background(), "chat_message_broadcast"); err != nil {
		logger.Fatal("Failed to subscribe to chat message broadcast", zap.Error(err))
	}
}

func startServers(
	logger *zap.Logger,
	httpServer *httpserver.HTTPServer,
	wsServer *wsserver.WSServer,
	httpPort int,
	wsPort int,
) {
	logger.Info("Starting gateway server",
		zap.String("http_port", fmt.Sprintf("%d", httpPort)),
		zap.String("ws_port", fmt.Sprintf("%d", wsPort)),
	)

	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	if err := wsServer.Start(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("WebSocket server error", zap.Error(err))
	}
}
