package user

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/pb"
)

// UserHandler 用户处理器接口
type UserHandler interface {
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error)
	GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error)
	AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error)
	RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error)
	UseItem(ctx context.Context, req *pb.UseItemRequest) (*pb.UseItemResponse, error)
	GetEquipments(ctx context.Context, req *pb.GetEquipmentsRequest) (*pb.GetEquipmentsResponse, error)
	EquipItem(ctx context.Context, req *pb.EquipItemRequest) (*pb.EquipItemResponse, error)
	UnequipItem(ctx context.Context, req *pb.UnequipItemRequest) (*pb.UnequipItemResponse, error)
	GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error)
	GetUserCards(ctx context.Context, req *pb.GetUserCardsRequest) (*pb.GetUserCardsResponse, error)
	ActivateCard(ctx context.Context, req *pb.ActivateCardRequest) (*pb.ActivateCardResponse, error)
	UpgradeCard(ctx context.Context, req *pb.UpgradeCardRequest) (*pb.UpgradeCardResponse, error)
	UpgradeCardStar(ctx context.Context, req *pb.UpgradeCardStarRequest) (*pb.UpgradeCardStarResponse, error)
	// 宠物相关方法
	GetUserPets(ctx context.Context, req *pb.GetUserPetsRequest) (*pb.GetUserPetsResponse, error)
	AddPet(ctx context.Context, req *pb.AddPetRequest) (*pb.AddPetResponse, error)
	SetPetBattleStatus(ctx context.Context, req *pb.SetPetBattleStatusRequest) (*pb.SetPetBattleStatusResponse, error)
	AddPetExp(ctx context.Context, req *pb.AddPetExpRequest) (*pb.AddPetExpResponse, error)
	// 月签到相关方法
	GetMonthlySignInfo(ctx context.Context, req *pb.GetMonthlySignInfoRequest) (*pb.GetMonthlySignInfoResponse, error)
	MonthlySign(ctx context.Context, req *pb.MonthlySignRequest) (*pb.MonthlySignResponse, error)
	ClaimMonthlySignReward(ctx context.Context, req *pb.ClaimMonthlySignRewardRequest) (*pb.ClaimMonthlySignRewardResponse, error)
}
