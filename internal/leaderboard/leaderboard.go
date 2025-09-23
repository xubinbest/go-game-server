package leaderboard

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/cache"

	"github.com/redis/go-redis/v9"
)

// Leaderboard 排行榜结构体
// cache 由外部注入，便于测试和复用

type Leaderboard struct {
	cache cache.Cache
}

func NewLeaderboard(cache cache.Cache) *Leaderboard {
	return &Leaderboard{cache: cache}
}

// ReportScore 上报分数
func (l *Leaderboard) ReportScore(ctx context.Context, leaderboard, userID string, score int64) error {
	key := fmt.Sprintf("leaderboard:%s", leaderboard)
	return l.cache.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: userID}).Err()
}

// GetLeaderboard 查询排行榜
func (l *Leaderboard) GetLeaderboard(ctx context.Context, leaderboard string, offset, limit int64) ([]redis.Z, error) {
	key := fmt.Sprintf("leaderboard:%s", leaderboard)
	return l.cache.ZRevRangeWithScores(ctx, key, offset, offset+limit-1).Result()
}

// GetRank 查询个人排名（返回排名和分数）
func (l *Leaderboard) GetRank(ctx context.Context, leaderboard, userID string) (rank int64, score int64, err error) {
	key := fmt.Sprintf("leaderboard:%s", leaderboard)
	rank, err = l.cache.ZRevRank(ctx, key, userID).Result()
	if err != nil {
		return
	}
	scoreF, err := l.cache.ZScore(ctx, key, userID).Result()
	if err != nil {
		return
	}
	score = int64(scoreF)
	return
}
