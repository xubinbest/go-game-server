package mongodb

import (
	"github.xubinbest.com/go-game-server/internal/db/interfaces"
	"github.xubinbest.com/go-game-server/internal/db/mongodb/guild"
	"github.xubinbest.com/go-game-server/internal/snowflake"

	"go.mongodb.org/mongo-driver/mongo"
)

// NewMongoDBGuildDatabase 创建 MongoDBGuildDatabase 实例
func NewMongoDBGuildDatabase(client *mongo.Client, database string, sf *snowflake.Snowflake) interfaces.GuildDatabase {
	return guild.NewMongoDBGuildDatabase(client, database, sf)
}
