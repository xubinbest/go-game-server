package gorm

import (
	"github.xubinbest.com/go-game-server/internal/db/gorm/inventory"
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"gorm.io/gorm"
)

// NewGormInventoryDatabase 创建GORM背包数据库实例
func NewGormInventoryDatabase(db *gorm.DB, sf *snowflake.Snowflake) interfaces.InventoryDatabase {
	return inventory.NewGormInventoryDatabase(db, sf)
}
