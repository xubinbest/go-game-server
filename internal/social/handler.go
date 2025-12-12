package social

import (
	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/common"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
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
