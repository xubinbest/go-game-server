package db

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db/gorm"
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/mongodb"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database 是所有数据库操作的主接口
type Database interface {
	// 基础操作
	Connect() error
	Close() error
	Ping(ctx context.Context) error

	// 用户相关方法
	interfaces.UserDatabase

	// 好友相关方法
	interfaces.FriendDatabase

	// 聊天消息相关操作
	interfaces.ChatDatabase

	// 帮派相关方法
	interfaces.GuildDatabase

	// 背包相关方法
	interfaces.InventoryDatabase

	// 卡牌相关方法
	interfaces.CardDatabase

	// 宠物相关方法
	interfaces.PetDatabase
}

// DatabaseClient 实现 Database 接口，使用组合模式
type DatabaseClient struct {
	cfg *config.Config
	sf  *snowflake.Snowflake

	// GORM客户端（用于MySQL）
	gormClient *gorm.GormDatabaseClient

	// MongoDB连接（用于MongoDB）
	mongoDB *mongo.Client

	// 各功能模块的接口实现
	userDB      interfaces.UserDatabase
	friendDB    interfaces.FriendDatabase
	chatDB      interfaces.ChatDatabase
	guildDB     interfaces.GuildDatabase
	inventoryDB interfaces.InventoryDatabase
	cardDB      interfaces.CardDatabase
	petDB       interfaces.PetDatabase
}

// NewDatabaseClient 创建数据库客户端实例
func NewDatabaseClient(sf *snowflake.Snowflake, cfg *config.Config) (Database, error) {
	client := &DatabaseClient{
		cfg: cfg,
		sf:  sf,
	}

	// 根据配置选择数据库实现
	if cfg.Database.MySQL.Enabled {
		// 初始化GORM（MySQL）
		if err := client.initGORM(); err != nil {
			return nil, err
		}
	} else if cfg.Database.MongoDB.Enabled {
		// 初始化MongoDB
		if err := client.initMongoDB(); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no database configured")
	}

	return client, nil
}

// initGORM 初始化GORM连接和相关实现
func (c *DatabaseClient) initGORM() error {
	// 初始化GORM客户端
	gormClient, err := gorm.NewGormDatabaseClient(c.sf, c.cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize GORM client: %w", err)
	}
	c.gormClient = gormClient

	// 获取GORM数据库实例
	gormDB := gormClient.GetDB()

	// 初始化各个模块的GORM实现
	c.userDB = gorm.NewGormUserDatabase(gormDB, c.sf)
	c.friendDB = gorm.NewGormFriendDatabase(gormDB, c.sf)
	c.chatDB = gorm.NewGormChatDatabase(gormDB, c.sf)
	c.guildDB = gorm.NewGormGuildDatabase(gormDB, c.sf)
	c.inventoryDB = gorm.NewGormInventoryDatabase(gormDB, c.sf)
	c.cardDB = gorm.NewGormCardDatabase(gormDB, c.sf)
	c.petDB = gorm.NewGormPetDatabase(gormDB, c.sf)

	return nil
}

// initMongoDB 初始化MongoDB连接和相关实现
func (c *DatabaseClient) initMongoDB() error {
	var err error

	// 创建MongoDB客户端
	clientOptions := options.Client().ApplyURI(c.cfg.Database.MongoDB.URI)
	c.mongoDB, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 验证连接
	if err = c.mongoDB.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 初始化各个模块的MongoDB实现
	dbName := c.cfg.Database.MongoDB.Database
	c.userDB = mongodb.NewMongoDBUserDatabase(c.mongoDB, dbName, c.sf)
	c.friendDB = mongodb.NewMongoDBFriendDatabase(c.mongoDB, dbName, c.sf)
	c.chatDB = mongodb.NewMongoDBChatDatabase(c.mongoDB, dbName, c.sf)
	c.guildDB = mongodb.NewMongoDBGuildDatabase(c.mongoDB, dbName, c.sf)
	c.inventoryDB = mongodb.NewMongoDBInventoryDatabase(c.mongoDB, dbName, c.sf)
	c.cardDB = mongodb.NewMongoDBCardDatabase(c.mongoDB, dbName, c.sf)

	return nil
}

// Connect 实现 Database 接口
func (c *DatabaseClient) Connect() error {
	// 连接已在初始化时建立
	return nil
}

// Close 关闭所有数据库连接
func (c *DatabaseClient) Close() error {
	var gormErr, mongoErr error

	if c.gormClient != nil {
		gormErr = c.gormClient.Close()
	}

	if c.mongoDB != nil {
		mongoErr = c.mongoDB.Disconnect(context.Background())
	}

	if gormErr != nil {
		return gormErr
	}
	return mongoErr
}

// Ping 检查数据库连接
func (c *DatabaseClient) Ping(ctx context.Context) error {
	if c.gormClient != nil {
		if err := c.gormClient.Ping(ctx); err != nil {
			return fmt.Errorf("GORM ping failed: %w", err)
		}
	}

	if c.mongoDB != nil {
		if err := c.mongoDB.Ping(ctx, readpref.Primary()); err != nil {
			return fmt.Errorf("MongoDB ping failed: %w", err)
		}
	}

	return nil
}
