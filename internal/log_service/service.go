package log_service

import (
	"context"
	"encoding/json"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/mq"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

type LogService struct {
	db *db.DatabaseClient
}

func NewLogService(db *db.DatabaseClient, cfg *config.Config) *LogService {
	consumer, err := mq.NewKafkaFactory(&cfg.KafkaConfigs).GetConsumer(mq.UserBehavior)
	if err != nil {
		utils.Fatal("Failed to create kafka consumer", zap.Error(err))
	}

	go startUserBehaviorConsumer(consumer)
	return &LogService{
		db: db,
	}
}

func startUserBehaviorConsumer(consumer *mq.KafkaConsumer) {
	type UserBehaviorMessage struct {
		UserID   string `json:"user_id"`
		Behavior string `json:"behavior"`
	}

	ctx := context.Background()
	err := consumer.ConsumeMessages(ctx, func(msg mq.Message) error {
		var userBehaviorMsg UserBehaviorMessage
		if err := json.Unmarshal(msg.Value, &userBehaviorMsg); err != nil {
			utils.Error("Failed to unmarshal user behavior message", zap.Error(err))
			return err
		}

		// TODO: 保存用户行为日志

		return nil
	})

	if err != nil {
		utils.Error("Failed to consume user behavior message", zap.Error(err))
	}

}
