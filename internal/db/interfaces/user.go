package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// UserDatabase 定义用户相关的数据库操作接口
type UserDatabase interface {
	// 根据用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)

	// 根据ID获取用户
	GetUser(ctx context.Context, userID int64) (*models.User, error)

	// 创建新用户
	CreateUser(ctx context.Context, user *models.User) error

	// 更新用户信息
	UpdateUser(ctx context.Context, user *models.User) error

	// 删除用户
	DeleteUser(ctx context.Context, userID int64) error

	// 月签到相关方法
	// 获取用户月签到信息
	GetMonthlySign(ctx context.Context, userID int64) (*models.MonthlySign, error)

	// 创建或更新月签到信息
	CreateOrUpdateMonthlySign(ctx context.Context, sign *models.MonthlySign) error

	// 创建月签到记录（用于首次签到）
	CreateMonthlySign(ctx context.Context, sign *models.MonthlySign) error

	// 获取用户月签到累计奖励记录
	GetMonthlySignReward(ctx context.Context, userID int64) (*models.MonthlySignReward, error)

	// 创建或更新月签到累计奖励记录
	CreateOrUpdateMonthlySignReward(ctx context.Context, reward *models.MonthlySignReward) error

	// 其他用户相关方法...
}
