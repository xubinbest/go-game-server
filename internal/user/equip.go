package user

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// GetEquipments 获取用户装备信息（带缓存）
func (h *Handler) GetEquipments(ctx context.Context, req *pb.GetEquipmentsRequest) (*pb.GetEquipmentsResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 使用缓存服务获取装备数据
	equipments, err := h.cacheService.GetEquipmentsWithCache(ctx, userID, func() ([]*models.Equipment, error) {
		return h.dbClient.GetEquipments(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get equipments: %w", err)
	}

	result := make([]*pb.Equipment, 0, len(equipments))
	for _, e := range equipments {
		// 从内存配置中获取模板数据
		template, err := h.getEquipmentTemplate(e.TemplateID)
		if err != nil {
			// 如果模板不存在，跳过这个装备
			continue
		}

		result = append(result, &pb.Equipment{
			Id:         e.ID,
			TemplateId: e.TemplateID,
			Slot:       e.Slot,
			Name:       template.Name,
			Properties: fmt.Sprintf(`{"atk":%d,"def":%d,"hpMax":%d}`, template.Attribute.Atk, template.Attribute.Def, template.Attribute.HpMax),
		})
	}

	return &pb.GetEquipmentsResponse{
		Equipments: result,
	}, nil
}

// EquipItem 装备物品（带缓存失效）
func (h *Handler) EquipItem(ctx context.Context, req *pb.EquipItemRequest) (*pb.EquipItemResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 检查物品是否存在且属于该用户
	if err := h.dbClient.EquipItem(ctx, userID, req.ItemId, req.Slot); err != nil {
		return &pb.EquipItemResponse{
			Success: false,
			Message: fmt.Sprintf("failed to equip item: %v", err),
		}, nil
	}

	// 失效装备缓存
	err := h.cacheService.InvalidateUserEquipmentsCache(ctx, userID)
	if err != nil {
		// 记录错误但不影响主流程
		fmt.Printf("Failed to invalidate equipments cache: %v", err)
	}

	return &pb.EquipItemResponse{
		Success: true,
		Message: "Item equipped successfully",
	}, nil
}

// UnequipItem 卸下装备（带缓存失效）
func (h *Handler) UnequipItem(ctx context.Context, req *pb.UnequipItemRequest) (*pb.UnequipItemResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	if err := h.dbClient.UnequipItem(ctx, userID, req.Slot); err != nil {
		return &pb.UnequipItemResponse{
			Success: false,
			Message: fmt.Sprintf("failed to unequip item: %v", err),
		}, nil
	}

	// 失效装备缓存
	err := h.cacheService.InvalidateUserEquipmentsCache(ctx, userID)
	if err != nil {
		// 记录错误但不影响主流程
		fmt.Printf("Failed to invalidate equipments cache: %v", err)
	}

	return &pb.UnequipItemResponse{
		Success: true,
		Message: "Item unequipped successfully",
	}, nil
}
