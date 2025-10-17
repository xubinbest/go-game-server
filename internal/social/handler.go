package social

import (
	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/handler"
	"github.xubinbest.com/go-game-server/internal/snowflake"
)

type Handler struct {
	deps          *handler.Dependencies
	dbClient      db.Database
	cacheClient   cache.Cache
	cacheManager  *cache.CacheManager
	cacheService  *CacheService
	cfg           *config.Config
	sf            *snowflake.Snowflake
	configManager *designconfig.DesignConfigManager
}

func NewHandler(dbClient db.Database, cacheClient cache.Cache, cacheManager *cache.CacheManager, sf *snowflake.Snowflake, cfg *config.Config, configManager *designconfig.DesignConfigManager) *Handler {
	// 使用统一的依赖容器创建和验证
	deps, err := handler.NewDependencies(dbClient, cacheClient, sf, cfg, configManager)
	if err != nil {
		panic(err)
	}

	cacheService := NewCacheService(deps.CacheManager)

	return &Handler{
		deps:          deps,
		dbClient:      deps.DBClient,
		cacheClient:   deps.CacheClient,
		cacheManager:  deps.CacheManager,
		cacheService:  cacheService,
		cfg:           deps.Cfg,
		sf:            deps.SF,
		configManager: deps.ConfigManager,
	}
}
