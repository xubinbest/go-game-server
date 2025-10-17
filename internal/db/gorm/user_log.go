package gorm

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
)

// 确保 GormDatabaseClient 实现了 UserLogDatabase 接口
var _ interfaces.UserLogDatabase = (*GormDatabaseClient)(nil)

// CreateUserCreateLog 创建用户创建日志
func (c *GormDatabaseClient) CreateUserCreateLog(ctx context.Context, log *models.UserCreateLog) error {
	if log.ID == 0 {
		id, err := c.sf.NextID()
		if err != nil {
			return err
		}
		log.ID = id
	}
	return c.db.WithContext(ctx).Create(log).Error
}

// CreateUserLoginLog 创建用户登录日志
func (c *GormDatabaseClient) CreateUserLoginLog(ctx context.Context, log *models.UserLoginLog) error {
	if log.ID == 0 {
		id, err := c.sf.NextID()
		if err != nil {
			return err
		}
		log.ID = id
	}
	return c.db.WithContext(ctx).Create(log).Error
}

// CreateUserLogoutLog 创建用户登出日志
func (c *GormDatabaseClient) CreateUserLogoutLog(ctx context.Context, log *models.UserLogoutLog) error {
	if log.ID == 0 {
		id, err := c.sf.NextID()
		if err != nil {
			return err
		}
		log.ID = id
	}
	return c.db.WithContext(ctx).Create(log).Error
}

// CreateUserItemLog 创建用户物品日志
func (c *GormDatabaseClient) CreateUserItemLog(ctx context.Context, log *models.UserItemLog) error {
	if log.ID == 0 {
		id, err := c.sf.NextID()
		if err != nil {
			return err
		}
		log.ID = id
	}
	return c.db.WithContext(ctx).Create(log).Error
}

// CreateUserMoneyLog 创建用户货币日志
func (c *GormDatabaseClient) CreateUserMoneyLog(ctx context.Context, log *models.UserMoneyLog) error {
	if log.ID == 0 {
		id, err := c.sf.NextID()
		if err != nil {
			return err
		}
		log.ID = id
	}
	return c.db.WithContext(ctx).Create(log).Error
}

// BatchCreateUserCreateLogs 批量创建用户创建日志
func (c *GormDatabaseClient) BatchCreateUserCreateLogs(ctx context.Context, logs []*models.UserCreateLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 为没有ID的日志生成ID
	for _, log := range logs {
		if log.ID == 0 {
			id, err := c.sf.NextID()
			if err != nil {
				return err
			}
			log.ID = id
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// BatchCreateUserLoginLogs 批量创建用户登录日志
func (c *GormDatabaseClient) BatchCreateUserLoginLogs(ctx context.Context, logs []*models.UserLoginLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 为没有ID的日志生成ID
	for _, log := range logs {
		if log.ID == 0 {
			id, err := c.sf.NextID()
			if err != nil {
				return err
			}
			log.ID = id
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// BatchCreateUserLogoutLogs 批量创建用户登出日志
func (c *GormDatabaseClient) BatchCreateUserLogoutLogs(ctx context.Context, logs []*models.UserLogoutLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 为没有ID的日志生成ID
	for _, log := range logs {
		if log.ID == 0 {
			id, err := c.sf.NextID()
			if err != nil {
				return err
			}
			log.ID = id
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// BatchCreateUserItemLogs 批量创建用户物品日志
func (c *GormDatabaseClient) BatchCreateUserItemLogs(ctx context.Context, logs []*models.UserItemLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 为没有ID的日志生成ID
	for _, log := range logs {
		if log.ID == 0 {
			id, err := c.sf.NextID()
			if err != nil {
				return err
			}
			log.ID = id
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// BatchCreateUserMoneyLogs 批量创建用户货币日志
func (c *GormDatabaseClient) BatchCreateUserMoneyLogs(ctx context.Context, logs []*models.UserMoneyLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 为没有ID的日志生成ID
	for _, log := range logs {
		if log.ID == 0 {
			id, err := c.sf.NextID()
			if err != nil {
				return err
			}
			log.ID = id
		}
	}

	return c.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// GetUserCreateLogs 查询用户创建日志
func (c *GormDatabaseClient) GetUserCreateLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserCreateLog, error) {
	var logs []*models.UserCreateLog
	err := c.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetUserLoginLogs 查询用户登录日志
func (c *GormDatabaseClient) GetUserLoginLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLoginLog, error) {
	var logs []*models.UserLoginLog
	err := c.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetUserLogoutLogs 查询用户登出日志
func (c *GormDatabaseClient) GetUserLogoutLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserLogoutLog, error) {
	var logs []*models.UserLogoutLog
	err := c.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetUserItemLogs 查询用户物品日志
func (c *GormDatabaseClient) GetUserItemLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserItemLog, error) {
	var logs []*models.UserItemLog
	err := c.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

// GetUserMoneyLogs 查询用户货币日志
func (c *GormDatabaseClient) GetUserMoneyLogs(ctx context.Context, userID int64, limit, offset int) ([]*models.UserMoneyLog, error) {
	var logs []*models.UserMoneyLog
	err := c.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
