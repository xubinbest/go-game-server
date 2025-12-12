package gorm

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GormDatabaseClient GORM数据库客户端
type GormDatabaseClient struct {
	cfg *config.Config
	sf  *snowflake.Snowflake
	db  *gorm.DB
}

// NewGormDatabaseClient 创建GORM数据库客户端
func NewGormDatabaseClient(sf *snowflake.Snowflake, cfg *config.Config) (*GormDatabaseClient, error) {
	client := &GormDatabaseClient{
		cfg: cfg,
		sf:  sf,
	}

	if err := client.initGORM(); err != nil {
		return nil, err
	}

	return client, nil
}

// initGORM 初始化GORM连接
func (c *GormDatabaseClient) initGORM() error {
	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch c.cfg.Database.MySQL.LogLevel {
	case "silent":
		logLevel = logger.Silent
	case "error":
		logLevel = logger.Error
	case "warn":
		logLevel = logger.Warn
	case "info":
		logLevel = logger.Info
	default:
		logLevel = logger.Warn
	}

	// 配置GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
		// 禁用外键约束（如果需要的话）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(c.cfg.Database.MySQL.DSN), config)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL with GORM: %w", err)
	}

	// 获取底层sql.DB进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxOpenConns(c.cfg.Database.MySQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(c.cfg.Database.MySQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(c.cfg.Database.MySQL.ConnMaxLifetime)

	c.db = db

	// 自动迁移（可选，建议在生产环境中手动管理）
	if c.cfg.Database.MySQL.AutoMigrate {
		if err := c.autoMigrate(); err != nil {
			return fmt.Errorf("failed to auto migrate: %w", err)
		}
	}

	return nil
}

// autoMigrate 自动迁移数据库表结构
func (c *GormDatabaseClient) autoMigrate() error {
	return c.db.AutoMigrate(
		// 用户相关表
		&models.User{},
		&models.MonthlySign{},
		&models.MonthlySignReward{},

		// 卡牌相关表
		&models.Card{},

		// 背包相关表
		&models.InventoryItem{},
		&models.Equipment{},

		// 宠物相关表
		&models.Pet{},

		// 好友相关表
		&models.Friend{},
		&models.FriendRequest{},

		// 公会相关表
		&models.Guild{},
		&models.GuildMember{},
		&models.GuildApplication{},
		&models.GuildInvitation{},

		// 聊天相关表
		&ChatMessage{},
	)
}

// GetDB 获取GORM数据库实例
func (c *GormDatabaseClient) GetDB() *gorm.DB {
	return c.db
}

// Ping 检查数据库连接
func (c *GormDatabaseClient) Ping(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.PingContext(ctx)
}

// Close 关闭数据库连接
func (c *GormDatabaseClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// HealthCheck 健康检查
func (c *GormDatabaseClient) HealthCheck(ctx context.Context) error {
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// 检查表是否存在
	var count int64
	if err := c.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count users table: %w", err)
	}

	return nil
}
