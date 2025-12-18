package wsserver

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/circuitbreaker"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/gateway/grpcpool"
	"github.xubinbest.com/go-game-server/internal/gateway/messagerouter"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/registry"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type WSServer struct {
	reg         registry.Registry
	logger      *zap.Logger
	cfg         *config.Config
	upgrader    websocket.Upgrader
	clients     sync.Map // map[*websocket.Conn]struct{}
	Addr        string
	grpcPools   sync.Map // map[string]*grpcpool.GRPCPool
	cacheClient cache.Cache
	cbManager   *circuitbreaker.Manager
}

func New(port int, reg registry.Registry, logger *zap.Logger, cfg *config.Config, cacheClient cache.Cache) *WSServer {
	ws := &WSServer{
		reg:         reg,
		logger:      logger,
		cfg:         cfg,
		Addr:        fmt.Sprintf(":%d", port),
		cacheClient: cacheClient,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     createOriginChecker(cfg),
		},
	}

	// 初始化熔断器管理器
	if cfg.CircuitBreaker.Enabled {
		ws.cbManager = circuitbreaker.NewManager(cfg.CircuitBreaker, logger)
	}

	return ws
}

// createOriginChecker 创建Origin检查函数
func createOriginChecker(cfg *config.Config) func(*http.Request) bool {
	// 如果未启用Origin检查，允许所有来源（仅用于开发环境）
	if !cfg.WebSocket.CheckOrigin {
		return func(r *http.Request) bool {
			return true
		}
	}

	// 如果没有配置允许的Origin列表，默认允许所有（不推荐生产环境）
	if len(cfg.WebSocket.AllowedOrigins) == 0 {
		return func(r *http.Request) bool {
			return true
		}
	}

	// 创建允许的Origin集合（用于快速查找）
	allowedOrigins := make(map[string]bool, len(cfg.WebSocket.AllowedOrigins))
	for _, origin := range cfg.WebSocket.AllowedOrigins {
		allowedOrigins[origin] = true
	}

	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// 如果没有Origin头，可能是同源请求（浏览器直接访问）
			// 可以根据需要决定是否允许
			return true
		}

		// 检查Origin是否在允许列表中
		if allowedOrigins[origin] {
			return true
		}

		// 支持通配符匹配（例如：*.example.com）
		for allowedOrigin := range allowedOrigins {
			if matchesWildcard(origin, allowedOrigin) {
				return true
			}
		}

		return false
	}
}

// matchesWildcard 检查origin是否匹配通配符模式
// 例如：*.example.com 匹配 https://sub.example.com
func matchesWildcard(origin, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return false
	}

	// 简单的通配符匹配实现
	// 将 *.example.com 转换为正则表达式
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")

	matched, err := regexp.MatchString("^"+pattern+"$", origin)
	return err == nil && matched
}

func (s *WSServer) selectInstance(instances []*registry.ServiceInstance) string {
	if len(instances) == 0 {
		return ""
	}

	// 使用轮询策略
	staticIndex := time.Now().UnixNano() % int64(len(instances))
	ins := instances[staticIndex]
	return ins.Ip + ":" + fmt.Sprint(ins.Port)
}

func (s *WSServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	// 分布式限流检查（按IP限流连接数）
	if s.cfg.DistributedRateLimit.Enabled && s.cacheClient != nil {
		allowed, _, err := s.checkConnectionRateLimit(r.Context(), r)
		if err != nil {
			s.logger.Warn("Rate limit check failed, allowing connection", zap.Error(err))
		} else if !allowed {
			s.logger.Warn("WebSocket connection rate limit exceeded",
				zap.String("remote", r.RemoteAddr))
			http.Error(w, "Connection rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	s.clients.Store(conn, struct{}{})
	defer s.clients.Delete(conn)

	// Message processing loop
	msgChan := make(chan []byte, 100)
	errChan := make(chan error, 1)
	done := make(chan struct{})

	// Reader goroutine
	go func() {
		defer close(errChan)
		for {
			// 读取消息类型
			_, msg, err := conn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			// 检查消息长度
			if len(msg) < 4 {
				s.logger.Error("Message too short", zap.Int("length", len(msg)))
				continue
			}

			// 解析消息长度
			msgLength := binary.BigEndian.Uint32(msg[:4])
			if uint32(len(msg)-4) != msgLength {
				s.logger.Error("Message length mismatch",
					zap.Uint32("expected", msgLength),
					zap.Int("actual", len(msg)-4))
				continue
			}

			msgChan <- msg
		}
	}()

	// Writer goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg := <-msgChan:
				s.logger.Debug("WebSocket message received",
					zap.ByteString("msg", msg),
					zap.String("remote", conn.RemoteAddr().String()))

				// 解析消息获取服务名
				var wsMsg pb.WSMessage
				if err := proto.Unmarshal(msg[4:], &wsMsg); err != nil {
					s.logger.Error("Failed to unmarshal message", zap.Error(err))
					continue
				}

				// 服务发现
				instances, err := s.reg.Discover(r.Context(), wsMsg.Service)
				if err != nil {
					s.logger.Error("Service discovery failed",
						zap.String("service", wsMsg.Service),
						zap.Error(err))
					continue
				}

				if len(instances) == 0 {
					s.logger.Error("No available instances",
						zap.String("service", wsMsg.Service))
					continue
				}

				// 选择服务实例
				target := s.selectInstance(instances)

				// 获取或创建gRPC连接池
				pool, _ := s.grpcPools.LoadOrStore(wsMsg.Service, grpcpool.New(100, 30*time.Minute))
				grpcPool := pool.(*grpcpool.GRPCPool)

				// 获取gRPC连接
				pc, err := grpcPool.GetConn(target, s.cfg)
				if err != nil {
					s.logger.Error("Failed to get gRPC connection",
						zap.String("service", wsMsg.Service),
						zap.Error(err))
					continue
				}
				defer pc.Close()

				// 处理消息（使用熔断器保护）
				var resp []byte
				if s.cbManager != nil {
					resp, err = messagerouter.HandleWSMessage(r.Context(), msg, pc.ClientConn, s.cbManager)
				} else {
					resp, err = messagerouter.HandleWSMessage(r.Context(), msg, pc.ClientConn, nil)
				}
				if err != nil {
					s.logger.Error("Failed to handle message", zap.Error(err))
					// 发送错误响应给客户端
					errorResp := s.buildErrorMessage(wsMsg.Service, wsMsg.Method, err.Error())
					if writeErr := conn.WriteMessage(websocket.BinaryMessage, errorResp); writeErr != nil {
						s.logger.Error("Failed to send error message", zap.Error(writeErr))
					}
					continue
				}

				// 发送响应
				if err := conn.WriteMessage(websocket.BinaryMessage, resp); err != nil {
					s.logger.Error("Failed to send WebSocket message", zap.Error(err))
					close(done)
					return
				}

			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage,
					[]byte("heartbeat"), time.Now().Add(time.Second*5)); err != nil {
					s.logger.Error("Failed to send ping", zap.Error(err))
					close(done)
					return
				}

			case err := <-errChan:
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseNormalClosure) {
					s.logger.Error("WebSocket error", zap.Error(err))
				}
				close(done)
				return

			case <-done:
				return
			}
		}
	}()

	<-done
}

