package guild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 获取帮派申请集合
func (m *MongoDBGuildDatabase) applicationCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_applications")
}

// CreateGuildApplication 创建帮派申请
func (m *MongoDBGuildDatabase) CreateGuildApplication(ctx context.Context, application *models.GuildApplication) error {
	if application.ID == 0 {
		var err error
		application.ID, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	if application.Time.IsZero() {
		application.Time = time.Now()
	}

	// 检查用户是否已经是帮派成员
	member, err := m.GetGuildMember(ctx, application.GuildID, application.UserID)
	if err != nil {
		return err
	}

	if member != nil {
		return errors.New("user is already a member of this guild")
	}

	// 检查是否已经有待处理的申请
	count, err := m.applicationCollection().CountDocuments(ctx, bson.M{
		"guild_id": application.GuildID,
		"user_id":  application.UserID,
		"status":   1, // 待处理状态
	})
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("application already exists")
	}

	// 创建申请
	document := bson.M{
		"_id":         application.ID,
		"guild_id":    application.GuildID,
		"user_id":     application.UserID,
		"time":        application.Time,
		"expire_time": application.ExpireTime,
	}

	_, err = m.applicationCollection().InsertOne(ctx, document)
	return err
}

// GetGuildApplication 获取帮派申请
func (m *MongoDBGuildDatabase) GetGuildApplication(ctx context.Context, appID int64) (*models.GuildApplication, error) {
	filter := bson.M{"_id": appID}

	var application models.GuildApplication
	err := m.applicationCollection().FindOne(ctx, filter).Decode(&application)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 申请不存在
		}
		return nil, err
	}

	return &application, nil
}

// DeleteGuildApplication 删除帮派申请
func (m *MongoDBGuildDatabase) DeleteGuildApplication(ctx context.Context, appID int64) error {
	filter := bson.M{
		"_id": appID,
	}

	result, err := m.applicationCollection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("guild application not found")
	}

	return nil
}

// GetGuildApplications 获取帮派申请列表
func (m *MongoDBGuildDatabase) GetGuildApplications(ctx context.Context, guildID int64) ([]*models.GuildApplication, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"guild_id": guildID,
			"status":   1, // 待处理状态
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user_info",
		}}},
		{{Key: "$unwind", Value: "$user_info"}},
		{{Key: "$project", Value: bson.M{
			"_id":        1,
			"guild_id":   1,
			"user_id":    1,
			"message":    1,
			"apply_time": 1,
			"status":     1,
			"username":   "$user_info.username",
		}}},
		{{Key: "$sort", Value: bson.M{"apply_time": 1}}},
	}

	cursor, err := m.applicationCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var applications []*models.GuildApplication
	if err = cursor.All(ctx, &applications); err != nil {
		return nil, err
	}

	return applications, nil
}
