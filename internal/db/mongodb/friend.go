package mongodb

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBFriendDatabase 实现 FriendDatabase 接口
type MongoDBFriendDatabase struct {
	client   *mongo.Client
	database string
	sf       *snowflake.Snowflake
}

// NewMongoDBFriendDatabase 创建 MongoDBFriendDatabase 实例
func NewMongoDBFriendDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.FriendDatabase {
	return &MongoDBFriendDatabase{
		client:   client,
		database: database,
		sf:       sf,
	}
}

// 获取好友集合
func (m *MongoDBFriendDatabase) friendCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("friends")
}

// 获取好友请求集合
func (m *MongoDBFriendDatabase) requestCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("friend_requests")
}

// 获取用户集合
func (m *MongoDBFriendDatabase) userCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("users")
}

// GetFriends 获取用户的好友列表
func (m *MongoDBFriendDatabase) GetFriends(ctx context.Context, userID int64) ([]*models.Friend, error) {
	filter := bson.M{"user_id": userID}
	cursor, err := m.friendCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var friends []*models.Friend
	for cursor.Next(ctx) {
		var friend models.Friend
		if err := cursor.Decode(&friend); err != nil {
			return nil, err
		}
		friends = append(friends, &friend)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return friends, nil
}

// GetFriend 获取用户的好友信息
func (m *MongoDBFriendDatabase) GetFriend(ctx context.Context, userID, friendID int64) (*models.Friend, error) {
	filter := bson.M{"user_id": userID, "friend_id": friendID}
	var friend models.Friend
	err := m.friendCollection().FindOne(ctx, filter).Decode(&friend)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 好友不存在
		}
		return nil, err
	}
	return &friend, nil
}

// CreateFriendRequest 创建好友请求
func (m *MongoDBFriendDatabase) CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) error {
	id, err := m.sf.NextID()
	if err != nil {
		return err
	}
	request := models.FriendRequest{
		ID:         id,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	}
	_, err = m.requestCollection().InsertOne(ctx, request)
	return err
}

// GetFriendRequests 获取好友请求列表
func (m *MongoDBFriendDatabase) GetFriendRequests(ctx context.Context, userID int64) ([]*models.FriendRequest, error) {
	filter := bson.M{"to_user_id": userID}
	cursor, err := m.requestCollection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*models.FriendRequest
	for cursor.Next(ctx) {
		var request models.FriendRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return requests, nil
}

// GetFriendRequest 获取单个好友请求
func (m *MongoDBFriendDatabase) GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	filter := bson.M{"_id": requestID}
	var request models.FriendRequest
	err := m.requestCollection().FindOne(ctx, filter).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // 好友请求不存在
		}
		return nil, err
	}
	return &request, nil
}

// AddFriend 添加好友
func (m *MongoDBFriendDatabase) AddFriend(ctx context.Context, userID, friendID int64) error {
	id, err := m.sf.NextID()
	if err != nil {
		return err
	}
	friend := models.Friend{
		ID:       id,
		UserID:   userID,
		FriendID: friendID,
	}
	_, err = m.friendCollection().InsertOne(ctx, friend)
	return err
}

// RemoveFriend 移除好友
func (m *MongoDBFriendDatabase) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	filter := bson.M{"user_id": userID, "friend_id": friendID}
	_, err := m.friendCollection().DeleteOne(ctx, filter)
	return err
}

// DeleteFriendRequest 删除好友请求
func (m *MongoDBFriendDatabase) DeleteFriendRequest(ctx context.Context, requestID int64) error {
	filter := bson.M{"_id": requestID}
	_, err := m.requestCollection().DeleteOne(ctx, filter)
	return err
}
