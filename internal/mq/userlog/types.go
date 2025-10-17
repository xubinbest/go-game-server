package userlog

import "time"

// 用户创建日志
type UserCreateLog struct {
	UserName     string    `json:"user_name"`
	Time         time.Time `json:"time"`
	CreateIP     string    `json:"create_ip"`
	CreateDevice string    `json:"create_device"`
}

// 用户登录日志
type UserLoginLog struct {
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	Time        time.Time `json:"time"`
	LoginIP     string    `json:"login_ip"`
	LoginDevice string    `json:"login_device"`
}

// 用户登出日志
type UserLogoutLog struct {
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name"`
	Time         time.Time `json:"time"`
	LogoutIP     string    `json:"logout_ip"`
	LogoutDevice string    `json:"logout_device"`
}

// 用户物品日志
// opt: 1: 获取, 2: 消耗
type UserItemLog struct {
	UserID     int64     `json:"user_id"`
	UserName   string    `json:"user_name"`
	ItemID     int64     `json:"item_id"`
	ItemAmount int64     `json:"item_amount"`
	Opt        int32     `json:"opt"`
	Time       time.Time `json:"time"`
	ItemIP     string    `json:"item_ip"`
	ItemDevice string    `json:"item_device"`
}

// 用户货币日志
// opt: 1: 获取, 2: 消耗
type UserMoneyLog struct {
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	Money       int64     `json:"money"`
	MoneyType   int32     `json:"money_type"`
	Opt         int32     `json:"opt"`
	Time        time.Time `json:"time"`
	MoneyIP     string    `json:"money_ip"`
	MoneyDevice string    `json:"money_device"`
}
