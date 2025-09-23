package user

// 月签到功能实现
import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// MonthlySignService 月签到服务
type MonthlySignService struct {
	dbClient      interfaces.UserDatabase
	configManager *designconfig.DesignConfigManager
	cache         cache.Cache
	cacheService  *CacheService
}

// NewMonthlySignService 创建月签到服务
func NewMonthlySignService(dbClient interfaces.UserDatabase, configManager *designconfig.DesignConfigManager, cache cache.Cache, cacheService *CacheService) *MonthlySignService {
	return &MonthlySignService{
		dbClient:      dbClient,
		configManager: configManager,
		cache:         cache,
		cacheService:  cacheService,
	}
}

// GetMonthlySignInfo 获取月签到信息
func (s *MonthlySignService) GetMonthlySignInfo(ctx context.Context, userID int64) (*pb.MonthlySignInfo, error) {
	now := time.Now()
	year := int32(now.Year())
	month := int32(now.Month())
	today := int32(now.Day())

	// 使用缓存获取用户月签到记录
	sign, err := s.cacheService.GetMonthlySignWithCache(ctx, userID, func() (*models.MonthlySign, error) {
		return s.dbClient.GetMonthlySign(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly sign: %w", err)
	}

	// 如果没有签到记录，创建新的
	if sign == nil {
		sign = &models.MonthlySign{
			UserID:    userID,
			Year:      year,
			Month:     month,
			SignDays:  0, // 位图初始化为0
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	// 检查今日是否可以签到
	canSignToday := s.canSignToday(sign)

	// 计算累计签到天数
	totalSignDays := CountBits(sign.SignDays)

	// 获取已签到的日期列表（用于返回给客户端）
	signDaysList := GetSetBits(sign.SignDays)

	return &pb.MonthlySignInfo{
		Year:          year,
		Month:         month,
		SignDays:      signDaysList,
		TotalSignDays: totalSignDays,
		CanSignToday:  canSignToday,
		Today:         today,
	}, nil
}

// MonthlySign 执行月签到
func (s *MonthlySignService) MonthlySign(ctx context.Context, userID int64) (*pb.MonthlySignResponse, error) {
	now := time.Now()
	year := int32(now.Year())
	month := int32(now.Month())
	today := int32(now.Day())

	// 使用分布式锁确保并发安全
	lockKey := fmt.Sprintf("monthly_sign:%d:%d:%d", userID, year, month)
	err := s.cache.Lock(ctx, lockKey, 30*time.Second, 5*time.Second)
	if err != nil {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "获取签到锁失败",
		}, err
	}
	defer s.cache.Unlock(ctx, lockKey)

	// 使用缓存获取用户月签到记录
	sign, err := s.cacheService.GetMonthlySignWithCache(ctx, userID, func() (*models.MonthlySign, error) {
		return s.dbClient.GetMonthlySign(ctx, userID)
	})
	if err != nil {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "获取签到信息失败",
		}, err
	}

	// 如果没有签到记录，创建新的
	if sign == nil {
		sign = &models.MonthlySign{
			UserID:       userID,
			Year:         year,
			Month:        month,
			SignDays:     0, // 位图初始化为0
			LastSignTime: time.Time{},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	// 检查是否可以签到
	if !s.canSignToday(sign) {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "今日已签到",
		}, nil
	}

	// 添加今日签到（使用位运算）
	sign.SignDays = SetBit(sign.SignDays, today)
	sign.LastSignTime = now
	sign.UpdatedAt = now

	// 保存签到记录
	err = s.dbClient.CreateOrUpdateMonthlySign(ctx, sign)
	if err != nil {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "保存签到记录失败",
		}, err
	}

	// 失效签到缓存
	err = s.cacheService.InvalidateMonthlySignCache(ctx, userID)
	if err != nil {
		// 缓存失效失败不影响业务逻辑，只记录日志
		fmt.Printf("failed to invalidate monthly sign cache: %v\n", err)
	}

	return s.handleSignSuccess(ctx, userID, today)
}

// handleSignSuccess 处理签到成功后的逻辑
func (s *MonthlySignService) handleSignSuccess(ctx context.Context, userID int64, today int32) (*pb.MonthlySignResponse, error) {

	// 获取签到奖励
	rewards, err := s.getSignRewards(today)
	if err != nil {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "获取签到奖励失败",
		}, err
	}

	// 发放奖励到背包
	err = s.giveRewards(ctx, userID, rewards)
	if err != nil {
		return &pb.MonthlySignResponse{
			Success: false,
			Message: "发放奖励失败",
		}, err
	}

	return &pb.MonthlySignResponse{
		Success: true,
		Message: "签到成功",
		Rewards: rewards,
	}, nil
}

// ClaimMonthlySignReward 领取月签到累计奖励
func (s *MonthlySignService) ClaimMonthlySignReward(ctx context.Context, userID int64, days int32) (*pb.ClaimMonthlySignRewardResponse, error) {
	now := time.Now()
	year := int32(now.Year())
	month := int32(now.Month())

	// 使用分布式锁确保并发安全
	lockKey := fmt.Sprintf("monthly_sign_reward:%d:%d:%d", userID, year, month)
	err := s.cache.Lock(ctx, lockKey, 30*time.Second, 5*time.Second)
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "获取奖励锁失败",
		}, err
	}
	defer s.cache.Unlock(ctx, lockKey)

	// 使用缓存获取用户月签到记录
	sign, err := s.cacheService.GetMonthlySignWithCache(ctx, userID, func() (*models.MonthlySign, error) {
		return s.dbClient.GetMonthlySign(ctx, userID)
	})
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "获取签到信息失败",
		}, err
	}

	if sign == nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "本月未签到",
		}, nil
	}

	// 检查累计签到天数是否足够
	totalSignDays := CountBits(sign.SignDays)
	if totalSignDays < days {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "累计签到天数不足",
		}, nil
	}

	// 使用缓存获取累计奖励记录
	reward, err := s.cacheService.GetMonthlySignRewardWithCache(ctx, userID, func() (*models.MonthlySignReward, error) {
		return s.dbClient.GetMonthlySignReward(ctx, userID)
	})
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "获取奖励记录失败",
		}, err
	}

	// 如果没有奖励记录，创建新的
	if reward == nil {
		reward = &models.MonthlySignReward{
			UserID:     userID,
			Year:       year,
			Month:      month,
			RewardDays: 0, // 位图初始化为0
			CreatedAt:  now,
			UpdatedAt:  now,
		}
	}

	// 检查是否已经领取过该天数的奖励
	if s.hasClaimedReward(reward, days) {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "该奖励已领取",
		}, nil
	}

	// 获取累计奖励
	rewards, err := s.getCumulativeRewards(days)
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "获取累计奖励失败",
		}, err
	}

	// 发放奖励到背包
	err = s.giveRewards(ctx, userID, rewards)
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "发放奖励失败",
		}, err
	}

	// 记录已领取的奖励（使用位运算）
	reward.RewardDays = SetBit(reward.RewardDays, days)
	reward.UpdatedAt = now

	err = s.dbClient.CreateOrUpdateMonthlySignReward(ctx, reward)
	if err != nil {
		return &pb.ClaimMonthlySignRewardResponse{
			Success: false,
			Message: "保存奖励记录失败",
		}, err
	}

	// 失效奖励缓存
	err = s.cacheService.InvalidateMonthlySignRewardCache(ctx, userID)
	if err != nil {
		// 缓存失效失败不影响业务逻辑，只记录日志
		fmt.Printf("failed to invalidate monthly sign reward cache: %v\n", err)
	}

	return &pb.ClaimMonthlySignRewardResponse{
		Success: true,
		Message: "领取奖励成功",
		Rewards: rewards,
	}, nil
}

