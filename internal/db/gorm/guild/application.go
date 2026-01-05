package guild

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"gorm.io/gorm"
)

// CreateGuildApplication 创建公会申请
func (g *GormGuildDatabase) CreateGuildApplication(ctx context.Context, application *models.GuildApplication) error {
	// 生成ID
	appID, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate application ID: %w", err)
	}
	application.ID = appID
	application.Time = time.Now()

	if err := g.db.WithContext(ctx).Create(application).Error; err != nil {
		return fmt.Errorf("failed to create guild application: %w", err)
	}

	return nil
}

// GetGuildApplication 获取公会申请
func (g *GormGuildDatabase) GetGuildApplication(ctx context.Context, appID int64) (*models.GuildApplication, error) {
	var application models.GuildApplication

	err := g.db.WithContext(ctx).
		Where("id = ?", appID).
		First(&application).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 申请不存在
		}
		return nil, fmt.Errorf("failed to get guild application: %w", err)
	}

	return &application, nil
}

// GetGuildApplications 获取公会的所有申请
func (g *GormGuildDatabase) GetGuildApplications(ctx context.Context, guildID int64) ([]*models.GuildApplication, error) {
	var applications []*models.GuildApplication

	err := g.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Find(&applications).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get guild applications: %w", err)
	}

	return applications, nil
}

// DeleteGuildApplication 删除公会申请
func (g *GormGuildDatabase) DeleteGuildApplication(ctx context.Context, appID int64) error {
	err := g.db.WithContext(ctx).Where("id = ?", appID).Delete(&models.GuildApplication{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete guild application: %w", err)
	}

	return nil
}
