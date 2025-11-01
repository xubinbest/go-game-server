package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Subscribe 订阅频道（每次订阅创建独立的 PubSub，便于广播与独立关闭）
func (r *RedisCache) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	var ps *redis.PubSub
	switch c := r.client.(type) {
	case *redis.Client:
		ps = c.Subscribe(ctx, channel)
	case *redis.ClusterClient:
		ps = c.Subscribe(ctx, channel)
	default:
		return nil, fmt.Errorf("unsupported client type for Subscribe")
	}

	// 记录该频道的 PubSub 列表
	r.mu.Lock()
	r.pubsubs[channel] = append(r.pubsubs[channel], ps)
	r.mu.Unlock()

	ch := make(chan interface{})

	go func(ps *redis.PubSub, out chan interface{}) {
		for msg := range ps.Channel() {
			out <- msg.Payload
		}
		close(out)
	}(ps, ch)

	return ch, nil
}

// Unsubscribe 取消订阅频道（将关闭该频道下所有订阅者的 PubSub）
func (r *RedisCache) Unsubscribe(ctx context.Context, channel string) error {
	r.mu.Lock()
	pss, ok := r.pubsubs[channel]
	if ok {
		delete(r.pubsubs, channel)
	}
	r.mu.Unlock()
	if !ok || len(pss) == 0 {
		// 没有找到对应订阅，视为已退订
		return nil
	}
	var firstErr error
	for _, ps := range pss {
		if ps == nil {
			continue
		}
		if err := ps.Unsubscribe(ctx, channel); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := ps.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Publish 发布消息到频道
func (r *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
