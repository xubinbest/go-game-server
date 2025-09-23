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

// getPetTemplate 从内存配置中获取宠物模板
func (h *Handler) getPetTemplate(templateID int64) (*designconfig.PetData, error) {
	pets := h.configManager.GetConfig("pet")
	if pets == nil {
		return nil, fmt.Errorf("pet config not found")
	}

	petsSlice := reflect.ValueOf(pets)
	for i := 0; i < petsSlice.Len(); i++ {
		pet := petsSlice.Index(i).Interface().(designconfig.PetData)
		if int64(pet.ID) == templateID {
			return &pet, nil
		}
	}
	return nil, fmt.Errorf("pet template not found: %d", templateID)
}

// getPetLevelTemplate 从内存配置中获取宠物等级模板
func (h *Handler) getPetLevelTemplate(templateID int64, level int32) (*designconfig.PetLevelData, error) {
	petLevels := h.configManager.GetConfig("pet_level")
	if petLevels == nil {
		return nil, fmt.Errorf("pet level config not found")
	}

	petLevelsSlice := reflect.ValueOf(petLevels)
	for i := 0; i < petLevelsSlice.Len(); i++ {
		petLevel := petLevelsSlice.Index(i).Interface().(designconfig.PetLevelData)
		if int64(petLevel.PetId) == templateID && int32(petLevel.Level) == level {
			return &petLevel, nil
		}
	}
	return nil, fmt.Errorf("pet level template not found: pet_id=%d, level=%d", templateID, level)
}

// GetUserPets 获取玩家所有宠物信息（带缓存）
func (h *Handler) GetUserPets(ctx context.Context, req *pb.GetUserPetsRequest) (*pb.GetUserPetsResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 使用缓存服务获取宠物数据
	pets, err := h.cacheService.GetUserPetsWithCache(ctx, userID, func() ([]*models.Pet, error) {
		return h.dbClient.GetUserPets(ctx, userID)
	})
	if err != nil {
		utils.Error("GetUserPets error", zap.Error(err))
		return nil, fmt.Errorf("failed to get user pets")
	}

	pbPets := make([]*pb.Pet, 0, len(pets))
	for _, pet := range pets {
		// 从内存配置中获取模板数据
		template, err := h.getPetTemplate(pet.TemplateID)
		if err != nil {
			utils.Error("Failed to get pet template", zap.Int64("template_id", pet.TemplateID), zap.Error(err))
			continue
		}

		// 构建宠物属性JSON
		properties := map[string]interface{}{
			"name":  template.Name,
			"color": template.Color,
		}
		propertiesJSON, _ := json.Marshal(properties)

		pbPets = append(pbPets, &pb.Pet{
			Id:         pet.ID,
			TemplateId: pet.TemplateID,
			Name:       pet.Name,
			Level:      pet.Level,
			Exp:        pet.Exp,
			IsBattle:   pet.IsBattle,
			Properties: string(propertiesJSON),
		})
	}

	return &pb.GetUserPetsResponse{
		Pets: pbPets,
	}, nil
}

// AddPet 为玩家添加宠物
func (h *Handler) AddPet(ctx context.Context, req *pb.AddPetRequest) (*pb.AddPetResponse, error) {
	userID := req.UserId
	templateID := req.TemplateId

	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	if templateID == 0 {
		return nil, fmt.Errorf("invalid template id")
	}

	// 获取宠物模板
	template, err := h.getPetTemplate(templateID)
	if err != nil {
		utils.Error("Failed to get pet template", zap.Error(err))
		return &pb.AddPetResponse{Success: false, Message: "宠物模板不存在"}, nil
	}

	// 创建宠物数据
	pet := &models.Pet{
		UserID:     userID,
		TemplateID: templateID,
		Name:       template.Name,
		Level:      1,
		Exp:        0,
		IsBattle:   false,
	}

	// 生成宠物ID
	petID, err := h.sf.NextID()
	if err != nil {
		utils.Error("Failed to generate pet ID", zap.Error(err))
		return nil, fmt.Errorf("failed to generate pet ID")
	}
	pet.ID = petID

	// 保存到数据库
	err = h.dbClient.CreatePet(ctx, pet)
	if err != nil {
		utils.Error("CreatePet error", zap.Error(err))
		return nil, fmt.Errorf("failed to create pet")
	}

	// 失效宠物缓存
	err = h.cacheService.InvalidateUserPetsCache(ctx, userID)
	if err != nil {
		utils.Error("Failed to invalidate pets cache", zap.Error(err))
	}

	return &pb.AddPetResponse{Success: true, Message: "宠物添加成功"}, nil
}

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
