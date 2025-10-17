package log_service

import (
	"context"
	"encoding/json"

	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/mq"
	"github.xubinbest.com/go-game-server/internal/mq/userlog"
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

	logService := &LogService{
		db: db,
	}

	go logService.startUserBehaviorConsumer(consumer)
	return logService
}

// UserBehaviorMessage 用户行为消息结构
type UserBehaviorMessage struct {
	Type string          `json:"type"` // 日志类型：create, login, logout, item, money
	Data json.RawMessage `json:"data"` // 具体的日志数据
}

func (s *LogService) startUserBehaviorConsumer(consumer *mq.KafkaConsumer) {
	ctx := context.Background()
	err := consumer.ConsumeMessages(ctx, func(msg mq.Message) error {
		var behaviorMsg UserBehaviorMessage
		if err := json.Unmarshal(msg.Value, &behaviorMsg); err != nil {
			utils.Error("Failed to unmarshal user behavior message", zap.Error(err))
			return err
		}

		// 根据日志类型处理不同的日志
		switch behaviorMsg.Type {
		case "create":
			return s.handleUserCreateLog(ctx, behaviorMsg.Data)
		case "login":
			return s.handleUserLoginLog(ctx, behaviorMsg.Data)
		case "logout":
			return s.handleUserLogoutLog(ctx, behaviorMsg.Data)
		case "item":
			return s.handleUserItemLog(ctx, behaviorMsg.Data)
		case "money":
			return s.handleUserMoneyLog(ctx, behaviorMsg.Data)
		default:
			utils.Error("Unknown user behavior log type", zap.String("type", behaviorMsg.Type))
			return nil
		}
	})

	if err != nil {
		utils.Error("Failed to consume user behavior message", zap.Error(err))
	}
}

// handleUserCreateLog 处理用户创建日志
func (s *LogService) handleUserCreateLog(ctx context.Context, data json.RawMessage) error {
	var createLog userlog.UserCreateLog
	if err := json.Unmarshal(data, &createLog); err != nil {
		utils.Error("Failed to unmarshal user create log", zap.Error(err))
		return err
	}

	dbLog := &models.UserCreateLog{
		UserName:     createLog.UserName,
		Time:         createLog.Time,
		CreateIP:     createLog.CreateIP,
		CreateDevice: createLog.CreateDevice,
	}

	return s.db.CreateUserCreateLog(ctx, dbLog)
}

// handleUserLoginLog 处理用户登录日志
func (s *LogService) handleUserLoginLog(ctx context.Context, data json.RawMessage) error {
	var loginLog userlog.UserLoginLog
	if err := json.Unmarshal(data, &loginLog); err != nil {
		utils.Error("Failed to unmarshal user login log", zap.Error(err))
		return err
	}

	dbLog := &models.UserLoginLog{
		UserID:      loginLog.UserID,
		UserName:    loginLog.UserName,
		Time:        loginLog.Time,
		LoginIP:     loginLog.LoginIP,
		LoginDevice: loginLog.LoginDevice,
	}

	return s.db.CreateUserLoginLog(ctx, dbLog)
}

// handleUserLogoutLog 处理用户登出日志
func (s *LogService) handleUserLogoutLog(ctx context.Context, data json.RawMessage) error {
	var logoutLog userlog.UserLogoutLog
	if err := json.Unmarshal(data, &logoutLog); err != nil {
		utils.Error("Failed to unmarshal user logout log", zap.Error(err))
		return err
	}

	dbLog := &models.UserLogoutLog{
		UserID:       logoutLog.UserID,
		UserName:     logoutLog.UserName,
		Time:         logoutLog.Time,
		LogoutIP:     logoutLog.LogoutIP,
		LogoutDevice: logoutLog.LogoutDevice,
	}

	return s.db.CreateUserLogoutLog(ctx, dbLog)
}

// handleUserItemLog 处理用户物品日志
func (s *LogService) handleUserItemLog(ctx context.Context, data json.RawMessage) error {
	var itemLog userlog.UserItemLog
	if err := json.Unmarshal(data, &itemLog); err != nil {
		utils.Error("Failed to unmarshal user item log", zap.Error(err))
		return err
	}

	dbLog := &models.UserItemLog{
		UserID:     itemLog.UserID,
		UserName:   itemLog.UserName,
		ItemID:     itemLog.ItemID,
		ItemAmount: itemLog.ItemAmount,
		Opt:        itemLog.Opt,
		Time:       itemLog.Time,
		ItemIP:     itemLog.ItemIP,
		ItemDevice: itemLog.ItemDevice,
	}

	return s.db.CreateUserItemLog(ctx, dbLog)
}

// handleUserMoneyLog 处理用户货币日志
func (s *LogService) handleUserMoneyLog(ctx context.Context, data json.RawMessage) error {
	var moneyLog userlog.UserMoneyLog
	if err := json.Unmarshal(data, &moneyLog); err != nil {
		utils.Error("Failed to unmarshal user money log", zap.Error(err))
		return err
	}

	dbLog := &models.UserMoneyLog{
		UserID:      moneyLog.UserID,
		UserName:    moneyLog.UserName,
		Money:       moneyLog.Money,
		MoneyType:   moneyLog.MoneyType,
		Opt:         moneyLog.Opt,
		Time:        moneyLog.Time,
		MoneyIP:     moneyLog.MoneyIP,
		MoneyDevice: moneyLog.MoneyDevice,
	}

	return s.db.CreateUserMoneyLog(ctx, dbLog)
}
