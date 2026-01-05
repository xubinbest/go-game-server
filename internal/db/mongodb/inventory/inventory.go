package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDBInventoryDatabase 实现 InventoryDatabase 接口
type MongoDBInventoryDatabase struct {
	client   *mongo.Client
	database string
	sf       *snowflake.Snowflake
}

// NewMongoDBInventoryDatabase 创建 MongoDBInventoryDatabase 实例
func NewMongoDBInventoryDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) *MongoDBInventoryDatabase {
	return &MongoDBInventoryDatabase{
		client:   client,
		database: database,
		sf:       sf,
	}
}

// 获取背包集合
func (m *MongoDBInventoryDatabase) collection() *mongo.Collection {
	return m.client.Database(m.database).Collection("inventory_items")
}

// GetInventory 获取用户背包
func (m *MongoDBInventoryDatabase) GetInventory(ctx context.Context, userID int64) (*models.Inventory, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := m.collection().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	inventory := &models.Inventory{
		UserID: userID,
		Items:  make([]*models.InventoryItem, 0),
	}

	for cursor.Next(ctx) {
		var item models.InventoryItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		inventory.Items = append(inventory.Items, &item)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return inventory, nil
}

// AddItemByTemplate 通过模板ID添加物品到背包
func (m *MongoDBInventoryDatabase) AddItemByTemplate(ctx context.Context, userID int64, templateID int64, count int32) error {
	if count <= 0 {
		return errors.New("count must be positive")
	}

	// 使用事务确保原子性
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 在事务中执行所有操作
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// 检查是否已存在相同模板的物品
		filter := bson.M{
			"user_id":     userID,
			"template_id": templateID,
		}

		var existingItem models.InventoryItem
		err := m.collection().FindOne(sc, filter).Decode(&existingItem)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return err
			}

			// 物品不存在，创建新实例
			now := time.Now().Unix()
			itemID, err := m.sf.NextID()
			if err != nil {
				return fmt.Errorf("failed to generate item ID: %w", err)
			}

			newItem := bson.M{
				"_id":         itemID,
				"user_id":     userID,
				"template_id": templateID,
				"count":       count,
				"equipped":    false,
				"created_at":  now,
				"updated_at":  now,
			}

			_, err = m.collection().InsertOne(sc, newItem)
			if err != nil {
				return fmt.Errorf("failed to insert item: %w", err)
			}
		} else {
			// 更新现有物品数量
			update := bson.M{
				"$inc": bson.M{"count": count},
				"$set": bson.M{"updated_at": time.Now().Unix()},
			}

			_, err = m.collection().UpdateOne(sc, filter, update)
			if err != nil {
				return fmt.Errorf("failed to update item count: %w", err)
			}
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// AddItem 通过实例ID添加物品到背包
func (m *MongoDBInventoryDatabase) AddItem(ctx context.Context, userID int64, itemID int64, count int32) error {
	if count <= 0 {
		return errors.New("count must be positive")
	}

	// 使用事务确保原子性
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 在事务中执行所有操作
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		// 检查物品是否存在且属于该用户
		filter := bson.M{
			"user_id": userID,
			"_id":     itemID,
		}

		var existingItem models.InventoryItem
		err := m.collection().FindOne(sc, filter).Decode(&existingItem)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("item not found: %d", itemID)
			}
			return err
		}

		// 更新物品数量
		update := bson.M{
			"$inc": bson.M{"count": count},
			"$set": bson.M{"updated_at": time.Now().Unix()},
		}

		_, err = m.collection().UpdateOne(sc, filter, update)
		if err != nil {
			return fmt.Errorf("failed to update item count: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	return err
}
