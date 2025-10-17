package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// UserLogDatabase 定义用户日志相关的数据库操作接口
type UserLogDatabase interface {
	// 用户创建日志
	CreateUserCreateLog(ctx context.Context, log *models.UserCreateLog) error

	// 用户登录日志
	CreateUserLoginLog(ctx context.Context, log *models.UserLoginLog) error

	// 用户登出日志
	CreateUserLogoutLog(ctx context.Context, log *models.UserLogoutLog) error

	// 用户物品日志
	CreateUserItemLog(ctx context.Context, log *models.UserItemLog) error

	// 用户货币日志
	CreateUserMoneyLog(ctx context.Context, log *models.UserMoneyLog) error

	// 批量创建用户创建日志
	BatchCreateUserCreateLogs(ctx context.Context, logs []*models.UserCreateLog) error

	// 批量创建用户登录日志
	BatchCreateUserLoginLogs(ctx context.Context, logs []*models.UserLoginLog) error

	// 批量创建用户登出日志
	BatchCreateUserLogoutLogs(ctx context.Context, logs []*models.UserLogoutLog) error

	// 批量创建用户物品日志
	BatchCreateUserItemLogs(ctx context.Context, logs []*models.UserItemLog) error

	// 批量创建用户货币日志
	BatchCreateUserMoneyLogs(ctx context.Context, logs []*models.UserMoneyLog) error

	// 查询用户创建日志
	GetUserCreateLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserCreateLog, error)

	// 查询用户登录日志
	GetUserLoginLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLoginLog, error)

	// 查询用户登出日志
	GetUserLogoutLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLogoutLog, error)

	// 查询用户物品日志
	GetUserItemLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserItemLog, error)

	// 查询用户货币日志
	GetUserMoneyLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserMoneyLog, error)
}
