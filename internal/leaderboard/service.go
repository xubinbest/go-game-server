package leaderboard

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

type LeaderboadGRPCService struct {
	pb.UnimplementedLeaderboardServiceServer
	LB *Leaderboard
}

func NewLeaderboadGRPCService(cache cache.Cache, cfg *config.Config) *LeaderboadGRPCService {
	lb := NewLeaderboard(cache)

	// 启动Kafka消费者，异步写入排行榜
	consumer, err := mq.NewKafkaFactory(&cfg.KafkaConfigs).GetConsumer(mq.GameScore)
	if err != nil {
		utils.Fatal("Failed to create kafka consumer", zap.Error(err))
	}

	go startScoreConsumer(consumer, lb)

	return &LeaderboadGRPCService{
		UnimplementedLeaderboardServiceServer: pb.UnimplementedLeaderboardServiceServer{},
		LB:                                    lb,
	}
}

// GetLeaderboard 获取排行榜
func (s *LeaderboadGRPCService) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	entries, err := s.LB.GetLeaderboard(ctx, req.Leaderboard, int64(req.Offset), int64(req.Limit))
	if err != nil {
		return nil, err
	}
	resp := &pb.GetLeaderboardResponse{}
	for i, e := range entries {
		resp.Entries = append(resp.Entries, &pb.LeaderboardEntry{
			UserId: e.Member.(string),
			Score:  int64(e.Score),
			Rank:   int32(req.Offset + int32(i) + 1),
		})
	}
	return resp, nil
}

// GetRank 获取指定用户的排名
func (s *LeaderboadGRPCService) GetRank(ctx context.Context, req *pb.GetRankRequest) (*pb.GetRankResponse, error) {
	rank, score, err := s.LB.GetRank(ctx, req.Leaderboard, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.GetRankResponse{Rank: int32(rank + 1), Score: score}, nil
}

// startScoreConsumer 启动分数消费者
func startScoreConsumer(consumer *mq.KafkaConsumer, lb *Leaderboard) {
	ctx := context.Background()
	err := consumer.ConsumeMessages(ctx, func(msg mq.Message) error {
		var scoreMsg leaderboard.GameScoreMessage
		if err := json.Unmarshal(msg.Value, &scoreMsg); err != nil {
			utils.Error("Failed to unmarshal score message", zap.Error(err))
			return err
		}

		// 更新排行榜
		if err := lb.ReportScore(ctx, "game_progress", scoreMsg.UserID, int64(scoreMsg.Score)); err != nil {
			utils.Error("Failed to update leaderboard", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		utils.Error("Score consumer error", zap.Error(err))
	}
}
