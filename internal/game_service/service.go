package game_service

import (
	"context"
	"encoding/json"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/mq"
	"github.xubinbest.com/go-game-server/internal/mq/leaderboard"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

type GameGRPCService struct {
	pb.UnimplementedGameServiceServer
	handler      *Handler
	kafkaFactory *mq.KafkaFactory
}

func NewGameGRPCService(ctx context.Context, cache cache.Cache, cfg *config.Config) *GameGRPCService {
	kafkaFactory := mq.NewKafkaFactory(&cfg.KafkaConfigs)

	return &GameGRPCService{
		UnimplementedGameServiceServer: pb.UnimplementedGameServiceServer{},
		handler:                        NewHandler(cache, cfg),
		kafkaFactory:                   kafkaFactory,
	}
}

func (s *GameGRPCService) JoinGame(ctx context.Context, req *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	utils.Info("Player joining game", zap.String("playerId", req.PlayerId), zap.String("gameId", req.GameId))
	return &pb.JoinGameResponse{
		Success: true,
		Message: "Welcome to the game!",
	}, nil
}

func (s *GameGRPCService) LeaveGame(ctx context.Context, req *pb.LeaveGameRequest) (*pb.LeaveGameResponse, error) {
	utils.Info("Player leaving game", zap.String("playerId", req.PlayerId), zap.String("gameId", req.GameId))
	return &pb.LeaveGameResponse{Success: true}, nil
}

func (s *GameGRPCService) GetGameState(ctx context.Context, req *pb.GameStateRequest) (*pb.GameStateResponse, error) {
	return &pb.GameStateResponse{
		State:   "running",
		Players: []string{"player1", "player2"},
	}, nil
}

func (s *GameGRPCService) PlayerAction(ctx context.Context, req *pb.PlayerActionRequest) (*pb.PlayerActionResponse, error) {
	utils.Info("Player action", zap.String("playerId", req.PlayerId), zap.String("gameId", req.GameId), zap.String("action", req.Action))

	// 示例：当玩家完成某个动作时，更新他们的分数
	if req.Action == "complete_level" {
		producer, err := s.kafkaFactory.GetProducer(mq.GameScore)
		if err != nil {
			utils.Error("Failed to get score producer", zap.Error(err))
			// 继续处理，不让Kafka错误影响游戏流程
		} else {
			scoreData := leaderboard.GameScoreMessage{
				UserID: req.PlayerId,
				Score:  100.0,
			}

			value, err := json.Marshal(scoreData)
			if err != nil {
				utils.Error("Failed to marshal score data", zap.Error(err))
			} else {
				key := []byte(req.PlayerId)
				if err := producer.SendMessage(ctx, key, value); err != nil {
					utils.Error("Failed to send score to kafka", zap.Error(err))
				}
			}
		}
	}

	return &pb.PlayerActionResponse{
		Success: true,
		Message: "Action processed",
	}, nil
}
