package social

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/cache"
	"github.xubinbest.com/go-game-server/internal/config"
	"github.xubinbest.com/go-game-server/internal/db"
	"github.xubinbest.com/go-game-server/internal/designconfig"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/snowflake"
)

type SocialGRPCServer struct {
	pb.UnimplementedSocialServiceServer
	handler SocialHandler
}

func NewSocialGRPCServer(dbClient db.Database, cacheClient cache.Cache, sf *snowflake.Snowflake, cfg *config.Config, configManager *designconfig.DesignConfigManager) (*SocialGRPCServer, error) {
	cacheManager := cache.NewCacheManager(cacheClient)
	handler, err := NewHandler(dbClient, cacheClient, cacheManager, sf, cfg, configManager)
	if err != nil {
		return nil, err
	}
	return &SocialGRPCServer{
		UnimplementedSocialServiceServer: pb.UnimplementedSocialServiceServer{},
		handler:                          handler,
	}, nil
}

// 好友相关接口
func (s *SocialGRPCServer) GetFriendList(ctx context.Context, req *pb.GetFriendListRequest) (*pb.GetFriendListResponse, error) {
	return s.handler.GetFriendList(ctx, req)
}

func (s *SocialGRPCServer) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	return s.handler.SendFriendRequest(ctx, req)
}

func (s *SocialGRPCServer) GetFriendRequestList(ctx context.Context, req *pb.GetFriendRequestListRequest) (*pb.GetFriendRequestListResponse, error) {
	return s.handler.GetFriendRequestList(ctx, req)
}

func (s *SocialGRPCServer) HandleFriendRequest(ctx context.Context, req *pb.HandleFriendRequestRequest) (*pb.HandleFriendRequestResponse, error) {
	return s.handler.HandleFriendRequest(ctx, req)
}

func (s *SocialGRPCServer) BatchHandleFriendRequest(ctx context.Context, req *pb.BatchHandleFriendRequestRequest) (*pb.BatchHandleFriendRequestResponse, error) {
	return s.handler.BatchHandleFriendRequest(ctx, req)
}

func (s *SocialGRPCServer) DeleteFriend(ctx context.Context, req *pb.DeleteFriendRequest) (*pb.DeleteFriendResponse, error) {
	return s.handler.DeleteFriend(ctx, req)
}

// 公会相关接口
func (s *SocialGRPCServer) CreateGuild(ctx context.Context, req *pb.CreateGuildRequest) (*pb.CreateGuildResponse, error) {
	return s.handler.CreateGuild(ctx, req)
}

func (s *SocialGRPCServer) GetGuildInfo(ctx context.Context, req *pb.GetGuildInfoRequest) (*pb.GetGuildInfoResponse, error) {
	return s.handler.GetGuildInfo(ctx, req)
}

func (s *SocialGRPCServer) GetGuildMembers(ctx context.Context, req *pb.GetGuildMembersRequest) (*pb.GetGuildMembersResponse, error) {
	return s.handler.GetGuildMembers(ctx, req)
}

func (s *SocialGRPCServer) ApplyToGuild(ctx context.Context, req *pb.ApplyToGuildRequest) (*pb.ApplyToGuildResponse, error) {
	return s.handler.ApplyToGuild(ctx, req)
}

func (s *SocialGRPCServer) InviteToGuild(ctx context.Context, req *pb.InviteToGuildRequest) (*pb.InviteToGuildResponse, error) {
	return s.handler.InviteToGuild(ctx, req)
}

func (s *SocialGRPCServer) GetGuildApplications(ctx context.Context, req *pb.GetGuildApplicationsRequest) (*pb.GetGuildApplicationsResponse, error) {
	return s.handler.GetGuildApplications(ctx, req)
}

func (s *SocialGRPCServer) HandleGuildApplication(ctx context.Context, req *pb.HandleGuildApplicationRequest) (*pb.HandleGuildApplicationResponse, error) {
	return s.handler.HandleGuildApplication(ctx, req)
}

func (s *SocialGRPCServer) KickGuildMember(ctx context.Context, req *pb.KickGuildMemberRequest) (*pb.KickGuildMemberResponse, error) {
	return s.handler.KickGuildMember(ctx, req)
}

func (s *SocialGRPCServer) ChangeMemberRole(ctx context.Context, req *pb.ChangeMemberRoleRequest) (*pb.ChangeMemberRoleResponse, error) {
	return s.handler.ChangeMemberRole(ctx, req)
}

func (s *SocialGRPCServer) TransferGuildMaster(ctx context.Context, req *pb.TransferGuildMasterRequest) (*pb.TransferGuildMasterResponse, error) {
	return s.handler.TransferGuildMaster(ctx, req)
}

func (s *SocialGRPCServer) DisbandGuild(ctx context.Context, req *pb.DisbandGuildRequest) (*pb.DisbandGuildResponse, error) {
	return s.handler.DisbandGuild(ctx, req)
}

func (s *SocialGRPCServer) LeaveGuild(ctx context.Context, req *pb.LeaveGuildRequest) (*pb.LeaveGuildResponse, error) {
	return s.handler.LeaveGuild(ctx, req)
}

// 聊天相关接口
func (s *SocialGRPCServer) SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.SendChatMessageResponse, error) {
	return s.handler.SendChatMessage(ctx, req)
}

func (s *SocialGRPCServer) GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error) {
	return s.handler.GetChatMessages(ctx, req)
}
