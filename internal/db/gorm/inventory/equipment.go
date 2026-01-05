package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"gorm.io/gorm"
)

// GetEquipments 获取用户装备
func (g *GormInventoryDatabase) GetEquipments(ctx context.Context, userID int64) ([]*models.Equipment, error) {
	var equipments []*models.Equipment

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&equipments).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get equipments: %w", err)
	}

	return equipments, nil
}

// EquipItem 装备物品
func (g *GormInventoryDatabase) EquipItem(ctx context.Context, userID int64, itemID int64, slot int32) error {
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

	// 取消该槽位的其他装备
	if err := tx.Model(&models.Equipment{}).
		Where("user_id = ? AND slot = ?", userID, slot).
		Delete(&models.Equipment{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unequip existing item: %w", err)
	}

	// 获取物品信息
	var item models.InventoryItem
	if err := tx.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get item: %w", err)
	}

	// 生成装备ID
	equipmentID, err := g.sf.NextID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate equipment ID: %w", err)
	}

	// 创建装备记录
	equipment := &models.Equipment{
		ID:         equipmentID,
		UserID:     userID,
		TemplateID: item.TemplateID,
		Slot:       slot,
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}

	if err := tx.Create(equipment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create equipment: %w", err)
	}

	// 更新物品装备状态
	if err := tx.Model(&models.InventoryItem{}).
		Where("id = ?", itemID).
		Update("equipped", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update item equipped status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UnequipItem 卸下装备
func (g *GormInventoryDatabase) UnequipItem(ctx context.Context, userID int64, slot int32) error {
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

	// 获取装备信息
	var equipment models.Equipment
	if err := tx.Where("user_id = ? AND slot = ?", userID, slot).First(&equipment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no equipment in slot %d", slot)
		}
		return fmt.Errorf("failed to get equipment: %w", err)
	}

	// 删除装备记录
	if err := tx.Delete(&equipment).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	// 更新物品装备状态
	if err := tx.Model(&models.InventoryItem{}).
		Where("user_id = ? AND template_id = ?", userID, equipment.TemplateID).
		Update("equipped", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update item equipped status: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetEquipmentBySlot 获取指定槽位的装备
func (g *GormInventoryDatabase) GetEquipmentBySlot(ctx context.Context, userID int64, slot int32) (*models.Equipment, error) {
	var equipment models.Equipment

	err := g.db.WithContext(ctx).
		Where("user_id = ? AND slot = ?", userID, slot).
		First(&equipment).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 槽位无装备
		}
		return nil, fmt.Errorf("failed to get equipment by slot: %w", err)
	}

	return &equipment, nil
}
