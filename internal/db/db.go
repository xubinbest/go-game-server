package db

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db/gorm"
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/db/mongodb"
	"github.xubinbest.com/go-game-server/internal/pb"
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

// FriendDatabase 接口方法实现

func (c *DatabaseClient) GetFriends(ctx context.Context, userID int64) ([]*models.Friend, error) {
	return c.friendDB.GetFriends(ctx, userID)
}

func (c *DatabaseClient) GetFriend(ctx context.Context, userID, friendID int64) (*models.Friend, error) {
	return c.friendDB.GetFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) error {
	return c.friendDB.CreateFriendRequest(ctx, fromUserID, toUserID)
}

func (c *DatabaseClient) GetFriendRequests(ctx context.Context, userID int64) ([]*models.FriendRequest, error) {
	return c.friendDB.GetFriendRequests(ctx, userID)
}

func (c *DatabaseClient) GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	return c.friendDB.GetFriendRequest(ctx, requestID)
}

func (c *DatabaseClient) AddFriend(ctx context.Context, userID, friendID int64) error {
	return c.friendDB.AddFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	return c.friendDB.RemoveFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) DeleteFriendRequest(ctx context.Context, requestID int64) error {
	return c.friendDB.DeleteFriendRequest(ctx, requestID)
}

// GuildDatabase 接口方法实现

func (c *DatabaseClient) CreateGuild(ctx context.Context, guild *models.Guild) error {
	return c.guildDB.CreateGuild(ctx, guild)
}

func (c *DatabaseClient) CreateGuildWithMaster(ctx context.Context, guild *models.Guild, master *models.GuildMember) error {
	return c.guildDB.CreateGuildWithMaster(ctx, guild, master)
}

func (c *DatabaseClient) GetGuildByName(ctx context.Context, name string) (*models.Guild, error) {
	return c.guildDB.GetGuildByName(ctx, name)
}

func (c *DatabaseClient) GetGuild(ctx context.Context, guildID int64) (*models.Guild, error) {
	return c.guildDB.GetGuild(ctx, guildID)
}

func (c *DatabaseClient) AddGuildMember(ctx context.Context, member *models.GuildMember) error {
	return c.guildDB.AddGuildMember(ctx, member)
}

func (c *DatabaseClient) UpdateGuild(ctx context.Context, guild *models.Guild) error {
	return c.guildDB.UpdateGuild(ctx, guild)
}

func (c *DatabaseClient) DeleteGuild(ctx context.Context, guildID int64) error {
	return c.guildDB.DeleteGuild(ctx, guildID)
}

func (c *DatabaseClient) GetGuildMember(ctx context.Context, guildID, userID int64) (*models.GuildMember, error) {
	return c.guildDB.GetGuildMember(ctx, guildID, userID)
}

func (c *DatabaseClient) GetGuildMembers(ctx context.Context, guildID int64) ([]*models.GuildMember, error) {
	return c.guildDB.GetGuildMembers(ctx, guildID)
}

func (c *DatabaseClient) UpdateGuildMemberRole(ctx context.Context, guildID, userID int64, newRole int) error {
	return c.guildDB.UpdateGuildMemberRole(ctx, guildID, userID, newRole)
}

func (c *DatabaseClient) RemoveGuildMember(ctx context.Context, guildID, userID int64) error {
	return c.guildDB.RemoveGuildMember(ctx, guildID, userID)
}

func (c *DatabaseClient) CreateGuildApplication(ctx context.Context, application *models.GuildApplication) error {
	return c.guildDB.CreateGuildApplication(ctx, application)
}

func (c *DatabaseClient) GetGuildApplication(ctx context.Context, appID int64) (*models.GuildApplication, error) {
	return c.guildDB.GetGuildApplication(ctx, appID)
}

func (c *DatabaseClient) GetGuildApplications(ctx context.Context, guildID int64) ([]*models.GuildApplication, error) {
	return c.guildDB.GetGuildApplications(ctx, guildID)
}

func (c *DatabaseClient) DeleteGuildApplication(ctx context.Context, appID int64) error {
	return c.guildDB.DeleteGuildApplication(ctx, appID)
}

func (c *DatabaseClient) CreateGuildInvitation(ctx context.Context, invitation *models.GuildInvitation) error {
	return c.guildDB.CreateGuildInvitation(ctx, invitation)
}

func (c *DatabaseClient) GetGuildInvitations(ctx context.Context, guildID int64) ([]*models.GuildInvitation, error) {
	return c.guildDB.GetGuildInvitations(ctx, guildID)
}

