package models

import (
	"time"
)

// User 用户模型 - 只包含基础字段，不包含关联关系
type User struct {
	ID           int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	Username     string    `json:"username" bson:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Level        int32     `json:"level" bson:"level" gorm:"type:int;default:1;not null"`
	Exp          int32     `json:"exp" bson:"exp" gorm:"type:int;default:0;not null"`
	Email        string    `json:"email" bson:"email" gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string    `json:"-" bson:"password_hash" gorm:"type:varchar(255);not null"`
	Salt         string    `json:"-" bson:"salt" gorm:"type:varchar(50);not null"`
	Role         string    `json:"role" bson:"role" gorm:"type:varchar(20);default:'user';not null"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// MonthlySign 月签到模型
type MonthlySign struct {
	UserID       int64     `json:"user_id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	Year         int32     `json:"year" bson:"year" gorm:"type:int;not null"`
	Month        int32     `json:"month" bson:"month" gorm:"type:int;not null"`
	SignDays     int32     `json:"sign_days" bson:"sign_days" gorm:"type:int;default:0;not null"`
	LastSignTime time.Time `json:"last_sign_time" bson:"last_sign_time" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at" gorm:"autoUpdateTime"`
}

func (MonthlySign) TableName() string {
	return "monthly_signs"
}

// MonthlySignReward 月签到奖励模型
type MonthlySignReward struct {
	UserID     int64     `json:"user_id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	Year       int32     `json:"year" bson:"year" gorm:"type:int;not null"`
	Month      int32     `json:"month" bson:"month" gorm:"type:int;not null"`
	RewardDays int32     `json:"reward_days" bson:"reward_days" gorm:"type:int;default:0;not null"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at" gorm:"autoUpdateTime"`
}

func (MonthlySignReward) TableName() string {
	return "monthly_sign_rewards"
}
