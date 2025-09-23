package social

import (
	"context"
	"time"

	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// 发送世界消息
func (h *Handler) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.SendChatMessageResponse, error) {
	msgID, err := h.sf.NextID()
	if err != nil {
		return nil, err
	}

	msg := &pb.ChatMessage{
		Id:         msgID,
		Channel:    int32(req.Channel),
		SenderId:   req.SenderId,
		Content:    req.Content,
		ReceiverId: req.TargetId,
		Timestamp:  time.Now().Unix(),
		ExtraData:  req.ExtraData,
	}

	// 存储消息到数据库
	if err := h.dbClient.SaveChatMessage(ctx, msg); err != nil {
		utils.Error("failed to save chat message", zap.Error(err))
		return nil, err
	}

	// 广播消息给所有在线玩家
	if err := h.broadcastWorldMessage(ctx, msg); err != nil {
		utils.Error("failed to broadcast chat message", zap.Error(err))
		return nil, err
	}

	// 失效聊天消息缓存
	_ = h.invalidateChatMessagesCache(ctx, req.Channel, req.TargetId)

	return &pb.SendChatMessageResponse{
		Success:   true,
		MessageId: msgID,
	}, nil
}

// 获取世界消息历史
func (h *Handler) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	// 这里可以添加缓存逻辑，但由于聊天消息是实时性较强的数据，
	// 通常不建议缓存，直接查询数据库
	messages, _, err := h.dbClient.GetChatMessages(ctx, req.Channel, req.TargetId, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	return &pb.GetChatMessagesResponse{
		Messages: messages,
	}, nil
}

// 广播世界消息
func (h *Handler) broadcastWorldMessage(ctx context.Context, msg *pb.ChatMessage) error {
	// 实现消息广播逻辑
	// 可以通过websocket或gRPC流发送给所有在线玩家
	h.cacheClient.Publish(ctx, "chat_message_broadcast", msg)
	return nil
}

// 失效聊天消息缓存
func (h *Handler) invalidateChatMessagesCache(ctx context.Context, channel int32, targetID int64) error {
	// 这里可以根据具体的缓存策略来失效相关缓存
	// 由于聊天消息通常不缓存，这里暂时不实现
	return nil
}
