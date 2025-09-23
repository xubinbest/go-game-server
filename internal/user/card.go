package user

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// getCardTemplate 从内存配置中获取卡牌模板
func (h *Handler) getCardTemplate(templateID int64) (*designconfig.CardData, error) {
	cards := h.configManager.GetConfig("card")
	if cards == nil {
		return nil, fmt.Errorf("card config not found")
	}

	cardsSlice := reflect.ValueOf(cards)
	for i := 0; i < cardsSlice.Len(); i++ {
		card := cardsSlice.Index(i).Interface().(designconfig.CardData)
		if int64(card.ID) == templateID {
			return &card, nil
		}
	}
	return nil, fmt.Errorf("card template not found: %d", templateID)
}

// getCardStarTemplate 从内存配置中获取卡牌星级模板
func (h *Handler) getCardStarTemplate(cardID int64, star int32) (*designconfig.CardStarData, error) {
	cardStars := h.configManager.GetConfig("card_star")
	if cardStars == nil {
		return nil, fmt.Errorf("card star config not found")
	}

	cardStarsSlice := reflect.ValueOf(cardStars)
	for i := 0; i < cardStarsSlice.Len(); i++ {
		cardStar := cardStarsSlice.Index(i).Interface().(designconfig.CardStarData)
		if int64(cardStar.CardId) == cardID && int32(cardStar.Star) == star {
			return &cardStar, nil
		}
	}
	return nil, fmt.Errorf("card star template not found: card_id=%d, star=%d", cardID, star)
}

// getCardLevelTemplate 从内存配置中获取卡牌等级模板
func (h *Handler) getCardLevelTemplate(cardID int64, level int32) (*designconfig.CardLevelData, error) {
	cardLevels := h.configManager.GetConfig("card_level")
	if cardLevels == nil {
		return nil, fmt.Errorf("card level config not found")
	}

	cardLevelsSlice := reflect.ValueOf(cardLevels)
	for i := 0; i < cardLevelsSlice.Len(); i++ {
		cardLevel := cardLevelsSlice.Index(i).Interface().(designconfig.CardLevelData)
		if int64(cardLevel.CardId) == cardID && int32(cardLevel.Level) == level {
			return &cardLevel, nil
		}
	}
	return nil, fmt.Errorf("card level template not found: card_id=%d, level=%d", cardID, level)
}

// checkAndConsumeItems 检查并消耗物品
func (h *Handler) checkAndConsumeItems(ctx context.Context, userID int64, costs []designconfig.BaseItemCost) error {
	for _, cost := range costs {
		// 检查物品是否足够
		hasEnough, err := h.dbClient.HasEnoughItems(ctx, userID, int64(cost.ItemId), int32(cost.Count))
		if err != nil {
			return fmt.Errorf("failed to check item count: %w", err)
		}
		if !hasEnough {
			return fmt.Errorf("insufficient items: item_id=%d, required=%d", cost.ItemId, cost.Count)
		}

		// 消耗物品
		err = h.dbClient.RemoveItem(ctx, userID, int64(cost.ItemId), int32(cost.Count))
		if err != nil {
			return fmt.Errorf("failed to consume item: %w", err)
		}
	}
	return nil
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

// ActivateCard 激活卡牌
func (h *Handler) ActivateCard(ctx context.Context, req *pb.ActivateCardRequest) (*pb.ActivateCardResponse, error) {
	userID := req.UserId
	templateID := req.TemplateId

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if templateID == 0 {
		return nil, fmt.Errorf("invalid template id")
	}

	// 检查是否已经有该模板的卡牌
	exists, err := h.dbClient.CardExists(ctx, userID, templateID)
	if err != nil {
		utils.Error("CardExists error", zap.Error(err))
		return nil, fmt.Errorf("failed to check card exists")
	}
	if exists {
		return &pb.ActivateCardResponse{Success: false, Message: "卡牌已经激活"}, nil
	}

	// 获取卡牌模板
	template, err := h.getCardTemplate(templateID)
	if err != nil {
		utils.Error("Failed to get card template", zap.Error(err))
		return &pb.ActivateCardResponse{Success: false, Message: "卡牌模板不存在"}, nil
	}

	// 检查并消耗激活所需物品
	err = h.checkAndConsumeItems(ctx, userID, template.Cost)
	if err != nil {
		utils.Error("Failed to consume activation items", zap.Error(err))
		return &pb.ActivateCardResponse{Success: false, Message: fmt.Sprintf("激活失败: %v", err)}, nil
	}

	// 创建卡牌数据（激活）
	card := &models.Card{
		UserID:     userID,
		TemplateID: templateID,
		Level:      1,
		Star:       0,
	}
	err = h.dbClient.CreateCard(ctx, card)
	if err != nil {
		utils.Error("CreateCard error", zap.Error(err))
		return nil, fmt.Errorf("failed to create card")
	}

	// 失效卡牌缓存
	err = h.cacheService.InvalidateUserCardsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate cards cache", zap.Error(err))
	}

	return &pb.ActivateCardResponse{Success: true, Message: "卡牌激活成功"}, nil
}

