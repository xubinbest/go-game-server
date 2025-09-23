package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// CardDatabase 定义卡牌相关的数据库操作接口
type CardDatabase interface {
	// 获取用户所有卡牌
	GetUserCards(ctx context.Context, userID int64) ([]*models.Card, error)

	// 获取用户指定卡牌
	GetUserCard(ctx context.Context, userID int64, cardID int64) (*models.Card, error)

	// 创建卡牌
	CreateCard(ctx context.Context, card *models.Card) error

	// 更新卡牌信息
	UpdateCard(ctx context.Context, card *models.Card) error

	// 升级卡牌
	UpgradeCard(ctx context.Context, userID int64, cardID int64, newLevel int32) error

	// 升星卡牌
	UpgradeCardStar(ctx context.Context, userID int64, cardID int64, newStar int32) error

	// 检查卡牌是否存在
	CardExists(ctx context.Context, userID int64, templateID int64) (bool, error)
}
