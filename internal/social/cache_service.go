package social

import (
	"context"
	"encoding/json"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
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

// GetFriendsWithCache 带缓存的获取好友列表
func (cs *CacheService) GetFriendsWithCache(ctx context.Context, userID int64, getFunc func() ([]*models.Friend, error)) ([]*models.Friend, error) {
	strategy := cache.Strategies["user_friends"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.Friend{}, nil
	}

	// 如果是从数据库直接返回的数据
	if friends, ok := data.([]*models.Friend); ok {
		return friends, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var friends []*models.Friend
		if err := json.Unmarshal(jsonData, &friends); err == nil {
			return friends, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal friends from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid friends data format")
	}

	return nil, fmt.Errorf("invalid friends data format")
}

// InvalidateFriendsCache 失效好友缓存
func (cs *CacheService) InvalidateFriendsCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["user_friends"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetFriendRequestsWithCache 带缓存的获取好友申请列表
func (cs *CacheService) GetFriendRequestsWithCache(ctx context.Context, userID int64, getFunc func() ([]*models.FriendRequest, error)) ([]*models.FriendRequest, error) {
	strategy := cache.Strategies["friend_requests"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.FriendRequest{}, nil
	}

	// 如果是从数据库直接返回的数据
	if requests, ok := data.([]*models.FriendRequest); ok {
		return requests, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var requests []*models.FriendRequest
		if err := json.Unmarshal(jsonData, &requests); err == nil {
			return requests, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal friend requests from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid friend requests data format")
	}

	return nil, fmt.Errorf("invalid friend requests data format")
}

// InvalidateFriendRequestsCache 失效好友申请缓存
func (cs *CacheService) InvalidateFriendRequestsCache(ctx context.Context, userID int64) error {
	strategy := cache.Strategies["friend_requests"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, userID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetGuildWithCache 带缓存的获取公会信息
func (cs *CacheService) GetGuildWithCache(ctx context.Context, guildID int64, getFunc func() (*models.Guild, error)) (*models.Guild, error) {
	strategy := cache.Strategies["guild_info"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)

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
	if guild, ok := data.(*models.Guild); ok {
		return guild, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var guild models.Guild
		if err := json.Unmarshal(jsonData, &guild); err == nil {
			return &guild, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal guild from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid guild data format")
	}

	return nil, fmt.Errorf("invalid guild data format")
}

// InvalidateGuildCache 失效公会缓存
func (cs *CacheService) InvalidateGuildCache(ctx context.Context, guildID int64) error {
	strategy := cache.Strategies["guild_info"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetGuildMembersWithCache 带缓存的获取公会成员
func (cs *CacheService) GetGuildMembersWithCache(ctx context.Context, guildID int64, getFunc func() (*pb.GetGuildMembersResponse, error)) (*pb.GetGuildMembersResponse, error) {
	strategy := cache.Strategies["guild_members"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return &pb.GetGuildMembersResponse{Members: []*pb.GuildMember{}}, nil
	}

	// 如果是从数据库直接返回的数据
	if resp, ok := data.(*pb.GetGuildMembersResponse); ok {
		return resp, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var resp pb.GetGuildMembersResponse
		if err := json.Unmarshal(jsonData, &resp); err == nil {
			return &resp, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal guild members from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid guild members data format")
	}

	return nil, fmt.Errorf("invalid guild members data format")
}

// InvalidateGuildMembersCache 失效公会成员缓存
func (cs *CacheService) InvalidateGuildMembersCache(ctx context.Context, guildID int64) error {
	strategy := cache.Strategies["guild_members"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)
	return cs.cacheManager.Invalidate(ctx, key)
}

// GetGuildApplicationsWithCache 带缓存的获取公会申请列表
func (cs *CacheService) GetGuildApplicationsWithCache(ctx context.Context, guildID int64, getFunc func() ([]*models.GuildApplication, error)) ([]*models.GuildApplication, error) {
	strategy := cache.Strategies["guild_applications"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)

	data, err := cs.cacheManager.GetOrSet(ctx, key, strategy, func() (interface{}, error) {
		return getFunc()
	})

	if err != nil {
		return nil, err
	}

	if data == nil {
		return []*models.GuildApplication{}, nil
	}

	// 如果是从数据库直接返回的数据
	if applications, ok := data.([]*models.GuildApplication); ok {
		return applications, nil
	}

	// 如果是从缓存返回的JSON数据
	if jsonData, ok := data.([]byte); ok {
		var applications []*models.GuildApplication
		if err := json.Unmarshal(jsonData, &applications); err == nil {
			return applications, nil
		}
		// 如果反序列化失败，记录错误并返回nil
		utils.Error("failed to unmarshal guild applications from cache", zap.Error(err))
		return nil, fmt.Errorf("invalid guild applications data format")
	}

	return nil, fmt.Errorf("invalid guild applications data format")
}

// InvalidateGuildApplicationsCache 失效公会申请缓存
func (cs *CacheService) InvalidateGuildApplicationsCache(ctx context.Context, guildID int64) error {
	strategy := cache.Strategies["guild_applications"]
	key := fmt.Sprintf("%s%d", strategy.KeyPrefix, guildID)
	return cs.cacheManager.Invalidate(ctx, key)
}
