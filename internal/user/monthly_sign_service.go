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
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
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
		utils.Error("Failed to invalidate monthly sign cache", zap.Int64("userID", userID), zap.Error(err))
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
