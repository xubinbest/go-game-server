package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/pb"
)

// ChatDatabase 接口方法实现

func (c *DatabaseClient) SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error {
	return c.chatDB.SaveChatMessage(ctx, message)
}

func (c *DatabaseClient) GetChatMessages(ctx context.Context, channel int32, target_id int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error) {
	return c.chatDB.GetChatMessages(ctx, channel, target_id, page, pageSize)
}
