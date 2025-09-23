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

// GormUserDatabase GORM用户数据库实现
type GormUserDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormUserDatabase 创建GORM用户数据库实例
func NewGormUserDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.UserDatabase {
	return &GormUserDatabase{
		db: db,
		sf: sf,
	}
}

// GetUserByUsername 根据用户名获取用户
func (g *GormUserDatabase) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User

	err := g.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetUser 根据ID获取用户
func (g *GormUserDatabase) GetUser(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User

	err := g.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 用户不存在
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// CreateUser 创建新用户
func (g *GormUserDatabase) CreateUser(ctx context.Context, user *models.User) error {
	// 生成ID
	id, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate user ID: %w", err)
	}
	user.ID = id

	// 设置创建时间
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := g.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdateUser 更新用户信息
func (g *GormUserDatabase) UpdateUser(ctx context.Context, user *models.User) error {
	// 设置更新时间
	user.UpdatedAt = time.Now()

	err := g.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser 删除用户
func (g *GormUserDatabase) DeleteUser(ctx context.Context, userID int64) error {
	err := g.db.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetMonthlySign 获取用户月签到信息
func (g *GormUserDatabase) GetMonthlySign(ctx context.Context, userID int64) (*models.MonthlySign, error) {
	var sign models.MonthlySign

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&sign).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 签到记录不存在
		}
		return nil, fmt.Errorf("failed to get monthly sign: %w", err)
	}

	return &sign, nil
}

// CreateOrUpdateMonthlySign 创建或更新月签到信息
func (g *GormUserDatabase) CreateOrUpdateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	// 设置更新时间
	sign.UpdatedAt = time.Now()

	err := g.db.WithContext(ctx).Save(sign).Error
	if err != nil {
		return fmt.Errorf("failed to create or update monthly sign: %w", err)
	}

	return nil
}

// CreateMonthlySign 创建月签到信息
func (g *GormUserDatabase) CreateMonthlySign(ctx context.Context, sign *models.MonthlySign) error {
	// 设置创建时间
	now := time.Now()
	sign.CreatedAt = now
	sign.UpdatedAt = now

	err := g.db.WithContext(ctx).Create(sign).Error
	if err != nil {
		return fmt.Errorf("failed to create monthly sign: %w", err)
	}

	return nil
}

// GetMonthlySignReward 获取用户月签到奖励信息
func (g *GormUserDatabase) GetMonthlySignReward(ctx context.Context, userID int64) (*models.MonthlySignReward, error) {
	var reward models.MonthlySignReward

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&reward).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 奖励记录不存在
		}
		return nil, fmt.Errorf("failed to get monthly sign reward: %w", err)
	}

	return &reward, nil
}

// CreateOrUpdateMonthlySignReward 创建或更新月签到奖励信息
func (g *GormUserDatabase) CreateOrUpdateMonthlySignReward(ctx context.Context, reward *models.MonthlySignReward) error {
	// 设置更新时间
	reward.UpdatedAt = time.Now()

	err := g.db.WithContext(ctx).Save(reward).Error
	if err != nil {
		return fmt.Errorf("failed to create or update monthly sign reward: %w", err)
	}

	return nil
}
