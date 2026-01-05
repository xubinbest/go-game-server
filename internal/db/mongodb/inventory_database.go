package mongodb

import (
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/mongodb/inventory"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/mongo"
)

// NewMongoDBInventoryDatabase 创建 MongoDBInventoryDatabase 实例
func NewMongoDBInventoryDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.InventoryDatabase {
	return inventory.NewMongoDBInventoryDatabase(client, database, sf)
}
