package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/interfaces"
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
func NewMongoDBInventoryDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.InventoryDatabase {
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

// 获取装备集合
func (m *MongoDBInventoryDatabase) equipmentCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("equipments")
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

// RemoveItem 从背包移除物品
func (m *MongoDBInventoryDatabase) RemoveItem(ctx context.Context, userID int64, itemID int64, count int32) error {
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

		if existingItem.Count < count {
			return fmt.Errorf("not enough items: have %d, need %d", existingItem.Count, count)
		}

		// 更新或删除物品
		newCount := existingItem.Count - count
		if newCount > 0 {
			update := bson.M{
				"$set": bson.M{
					"count":      newCount,
					"updated_at": time.Now().Unix(),
				},
			}

			_, err = m.collection().UpdateOne(sc, filter, update)
		} else {
			_, err = m.collection().DeleteOne(sc, filter)
		}

		if err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// UpdateItemCount 更新物品数量
func (m *MongoDBInventoryDatabase) UpdateItemCount(ctx context.Context, userID int64, itemID int64, newCount int32) error {
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

		filter := bson.M{
			"user_id": userID,
			"_id":     itemID,
		}

		if newCount <= 0 {
			// 如果数量为0或负数，则删除物品
			result, err := m.collection().DeleteOne(sc, filter)
			if err != nil {
				return err
			}

			if result.DeletedCount == 0 {
				return errors.New("item not found in inventory")
			}
		} else {
			// 检查物品是否存在
			var exists bool
			err := m.collection().FindOne(sc, filter).Decode(&exists)
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return errors.New("item not found in inventory")
				}
				return err
			}

			// 更新物品数量
			update := bson.M{
				"$set": bson.M{
					"count":      newCount,
					"updated_at": time.Now().Unix(),
				},
			}

			_, err = m.collection().UpdateOne(sc, filter, update)
			if err != nil {
				return err
			}
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// HasEnoughItems 检查物品是否足够
func (m *MongoDBInventoryDatabase) HasEnoughItems(ctx context.Context, userID int64, itemID int64, requiredCount int32) (bool, error) {
	if requiredCount <= 0 {
		return true, nil
	}

	filter := bson.M{
		"user_id": userID,
		"_id":     itemID,
	}

	var item models.InventoryItem
	err := m.collection().FindOne(ctx, filter).Decode(&item)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil // 物品不存在
		}
		return false, err
	}

	return item.Count >= requiredCount, nil
}

// GetEquipments 获取用户所有装备信息
func (m *MongoDBInventoryDatabase) GetEquipments(ctx context.Context, userID int64) ([]*models.Equipment, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := m.equipmentCollection().Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query equipments: %w", err)
	}
	defer cursor.Close(ctx)

	var equipments []*models.Equipment
	for cursor.Next(ctx) {
		var equipment models.Equipment
		if err := cursor.Decode(&equipment); err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipments = append(equipments, &equipment)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("error iterating equipment rows: %w", err)
	}

	return equipments, nil
}

// EquipItem 装备物品
func (m *MongoDBInventoryDatabase) EquipItem(ctx context.Context, userID int64, itemID int64, slot int32) error {
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

		// 1. 检查物品是否存在且属于该用户
		itemFilter := bson.M{
			"user_id": userID,
			"_id":     itemID,
		}

		var item models.InventoryItem
		err := m.collection().FindOne(sc, itemFilter).Decode(&item)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("item not found or not owned by user")
			}
			return fmt.Errorf("failed to check item ownership: %w", err)
		}

		// 2. 卸下当前槽位的装备
		equipmentFilter := bson.M{
			"user_id": userID,
			"slot":    slot,
		}

		_, err = m.equipmentCollection().DeleteOne(sc, equipmentFilter)
		if err != nil {
			return fmt.Errorf("failed to unequip current item: %w", err)
		}

		// 3. 装备新物品
		now := time.Now().Unix()
		equipmentID, err := m.sf.NextID()
		if err != nil {
			return fmt.Errorf("failed to generate equipment ID: %w", err)
		}

		newEquipment := bson.M{
			"_id":         equipmentID,
			"user_id":     userID,
			"template_id": item.TemplateID,
			"slot":        slot,
			"created_at":  now,
			"updated_at":  now,
		}

		_, err = m.equipmentCollection().InsertOne(sc, newEquipment)
		if err != nil {
			return fmt.Errorf("failed to equip new item: %w", err)
		}

		// 4. 更新物品装备状态
		update := bson.M{
			"$set": bson.M{
				"equipped":   true,
				"updated_at": now,
			},
		}

		_, err = m.collection().UpdateOne(sc, itemFilter, update)
		if err != nil {
			return fmt.Errorf("failed to update item equipped status: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// UnequipItem 卸下装备
func (m *MongoDBInventoryDatabase) UnequipItem(ctx context.Context, userID int64, slot int32) error {
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

		// 1. 获取当前槽位的装备
		equipmentFilter := bson.M{
			"user_id": userID,
			"slot":    slot,
		}

		var equipment models.Equipment
		err := m.equipmentCollection().FindOne(sc, equipmentFilter).Decode(&equipment)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return fmt.Errorf("equipment not found in slot: %d", slot)
			}
			return fmt.Errorf("failed to get equipment: %w", err)
		}

		// 2. 删除装备记录
		_, err = m.equipmentCollection().DeleteOne(sc, equipmentFilter)
		if err != nil {
			return fmt.Errorf("failed to delete equipment: %w", err)
		}

		// 3. 更新物品装备状态（如果有对应的背包物品）
		now := time.Now().Unix()
		update := bson.M{
			"$set": bson.M{
				"equipped":   false,
				"updated_at": now,
			},
		}

		itemFilter := bson.M{
			"user_id":     userID,
			"template_id": equipment.TemplateID,
		}

		_, err = m.collection().UpdateMany(sc, itemFilter, update)
		if err != nil {
			return fmt.Errorf("failed to update item equipped status: %w", err)
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// GetEquipmentBySlot 获取指定槽位的装备
func (m *MongoDBInventoryDatabase) GetEquipmentBySlot(ctx context.Context, userID int64, slot int32) (*models.Equipment, error) {
	filter := bson.M{
		"user_id": userID,
		"slot":    slot,
	}

	var equipment models.Equipment
	err := m.equipmentCollection().FindOne(ctx, filter).Decode(&equipment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	return &equipment, nil
}