func (c *DatabaseClient) GetUserPendingInvitations(ctx context.Context, userID int64) ([]*models.GuildInvitation, error) {
	return c.guildDB.GetUserPendingInvitations(ctx, userID)
}

func (c *DatabaseClient) GetUserGuilds(ctx context.Context, userID int64) ([]*models.Guild, error) {
	return c.guildDB.GetUserGuilds(ctx, userID)
}

func (c *DatabaseClient) GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error) {
	return c.guildDB.GetGuildMemberCount(ctx, guildID)
}

func (c *DatabaseClient) GetGuildList(ctx context.Context, page, pageSize int32) ([]*models.Guild, int32, error) {
	return c.guildDB.GetGuildList(ctx, page, pageSize)
}

// InventoryDatabase 接口方法实现

func (c *DatabaseClient) GetInventory(ctx context.Context, userID int64) (*models.Inventory, error) {
	return c.inventoryDB.GetInventory(ctx, userID)
}

func (c *DatabaseClient) AddItemByTemplate(ctx context.Context, userID int64, templateID int64, count int32) error {
	return c.inventoryDB.AddItemByTemplate(ctx, userID, templateID, count)
}

func (c *DatabaseClient) AddItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	return c.inventoryDB.AddItem(ctx, userID, itemID, count)
}

func (c *DatabaseClient) RemoveItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	return c.inventoryDB.RemoveItem(ctx, userID, itemID, count)
}

func (c *DatabaseClient) UpdateItemCount(ctx context.Context, userID int64, itemID int64, newCount int32) error {
	return c.inventoryDB.UpdateItemCount(ctx, userID, itemID, newCount)
}

func (c *DatabaseClient) HasEnoughItems(ctx context.Context, userID int64, itemID int64, requiredCount int32) (bool, error) {
	return c.inventoryDB.HasEnoughItems(ctx, userID, itemID, requiredCount)
}

func (c *DatabaseClient) GetEquipments(ctx context.Context, userID int64) ([]*models.Equipment, error) {
	return c.inventoryDB.GetEquipments(ctx, userID)
}

func (c *DatabaseClient) EquipItem(ctx context.Context, userID int64, itemID int64, slot int32) error {
	return c.inventoryDB.EquipItem(ctx, userID, itemID, slot)
}

func (c *DatabaseClient) UnequipItem(ctx context.Context, userID int64, slot int32) error {
	return c.inventoryDB.UnequipItem(ctx, userID, slot)
}

func (c *DatabaseClient) GetEquipmentBySlot(ctx context.Context, userID int64, slot int32) (*models.Equipment, error) {
	return c.inventoryDB.GetEquipmentBySlot(ctx, userID, slot)
}

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

// PetDatabase 接口方法实现

func (c *DatabaseClient) GetPet(ctx context.Context, petID int64) (*models.Pet, error) {
	return c.petDB.GetPet(ctx, petID)
}

func (c *DatabaseClient) GetUserPets(ctx context.Context, userID int64) ([]*models.Pet, error) {
	return c.petDB.GetUserPets(ctx, userID)
}

func (c *DatabaseClient) GetUserBattlePet(ctx context.Context, userID int64) (*models.Pet, error) {
	return c.petDB.GetUserBattlePet(ctx, userID)
}

func (c *DatabaseClient) CreatePet(ctx context.Context, pet *models.Pet) error {
	return c.petDB.CreatePet(ctx, pet)
}

func (c *DatabaseClient) UpdatePet(ctx context.Context, pet *models.Pet) error {
	return c.petDB.UpdatePet(ctx, pet)
}

func (c *DatabaseClient) DeletePet(ctx context.Context, petID int64) error {
	return c.petDB.DeletePet(ctx, petID)
}

func (c *DatabaseClient) SetPetBattleStatus(ctx context.Context, userID int64, petID int64, isBattle bool) error {
	return c.petDB.SetPetBattleStatus(ctx, userID, petID, isBattle)
}

func (c *DatabaseClient) CancelAllPetBattleStatus(ctx context.Context, userID int64) error {
	return c.petDB.CancelAllPetBattleStatus(ctx, userID)
}

// ChatDatabase 接口方法实现

func (c *DatabaseClient) SaveChatMessage(ctx context.Context, message *pb.ChatMessage) error {
	return c.chatDB.SaveChatMessage(ctx, message)
}

func (c *DatabaseClient) GetChatMessages(ctx context.Context, channel int32, target_id int64, page, pageSize int32) ([]*pb.ChatMessage, int32, error) {
	return c.chatDB.GetChatMessages(ctx, channel, target_id, page, pageSize)
}
