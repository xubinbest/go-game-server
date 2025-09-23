package user

import (
	"context"
	"fmt"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// GetUserInfo 获取用户信息（带缓存）
func (h *Handler) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	userID := req.UserId
	if userID == 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	// 使用缓存服务获取用户数据
	user, err := h.cacheService.GetUserInfoWithCache(ctx, userID, func() (*models.User, error) {
		return h.dbClient.GetUser(ctx, userID)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return &pb.GetUserInfoResponse{
		UserId: user.ID,
		Name:   user.Username,
		Level:  user.Level,
	}, nil
}
