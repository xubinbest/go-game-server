package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// InventoryDatabase 定义背包相关的数据库操作接口
type InventoryDatabase interface {
	// 获取用户背包
	GetInventory(ctx context.Context, userID int64) (*models.Inventory, error)

	// 添加物品到背包（通过模板ID）
	AddItemByTemplate(ctx context.Context, userID int64, templateID int64, count int32) error

	// 添加物品到背包（通过实例ID）
	AddItem(ctx context.Context, userID int64, itemID int64, count int32) error

	// 从背包移除物品
	RemoveItem(ctx context.Context, userID int64, itemID int64, count int32) error

	// 更新物品数量
	UpdateItemCount(ctx context.Context, userID int64, itemID int64, newCount int32) error

	// 检查物品是否足够
	HasEnoughItems(ctx context.Context, userID int64, itemID int64, requiredCount int32) (bool, error)

	// 装备相关方法
	GetEquipments(ctx context.Context, userID int64) ([]*models.Equipment, error)
	EquipItem(ctx context.Context, userID int64, itemID int64, slot int32) error
	UnequipItem(ctx context.Context, userID int64, slot int32) error
	GetEquipmentBySlot(ctx context.Context, userID int64, slot int32) (*models.Equipment, error)
}
