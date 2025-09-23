package mongodb

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBChatDatabase 实现 ChatDatabase 接口
type MongoDBChatDatabase struct {
	client   *mongo.Client
	database string
	sf       *snowflake.Snowflake
}

// NewMongoDBChatDatabase 创建 MongoDBChatDatabase 实例
func NewMongoDBChatDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.ChatDatabase {
	return &MongoDBChatDatabase{
		client:   client,
		database: database,
		sf:       sf,
	}
}

// 获取聊天消息集合
func (m *MongoDBChatDatabase) collection() *mongo.Collection {
	return m.client.Database(m.database).Collection("chat_messages")
}

// SaveChatMessage 保存聊天消息
func (m *MongoDBChatDatabase) SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error {
	if message.Id == 0 {
		var err error
		message.Id, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	document := bson.M{
		"_id":         message.Id,
		"sender_id":   message.SenderId,
		"sender_name": message.SenderName,
		"channel":     message.Channel,
		"content":     message.Content,
		"send_time":   message.Timestamp,
		"extra_data":  message.ExtraData,
	}

	_, err := m.collection().InsertOne(ctx, document)
	return err
}

// GetChatMessages 获取聊天消息
func (m *MongoDBChatDatabase) GetChatMessages(ctx context.Context, channel int32, target_id int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error) {
	filter := bson.M{
		"channel":     channel,
		"receiver_id": target_id,
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "send_time", Value: -1}})
	opts.SetSkip(int64(page * pageSize))
	opts.SetLimit(int64(pageSize))

	cursor, err := m.collection().Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var messages []*pb.ChatMessage
	for cursor.Next(ctx) {
		var msg pb.ChatMessage
		if err := cursor.Decode(&msg); err != nil {
			return nil, 0, err
		}
		messages = append(messages, &msg)
	}

	count, err := m.collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return messages, int32(count), nil
}
