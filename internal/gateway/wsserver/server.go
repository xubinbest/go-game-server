package wsserver

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/gateway/grpcpool"
	"github.xubinbest.com/go-game-server/internal/gateway/messagerouter"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/registry"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type WSServer struct {
	reg       registry.Registry
	logger    *zap.Logger
	cfg       *config.Config
	upgrader  websocket.Upgrader
	clients   sync.Map // map[*websocket.Conn]struct{}
	Addr      string
	grpcPools sync.Map // map[string]*grpcpool.GRPCPool
}

func New(port int, reg registry.Registry, logger *zap.Logger, cfg *config.Config) *WSServer {
	return &WSServer{
		reg:    reg,
		logger: logger,
		cfg:    cfg,
		Addr:   fmt.Sprintf(":%d", port),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: implement proper origin check
			},
		},
	}
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

				// 处理消息
				resp, err := messagerouter.HandleWSMessage(r.Context(), msg, pc.ClientConn)
				if err != nil {
					s.logger.Error("Failed to handle message", zap.Error(err))
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
