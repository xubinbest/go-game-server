package guild

import (
	"context"
	"errors"
	"fmt"
	"time"

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
func NewMongoDBGuildDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) *MongoDBGuildDatabase {
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

		memberCollection := m.client.Database(m.database).Collection("guild_members")
		_, err = memberCollection.InsertOne(sc, memberDoc)
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

		memberCollection := m.client.Database(m.database).Collection("guild_members")
		applicationCollection := m.client.Database(m.database).Collection("guild_applications")
		invitationCollection := m.client.Database(m.database).Collection("guild_invitations")

		// 删除帮派成员
		_, err := memberCollection.DeleteMany(sc, bson.M{"guild_id": guildID})
		if err != nil {
			return err
		}

		// 删除帮派申请
		_, err = applicationCollection.DeleteMany(sc, bson.M{"guild_id": guildID})
		if err != nil {
			return err
		}

		// 删除帮派邀请
		_, err = invitationCollection.DeleteMany(sc, bson.M{"guild_id": guildID})
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
