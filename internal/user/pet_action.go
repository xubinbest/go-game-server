package user

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

// SetPetBattleStatus 设置宠物出战状态
func (h *Handler) SetPetBattleStatus(ctx context.Context, req *pb.SetPetBattleStatusRequest) (*pb.SetPetBattleStatusResponse, error) {
	userID := req.UserId
	petID := req.PetId
	isBattle := req.IsBattle

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if petID == 0 {
		return nil, fmt.Errorf("invalid pet id")
	}

	// 检查宠物是否存在并属于该用户
	pet, err := h.dbClient.GetPet(ctx, petID)
	if err != nil {
		utils.Error("GetPet error", zap.Error(err))
		return &pb.SetPetBattleStatusResponse{Success: false, Message: "宠物不存在"}, nil
	}
	if pet == nil || pet.UserID != userID {
		return &pb.SetPetBattleStatusResponse{Success: false, Message: "宠物不存在或不属于该用户"}, nil
	}

	// 设置出战状态
	err = h.dbClient.SetPetBattleStatus(ctx, userID, petID, isBattle)
	if err != nil {
		utils.Error("SetPetBattleStatus error", zap.Error(err))
		return nil, fmt.Errorf("failed to set pet battle status")
	}

	// 失效宠物缓存
	err = h.cacheService.InvalidateUserPetsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate pets cache", zap.Error(err))
	}

	statusText := "出战"
	if !isBattle {
		statusText = "休战"
	}

	return &pb.SetPetBattleStatusResponse{Success: true, Message: fmt.Sprintf("宠物%s成功", statusText)}, nil
}

// AddPetExp 增加宠物经验
func (h *Handler) AddPetExp(ctx context.Context, req *pb.AddPetExpRequest) (*pb.AddPetExpResponse, error) {
	userID := req.UserId
	petID := req.PetId
	exp := req.Exp

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if petID == 0 {
		return nil, fmt.Errorf("invalid pet id")
	}
	if exp <= 0 {
		return nil, fmt.Errorf("invalid exp value")
	}

	// 检查宠物是否存在并属于该用户
	pet, err := h.dbClient.GetPet(ctx, petID)
	if err != nil {
		utils.Error("GetPet error", zap.Error(err))
		return &pb.AddPetExpResponse{Success: false, Message: "宠物不存在"}, nil
	}
	if pet == nil || pet.UserID != userID {
		return &pb.AddPetExpResponse{Success: false, Message: "宠物不存在或不属于该用户"}, nil
	}

	// 业务层处理加经验和升级
	pet.Exp += exp
	for {
		levelTemplate, err := h.getPetLevelTemplate(pet.TemplateID, pet.Level)
		if err != nil {
			break // 没有下一级配置，停止升级
		}
		if pet.Exp >= int32(levelTemplate.Exp) {
			pet.Level++
			pet.Exp -= int32(levelTemplate.Exp)
		} else {
			break
		}
	}

	// 更新宠物数据
	err = h.dbClient.UpdatePet(ctx, pet)
	if err != nil {
		utils.Error("UpdatePet error", zap.Error(err))
		return nil, fmt.Errorf("failed to update pet")
	}

	// 失效宠物缓存
	err = h.cacheService.InvalidateUserPetsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate pets cache", zap.Error(err))
	}

	return &pb.AddPetExpResponse{Success: true, Message: "经验增加成功"}, nil
}
