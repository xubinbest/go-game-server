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
