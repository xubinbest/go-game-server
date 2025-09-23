package models

import (
	"time"
)

// Friend 好友模型
type Friend struct {
	ID        int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	UserID    int64     `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	FriendID  int64     `json:"friend_id" bson:"friend_id" gorm:"type:bigint;not null;index"`
	CreatedAt time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
}

func (Friend) TableName() string {
	return "friends"
}

// FriendRequest 好友请求模型
type FriendRequest struct {
	ID           int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	FromUserID   int64     `json:"from_user_id" bson:"from_user_id" gorm:"type:bigint;not null;index"`
	ToUserID     int64     `json:"to_user_id" bson:"to_user_id" gorm:"type:bigint;not null;index"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty" gorm:"autoUpdateTime"`
	Status       int       `json:"status" bson:"status" gorm:"type:int;default:1;not null"` // 1: 待处理, 2: 已接受, 3: 已拒绝
	FromUsername string    `json:"from_username" bson:"from_username,omitempty" gorm:"type:varchar(50);not null"`
}

func (FriendRequest) TableName() string {
	return "friend_requests"
}
