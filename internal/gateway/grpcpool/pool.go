package grpcpool

import (
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type pooledConn struct {
	*grpc.ClientConn
	pool     *GRPCPool
	addr     string
	lastUsed time.Time
}

func (c *pooledConn) Close() error {
	c.pool.mu.Lock()
	defer c.pool.mu.Unlock()

	// Check if connection is still valid
	if c.ClientConn.GetState() == connectivity.Shutdown {
		return nil
	}

	c.lastUsed = time.Now()
	c.pool.clients[c.addr] = c
	return nil
}

type GRPCPool struct {
	clients  map[string]*pooledConn
	mu       sync.RWMutex
	maxIdle  time.Duration
	maxConns int
}

func New(maxConns int, maxIdle time.Duration) *GRPCPool {
	pool := &GRPCPool{
		clients:  make(map[string]*pooledConn),
		maxIdle:  maxIdle,
		maxConns: maxConns,
	}

	// Start background cleaner
	go pool.cleanStaleConnections()
	return pool
}

func (p *GRPCPool) cleanStaleConnections() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.Lock()
		for addr, conn := range p.clients {
			if time.Since(conn.lastUsed) > p.maxIdle {
				conn.ClientConn.Close()
				delete(p.clients, addr)
			}
		}
		p.mu.Unlock()
	}
}

func (p *GRPCPool) GetConn(addr string, cfg *config.Config) (*pooledConn, error) {
	p.mu.RLock()
	if pc, ok := p.clients[addr]; ok {
		if pc.ClientConn.GetState() == connectivity.Ready {
			pc.lastUsed = time.Now()
			p.mu.RUnlock()
			return pc, nil
		}
		// Remove bad connection
		pc.ClientConn.Close()
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check again after acquiring write lock
	if pc, ok := p.clients[addr]; ok && pc.ClientConn.GetState() == connectivity.Ready {
		pc.lastUsed = time.Now()
		return pc, nil
	}

	// Check max connections
	if len(p.clients) >= p.maxConns {
		// Evict oldest connection
		var oldestAddr string
		var oldestTime time.Time
		for addr, conn := range p.clients {
			if oldestTime.IsZero() || conn.lastUsed.Before(oldestTime) {
				oldestAddr = addr
				oldestTime = conn.lastUsed
			}
		}
		if oldestAddr != "" {
			p.clients[oldestAddr].ClientConn.Close()
			delete(p.clients, oldestAddr)
		}
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    cfg.GRPC.KeepAlive,
			Timeout: cfg.GRPC.Timeout,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.GRPC.MaxSendMsgSize),
		),
	)
	if err != nil {
		return nil, err
	}

	pc := &pooledConn{
		ClientConn: conn,
		pool:       p,
		addr:       addr,
		lastUsed:   time.Now(),
	}
	p.clients[addr] = pc
	return pc, nil
}
