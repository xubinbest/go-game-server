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

// 获取帮派成员集合
func (m *MongoDBGuildDatabase) memberCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_members")
}

// AddGuildMember 添加帮派成员
func (m *MongoDBGuildDatabase) AddGuildMember(ctx context.Context, member *models.GuildMember) error {
	if member.ID == 0 {
		var err error
		member.ID, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	if member.JoinTime.IsZero() {
		member.JoinTime = time.Now()
	}

	// 检查用户是否已经是帮派成员
	count, err := m.memberCollection().CountDocuments(ctx, bson.M{
		"guild_id": member.GuildID,
		"user_id":  member.UserID,
	})
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("user is already a member of this guild")
	}

	// 添加成员
	document := bson.M{
		"_id":        member.ID,
		"guild_id":   member.GuildID,
		"user_id":    member.UserID,
		"role":       member.Role,
		"join_time":  member.JoinTime,
		"last_login": member.LastLogin,
	}

	_, err = m.memberCollection().InsertOne(ctx, document)
	return err
}

// GetGuildMember 获取帮派成员
func (m *MongoDBGuildDatabase) GetGuildMember(ctx context.Context, guildID, userID int64) (*models.GuildMember, error) {
	filter := bson.M{
		"guild_id": guildID,
		"user_id":  userID,
	}

	var member models.GuildMember
	err := m.memberCollection().FindOne(ctx, filter).Decode(&member)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 成员不存在
		}
		return nil, err
	}

	return &member, nil
}

// GetGuildMembers 获取帮派所有成员
func (m *MongoDBGuildDatabase) GetGuildMembers(ctx context.Context, guildID int64) ([]*models.GuildMember, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"guild_id": guildID}}},
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
			"role":       1,
			"join_time":  1,
			"last_login": 1,
			"username":   "$user_info.username",
		}}},
		{{Key: "$sort", Value: bson.M{
			"role":      1,
			"join_time": 1,
		}}},
	}

	cursor, err := m.memberCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []*models.GuildMember
	if err = cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// UpdateGuildMemberRole 更新帮派成员角色
func (m *MongoDBGuildDatabase) UpdateGuildMemberRole(ctx context.Context, guildID, userID int64, newRole int) error {
	filter := bson.M{
		"guild_id": guildID,
		"user_id":  userID,
	}

	update := bson.M{
		"$set": bson.M{
			"role": newRole,
		},
	}

	result, err := m.memberCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("guild member not found")
	}

	return nil
}

// RemoveGuildMember 移除帮派成员
func (m *MongoDBGuildDatabase) RemoveGuildMember(ctx context.Context, guildID, userID int64) error {
	// 检查是否是帮主
	guild, err := m.GetGuild(ctx, guildID)
	if err != nil {
		return err
	}

	if guild == nil {
		return errors.New("guild not found")
	}

	if guild.MasterID == userID {
		return errors.New("cannot remove guild master")
	}

	filter := bson.M{
		"guild_id": guildID,
		"user_id":  userID,
	}

	result, err := m.memberCollection().DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("guild member not found")
	}

	return nil
}
