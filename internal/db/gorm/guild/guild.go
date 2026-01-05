package guild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/gorm"
)

// GormGuildDatabase GORM公会数据库实现
type GormGuildDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormGuildDatabase 创建GORM公会数据库实例
func NewGormGuildDatabase(db *gorm.DB, sf *snowflake.Snowflake) *GormGuildDatabase {
	return &GormGuildDatabase{
		db: db,
		sf: sf,
	}
}

// CreateGuild 创建公会
func (g *GormGuildDatabase) CreateGuild(ctx context.Context, guild *models.Guild) error {
	// 生成ID
	guildID, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate guild ID: %w", err)
	}
	guild.ID = guildID

	// 设置创建时间
	guild.CreatedAt = time.Now()

	if err := g.db.WithContext(ctx).Create(guild).Error; err != nil {
		return fmt.Errorf("failed to create guild: %w", err)
	}

	return nil
}

// CreateGuildWithMaster 创建公会并添加会长
func (g *GormGuildDatabase) CreateGuildWithMaster(ctx context.Context, guild *models.Guild, master *models.GuildMember) error {
	// 开始事务
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成公会ID
	guildID, err := g.sf.NextID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate guild ID: %w", err)
	}
	guild.ID = guildID
	guild.CreatedAt = time.Now()

	// 创建公会
	if err := tx.Create(guild).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create guild: %w", err)
	}

	// 生成成员ID并设置公会ID
	masterID, err := g.sf.NextID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate master ID: %w", err)
	}
	master.ID = masterID
	master.GuildID = guild.ID
	master.JoinTime = time.Now()
	master.LastLogin = time.Now()

	// 创建会长成员记录
	if err := tx.Create(master).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create guild master: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetGuildByName 根据名称获取公会
func (g *GormGuildDatabase) GetGuildByName(ctx context.Context, name string) (*models.Guild, error) {
	var guild models.Guild

	err := g.db.WithContext(ctx).
		Where("name = ?", name).
		First(&guild).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 公会不存在
		}
		return nil, fmt.Errorf("failed to get guild by name: %w", err)
	}

	return &guild, nil
}

// GetGuild 根据ID获取公会
func (g *GormGuildDatabase) GetGuild(ctx context.Context, guildID int64) (*models.Guild, error) {
	var guild models.Guild

	err := g.db.WithContext(ctx).
		Where("id = ?", guildID).
		First(&guild).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 公会不存在
		}
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	return &guild, nil
}

// UpdateGuild 更新公会信息
func (g *GormGuildDatabase) UpdateGuild(ctx context.Context, guild *models.Guild) error {
	err := g.db.WithContext(ctx).Save(guild).Error
	if err != nil {
		return fmt.Errorf("failed to update guild: %w", err)
	}

	return nil
}

// DeleteGuild 删除公会
func (g *GormGuildDatabase) DeleteGuild(ctx context.Context, guildID int64) error {
	// 开始事务
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除公会成员
	if err := tx.Where("guild_id = ?", guildID).Delete(&models.GuildMember{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete guild members: %w", err)
	}

	// 删除公会申请
	if err := tx.Where("guild_id = ?", guildID).Delete(&models.GuildApplication{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete guild applications: %w", err)
	}

	// 删除公会邀请
	if err := tx.Where("guild_id = ?", guildID).Delete(&models.GuildInvitation{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete guild invitations: %w", err)
	}

	// 删除公会
	if err := tx.Where("id = ?", guildID).Delete(&models.Guild{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete guild: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetUserGuilds 获取用户的所有公会
func (g *GormGuildDatabase) GetUserGuilds(ctx context.Context, userID int64) ([]*models.Guild, error) {
	var guilds []*models.Guild

	err := g.db.WithContext(ctx).
		Joins("JOIN guild_members ON guilds.id = guild_members.guild_id").
		Where("guild_members.user_id = ?", userID).
		Find(&guilds).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user guilds: %w", err)
	}

	return guilds, nil
}

// GetGuildMemberCount 获取公会成员数量
func (g *GormGuildDatabase) GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error) {
	var count int64

	err := g.db.WithContext(ctx).
		Model(&models.GuildMember{}).
		Where("guild_id = ?", guildID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to get guild member count: %w", err)
	}

	return int32(count), nil
}

// GetGuildList 获取公会列表
func (g *GormGuildDatabase) GetGuildList(ctx context.Context, page, pageSize int32) ([]*models.Guild, int32, error) {
	var guilds []*models.Guild
	var total int64

	// 获取总数
	err := g.db.WithContext(ctx).Model(&models.Guild{}).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count guilds: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = g.db.WithContext(ctx).
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&guilds).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get guild list: %w", err)
	}

	return guilds, int32(total), nil
}
