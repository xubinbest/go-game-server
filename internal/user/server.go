package user

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/snowflake"
)

type UserGRPCServer struct {
	pb.UnimplementedUserServiceServer
	handler UserHandler
}

func NewUserGRPCServer(dbClient db.Database, cacheClient cache.Cache, sf *snowflake.Snowflake, cfg *config.Config, configManager *designconfig.DesignConfigManager) *UserGRPCServer {
	cacheManager := cache.NewCacheManager(cacheClient)
	handler := NewHandler(dbClient, cacheClient, cacheManager, sf, cfg, configManager)
	return &UserGRPCServer{
		UnimplementedUserServiceServer: pb.UnimplementedUserServiceServer{},
		handler:                        handler,
	}
}

func (s *UserGRPCServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return s.handler.Register(ctx, req)
}

func (s *UserGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.handler.Login(ctx, req)
}

func (s *UserGRPCServer) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	return s.handler.GetInventory(ctx, req)
}

func (s *UserGRPCServer) AddItem(ctx context.Context, req *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	return s.handler.AddItem(ctx, req)
}

func (s *UserGRPCServer) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*pb.RemoveItemResponse, error) {
	return s.handler.RemoveItem(ctx, req)
}

func (s *UserGRPCServer) UseItem(ctx context.Context, req *pb.UseItemRequest) (*pb.UseItemResponse, error) {
	return s.handler.UseItem(ctx, req)
}

func (s *UserGRPCServer) GetEquipments(ctx context.Context, req *pb.GetEquipmentsRequest) (*pb.GetEquipmentsResponse, error) {
	return s.handler.GetEquipments(ctx, req)
}

func (s *UserGRPCServer) EquipItem(ctx context.Context, req *pb.EquipItemRequest) (*pb.EquipItemResponse, error) {
	return s.handler.EquipItem(ctx, req)
}

func (s *UserGRPCServer) UnequipItem(ctx context.Context, req *pb.UnequipItemRequest) (*pb.UnequipItemResponse, error) {
	return s.handler.UnequipItem(ctx, req)
}

func (s *UserGRPCServer) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	return s.handler.GetUserInfo(ctx, req)
}

func (s *UserGRPCServer) GetUserCards(ctx context.Context, req *pb.GetUserCardsRequest) (*pb.GetUserCardsResponse, error) {
	return s.handler.GetUserCards(ctx, req)
}

func (s *UserGRPCServer) ActivateCard(ctx context.Context, req *pb.ActivateCardRequest) (*pb.ActivateCardResponse, error) {
	return s.handler.ActivateCard(ctx, req)
}

func (s *UserGRPCServer) UpgradeCard(ctx context.Context, req *pb.UpgradeCardRequest) (*pb.UpgradeCardResponse, error) {
	return s.handler.UpgradeCard(ctx, req)
}

func (s *UserGRPCServer) UpgradeCardStar(ctx context.Context, req *pb.UpgradeCardStarRequest) (*pb.UpgradeCardStarResponse, error) {
	return s.handler.UpgradeCardStar(ctx, req)
}

// 宠物相关方法

func (s *UserGRPCServer) GetUserPets(ctx context.Context, req *pb.GetUserPetsRequest) (*pb.GetUserPetsResponse, error) {
	return s.handler.GetUserPets(ctx, req)
}

func (s *UserGRPCServer) AddPet(ctx context.Context, req *pb.AddPetRequest) (*pb.AddPetResponse, error) {
	return s.handler.AddPet(ctx, req)
}

func (s *UserGRPCServer) SetPetBattleStatus(ctx context.Context, req *pb.SetPetBattleStatusRequest) (*pb.SetPetBattleStatusResponse, error) {
	return s.handler.SetPetBattleStatus(ctx, req)
}

func (s *UserGRPCServer) AddPetExp(ctx context.Context, req *pb.AddPetExpRequest) (*pb.AddPetExpResponse, error) {
	return s.handler.AddPetExp(ctx, req)
}

// 月签到相关方法
func (s *UserGRPCServer) GetMonthlySignInfo(ctx context.Context, req *pb.GetMonthlySignInfoRequest) (*pb.GetMonthlySignInfoResponse, error) {
	return s.handler.GetMonthlySignInfo(ctx, req)
}

func (s *UserGRPCServer) MonthlySign(ctx context.Context, req *pb.MonthlySignRequest) (*pb.MonthlySignResponse, error) {
	return s.handler.MonthlySign(ctx, req)
}

func (s *UserGRPCServer) ClaimMonthlySignReward(ctx context.Context, req *pb.ClaimMonthlySignRewardRequest) (*pb.ClaimMonthlySignRewardResponse, error) {
	return s.handler.ClaimMonthlySignReward(ctx, req)
}
