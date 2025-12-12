package user

import (
	"context"
	"fmt"
	"reflect"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// getItemTemplate 从内存配置中获取物品模板
func (h *Handler) getItemTemplate(templateID int64) (*designconfig.ItemData, error) {
	items := h.configManager.GetConfig("item")
	if items == nil {
		return nil, fmt.Errorf("item config not found")
	}

	itemsSlice := reflect.ValueOf(items)
	for i := 0; i < itemsSlice.Len(); i++ {
		item := itemsSlice.Index(i).Interface().(designconfig.ItemData)
		if int64(item.ID) == templateID {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("item template not found: %d", templateID)
}

// getEquipmentTemplate 从内存配置中获取装备模板
func (h *Handler) getEquipmentTemplate(templateID int64) (*designconfig.EquipmentData, error) {
	equipments := h.configManager.GetConfig("equip")
	if equipments == nil {
		return nil, fmt.Errorf("equipment config not found")
	}

	equipmentsSlice := reflect.ValueOf(equipments)
	for i := 0; i < equipmentsSlice.Len(); i++ {
		equipment := equipmentsSlice.Index(i).Interface().(designconfig.EquipmentData)
		if int64(equipment.ID) == templateID {
			return &equipment, nil
		}
	}
	return nil, fmt.Errorf("equipment template not found: %d", templateID)
}

// GetInventory 获取背包信息（带缓存）
func (h *Handler) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 使用缓存服务获取背包数据
	inventory, err := h.cacheService.GetInventoryWithCache(ctx, userID, func() (*models.Inventory, error) {
		return h.dbClient.GetInventory(ctx, userID)
	})
	if err != nil {
		utils.Error("GetInventory error", zap.Error(err))
		return nil, fmt.Errorf("failed to get inventory")
	}

	pbItems := make([]*pb.Item, 0, len(inventory.Items))
	for _, item := range inventory.Items {
		// 从内存配置中获取模板数据
		template, err := h.getItemTemplate(item.TemplateID)
		if err != nil {
			utils.Error("Failed to get item template", zap.Int64("template_id", item.TemplateID), zap.Error(err))
			continue
		}

		pbItems = append(pbItems, &pb.Item{
			ItemId:     item.ID,
			TemplateId: item.TemplateID,
			Name:       template.Name,
			Count:      item.Count,
			Type:       int32(template.Type),
			SubType:    int32(template.Subtype),
			Color:      int32(template.Color),
			Stack:      int32(template.Stack),
			Equipped:   item.Equipped,
		})
	}

	return &pb.GetInventoryResponse{
		Inventory: &pb.Inventory{
			Items:    pbItems,
			Capacity: inventory.Capacity,
		},
	}, nil
}

// AddItem 添加物品（带缓存失效）
func (h *Handler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	tplID := req.TplId
	count := req.Count

	// 优先使用模板ID添加物品
	if tplID > 0 {
		// 验证模板是否存在
		_, err := h.getItemTemplate(tplID)
		if err != nil {
			return nil, fmt.Errorf("invalid template id: %v", err)
		}

		err = h.dbClient.AddItemByTemplate(ctx, userID, tplID, count)
		if err != nil {
			utils.Error("AddItemByTemplate error", zap.Error(err))
			return nil, fmt.Errorf("failed to add item by template: %v", err)
		}

		// 失效背包缓存
		err = h.cacheService.InvalidateInventoryCache(ctx, userID)
		if err != nil {
			utils.Error("Failed to invalidate inventory cache", zap.Error(err))
		}
	} else {
		return &pb.AddItemResponse{Success: false}, fmt.Errorf("template_id must be provided")
	}

	return &pb.AddItemResponse{Success: true}, nil
}

// RemoveItem 移除物品（带缓存失效）
func (h *Handler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	itemID := req.ItemId
	count := req.Count
	if count <= 0 {
		return nil, fmt.Errorf("invalid count")
	}

	err := h.dbClient.RemoveItem(ctx, userID, itemID, count)
	if err != nil {
		utils.Error("RemoveItem error", zap.Error(err))
		return nil, fmt.Errorf("failed to remove item: %v", err)
	}

	// 失效背包缓存
	err = h.cacheService.InvalidateInventoryCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate inventory cache", zap.Error(err))
	}

	return &pb.RemoveItemResponse{Success: true}, nil
}

// UseItem 使用物品（带缓存失效）
func (h *Handler) UseItem(ctx context.Context, req *pb.UseItemRequest) (*pb.UseItemResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	itemID := req.ItemId
	count := req.Count
	if count <= 0 {
		return nil, fmt.Errorf("invalid count")
	}

	err := h.dbClient.RemoveItem(ctx, userID, itemID, count)
	if err != nil {
		utils.Error("UseItem error", zap.Error(err))
		return nil, fmt.Errorf("failed to use item: %v", err)
	}

	// 失效背包缓存
	err = h.cacheService.InvalidateInventoryCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate inventory cache", zap.Error(err))
	}

	return &pb.UseItemResponse{Success: true}, nil
}
