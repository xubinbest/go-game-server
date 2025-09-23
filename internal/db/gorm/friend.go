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

// GormFriendDatabase GORM好友数据库实现
type GormFriendDatabase struct {
	db *gorm.DB
	sf *snowflake.Snowflake
}

// NewGormFriendDatabase 创建GORM好友数据库实例
func NewGormFriendDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.FriendDatabase {
	return &GormFriendDatabase{
		db: db,
		sf: sf,
	}
}

// GetFriends 获取用户的好友列表
func (g *GormFriendDatabase) GetFriends(ctx context.Context, userID int64) ([]*models.Friend, error) {
	var friends []*models.Friend

	err := g.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&friends).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	return friends, nil
}

// GetFriend 获取特定好友关系
func (g *GormFriendDatabase) GetFriend(ctx context.Context, userID, friendID int64) (*models.Friend, error) {
	var friend models.Friend

	err := g.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&friend).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 好友关系不存在
		}
		return nil, fmt.Errorf("failed to get friend: %w", err)
	}

	return &friend, nil
}

// CreateFriendRequest 创建好友请求
func (g *GormFriendDatabase) CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) error {
	// 生成ID
	requestID, err := g.sf.NextID()
	if err != nil {
		return fmt.Errorf("failed to generate request ID: %w", err)
	}

	// 获取发送者用户名
	var fromUser models.User
	if err := g.db.WithContext(ctx).
		Where("id = ?", fromUserID).
		First(&fromUser).Error; err != nil {
		return fmt.Errorf("failed to get from user: %w", err)
	}

	request := &models.FriendRequest{
		ID:           requestID,
		FromUserID:   fromUserID,
		ToUserID:     toUserID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       1, // 待处理
		FromUsername: fromUser.Username,
	}

	if err := g.db.WithContext(ctx).Create(request).Error; err != nil {
		return fmt.Errorf("failed to create friend request: %w", err)
	}

	return nil
}

// GetFriendRequests 获取用户的好友请求列表
func (g *GormFriendDatabase) GetFriendRequests(ctx context.Context, userID int64) ([]*models.FriendRequest, error) {
	var requests []*models.FriendRequest

	err := g.db.WithContext(ctx).
		Where("to_user_id = ? AND status = ?", userID, 1). // 只获取待处理的请求
		Find(&requests).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get friend requests: %w", err)
	}

	return requests, nil
}

// GetFriendRequest 获取特定的好友请求
func (g *GormFriendDatabase) GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	var request models.FriendRequest

	err := g.db.WithContext(ctx).
		Where("id = ?", requestID).
		First(&request).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 请求不存在
		}
		return nil, fmt.Errorf("failed to get friend request: %w", err)
	}

	return &request, nil
}

// AddFriend 添加好友关系
func (g *GormFriendDatabase) AddFriend(ctx context.Context, userID, friendID int64) error {
	// 开始事务
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 生成ID
	friend1ID, err := g.sf.NextID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate friend1 ID: %w", err)
	}

	friend2ID, err := g.sf.NextID()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate friend2 ID: %w", err)
	}

	// 创建双向好友关系
	friend1 := &models.Friend{
		ID:        friend1ID,
		UserID:    userID,
		FriendID:  friendID,
		CreatedAt: time.Now(),
	}

	friend2 := &models.Friend{
		ID:        friend2ID,
		UserID:    friendID,
		FriendID:  userID,
		CreatedAt: time.Now(),
	}

	if err := tx.Create(friend1).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create friend relationship 1: %w", err)
	}

	if err := tx.Create(friend2).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create friend relationship 2: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveFriend 移除好友关系
func (g *GormFriendDatabase) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	// 开始事务
	tx := g.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除双向好友关系
	if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&models.Friend{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove friend relationship 1: %w", err)
	}

	if err := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).Delete(&models.Friend{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove friend relationship 2: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteFriendRequest 删除好友请求
func (g *GormFriendDatabase) DeleteFriendRequest(ctx context.Context, requestID int64) error {
	err := g.db.WithContext(ctx).Where("id = ?", requestID).Delete(&models.FriendRequest{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete friend request: %w", err)
	}

	return nil
}
