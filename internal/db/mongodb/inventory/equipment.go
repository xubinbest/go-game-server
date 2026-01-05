package inventory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 获取装备集合
func (m *MongoDBInventoryDatabase) equipmentCollection() *mongo.Collection {
	return m.client.Database(m.database).Collection("equipments")
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
