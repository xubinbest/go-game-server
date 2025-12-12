package user

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/common"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/snowflake"
)

type Handler struct {
	dbClient      db.Database
	cacheClient   cache.Cache
	cacheManager  *cache.CacheManager
	cacheService  *CacheService
	cfg           *config.Config
	sf            *snowflake.Snowflake
	configManager *designconfig.DesignConfigManager
}

func NewHandler(dbClient db.Database, cacheClient cache.Cache, cacheManager *cache.CacheManager, sf *snowflake.Snowflake, cfg *config.Config, configManager *designconfig.DesignConfigManager) (*Handler, error) {
	if err := common.ValidateHandlerDependencies(dbClient, cacheClient, cacheManager, cfg, configManager); err != nil {
		return nil, err
	}

	cacheService := NewCacheService(cacheManager)

	return &Handler{
		dbClient:      dbClient,
		cacheClient:   cacheClient,
		cacheManager:  cacheManager,
		cacheService:  cacheService,
		cfg:           cfg,
		sf:            sf,
		configManager: configManager,
	}, nil
}

// 月签到相关方法实现
// GetMonthlySignInfo 获取月签到信息
func (h *Handler) GetMonthlySignInfo(ctx context.Context, req *pb.GetMonthlySignInfoRequest) (*pb.GetMonthlySignInfoResponse, error) {
	// 这里需要将db.Database转换为interfaces.UserDatabase
	// 由于架构限制，这里简化处理，实际实现中需要正确的类型转换
	monthlySignService := NewMonthlySignService(h.dbClient, h.configManager, h.cacheClient, h.cacheService)
	info, err := monthlySignService.GetMonthlySignInfo(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetMonthlySignInfoResponse{
		Info: info,
	}, nil
}

// MonthlySign 执行月签到
func (h *Handler) MonthlySign(ctx context.Context, req *pb.MonthlySignRequest) (*pb.MonthlySignResponse, error) {
	// 这里需要将db.Database转换为interfaces.UserDatabase
	// 由于架构限制，这里简化处理，实际实现中需要正确的类型转换
	monthlySignService := NewMonthlySignService(h.dbClient, h.configManager, h.cacheClient, h.cacheService)
	return monthlySignService.MonthlySign(ctx, req.UserId)
}

// ClaimMonthlySignReward 领取月签到累计奖励
func (h *Handler) ClaimMonthlySignReward(ctx context.Context, req *pb.ClaimMonthlySignRewardRequest) (*pb.ClaimMonthlySignRewardResponse, error) {
	// 这里需要将db.Database转换为interfaces.UserDatabase
	// 由于架构限制，这里简化处理，实际实现中需要正确的类型转换
	monthlySignService := NewMonthlySignService(h.dbClient, h.configManager, h.cacheClient, h.cacheService)
	return monthlySignService.ClaimMonthlySignReward(ctx, req.UserId, req.Days)
}