// UpgradeCard 卡牌升级
func (h *Handler) UpgradeCard(ctx context.Context, req *pb.UpgradeCardRequest) (*pb.UpgradeCardResponse, error) {
	userID := req.UserId
	cardID := req.CardId

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if cardID == 0 {
		return nil, fmt.Errorf("invalid card id")
	}

	// 检查卡牌是否存在并获取卡牌信息
	card, err := h.dbClient.GetUserCard(ctx, userID, cardID)
	if err != nil {
		utils.Error("GetUserCard error", zap.Error(err))
		return &pb.UpgradeCardResponse{Success: false, Message: "卡牌不存在"}, nil
	}
	if card == nil {
		return &pb.UpgradeCardResponse{Success: false, Message: "卡牌不存在"}, nil
	}

	// 获取升级模板
	levelTemplate, err := h.getCardLevelTemplate(card.TemplateID, card.Level)
	if err != nil {
		utils.Error("Failed to get card level template", zap.Error(err))
		return &pb.UpgradeCardResponse{Success: false, Message: "升级模板不存在"}, nil
	}

	// 检查并消耗升级所需物品
	err = h.checkAndConsumeItems(ctx, userID, levelTemplate.Cost)
	if err != nil {
		utils.Error("Failed to consume upgrade items", zap.Error(err))
		return &pb.UpgradeCardResponse{Success: false, Message: fmt.Sprintf("升级失败: %v", err)}, nil
	}

	// 升级卡牌
	card.Level++
	err = h.dbClient.UpdateCard(ctx, card)
	if err != nil {
		utils.Error("UpdateCard error", zap.Error(err))
		return nil, fmt.Errorf("failed to update card")
	}

	// 失效卡牌缓存
	err = h.cacheService.InvalidateUserCardsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate cards cache", zap.Error(err))
	}

	return &pb.UpgradeCardResponse{Success: true, Message: "卡牌升级成功"}, nil
}

// UpgradeCardStar 卡牌升星
func (h *Handler) UpgradeCardStar(ctx context.Context, req *pb.UpgradeCardStarRequest) (*pb.UpgradeCardStarResponse, error) {
	userID := req.UserId
	cardID := req.CardId

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if cardID == 0 {
		return nil, fmt.Errorf("invalid card id")
	}

	// 检查卡牌是否存在并获取卡牌信息
	card, err := h.dbClient.GetUserCard(ctx, userID, cardID)
	if err != nil {
		utils.Error("GetUserCard error", zap.Error(err))
		return &pb.UpgradeCardStarResponse{Success: false, Message: "卡牌不存在"}, nil
	}
	if card == nil {
		return &pb.UpgradeCardStarResponse{Success: false, Message: "卡牌不存在"}, nil
	}

	// 获取升星模板
	starTemplate, err := h.getCardStarTemplate(card.TemplateID, card.Star)
	if err != nil {
		utils.Error("Failed to get card star template", zap.Error(err))
		return &pb.UpgradeCardStarResponse{Success: false, Message: "升星模板不存在"}, nil
	}

	// 检查并消耗升星所需物品
	err = h.checkAndConsumeItems(ctx, userID, starTemplate.Cost)
	if err != nil {
		utils.Error("Failed to consume star upgrade items", zap.Error(err))
		return &pb.UpgradeCardStarResponse{Success: false, Message: fmt.Sprintf("升星失败: %v", err)}, nil
	}

	// 升星卡牌
	card.Star++
	err = h.dbClient.UpdateCard(ctx, card)
	if err != nil {
		utils.Error("UpdateCard error", zap.Error(err))
		return nil, fmt.Errorf("failed to update card")
	}

	// 失效卡牌缓存
	err = h.cacheService.InvalidateUserCardsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate cards cache", zap.Error(err))
	}

	return &pb.UpgradeCardStarResponse{Success: true, Message: "卡牌升星成功"}, nil
}
