package handler

import (
	"fmt"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/snowflake"
)

// Dependencies 统一的依赖容器，避免数据泥团
type Dependencies struct {
	DBClient      db.Database
	CacheClient   cache.Cache
	CacheManager  *cache.CacheManager
	Cfg           *config.Config
	SF            *snowflake.Snowflake
	ConfigManager *designconfig.DesignConfigManager
}

// NewDependencies 创建依赖容器，统一验证和初始化
func NewDependencies(
	dbClient db.Database,
	cacheClient cache.Cache,
	sf *snowflake.Snowflake,
	cfg *config.Config,
	configManager *designconfig.DesignConfigManager,
) (*Dependencies, error) {
	// 统一参数验证
	if dbClient == nil {
		return nil, fmt.Errorf("dbClient cannot be nil")
	}
	if cacheClient == nil {
		return nil, fmt.Errorf("cacheClient cannot be nil")
	}
	if sf == nil {
		return nil, fmt.Errorf("snowflake cannot be nil")
	}
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if configManager == nil {
		return nil, fmt.Errorf("configManager cannot be nil")
	}

	// 创建CacheManager
	cacheManager := cache.NewCacheManager(cacheClient)

	return &Dependencies{
		DBClient:      dbClient,
		CacheClient:   cacheClient,
		CacheManager:  cacheManager,
		Cfg:           cfg,
		SF:            sf,
		ConfigManager: configManager,
	}, nil
}

// BaseHandler 基础Handler，提供通用功能
type BaseHandler struct {
	deps *Dependencies
}

// NewBaseHandler 创建基础Handler
func NewBaseHandler(deps *Dependencies) *BaseHandler {
	return &BaseHandler{
		deps: deps,
	}
}

// GetDependencies 获取依赖容器
func (h *BaseHandler) GetDependencies() *Dependencies {
	return h.deps
}
