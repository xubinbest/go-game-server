package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Subscribe 订阅频道
func (r *RedisCache) Subscribe(ctx context.Context, channel string) (<-chan interface{}, error) {
	// 复用或创建该频道的 PubSub
	r.mu.Lock()
	pubsub, ok := r.pubsubs[channel]
	if !ok {
		switch c := r.client.(type) {
		case *redis.Client:
			pubsub = c.Subscribe(ctx, channel)
		case *redis.ClusterClient:
			pubsub = c.Subscribe(ctx, channel)
		default:
			r.mu.Unlock()
			return nil, fmt.Errorf("unsupported client type for Subscribe")
		}
		r.pubsubs[channel] = pubsub
	}
	r.mu.Unlock()

	ch := make(chan interface{})

	go func(ps *redis.PubSub, out chan interface{}) {
		for msg := range ps.Channel() {
			out <- msg.Payload
		}
		close(out)
	}(pubsub, ch)

	return ch, nil
}

// Unsubscribe 取消订阅频道
func (r *RedisCache) Unsubscribe(ctx context.Context, channel string) error {
	r.mu.Lock()
	ps, ok := r.pubsubs[channel]
	if ok {
		delete(r.pubsubs, channel)
	}
	r.mu.Unlock()
	if !ok || ps == nil {
		// 没有找到对应订阅，视为已退订
		return nil
	}
	// 退订该频道并关闭 PubSub（单通道场景可直接关闭）
	if err := ps.Unsubscribe(ctx, channel); err != nil {
		return err
	}
	return ps.Close()
}

// Publish 发布消息到频道
func (r *RedisCache) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}