// canSignToday 检查今日是否可以签到
func (s *MonthlySignService) canSignToday(sign *models.MonthlySign) bool {
	now := time.Now()
	year := int32(now.Year())
	month := int32(now.Month())
	today := int32(now.Day())

	// 如果年份或月份不匹配，则可以签到
	if sign.Year != year || sign.Month != month {
		return true
	}

	// 检查今日是否已经签到（使用位运算）
	return !HasBit(sign.SignDays, today)
}

// hasClaimedReward 检查是否已经领取过指定天数的奖励
func (s *MonthlySignService) hasClaimedReward(reward *models.MonthlySignReward, days int32) bool {
	return HasBit(reward.RewardDays, days)
}

// getSignRewards 获取签到奖励
func (s *MonthlySignService) getSignRewards(day int32) ([]*pb.Item, error) {
	// 从配置表获取签到奖励
	configData := s.configManager.GetConfig("monthly_sign")
	if configData == nil {
		return nil, fmt.Errorf("未找到月签到配置")
	}

	// 配置数据是切片类型，需要遍历查找对应天数的配置
	signDataSlice, ok := configData.([]designconfig.MonthlySignData)
	if !ok {
		return nil, fmt.Errorf("签到配置格式错误")
	}

	var signData *designconfig.MonthlySignData
	for i := range signDataSlice {
		if signDataSlice[i].ID == int(day) {
			signData = &signDataSlice[i]
			break
		}
	}

	if signData == nil {
		return nil, fmt.Errorf("未找到第%d天的签到配置", day)
	}

	// 转换为协议格式
	var rewards []*pb.Item
	for _, reward := range signData.Reward {
		item := &pb.Item{
			ItemId:     0, // 这里需要生成物品实例ID
			TemplateId: int64(reward.ItemId),
			Count:      int32(reward.Count),
			// 其他字段需要从物品模板获取
		}
		rewards = append(rewards, item)
	}

	return rewards, nil
}

