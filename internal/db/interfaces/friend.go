package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// FriendDatabase 定义好友相关的数据库操作接口
type FriendDatabase interface {
	GetFriends(ctx context.Context, userID int64) ([]*models.Friend, error)
	GetFriend(ctx context.Context, userID, friendID int64) (*models.Friend, error)
	CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) error
	GetFriendRequests(ctx context.Context, userID int64) ([]*models.FriendRequest, error)
	GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error)
	AddFriend(ctx context.Context, userID, friendID int64) error
	RemoveFriend(ctx context.Context, userID, friendID int64) error
	DeleteFriendRequest(ctx context.Context, requestID int64) error

	// 其他好友相关方法...
}
