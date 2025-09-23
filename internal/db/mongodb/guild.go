package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBGuildDatabase 实现 GuildDatabase 接口
type MongoDBGuildDatabase struct {
	client   *mongo.Client
	database string
	sf       *snowflake.Snowflake
}

// NewMongoDBGuildDatabase 创建 MongoDBGuildDatabase 实例
func NewMongoDBGuildDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.GuildDatabase {
	return &MongoDBGuildDatabase{
		client:   client,
		database: database,
		sf:       sf,
	}
}

// 获取帮派集合
func (m *MongoDBGuildDatabase) guildCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guilds")
}

// 获取帮派成员集合
func (m *MongoDBGuildDatabase) memberCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_members")
}

// 获取帮派申请集合
func (m *MongoDBGuildDatabase) applicationCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_applications")
}

// 获取帮派邀请集合
func (m *MongoDBGuildDatabase) invitationCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("guild_invitations")
}

// CreateGuild 创建帮派
func (m *MongoDBGuildDatabase) CreateGuild(ctx context.Context, guild *models.Guild) error {
	if guild.ID == 0 {
		var err error
		guild.ID, err = m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}
	}

	if guild.CreatedAt.IsZero() {
		guild.CreatedAt = time.Now()
	}

	// 检查帮派名称是否已存在
	count, err := m.guildCollection().CountDocuments(ctx, bson.M{"name": guild.Name})
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("guild name already exists")
	}

	// 创建帮派
	document := bson.M{
		"_id":          guild.ID,
		"name":         guild.Name,
		"description":  guild.Description,
		"announcement": guild.Announcement,
		"master_id":    guild.MasterID,
		"created_at":   guild.CreatedAt,
		"max_members":  guild.MaxMembers,
		"version":      1,
	}

	_, err = m.guildCollection().InsertOne(ctx, document)
	return err
}

// CreateGuildWithMaster 创建帮派并添加帮主
func (m *MongoDBGuildDatabase) CreateGuildWithMaster(ctx context.Context, guild *models.Guild, master *models.GuildMember) error {
	// 使用事务确保原子性
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 在事务中执行所有操作
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// 生成ID
		if guild.ID == 0 {
			guild.ID, err = m.sf.NextID()
			if err != nil {
				return fmt.Errorf("failed to generate guild ID: %w", err)
			}
		}

		if master.ID == 0 {
			master.ID, err = m.sf.NextID()
			if err != nil {
				return fmt.Errorf("failed to generate member ID: %w", err)
			}
		}

		// 设置时间
		now := time.Now()
		if guild.CreatedAt.IsZero() {
			guild.CreatedAt = now
		}

		if master.JoinTime.IsZero() {
			master.JoinTime = now
		}

		// 检查帮派名称是否已存在
		count, err := m.guildCollection().CountDocuments(sc, bson.M{"name": guild.Name})
		if err != nil {
			return err
		}

		if count > 0 {
			return errors.New("guild name already exists")
		}

		// 创建帮派
		guildDoc := bson.M{
			"_id":          guild.ID,
			"name":         guild.Name,
			"description":  guild.Description,
			"announcement": guild.Announcement,
			"master_id":    guild.MasterID,
			"created_at":   guild.CreatedAt,
			"max_members":  guild.MaxMembers,
			"version":      1,
		}

		_, err = m.guildCollection().InsertOne(sc, guildDoc)
		if err != nil {
			return fmt.Errorf("failed to create guild: %w", err)
		}

		// 添加帮主
		master.GuildID = guild.ID
		master.Role = models.GuildRoleMaster

		memberDoc := bson.M{
			"_id":        master.ID,
			"guild_id":   master.GuildID,
			"user_id":    master.UserID,
			"role":       master.Role,
			"join_time":  master.JoinTime,
			"last_login": master.LastLogin,
		}

		_, err = m.memberCollection().InsertOne(sc, memberDoc)
		if err != nil {
			return fmt.Errorf("failed to add guild master: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// GetGuildByName 根据名称获取帮派
func (m *MongoDBGuildDatabase) GetGuildByName(ctx context.Context, name string) (*models.Guild, error) {
	filter := bson.M{"name": name}

	var guild models.Guild
	err := m.guildCollection().FindOne(ctx, filter).Decode(&guild)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 帮派不存在
		}
		return nil, err
	}

	return &guild, nil
}

// GetGuild 根据ID获取帮派
func (m *MongoDBGuildDatabase) GetGuild(ctx context.Context, guildID int64) (*models.Guild, error) {
	filter := bson.M{"_id": guildID}

	var guild models.Guild
	err := m.guildCollection().FindOne(ctx, filter).Decode(&guild)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // 帮派不存在
		}
		return nil, err
	}

	return &guild, nil
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

// UpdateGuild 更新帮派信息
func (m *MongoDBGuildDatabase) UpdateGuild(ctx context.Context, guild *models.Guild) error {
	filter := bson.M{
		"_id":     guild.ID,
		"version": guild.Version,
	}

	update := bson.M{
		"$set": bson.M{
			"name":         guild.Name,
			"description":  guild.Description,
			"announcement": guild.Announcement,
			"master_id":    guild.MasterID,
			"max_members":  guild.MaxMembers,
			"version":      guild.Version + 1,
		},
	}

	result, err := m.guildCollection().UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("guild not found or data has been modified")
	}

	// 更新版本号
	guild.Version++

	return nil
}

// DeleteGuild 删除帮派
func (m *MongoDBGuildDatabase) DeleteGuild(ctx context.Context, guildID int64) error {
	// 使用事务确保原子性
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 在事务中执行所有操作
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// 删除帮派成员
		_, err := m.memberCollection().DeleteMany(sc, bson.M{"guild_id": guildID})
		if err != nil {
			return err
		}

		// 删除帮派申请
		_, err = m.applicationCollection().DeleteMany(sc, bson.M{"guild_id": guildID})
		if err != nil {
			return err
		}

		// 删除帮派邀请
		_, err = m.invitationCollection().DeleteMany(sc, bson.M{"guild_id": guildID})
		if err != nil {
			return err
		}

		// 删除帮派
		result, err := m.guildCollection().DeleteOne(sc, bson.M{"_id": guildID})
		if err != nil {
			return err
		}

		if result.DeletedCount == 0 {
			return errors.New("guild not found")
		}

		return session.CommitTransaction(sc)
	})

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

// GetGuildMemberCount 获取帮派成员数量
func (m *MongoDBGuildDatabase) GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error) {
	filter := bson.M{"guild_id": guildID}
	count, err := m.memberCollection().CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
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
