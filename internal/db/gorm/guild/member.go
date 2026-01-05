package guild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"gorm.io/gorm"
)

// AddGuildMember 添加公会成员
func (g *GormGuildDatabase) AddGuildMember(ctx context.Context, member *models.GuildMember) error {
	// 生成ID
	memberID, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate member ID: %w", err)
	}
	member.ID = memberID
	member.JoinTime = time.Now()
	member.LastLogin = time.Now()

	if err := g.db.WithContext(ctx).Create(member).Error; err != nil {
		return fmt.Errorf("failed to add guild member: %w", err)
	}

	return nil
}

// GetGuildMember 获取公会成员
func (g *GormGuildDatabase) GetGuildMember(ctx context.Context, guildID, userID int64) (*models.GuildMember, error) {
	var member models.GuildMember

	err := g.db.WithContext(ctx).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 成员不存在
		}
		return nil, fmt.Errorf("failed to get guild member: %w", err)
	}

	return &member, nil
}

// GetGuildMembers 获取公会所有成员
func (g *GormGuildDatabase) GetGuildMembers(ctx context.Context, guildID int64) ([]*models.GuildMember, error) {
	var members []*models.GuildMember

	err := g.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Find(&members).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get guild members: %w", err)
	}

	return members, nil
}

// UpdateGuildMemberRole 更新公会成员角色
func (g *GormGuildDatabase) UpdateGuildMemberRole(ctx context.Context, guildID, userID int64, newRole int) error {
	err := g.db.WithContext(ctx).
		Model(&models.GuildMember{}).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		Update("role", newRole).Error

	if err != nil {
		return fmt.Errorf("failed to update guild member role: %w", err)
	}

	return nil
}

// RemoveGuildMember 移除公会成员
func (g *GormGuildDatabase) RemoveGuildMember(ctx context.Context, guildID, userID int64) error {
	err := g.db.WithContext(ctx).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		Delete(&models.GuildMember{}).Error

	if err != nil {
		return fmt.Errorf("failed to remove guild member: %w", err)
	}

	return nil
}
