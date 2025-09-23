package chatmessage

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

type ChatMessageBroadcast struct {
	r cache.Cache
}

func NewChatMessageBroadcast(r cache.Cache) (*ChatMessageBroadcast, error) {
	return &ChatMessageBroadcast{r: r}, nil
}

func (c *ChatMessageBroadcast) SubscribeChatMessageBroadcast(ctx context.Context, channel string) error {
	ch, err := c.r.Subscribe(ctx, channel)
	if err != nil {
		return fmt.Errorf("subscribe chat message broadcast fail. err: %v", err)
	}
	go handlerChatMessageBroadcast(ch)
	return nil
}

func handlerChatMessageBroadcast(ch <-chan interface{}) {
	select {
	case msg := <-ch:
		// 处理消息
		utils.Info("收到广播消息", zap.Any("message", msg))
	default:
		utils.Info("没有收到消息")
	}

}
