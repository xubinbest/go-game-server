package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

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
