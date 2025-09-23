package social

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/pb"
)

// SocialHandler 社交处理器接口
type SocialHandler interface {
	// 好友相关
	GetFriendList(ctx context.Context, req *pb.GetFriendListRequest) (*pb.GetFriendListResponse, error)
	SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error)
	GetFriendRequestList(ctx context.Context, req *pb.GetFriendRequestListRequest) (*pb.GetFriendRequestListResponse, error)
	HandleFriendRequest(ctx context.Context, req *pb.HandleFriendRequestRequest) (*pb.HandleFriendRequestResponse, error)
	BatchHandleFriendRequest(ctx context.Context, req *pb.BatchHandleFriendRequestRequest) (*pb.BatchHandleFriendRequestResponse, error)
	DeleteFriend(ctx context.Context, req *pb.DeleteFriendRequest) (*pb.DeleteFriendResponse, error)

	// 公会相关
	CreateGuild(ctx context.Context, req *pb.CreateGuildRequest) (*pb.CreateGuildResponse, error)
	GetGuildInfo(ctx context.Context, req *pb.GetGuildInfoRequest) (*pb.GetGuildInfoResponse, error)
	GetGuildMembers(ctx context.Context, req *pb.GetGuildMembersRequest) (*pb.GetGuildMembersResponse, error)
	ApplyToGuild(ctx context.Context, req *pb.ApplyToGuildRequest) (*pb.ApplyToGuildResponse, error)
	InviteToGuild(ctx context.Context, req *pb.InviteToGuildRequest) (*pb.InviteToGuildResponse, error)
	GetGuildApplications(ctx context.Context, req *pb.GetGuildApplicationsRequest) (*pb.GetGuildApplicationsResponse, error)
	HandleGuildApplication(ctx context.Context, req *pb.HandleGuildApplicationRequest) (*pb.HandleGuildApplicationResponse, error)
	KickGuildMember(ctx context.Context, req *pb.KickGuildMemberRequest) (*pb.KickGuildMemberResponse, error)
	ChangeMemberRole(ctx context.Context, req *pb.ChangeMemberRoleRequest) (*pb.ChangeMemberRoleResponse, error)
	TransferGuildMaster(ctx context.Context, req *pb.TransferGuildMasterRequest) (*pb.TransferGuildMasterResponse, error)
	DisbandGuild(ctx context.Context, req *pb.DisbandGuildRequest) (*pb.DisbandGuildResponse, error)
	LeaveGuild(ctx context.Context, req *pb.LeaveGuildRequest) (*pb.LeaveGuildResponse, error)

	// 聊天相关
	SendChatMessage(ctx context.Context, req *pb.SendChatMessageRequest) (*pb.SendChatMessageResponse, error)
	GetChatMessages(ctx context.Context, req *pb.GetChatMessagesRequest) (*pb.GetChatMessagesResponse, error)
}
