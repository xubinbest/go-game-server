package game_service

import (
	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
)

type Handler struct {
	cache cache.Cache
	cfg   *config.Config
}

func NewHandler(cache cache.Cache, cfg *config.Config) *Handler {
	return &Handler{
		cache: cache,
		cfg:   cfg,
	}
}
