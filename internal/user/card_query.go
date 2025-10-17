package user

import (
	"context"
	"encoding/json"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// getCardTemplate 从内存配置中获取卡牌模板 (O(1)查询)
func (h *Handler) getCardTemplate(templateID int64) (*designconfig.CardData, error) {
	return h.configManager.GetCardByID(templateID)
}

// getCardStarTemplate 从内存配置中获取卡牌星级模板
// 注意：此方法需要通过cardID和star组合查找，保留原有的O(n)查询
func (h *Handler) getCardStarTemplate(cardID int64, star int32) (*designconfig.CardStarData, error) {
	cardStars := h.configManager.GetConfig("card_star")
	if cardStars == nil {
		return nil, fmt.Errorf("card star config not found")
	}

	// 类型断言获取切片
	cardStarsSlice, ok := cardStars.([]designconfig.CardStarData)
	if !ok {
		return nil, fmt.Errorf("card star config type assertion failed")
	}

	for i := range cardStarsSlice {
		if int64(cardStarsSlice[i].CardId) == cardID && int32(cardStarsSlice[i].Star) == star {
			return &cardStarsSlice[i], nil
		}
	}
	return nil, fmt.Errorf("card star template not found: card_id=%d, star=%d", cardID, star)
}

// getCardLevelTemplate 从内存配置中获取卡牌等级模板
// 注意：此方法需要通过cardID和level组合查找，保留原有的O(n)查询
func (h *Handler) getCardLevelTemplate(cardID int64, level int32) (*designconfig.CardLevelData, error) {
	cardLevels := h.configManager.GetConfig("card_level")
	if cardLevels == nil {
		return nil, fmt.Errorf("card level config not found")
	}

	// 类型断言获取切片
	cardLevelsSlice, ok := cardLevels.([]designconfig.CardLevelData)
	if !ok {
		return nil, fmt.Errorf("card level config type assertion failed")
	}

	for i := range cardLevelsSlice {
		if int64(cardLevelsSlice[i].CardId) == cardID && int32(cardLevelsSlice[i].Level) == level {
			return &cardLevelsSlice[i], nil
		}
	}
	return nil, fmt.Errorf("card level template not found: card_id=%d, level=%d", cardID, level)
}

// GetUserCards 获取玩家所有卡牌信息（带缓存）
func (h *Handler) GetUserCards(ctx context.Context, req *pb.GetUserCardsRequest) (*pb.GetUserCardsResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 使用缓存服务获取卡牌数据
	cards, err := h.cacheService.GetUserCardsWithCache(ctx, userID, func() ([]*models.Card, error) {
		return h.dbClient.GetUserCards(ctx, userID)
	})
	if err != nil {
		utils.Error("GetUserCards error", zap.Error(err))
		return nil, fmt.Errorf("failed to get user cards")
	}

	pbCards := make([]*pb.Card, 0, len(cards))
	for _, card := range cards {
		// 从内存配置中获取模板数据
		template, err := h.getCardTemplate(card.TemplateID)
		if err != nil {
			utils.Error("Failed to get card template", zap.Int64("template_id", card.TemplateID), zap.Error(err))
			continue
		}

		// 构建卡牌属性JSON
		properties := map[string]interface{}{
			"atk":   template.Attribute.Atk,
			"def":   template.Attribute.Def,
			"hpMax": template.Attribute.HpMax,
		}
		propertiesJSON, _ := json.Marshal(properties)

		pbCards = append(pbCards, &pb.Card{
			Id:         card.ID,
			TemplateId: card.TemplateID,
			Name:       template.Name,
			Level:      card.Level,
			Star:       card.Star,
			Activated:  true, // 有卡牌数据就表示已激活
			Properties: string(propertiesJSON),
		})
	}

	return &pb.GetUserCardsResponse{
		Cards: pbCards,
	}, nil
}
