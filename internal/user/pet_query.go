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

// getPetTemplate 从内存配置中获取宠物模板 (O(1)查询)
func (h *Handler) getPetTemplate(templateID int64) (*designconfig.PetData, error) {
	return h.configManager.GetPetByID(templateID)
}

// getPetLevelTemplate 从内存配置中获取宠物等级模板
// 注意：此方法需要通过petID和level组合查找，保留原有的O(n)查询
func (h *Handler) getPetLevelTemplate(templateID int64, level int32) (*designconfig.PetLevelData, error) {
	petLevels := h.configManager.GetConfig("pet_level")
	if petLevels == nil {
		return nil, fmt.Errorf("pet level config not found")
	}

	// 类型断言获取切片
	petLevelsSlice, ok := petLevels.([]designconfig.PetLevelData)
	if !ok {
		return nil, fmt.Errorf("pet level config type assertion failed")
	}

	for i := range petLevelsSlice {
		if int64(petLevelsSlice[i].PetId) == templateID && int32(petLevelsSlice[i].Level) == level {
			return &petLevelsSlice[i], nil
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
