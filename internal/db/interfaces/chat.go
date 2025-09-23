package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/pb"
)

// ChatDatabase 定义聊天相关的数据库操作接口
type ChatDatabase interface {
	// 保存聊天消息
	SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error

	// 获取聊天消息
	GetChatMessages(ctx context.Context, channel int32, target_id int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error)
	// 其他聊天相关方法...
}
