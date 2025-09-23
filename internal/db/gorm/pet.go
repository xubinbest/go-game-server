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

// GormPetDatabase GORM宠物数据库实现
type GormPetDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormPetDatabase 创建GORM宠物数据库实例
func NewGormPetDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.PetDatabase {
	return &GormPetDatabase{
		db: db,
		sf: sf,
	}
}

// GetPet 根据ID获取宠物
func (g *GormPetDatabase) GetPet(ctx context.Context, petID int64) (*models.Pet, error) {
	var pet models.Pet

	err := g.db.WithContext(ctx).
		Where("id = ?", petID).
		First(&pet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 宠物不存在
		}
		return nil, fmt.Errorf("failed to get pet: %w", err)
	}

	return &pet, nil
}

// GetUserPets 获取用户的所有宠物
func (g *GormPetDatabase) GetUserPets(ctx context.Context, userID int64) ([]*models.Pet, error) {
	var pets []*models.Pet

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&pets).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get user pets: %w", err)
	}

	return pets, nil
}

// GetUserBattlePet 获取用户的出战宠物
func (g *GormPetDatabase) GetUserBattlePet(ctx context.Context, userID int64) (*models.Pet, error) {
	var pet models.Pet

	err := g.db.WithContext(ctx).
		Where("user_id = ? AND is_battle = ?", userID, true).
		First(&pet).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有出战宠物
		}
		return nil, fmt.Errorf("failed to get user battle pet: %w", err)
	}

	return &pet, nil
}

// CreatePet 创建新宠物
func (g *GormPetDatabase) CreatePet(ctx context.Context, pet *models.Pet) error {
	// 生成ID
	id, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate pet ID: %w", err)
	}
	pet.ID = id

	// 设置创建时间
	now := time.Now()
	pet.CreatedAt = now
	pet.UpdatedAt = now

	if err := g.db.WithContext(ctx).Create(pet).Error; err != nil {
		return fmt.Errorf("failed to create pet: %w", err)
	}

	return nil
}

// UpdatePet 更新宠物信息
func (g *GormPetDatabase) UpdatePet(ctx context.Context, pet *models.Pet) error {
	// 设置更新时间
	pet.UpdatedAt = time.Now()

	err := g.db.WithContext(ctx).Save(pet).Error
	if err != nil {
		return fmt.Errorf("failed to update pet: %w", err)
	}

	return nil
}

// DeletePet 删除宠物
func (g *GormPetDatabase) DeletePet(ctx context.Context, petID int64) error {
	err := g.db.WithContext(ctx).Where("id = ?", petID).Delete(&models.Pet{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete pet: %w", err)
	}

	return nil
}

// SetPetBattleStatus 设置宠物出战状态
func (g *GormPetDatabase) SetPetBattleStatus(ctx context.Context, userID int64, petID int64, isBattle bool) error {
	// 如果设置为出战，先取消其他宠物的出战状态
	if isBattle {
		err := g.db.WithContext(ctx).
			Model(&models.Pet{}).
			Where("user_id = ?", userID).
			Update("is_battle", false).Error
		if err != nil {
			return fmt.Errorf("failed to cancel other pets battle status: %w", err)
		}
	}

	// 更新指定宠物的出战状态
	err := g.db.WithContext(ctx).
		Model(&models.Pet{}).
		Where("id = ? AND user_id = ?", petID, userID).
		Updates(map[string]interface{}{
			"is_battle":  isBattle,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to set pet battle status: %w", err)
	}

	return nil
}

// CancelAllPetBattleStatus 取消所有宠物的出战状态
func (g *GormPetDatabase) CancelAllPetBattleStatus(ctx context.Context, userID int64) error {
	err := g.db.WithContext(ctx).
		Model(&models.Pet{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"is_battle":  false,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to cancel all pets battle status: %w", err)
	}

	return nil
}