// getCumulativeRewards 获取累计奖励
func (s *MonthlySignService) getCumulativeRewards(days int32) ([]*pb.Item, error) {
	// 从配置表获取累计奖励
	configData := s.configManager.GetConfig("monthly_sign_cumulative")
	if configData == nil {
		return nil, fmt.Errorf("未找到月签到累计奖励配置")
	}

	// 配置数据是切片类型，需要遍历查找对应天数的配置
	cumulativeDataSlice, ok := configData.([]designconfig.MonthlySignCumulativeData)
	if !ok {
		return nil, fmt.Errorf("累计奖励配置格式错误")
	}

	var cumulativeData *designconfig.MonthlySignCumulativeData
	for i := range cumulativeDataSlice {
		if cumulativeDataSlice[i].ID == int(days) {
			cumulativeData = &cumulativeDataSlice[i]
			break
		}
	}

	if cumulativeData == nil {
		return nil, fmt.Errorf("未找到%d天累计奖励配置", days)
	}

	// 转换为协议格式
	var rewards []*pb.Item
	for _, reward := range cumulativeData.Reward {
		item := &pb.Item{
			ItemId:     0, // 这里需要生成物品实例ID
			TemplateId: int64(reward.ItemId),
			Count:      int32(reward.Count),
			// 其他字段需要从物品模板获取
		}
		rewards = append(rewards, item)
	}

	return rewards, nil
}

// giveRewards 发放奖励到背包
func (s *MonthlySignService) giveRewards(ctx context.Context, userID int64, rewards []*pb.Item) error {
	// 这里需要调用背包服务来添加物品
	// 由于当前架构限制，这里简化处理
	// 实际实现中应该调用背包服务的AddItem方法

	for _, item := range rewards {
		// 调用背包服务添加物品
		// 这里需要实现具体的背包操作逻辑
		_ = item // 避免未使用变量警告
	}

	return nil
}
