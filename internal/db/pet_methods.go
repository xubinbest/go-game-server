package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

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
