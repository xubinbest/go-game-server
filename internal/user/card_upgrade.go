package user

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

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
