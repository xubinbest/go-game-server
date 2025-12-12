package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// FriendDatabase 接口方法实现

func (c *DatabaseClient) GetFriends(ctx context.Context, userID int64) ([]*models.Friend, error) {
	return c.friendDB.GetFriends(ctx, userID)
}

func (c *DatabaseClient) GetFriend(ctx context.Context, userID, friendID int64) (*models.Friend, error) {
	return c.friendDB.GetFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) error {
	return c.friendDB.CreateFriendRequest(ctx, fromUserID, toUserID)
}

func (c *DatabaseClient) GetFriendRequests(ctx context.Context, userID int64) ([]*models.FriendRequest, error) {
	return c.friendDB.GetFriendRequests(ctx, userID)
}

func (c *DatabaseClient) GetFriendRequest(ctx context.Context, requestID int64) (*models.FriendRequest, error) {
	return c.friendDB.GetFriendRequest(ctx, requestID)
}

func (c *DatabaseClient) AddFriend(ctx context.Context, userID, friendID int64) error {
	return c.friendDB.AddFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	return c.friendDB.RemoveFriend(ctx, userID, friendID)
}

func (c *DatabaseClient) DeleteFriendRequest(ctx context.Context, requestID int64) error {
	return c.friendDB.DeleteFriendRequest(ctx, requestID)
}
