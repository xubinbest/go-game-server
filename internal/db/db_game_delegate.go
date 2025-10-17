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

// CardDatabase 接口方法实现

func (c *DatabaseClient) GetUserCards(ctx context.Context, userID int64) ([]*models.Card, error) {
	return c.cardDB.GetUserCards(ctx, userID)
}

func (c *DatabaseClient) GetUserCard(ctx context.Context, userID int64, cardID int64) (*models.Card, error) {
	return c.cardDB.GetUserCard(ctx, userID, cardID)
}

func (c *DatabaseClient) CreateCard(ctx context.Context, card *models.Card) error {
	return c.cardDB.CreateCard(ctx, card)
}

func (c *DatabaseClient) UpdateCard(ctx context.Context, card *models.Card) error {
	return c.cardDB.UpdateCard(ctx, card)
}

func (c *DatabaseClient) UpgradeCard(ctx context.Context, userID int64, cardID int64, newLevel int32) error {
	return c.cardDB.UpgradeCard(ctx, userID, cardID, newLevel)
}

func (c *DatabaseClient) UpgradeCardStar(ctx context.Context, userID int64, cardID int64, newStar int32) error {
	return c.cardDB.UpgradeCardStar(ctx, userID, cardID, newStar)
}

func (c *DatabaseClient) CardExists(ctx context.Context, userID int64, templateID int64) (bool, error) {
	return c.cardDB.CardExists(ctx, userID, templateID)
}

// PetDatabase 接口方法实现

func (c *DatabaseClient) GetPet(ctx context.Context, petID int64) (*models.Pet, error) {
	return c.petDB.GetPet(ctx, petID)
}

func (c *DatabaseClient) GetUserPets(ctx context.Context, userID int64) ([]*models.Pet, error) {
	return c.petDB.GetUserPets(ctx, userID)
}

func (c *DatabaseClient) GetUserBattlePet(ctx context.Context, userID int64) (*models.Pet, error) {
	return c.petDB.GetUserBattlePet(ctx, userID)
}

func (c *DatabaseClient) CreatePet(ctx context.Context, pet *models.Pet) error {
	return c.petDB.CreatePet(ctx, pet)
}

func (c *DatabaseClient) UpdatePet(ctx context.Context, pet *models.Pet) error {
	return c.petDB.UpdatePet(ctx, pet)
}

func (c *DatabaseClient) DeletePet(ctx context.Context, petID int64) error {
	return c.petDB.DeletePet(ctx, petID)
}

func (c *DatabaseClient) SetPetBattleStatus(ctx context.Context, userID int64, petID int64, isBattle bool) error {
	return c.petDB.SetPetBattleStatus(ctx, userID, petID, isBattle)
}

func (c *DatabaseClient) CancelAllPetBattleStatus(ctx context.Context, userID int64) error {
	return c.petDB.CancelAllPetBattleStatus(ctx, userID)
}
