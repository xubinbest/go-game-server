package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/gorm"
)

// GormCardDatabase GORM卡牌数据库实现
type GormCardDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormCardDatabase 创建GORM卡牌数据库实例
func NewGormCardDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.CardDatabase {
	return &GormCardDatabase{
		db: db,
		sf: sf,
	}
}

// GetUserCards 获取用户的所有卡牌
func (g *GormCardDatabase) GetUserCards(ctx context.Context, userID int64) ([]*models.Card, error) {
	var cards []*models.Card

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&cards).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user cards: %w", err)
	}

	return cards, nil
}

// GetUserCard 获取用户的特定卡牌
func (g *GormCardDatabase) GetUserCard(ctx context.Context, userID int64, cardID int64) (*models.Card, error) {
	var card models.Card

	err := g.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", cardID, userID).
		First(&card).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 卡牌不存在
		}
		return nil, fmt.Errorf("failed to get user card: %w", err)
	}

	return &card, nil
}

// CreateCard 创建新卡牌
func (g *GormCardDatabase) CreateCard(ctx context.Context, card *models.Card) error {
	// 生成ID
	id, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate card ID: %w", err)
	}
	card.ID = id

	// 设置创建时间
	now := time.Now().Unix()
	card.CreatedAt = now
	card.UpdatedAt = now

	if err := g.db.WithContext(ctx).Create(card).Error; err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

// UpdateCard 更新卡牌信息
func (g *GormCardDatabase) UpdateCard(ctx context.Context, card *models.Card) error {
	// 设置更新时间
	card.UpdatedAt = time.Now().Unix()

	err := g.db.WithContext(ctx).Save(card).Error
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	return nil
}

// UpgradeCard 升级卡牌等级
func (g *GormCardDatabase) UpgradeCard(ctx context.Context, userID int64, cardID int64, newLevel int32) error {
	err := g.db.WithContext(ctx).
		Model(&models.Card{}).
		Where("id = ? AND user_id = ?", cardID, userID).
		Updates(map[string]interface{}{
			"level":      newLevel,
			"updated_at": time.Now().Unix(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to upgrade card: %w", err)
	}

	return nil
}

// UpgradeCardStar 升级卡牌星级
func (g *GormCardDatabase) UpgradeCardStar(ctx context.Context, userID int64, cardID int64, newStar int32) error {
	err := g.db.WithContext(ctx).
		Model(&models.Card{}).
		Where("id = ? AND user_id = ?", cardID, userID).
		Updates(map[string]interface{}{
			"star":       newStar,
			"updated_at": time.Now().Unix(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to upgrade card star: %w", err)
	}

	return nil
}

// CardExists 检查卡牌是否存在
func (g *GormCardDatabase) CardExists(ctx context.Context, userID int64, templateID int64) (bool, error) {
	var count int64

	err := g.db.WithContext(ctx).
		Model(&models.Card{}).
		Where("user_id = ? AND template_id = ?", userID, templateID).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check card existence: %w", err)
	}

	return count > 0, nil
}
