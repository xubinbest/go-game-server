package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// CacheService 缓存服务
type CacheService struct {
	cacheManager *cache.CacheManager
}

// NewCacheService 创建缓存服务
func NewCacheService(cacheManager *cache.CacheManager) *CacheService {
	return &CacheService{
		cacheManager: cacheManager,
	}
}

// GetUserInfoWithCache 带缓存的获取用户信息
func (cs *CacheService) GetUserInfoWithCache(ctx context.Context, userID int64, getFunc func() (*models.User, error)) (*models.User, error) {
	strategy := cache.Strategies["user_info"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	// 如果是从数据库直接返回的数据
	if user, ok := data.(*models.User); ok {
		return user, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var user models.User
		if err := json.Unmarshal(jsonData, &user); err == nil {
			return &user, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal user from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid user data format")
	}

	return nil, fmt.Errorf("invalid user data format")
}

// UpdateUserWithCache 带缓存的更新用户信息
func (cs *CacheService) UpdateUserWithCache(ctx context.Context, user *models.User, updateFunc func() error) error {
	// 1. 更新数据库
	if err := updateFunc(); err != nil {
		return err
	}

	// 2. 失效相关缓存
	strategy := cache.Strategies["user_info"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, user.ID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetInventoryWithCache 带缓存的获取背包
func (cs *CacheService) GetInventoryWithCache(ctx context.Context, userID int64, getFunc func() (*models.Inventory, error)) (*models.Inventory, error) {
	strategy := cache.Strategies["user_inventory"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	// 如果是从数据库直接返回的数据
	if inventory, ok := data.(*models.Inventory); ok {
		return inventory, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var inventory models.Inventory
		if err := json.Unmarshal(jsonData, &inventory); err == nil {
			return &inventory, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal inventory from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid inventory data format")
	}

	return nil, fmt.Errorf("invalid inventory data format")
}

// InvalidateInventoryCache 失效背包缓存
func (cs *CacheService) InvalidateInventoryCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["user_inventory"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetUserCardsWithCache 带缓存的获取用户卡牌
func (cs *CacheService) GetUserCardsWithCache(ctx context.Context, userID int64, getFunc func() ([]*models.Card, error)) ([]*models.Card, error) {
	strategy := cache.Strategies["user_cards"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.Card{}, nil
	}

	// 如果是从数据库直接返回的数据
	if cards, ok := data.([]*models.Card); ok {
		return cards, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var cards []*models.Card
		if err := json.Unmarshal(jsonData, &cards); err == nil {
			return cards, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal cards from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid cards data format")
	}

	return nil, fmt.Errorf("invalid cards data format")
}

// InvalidateUserCardsCache 失效用户卡牌缓存
func (cs *CacheService) InvalidateUserCardsCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["user_cards"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetUserPetsWithCache 带缓存的获取用户宠物
func (cs *CacheService) GetUserPetsWithCache(ctx context.Context, userID int64, getFunc func() ([]*models.Pet, error)) ([]*models.Pet, error) {
	strategy := cache.Strategies["user_pets"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.Pet{}, nil
	}

	// 如果是从数据库直接返回的数据
	if pets, ok := data.([]*models.Pet); ok {
		return pets, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var pets []*models.Pet
		if err := json.Unmarshal(jsonData, &pets); err == nil {
			return pets, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal pets from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid pets data format")
	}

	return nil, fmt.Errorf("invalid pets data format")
}

// InvalidateUserPetsCache 失效用户宠物缓存
func (cs *CacheService) InvalidateUserPetsCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["user_pets"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetEquipmentsWithCache 带缓存的获取用户装备
func (cs *CacheService) GetEquipmentsWithCache(ctx context.Context, userID int64, getFunc func() ([]*models.Equipment, error)) ([]*models.Equipment, error) {
	strategy := cache.Strategies["user_equipments"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.Equipment{}, nil
	}

	// 如果是从数据库直接返回的数据
	if equipments, ok := data.([]*models.Equipment); ok {
		return equipments, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var equipments []*models.Equipment
		if err := json.Unmarshal(jsonData, &equipments); err == nil {
			return equipments, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal equipments from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid equipments data format")
	}

	return nil, fmt.Errorf("invalid equipments data format")
}

// InvalidateUserEquipmentsCache 失效用户装备缓存
func (cs *CacheService) InvalidateUserEquipmentsCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["user_equipments"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetMonthlySignWithCache 带缓存的获取月签到信息
func (cs *CacheService) GetMonthlySignWithCache(ctx context.Context, userID int64, getFunc func() (*models.MonthlySign, error)) (*models.MonthlySign, error) {
	strategy := cache.Strategies["monthly_sign"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	// 如果是从数据库直接返回的数据
	if sign, ok := data.(*models.MonthlySign); ok {
		return sign, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var sign models.MonthlySign
		if err := json.Unmarshal(jsonData, &sign); err == nil {
			return &sign, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal monthly sign from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid monthly sign data format")
	}

	return nil, fmt.Errorf("invalid monthly sign data format")
}

// InvalidateMonthlySignCache 失效月签到缓存
func (cs *CacheService) InvalidateMonthlySignCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["monthly_sign"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetMonthlySignRewardWithCache 带缓存的获取月签到奖励记录
func (cs *CacheService) GetMonthlySignRewardWithCache(ctx context.Context, userID int64, getFunc func() (*models.MonthlySignReward, error)) (*models.MonthlySignReward, error) {
	strategy := cache.Strategies["monthly_sign_reward"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	// 如果是从数据库直接返回的数据
	if reward, ok := data.(*models.MonthlySignReward); ok {
		return reward, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var reward models.MonthlySignReward
		if err := json.Unmarshal(jsonData, &reward); err == nil {
			return &reward, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal monthly sign reward from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid monthly sign reward data format")
	}

	return nil, fmt.Errorf("invalid monthly sign reward data format")
}

// InvalidateMonthlySignRewardCache 失效月签到奖励缓存
func (cs *CacheService) InvalidateMonthlySignRewardCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["monthly_sign_reward"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}
