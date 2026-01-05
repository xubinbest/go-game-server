package gorm

import (
	"github.xubinbest.com/go-game-server/internal/db/gorm/guild"
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/gorm"
)

// NewGormGuildDatabase 创建GORM公会数据库实例
func NewGormGuildDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.GuildDatabase {
	return guild.NewGormGuildDatabase(db, sf)
}
