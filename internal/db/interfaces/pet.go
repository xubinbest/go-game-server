package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// PetDatabase 定义宠物相关的数据库操作接口
type PetDatabase interface {
	// 根据ID获取宠物
	GetPet(ctx context.Context, petID int64) (*models.Pet, error)

	// 获取用户的所有宠物
	GetUserPets(ctx context.Context, userID int64) ([]*models.Pet, error)

	// 获取用户出战的宠物
	GetUserBattlePet(ctx context.Context, userID int64) (*models.Pet, error)

	// 创建新宠物
	CreatePet(ctx context.Context, pet *models.Pet) error

	// 更新宠物信息
	UpdatePet(ctx context.Context, pet *models.Pet) error

	// 删除宠物
	DeletePet(ctx context.Context, petID int64) error

	// 设置宠物出战状态
	SetPetBattleStatus(ctx context.Context, userID int64, petID int64, isBattle bool) error

	// 取消用户所有宠物的出战状态
	CancelAllPetBattleStatus(ctx context.Context, userID int64) error
}