func (s *WSServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.HandleWS)

	srv := &http.Server{
		Addr:         s.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return srv.ListenAndServe()
}

func (s *WSServer) Shutdown(ctx context.Context) error {
	var err error
	s.clients.Range(func(conn, _ interface{}) bool {
		if wsConn, ok := conn.(*websocket.Conn); ok {
			wsConn.WriteControl(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
				time.Now().Add(time.Second))
			wsConn.Close()
		}
		return true
	})
	return err
}

func (s *WSServer) Broadcast(msg []byte) {
	s.clients.Range(func(conn, _ interface{}) bool {
		if wsConn, ok := conn.(*websocket.Conn); ok {
			wsConn.WriteMessage(websocket.BinaryMessage, msg)
		}
		return true
	})
}

// checkConnectionRateLimit 检查连接限流（按IP限制连接数）
func (s *WSServer) checkConnectionRateLimit(ctx context.Context, r *http.Request) (bool, int, error) {
	// 获取客户端IP
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	key := fmt.Sprintf("ratelimit:ws:conn:%s", ip)

	// 使用滑动窗口算法检查限流
	now := time.Now()
	windowStart := now.Add(-s.cfg.DistributedRateLimit.Window)

	zsetKey := fmt.Sprintf("ratelimit:zset:%s", key)
	minScore := float64(windowStart.UnixMilli())

	count, err := s.cacheClient.ZCount(ctx, zsetKey, minScore, float64(now.UnixMilli())).Result()
	if err != nil {
		s.logger.Warn("Failed to count rate limit", zap.Error(err))
		return true, s.cfg.DistributedRateLimit.RequestsPerSecond, nil
	}

	if int(count) >= s.cfg.DistributedRateLimit.RequestsPerSecond {
		return false, 0, nil
	}

	member := fmt.Sprintf("%d:%d", now.UnixMilli(), now.Nanosecond())
	score := float64(now.UnixMilli())
	if err := s.cacheClient.ZAdd(ctx, zsetKey, redis.Z{
		Score:  score,
		Member: member,
	}).Err(); err != nil {
		s.logger.Warn("Failed to add rate limit record", zap.Error(err))
		return true, s.cfg.DistributedRateLimit.RequestsPerSecond, nil
	}

	if err := s.cacheClient.Expire(ctx, zsetKey, s.cfg.DistributedRateLimit.Window+time.Second); err != nil {
		s.logger.Warn("Failed to set expire for rate limit key", zap.Error(err))
	}

	remaining := s.cfg.DistributedRateLimit.RequestsPerSecond - int(count) - 1
	return true, remaining, nil
}

// buildErrorMessage 构建错误消息
func (s *WSServer) buildErrorMessage(service, method, errorMsg string) []byte {
	errorResp := &pb.WSMessage{
		Service: service,
		Method:  method,
		Payload: []byte(fmt.Sprintf(`{"error":"%s"}`, errorMsg)),
	}

	respMsgBytes, err := proto.Marshal(errorResp)
	if err != nil {
		s.logger.Error("Failed to marshal error message", zap.Error(err))
		return nil
	}

	msgLen := uint32(len(respMsgBytes))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, msgLen)

	return append(header, respMsgBytes...)
}
