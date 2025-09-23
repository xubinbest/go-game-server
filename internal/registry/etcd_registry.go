package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.xubinbest.com/go-game-server/internal/utils"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

type etcdRegistry struct {
	client      *clientv3.Client
	leaseID     clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
	logger      *zap.Logger
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewEtcdRegistry(cfg *ETCDConfig) (Registry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &etcdRegistry{
		client: client,
		logger: zap.NewExample(),
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (r *etcdRegistry) Register(ctx context.Context, instance *ServiceInstance) error {
	utils.Info("Registering service instance", zap.String("name", instance.Name), zap.String("id", instance.ID))
	key := path.Join("/services", instance.Name, instance.ID)
	value, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	// Grant lease
	resp, err := r.client.Grant(ctx, 10)
	if err != nil {
		return err
	}

	// Put with lease
	_, err = r.client.Put(ctx, key, string(value), clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// Keep alive
	keepAlive, err := r.client.KeepAlive(ctx, resp.ID)
	if err != nil {
		return err
	}

	r.leaseID = resp.ID
	r.keepAliveCh = keepAlive

	go r.keepAlive()

	return nil
}

func (r *etcdRegistry) keepAlive() {
	retryCount := 0
	maxRetries := 3
	retryInterval := 5 * time.Second

	for {
		select {
		case _, ok := <-r.keepAliveCh:
			if !ok {
				r.logger.Warn("keep alive channel closed, attempting to reconnect")

				if retryCount >= maxRetries {
					r.logger.Error("max retries reached for keep alive")
					return
				}

				// Try to re-establish the keep alive
				time.Sleep(retryInterval)
				keepAlive, err := r.client.KeepAlive(r.ctx, r.leaseID)
				if err != nil {
					r.logger.Error("failed to re-establish keep alive", zap.Error(err))
					retryCount++
					continue
				}

				r.keepAliveCh = keepAlive
				retryCount = 0
				r.logger.Info("successfully re-established keep alive")
				continue
			}
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *etcdRegistry) Deregister(ctx context.Context, instance *ServiceInstance) error {
	key := path.Join("/services", instance.Name, instance.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *etcdRegistry) Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error) {
	utils.Info("Discovering service instances", zap.String("serviceName", serviceName))
	key := path.Join("/services", serviceName)
	resp, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	instances := make([]*ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var instance ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		instances = append(instances, &instance)
	}

	if len(instances) == 0 {
		return nil, ErrServiceNotFound
	}

	return instances, nil
}

func (r *etcdRegistry) Watch(ctx context.Context, serviceName string) (chan []*ServiceInstance, error) {
	key := path.Join("/services", serviceName)
	watchCh := make(chan []*ServiceInstance, 10)

	go func() {
		watcher := clientv3.NewWatcher(r.client)
		defer watcher.Close()

		watchResp := watcher.Watch(ctx, key, clientv3.WithPrefix())

		for {
			select {
			case resp := <-watchResp:
				instances := make([]*ServiceInstance, 0)
				for _, event := range resp.Events {
					if event.Type == clientv3.EventTypePut {
						var instance ServiceInstance
						if err := json.Unmarshal(event.Kv.Value, &instance); err == nil {
							instances = append(instances, &instance)
						}
					}
				}
				if len(instances) > 0 {
					watchCh <- instances
				}
			case <-ctx.Done():
				close(watchCh)
				return
			}
		}
	}()

	return watchCh, nil
}

func (r *etcdRegistry) Close() error {
	r.cancel()
	return r.client.Close()
}

func (r *etcdRegistry) LoadConfig(ctx context.Context, dataId, group string) (string, error) {
	key := path.Join("/config", group, dataId)
	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("config not found: %s/%s", group, dataId)
	}

	return string(resp.Kvs[0].Value), nil
}

func (r *etcdRegistry) WatchConfig(dataId, group string, callback ConfigChangeCallback) {
}

// RoundRobinSelector implements load balancing
type RoundRobinSelector struct {
	services []*ServiceInstance
	index    int
	mu       sync.Mutex
}

func NewRoundRobinSelector(services []*ServiceInstance) *RoundRobinSelector {
	return &RoundRobinSelector{
		services: services,
		index:    0,
	}
}

func (s *RoundRobinSelector) Next() *ServiceInstance {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.services) == 0 {
		return nil
	}

	service := s.services[s.index]
	s.index = (s.index + 1) % len(s.services)
	return service
}
