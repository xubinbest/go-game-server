package guild

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetUserGuilds 获取用户的帮派
func (m *MongoDBGuildDatabase) GetUserGuilds(ctx context.Context, userID int64) ([]*models.Guild, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from":         "guild_members",
			"localField":   "_id",
			"foreignField": "guild_id",
			"as":           "members",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$members",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$match", Value: bson.M{
			"members.user_id": userID,
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":          1,
			"name":         1,
			"description":  1,
			"announcement": 1,
			"master_id":    1,
			"created_at":   1,
			"max_members":  1,
			"version":      1,
			"role":         "$members.role",
		}}},
		{{Key: "$sort", Value: bson.M{
			"created_at": 1,
		}}},
	}

	cursor, err := m.guildCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var guilds []*models.Guild
	if err = cursor.All(ctx, &guilds); err != nil {
		return nil, err
	}

	return guilds, nil
}

// GetGuildList 分页查询帮派列表
func (m *MongoDBGuildDatabase) GetGuildList(ctx context.Context, page, pageSize int32) ([]*models.Guild, int32, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{}}},
		{{Key: "$sort", Value: bson.M{"created_at": -1}}},
		{{Key: "$skip", Value: (page - 1) * pageSize}},
		{{Key: "$limit", Value: pageSize}},
		{{Key: "$project", Value: bson.M{
			"_id":          1,
			"name":         1,
			"description":  1,
			"announcement": 1,
			"master_id":    1,
			"created_at":   1,
			"max_members":  1,
			"version":      1,
		}}},
	}

	cursor, err := m.guildCollection().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var guilds []*models.Guild
	if err = cursor.All(ctx, &guilds); err != nil {
		return nil, 0, err
	}

	// 获取总数
	count, err := m.guildCollection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	return guilds, int32(count), nil
}

// GetGuildMemberCount 获取帮派成员数量
func (m *MongoDBGuildDatabase) GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error) {
	filter := bson.M{"guild_id": guildID}
	memberCollection := m.client.Database(m.database).Collection("guild_members")
	count, err := memberCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}
