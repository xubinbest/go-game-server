package guild

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// CreateGuildInvitation 创建公会邀请
func (g *GormGuildDatabase) CreateGuildInvitation(ctx context.Context, invitation *models.GuildInvitation) error {
	// 生成ID
	invitationID, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate invitation ID: %w", err)
	}
	invitation.ID = invitationID
	invitation.Time = time.Now()

	if err := g.db.WithContext(ctx).Create(invitation).Error; err != nil {
		return fmt.Errorf("failed to create guild invitation: %w", err)
	}

	return nil
}

// GetGuildInvitations 获取公会的所有邀请
func (g *GormGuildDatabase) GetGuildInvitations(ctx context.Context, guildID int64) ([]*models.GuildInvitation, error) {
	var invitations []*models.GuildInvitation

	err := g.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Find(&invitations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get guild invitations: %w", err)
	}

	return invitations, nil
}

// GetUserPendingInvitations 获取用户待处理的邀请
func (g *GormGuildDatabase) GetUserPendingInvitations(ctx context.Context, userID int64) ([]*models.GuildInvitation, error) {
	var invitations []*models.GuildInvitation

	err := g.db.WithContext(ctx).
		Where("user_id = ? AND expire_time > ?", userID, time.Now()).
		Find(&invitations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user pending invitations: %w", err)
	}

	return invitations, nil
}
