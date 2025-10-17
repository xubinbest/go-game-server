package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
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

// ChatDatabase 接口方法实现

func (c *DatabaseClient) SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error {
	return c.chatDB.SaveChatMessage(ctx, message)
}

func (c *DatabaseClient) GetChatMessages(ctx context.Context, channel int32, target_id int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error) {
	return c.chatDB.GetChatMessages(ctx, channel, target_id, page, pageSize)
}

// UserLogDatabase 接口方法实现

func (c *DatabaseClient) CreateUserCreateLog(ctx context.Context, log *models.UserCreateLog) error {
	return c.logDB.CreateUserCreateLog(ctx, log)
}

func (c *DatabaseClient) CreateUserLoginLog(ctx context.Context, log *models.UserLoginLog) error {
	return c.logDB.CreateUserLoginLog(ctx, log)
}

func (c *DatabaseClient) CreateUserLogoutLog(ctx context.Context, log *models.UserLogoutLog) error {
	return c.logDB.CreateUserLogoutLog(ctx, log)
}

func (c *DatabaseClient) CreateUserItemLog(ctx context.Context, log *models.UserItemLog) error {
	return c.logDB.CreateUserItemLog(ctx, log)
}

func (c *DatabaseClient) CreateUserMoneyLog(ctx context.Context, log *models.UserMoneyLog) error {
	return c.logDB.CreateUserMoneyLog(ctx, log)
}

func (c *DatabaseClient) BatchCreateUserCreateLogs(ctx context.Context, logs []*models.UserCreateLog) error {
	return c.logDB.BatchCreateUserCreateLogs(ctx, logs)
}

func (c *DatabaseClient) BatchCreateUserLoginLogs(ctx context.Context, logs []*models.UserLoginLog) error {
	return c.logDB.BatchCreateUserLoginLogs(ctx, logs)
}

func (c *DatabaseClient) BatchCreateUserLogoutLogs(ctx context.Context, logs []*models.UserLogoutLog) error {
	return c.logDB.BatchCreateUserLogoutLogs(ctx, logs)
}

func (c *DatabaseClient) BatchCreateUserItemLogs(ctx context.Context, logs []*models.UserItemLog) error {
	return c.logDB.BatchCreateUserItemLogs(ctx, logs)
}

func (c *DatabaseClient) BatchCreateUserMoneyLogs(ctx context.Context, logs []*models.UserMoneyLog) error {
	return c.logDB.BatchCreateUserMoneyLogs(ctx, logs)
}

func (c *DatabaseClient) GetUserCreateLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserCreateLog, error) {
	return c.logDB.GetUserCreateLogs(ctx, userID, limit, offset)
}

func (c *DatabaseClient) GetUserLoginLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLoginLog, error) {
	return c.logDB.GetUserLoginLogs(ctx, userID, limit, offset)
}

func (c *DatabaseClient) GetUserLogoutLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLogoutLog, error) {
	return c.logDB.GetUserLogoutLogs(ctx, userID, limit, offset)
}

func (c *DatabaseClient) GetUserItemLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserItemLog, error) {
	return c.logDB.GetUserItemLogs(ctx, userID, limit, offset)
}

func (c *DatabaseClient) GetUserMoneyLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserMoneyLog, error) {
	return c.logDB.GetUserMoneyLogs(ctx, userID, limit, offset)
}
