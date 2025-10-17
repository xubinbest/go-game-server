package models

import (
	"time"
)

// UserCreateLog 用户创建日志
type UserCreateLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserName     string    `json:"user_name" gorm:"type:varchar(50);not null;index"`
	Time         time.Time `json:"time" gorm:"not null;index"`
	CreateIP     string    `json:"create_ip" gorm:"type:varchar(45);not null"`
	CreateDevice string    `json:"create_device" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserCreateLog) TableName() string {
	return "user_create_logs"
}

// UserLoginLog 用户登录日志
type UserLoginLog struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID      int64     `json:"user_id" gorm:"type:bigint;not null;index"`
	UserName    string    `json:"user_name" gorm:"type:varchar(50);not null;index"`
	Time        time.Time `json:"time" gorm:"not null;index"`
	LoginIP     string    `json:"login_ip" gorm:"type:varchar(45);not null"`
	LoginDevice string    `json:"login_device" gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserLoginLog) TableName() string {
	return "user_login_logs"
}

// UserLogoutLog 用户登出日志
type UserLogoutLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID       int64     `json:"user_id" gorm:"type:bigint;not null;index"`
	UserName     string    `json:"user_name" gorm:"type:varchar(50);not null;index"`
	Time         time.Time `json:"time" gorm:"not null;index"`
	LogoutIP     string    `json:"logout_ip" gorm:"type:varchar(45);not null"`
	LogoutDevice string    `json:"logout_device" gorm:"type:varchar(255);not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserLogoutLog) TableName() string {
	return "user_logout_logs"
}

// UserItemLog 用户物品日志
type UserItemLog struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID     int64     `json:"user_id" gorm:"type:bigint;not null;index"`
	UserName   string    `json:"user_name" gorm:"type:varchar(50);not null;index"`
	ItemID     int64     `json:"item_id" gorm:"type:bigint;not null;index"`
	ItemAmount int64     `json:"item_amount" gorm:"type:bigint;not null"`
	Opt        int32     `json:"opt" gorm:"type:int;not null;index"` // 1: 获取, 2: 消耗
	Time       time.Time `json:"time" gorm:"not null;index"`
	ItemIP     string    `json:"item_ip" gorm:"type:varchar(45);not null"`
	ItemDevice string    `json:"item_device" gorm:"type:varchar(255);not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserItemLog) TableName() string {
	return "user_item_logs"
}

// UserMoneyLog 用户货币日志
type UserMoneyLog struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement:false"`
	UserID      int64     `json:"user_id" gorm:"type:bigint;not null;index"`
	UserName    string    `json:"user_name" gorm:"type:varchar(50);not null;index"`
	Money       int64     `json:"money" gorm:"type:bigint;not null"`
	MoneyType   int32     `json:"money_type" gorm:"type:int;not null;index"`
	Opt         int32     `json:"opt" gorm:"type:int;not null;index"` // 1: 获取, 2: 消耗
	Time        time.Time `json:"time" gorm:"not null;index"`
	MoneyIP     string    `json:"money_ip" gorm:"type:varchar(45);not null"`
	MoneyDevice string    `json:"money_device" gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserMoneyLog) TableName() string {
	return "user_money_logs"
}
