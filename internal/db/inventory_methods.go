package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// InventoryDatabase 接口方法实现

func (c *DatabaseClient) GetInventory(ctx context.Context, userID int64) (*models.Inventory, error) {
	return c.inventoryDB.GetInventory(ctx, userID)
}

func (c *DatabaseClient) AddItemByTemplate(ctx context.Context, userID int64, templateID int64, count int32) error {
	return c.inventoryDB.AddItemByTemplate(ctx, userID, templateID, count)
}

func (c *DatabaseClient) AddItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	return c.inventoryDB.AddItem(ctx, userID, itemID, count)
}

func (c *DatabaseClient) RemoveItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	return c.inventoryDB.RemoveItem(ctx, userID, itemID, count)
}

func (c *DatabaseClient) UpdateItemCount(ctx context.Context, userID int64, itemID int64, newCount int32) error {
	return c.inventoryDB.UpdateItemCount(ctx, userID, itemID, newCount)
}

func (c *DatabaseClient) HasEnoughItems(ctx context.Context, userID int64, itemID int64, requiredCount int32) (bool, error) {
	return c.inventoryDB.HasEnoughItems(ctx, userID, itemID, requiredCount)
}

func (c *DatabaseClient) GetEquipments(ctx context.Context, userID int64) ([]*models.Equipment, error) {
	return c.inventoryDB.GetEquipments(ctx, userID)
}

func (c *DatabaseClient) EquipItem(ctx context.Context, userID int64, itemID int64, slot int32) error {
	return c.inventoryDB.EquipItem(ctx, userID, itemID, slot)
}

func (c *DatabaseClient) UnequipItem(ctx context.Context, userID int64, slot int32) error {
	return c.inventoryDB.UnequipItem(ctx, userID, slot)
}

func (c *DatabaseClient) GetEquipmentBySlot(ctx context.Context, userID int64, slot int32) (*models.Equipment, error) {
	return c.inventoryDB.GetEquipmentBySlot(ctx, userID, slot)
}
