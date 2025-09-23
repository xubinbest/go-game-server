package gorm

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/gorm"
)

// GormChatDatabase GORM聊天数据库实现
type GormChatDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormChatDatabase 创建GORM聊天数据库实例
func NewGormChatDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.ChatDatabase {
	return &GormChatDatabase{
		db: db,
		sf: sf,
	}
}

// SaveChatMessage 保存聊天消息
func (g *GormChatDatabase) SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error {
	// 生成ID
	id, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate message ID: %w", err)
	}

	chatMessage := &ChatMessage{
		ID:         id,
		Channel:    message.Channel,
		SenderID:   message.SenderId,
		SenderName: message.SenderName,
		ReceiverID: message.ReceiverId,
		Content:    message.Content,
		SendTime:   time.Now().Unix(),
		ExtraData:  "", // 可以根据需要序列化额外数据
	}

	if err := g.db.WithContext(ctx).Create(chatMessage).Error; err != nil {
		return fmt.Errorf("failed to save chat message: %w", err)
	}

	return nil
}

// GetChatMessages 获取聊天消息
func (g *GormChatDatabase) GetChatMessages(ctx context.Context, channel int32, targetID int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error) {
	var messages []*ChatMessage
	var total int64

	// 构建查询条件
	query := g.db.WithContext(ctx).Model(&ChatMessage{}).
		Where("channel = ?", channel)

	// 根据频道类型添加额外条件
	switch channel {
	case 1: // 世界频道
		// 世界频道不需要额外条件
	case 2: // 公会频道
		query = query.Where("receiver_id = ?", targetID)
	case 3: // 私聊频道
		query = query.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			targetID, targetID, targetID, targetID)
	}

	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count chat messages: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Order("send_time DESC").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&messages).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chat messages: %w", err)
	}

	// 转换为protobuf消息
	pbMessages := make([]*pb.ChatMessage, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.ChatMessage{
			Id:         msg.ID,
			Channel:    msg.Channel,
			SenderId:   msg.SenderID,
			SenderName: msg.SenderName,
			ReceiverId: msg.ReceiverID,
			Content:    msg.Content,
			Timestamp:  msg.SendTime,
		}
	}

	return pbMessages, int32(total), nil
}

// ChatMessage 聊天消息模型
type ChatMessage struct {
	ID         int64  `json:"id" gorm:"primaryKey;autoIncrement:false"`
	Channel    int32  `json:"channel" gorm:"type:int;not null;index"`
	SenderID   int64  `json:"sender_id" gorm:"type:bigint;not null;index"`
	SenderName string `json:"sender_name" gorm:"type:varchar(50);not null"`
	ReceiverID int64  `json:"receiver_id" gorm:"type:bigint;index"`
	Content    string `json:"content" gorm:"type:text;not null"`
	SendTime   int64  `json:"send_time" gorm:"type:bigint;not null;index"`
	ExtraData  string `json:"extra_data" gorm:"type:text"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
