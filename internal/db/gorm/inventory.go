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

// GormInventoryDatabase GORM背包数据库实现
type GormInventoryDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormInventoryDatabase 创建GORM背包数据库实例
func NewGormInventoryDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.InventoryDatabase {
	return &GormInventoryDatabase{
		db: db,
		sf: sf,
	}
}

// GetInventory 获取用户背包
func (g *GormInventoryDatabase) GetInventory(ctx context.Context, userID int64) (*models.Inventory, error) {
	var items []*models.InventoryItem

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	inventory := &models.Inventory{
		UserID:   userID,
		Items:    items,
		Capacity: 100, // 默认背包容量
	}

	return inventory, nil
}

// AddItemByTemplate 根据模板ID添加物品
func (g *GormInventoryDatabase) AddItemByTemplate(ctx context.Context, userID int64, templateID int64, count int32) error {
	// 检查是否已存在该模板的物品
	var existingItem models.InventoryItem
	err := g.db.WithContext(ctx).
		Where("user_id = ? AND template_id = ?", userID, templateID).
		First(&existingItem).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to check existing item: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 生成ID
		itemID, err := g.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate item ID: %w", err)
		}

		// 创建新物品
		item := &models.InventoryItem{
			ID:         itemID,
			UserID:     userID,
			TemplateID: templateID,
			Count:      count,
			Equipped:   false,
			CreatedAt:  time.Now().Unix(),
			UpdatedAt:  time.Now().Unix(),
		}

		err = g.db.WithContext(ctx).Create(item).Error
		if err != nil {
			return fmt.Errorf("failed to create item: %w", err)
		}
	} else {
		// 更新现有物品数量
		existingItem.Count += count
		existingItem.UpdatedAt = time.Now().Unix()

		err = g.db.WithContext(ctx).Save(&existingItem).Error
		if err != nil {
			return fmt.Errorf("failed to update item count: %w", err)
		}
	}

	return nil
}

// AddItem 添加物品（根据物品ID）
func (g *GormInventoryDatabase) AddItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	err := g.db.WithContext(ctx).
		Model(&models.InventoryItem{}).
		Where("id = ? AND user_id = ?", itemID, userID).
		UpdateColumns(map[string]interface{}{
			"count":      gorm.Expr("count + ?", count),
			"updated_at": time.Now().Unix(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	return nil
}

// RemoveItem 移除物品
func (g *GormInventoryDatabase) RemoveItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	// 获取当前物品
	var item models.InventoryItem
	err := g.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", itemID, userID).
		First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("failed to get item: %w", err)
	}

	if item.Count < count {
		return fmt.Errorf("insufficient item count")
	}

	if item.Count == count {
		// 删除物品
		err = g.db.WithContext(ctx).Delete(&item).Error
		if err != nil {
			return fmt.Errorf("failed to delete item: %w", err)
		}
	} else {
		// 减少数量
		item.Count -= count
		item.UpdatedAt = time.Now().Unix()

		err = g.db.WithContext(ctx).Save(&item).Error
		if err != nil {
			return fmt.Errorf("failed to update item count: %w", err)
		}
	}

	return nil
}

// UpdateItemCount 更新物品数量
func (g *GormInventoryDatabase) UpdateItemCount(ctx context.Context, userID int64, itemID int64, newCount int32) error {
	err := g.db.WithContext(ctx).
		Model(&models.InventoryItem{}).
		Where("id = ? AND user_id = ?", itemID, userID).
		Updates(map[string]interface{}{
			"count":      newCount,
			"updated_at": time.Now().Unix(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update item count: %w", err)
	}

	return nil
}

// HasEnoughItems 检查是否有足够的物品
func (g *GormInventoryDatabase) HasEnoughItems(ctx context.Context, userID int64, itemID int64, requiredCount int32) (bool, error) {
	var item models.InventoryItem

	err := g.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", itemID, userID).
		First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // 物品不存在
		}
		return false, fmt.Errorf("failed to get item: %w", err)
	}

	return item.Count >= requiredCount, nil
}

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
