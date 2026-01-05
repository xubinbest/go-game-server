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

// 获取帮派邀请集合
func (m *MongoDBGuildDatabase) invitationCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_invitations")
}

// CreateGuildInvitation 创建帮派邀请
func (m *MongoDBGuildDatabase) CreateGuildInvitation(ctx context.Context, invitation *models.GuildInvitation) error {
	if invitation.ID == 0 {
		var err error
		invitation.ID, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	if invitation.Time.IsZero() {
		invitation.Time = time.Now()
	}

	// 检查用户是否已经是帮派成员
	member, err := m.GetGuildMember(ctx, invitation.GuildID, invitation.UserID)
	if err != nil {
		return err
	}

	if member != nil {
		return errors.New("user is already a member of this guild")
	}

	// 检查是否已经有待处理的邀请
	count, err := m.invitationCollection().CountDocuments(ctx, bson.M{
		"guild_id": invitation.GuildID,
		"user_id":  invitation.UserID,
		"status":   1, // 待处理状态
	})
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("invitation already exists")
	}

	// 创建邀请
	document := bson.M{
		"_id":         invitation.ID,
		"guild_id":    invitation.GuildID,
		"user_id":     invitation.UserID,
		"inviter_id":  invitation.InviterID,
		"time":        invitation.Time,
		"expire_time": invitation.ExpireTime,
	}

	_, err = m.invitationCollection().InsertOne(ctx, document)
	return err
}

// GetGuildInvitations 获取帮派邀请列表
func (m *MongoDBGuildDatabase) GetGuildInvitations(ctx context.Context, guildID int64) ([]*models.GuildInvitation, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"guild_id": guildID,
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "user_id",
			"foreignField": "_id",
			"as":           "user_info",
		}}},
		{{Key: "$unwind", Value: "$user_info"}},
		{{Key: "$project", Value: bson.M{
			"_id":         1,
			"guild_id":    1,
			"user_id":     1,
			"inviter_id":  1,
			"invite_time": 1,
			"username":    "$user_info.username",
		}}},
		{{Key: "$sort", Value: bson.M{"invite_time": 1}}},
	}

	cursor, err := m.invitationCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var invitations []*models.GuildInvitation
	if err = cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}

// GetUserPendingInvitations 获取用户的待处理邀请
func (m *MongoDBGuildDatabase) GetUserPendingInvitations(ctx context.Context, userID int64) ([]*models.GuildInvitation, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"status":  1, // 待处理状态
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "guilds",
			"localField":   "guild_id",
			"foreignField": "_id",
			"as":           "guild_info",
		}}},
		{{Key: "$unwind", Value: "$guild_info"}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "inviter_id",
			"foreignField": "_id",
			"as":           "inviter_info",
		}}},
		{{Key: "$unwind", Value: "$inviter_info"}},
		{{Key: "$project", Value: bson.M{
			"_id":          1,
			"guild_id":     1,
			"user_id":      1,
			"inviter_id":   1,
			"invite_time":  1,
			"status":       1,
			"guild_name":   "$guild_info.name",
			"inviter_name": "$inviter_info.username",
		}}},
		{{Key: "$sort", Value: bson.M{"invite_time": 1}}},
	}

	cursor, err := m.invitationCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var invitations []*models.GuildInvitation
	if err = cursor.All(ctx, &invitations); err != nil {
		return nil, err
	}

	return invitations, nil
}
