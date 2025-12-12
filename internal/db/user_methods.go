package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// UserDatabase 接口方法实现

func (c *DatabaseClient) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return c.userDB.GetUserByUsername(ctx, username)
}

func (c *DatabaseClient) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	return c.userDB.GetUser(ctx, userID)
}

func (c *DatabaseClient) CreateUser(ctx context.Context, user *models.User) error {
	return c.userDB.CreateUser(ctx, user)
}

func (c *DatabaseClient) UpdateUser(ctx context.Context, user *models.User) error {
	return c.userDB.UpdateUser(ctx, user)
}

func (c *DatabaseClient) DeleteUser(ctx context.Context, userID int64) error {
	return c.userDB.DeleteUser(ctx, userID)
}

// 月签到相关方法实现

func (c *DatabaseClient) GetMonthlySign(ctx context.Context, userID int64) (*models.MonthlySign, error) {
	return c.userDB.GetMonthlySign(ctx, userID)
}

func (c *DatabaseClient) CreateOrUpdateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	return c.userDB.CreateOrUpdateMonthlySign(ctx, sign)
}

func (c *DatabaseClient) CreateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	return c.userDB.CreateMonthlySign(ctx, sign)
}

func (c *DatabaseClient) GetMonthlySignReward(ctx context.Context, userID int64) (*models.MonthlySignReward, error) {
	return c.userDB.GetMonthlySignReward(ctx, userID)
}

func (c *DatabaseClient) CreateOrUpdateMonthlySignReward(ctx context.Context, reward *models.MonthlySignReward) error {
	return c.userDB.CreateOrUpdateMonthlySignReward(ctx, reward)
}
