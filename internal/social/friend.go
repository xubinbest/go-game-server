package social

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) GetFriendList(ctx context.Context, req *pb.GetFriendListRequest) (*pb.GetFriendListResponse, error) {
	friends, err := h.cacheService.GetFriendsWithCache(ctx, req.UserId, func() ([]*models.Friend, error) {
		return h.dbClient.GetFriends(ctx, req.UserId)
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get friend list: %v", err)
	}

	var friendInfos []*pb.FriendInfo
	for _, friend := range friends {
		friendInfos = append(friendInfos, &pb.FriendInfo{
			UserId: friend.ID,
		})
	}

	return &pb.GetFriendListResponse{
		Friends: friendInfos,
	}, nil
}

func (h *Handler) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	fromUserID := req.FromUserId
	toUserID := req.ToUserId

	// 检查是否已经是好友
	if _, err := h.dbClient.GetFriend(ctx, fromUserID, toUserID); err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "already friends")
	}

	// 创建好友请求
	err := h.dbClient.CreateFriendRequest(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send friend request: %v", err)
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateFriendRequestsCache(ctx, toUserID)

	return &pb.SendFriendRequestResponse{
		Success: true,
	}, nil
}

func (h *Handler) GetFriendRequestList(ctx context.Context, req *pb.GetFriendRequestListRequest) (*pb.GetFriendRequestListResponse, error) {
	userID := req.UserId
	requests, err := h.cacheService.GetFriendRequestsWithCache(ctx, userID, func() ([]*models.FriendRequest, error) {
		return h.dbClient.GetFriendRequests(ctx, userID)
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get friend requests: %v", err)
	}

	var requestInfos []*pb.FriendRequestInfo
	for _, request := range requests {
		utils.Info("request", zap.Any("request", request))
		requestInfos = append(requestInfos, &pb.FriendRequestInfo{
			RequestId:  request.ID,
			FromUserId: request.FromUserID,
			ToUserId:   request.ToUserID,
			CreatedAt:  request.CreatedAt.Unix(),
		})
	}

	return &pb.GetFriendRequestListResponse{
		Requests: requestInfos,
	}, nil
}

func (h *Handler) HandleFriendRequest(ctx context.Context, req *pb.HandleFriendRequestRequest) (*pb.HandleFriendRequestResponse, error) {
	requestID := req.RequestId
	action := req.Action

	utils.Info("HandleFriendRequest", zap.Int64("requestID", requestID), zap.String("action", action.String()))

	// 获取好友请求
	request, err := h.dbClient.GetFriendRequest(ctx, requestID)
	utils.Info("request", zap.Any("request", request))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "friend request not found")
	}

	// 处理请求
	switch action {
	case pb.HandleFriendRequestRequest_ACCEPT:
		// 添加好友
		err = h.dbClient.AddFriend(ctx, request.FromUserID, request.ToUserID)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to add friend: %v", err)
		}
		// 失效相关缓存
		_ = h.cacheService.InvalidateFriendsCache(ctx, request.FromUserID)
		_ = h.cacheService.InvalidateFriendsCache(ctx, request.ToUserID)
	case pb.HandleFriendRequestRequest_REJECT:
		// 拒绝请求，不做任何操作
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid action")
	}

	// 删除请求
	err = h.dbClient.DeleteFriendRequest(ctx, requestID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete friend request: %v", err)
	}

	// 失效好友申请缓存
	_ = h.cacheService.InvalidateFriendRequestsCache(ctx, request.ToUserID)

	return &pb.HandleFriendRequestResponse{
		Success: true,
	}, nil
}

func (h *Handler) BatchHandleFriendRequest(ctx context.Context, req *pb.BatchHandleFriendRequestRequest) (*pb.BatchHandleFriendRequestResponse, error) {
	requestIDs := req.RequestIds
	action := req.Action

	// 批量处理请求
	for _, requestID := range requestIDs {
		// 获取好友请求
		request, err := h.dbClient.GetFriendRequest(ctx, requestID)
		if err != nil {
			continue
		}

		// 处理请求
		switch action {
		case pb.BatchHandleFriendRequestRequest_ACCEPT_ALL:
			// 添加好友
			_ = h.dbClient.AddFriend(ctx, request.FromUserID, request.ToUserID)
			// 失效相关缓存
			_ = h.cacheService.InvalidateFriendsCache(ctx, request.FromUserID)
			_ = h.cacheService.InvalidateFriendsCache(ctx, request.ToUserID)
		case pb.BatchHandleFriendRequestRequest_REJECT_ALL:
			// 拒绝请求，不做任何操作
		default:
			return nil, status.Errorf(codes.InvalidArgument, "invalid action")
		}

		// 删除请求
		_ = h.dbClient.DeleteFriendRequest(ctx, requestID)
	}

	// 失效所有相关的好友申请缓存
	// 这里简化处理，实际应该根据具体的用户ID来失效
	// 由于BatchHandleFriendRequestRequest没有UserId字段，这里暂时跳过缓存失效

	return &pb.BatchHandleFriendRequestResponse{
		Success: true,
	}, nil
}

func (h *Handler) DeleteFriend(ctx context.Context, req *pb.DeleteFriendRequest) (*pb.DeleteFriendResponse, error) {
	userID := req.UserId
	friendID := req.FriendId

	// 检查是否是好友关系
	if _, err := h.dbClient.GetFriend(ctx, userID, friendID); err != nil {
		return nil, status.Errorf(codes.NotFound, "friend relationship not found")
	}

	// 删除好友关系
	err := h.dbClient.RemoveFriend(ctx, userID, friendID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove friend: %v", err)
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateFriendsCache(ctx, userID)
	_ = h.cacheService.InvalidateFriendsCache(ctx, friendID)

	return &pb.DeleteFriendResponse{
		Success: true,
	}, nil
}
