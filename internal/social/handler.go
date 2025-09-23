package social

import (
	"github.xubinbest.com/go-game-server/internal/cache"
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

func NewHandler(dbClient db.Database, cacheClient cache.Cache, cacheManager *cache.CacheManager, sf *snowflake.Snowflake, cfg *config.Config, configManager *designconfig.DesignConfigManager) *Handler {
	if dbClient == nil {
		panic("dbClient cannot be nil")
	}
	if cacheClient == nil {
		panic("cacheClient cannot be nil")
	}
	if cacheManager == nil {
		panic("cacheManager cannot be nil")
	}
	if cfg == nil {
		panic("cfg cannot be nil")
	}
	if configManager == nil {
		panic("configManager cannot be nil")
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
	}
}
