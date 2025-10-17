package errors

import (
	"fmt"
)

// ErrorCode 业务错误码
type ErrorCode int

const (
	// 通用错误 (1000-1999)
	ErrInvalidParam     ErrorCode = 1001 // 无效参数
	ErrInternalError    ErrorCode = 1002 // 内部错误
	ErrNotFound         ErrorCode = 1003 // 资源不找到
	ErrAlreadyExists    ErrorCode = 1004 // 资源已存在
	ErrPermissionDenied ErrorCode = 1005 // 权限不足

	// 用户相关错误 (2000-2999)
	ErrUserNotFound       ErrorCode = 2001 // 用户不存在
	ErrUserAlreadyExists  ErrorCode = 2002 // 用户已存在
	ErrInvalidCredentials ErrorCode = 2003 // 凭证无效

	// 物品相关错误 (3000-3999)
	ErrItemNotFound     ErrorCode = 3001 // 物品不存在
	ErrInsufficientItem ErrorCode = 3002 // 物品数量不足
	ErrInventoryFull    ErrorCode = 3003 // 背包已满

	// 卡牌相关错误 (4000-4999)
	ErrCardNotFound      ErrorCode = 4001 // 卡牌不存在
	ErrCardAlreadyExists ErrorCode = 4002 // 卡牌已存在
	ErrCardMaxLevel      ErrorCode = 4003 // 卡牌已达最大等级

	// 宠物相关错误 (5000-5999)
	ErrPetNotFound ErrorCode = 5001 // 宠物不存在
	ErrPetMaxLevel ErrorCode = 5002 // 宠物已达最大等级

	// 公会相关错误 (6000-6999)
	ErrGuildNotFound      ErrorCode = 6001 // 公会不存在
	ErrGuildAlreadyExists ErrorCode = 6002 // 公会名已存在
	ErrGuildFull          ErrorCode = 6003 // 公会已满
	ErrNotGuildMember     ErrorCode = 6004 // 不是公会成员
	ErrAlreadyInGuild     ErrorCode = 6005 // 已在公会中

	// 好友相关错误 (7000-7999)
	ErrFriendNotFound        ErrorCode = 7001 // 好友不存在
	ErrAlreadyFriends        ErrorCode = 7002 // 已经是好友
	ErrFriendRequestNotFound ErrorCode = 7003 // 好友请求不存在

	// 签到相关错误 (8000-8999)
	ErrAlreadySigned        ErrorCode = 8001 // 今日已签到
	ErrSignRewardClaimed    ErrorCode = 8002 // 奖励已领取
	ErrInsufficientSignDays ErrorCode = 8003 // 签到天数不足

	// 配置相关错误 (9000-9999)
	ErrTemplateNotFound ErrorCode = 9001 // 模板配置不存在
	ErrConfigLoadFailed ErrorCode = 9002 // 配置加载失败
)

// BusinessError 业务错误
type BusinessError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error 实现error接口
func (e *BusinessError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 支持errors.Unwrap
func (e *BusinessError) Unwrap() error {
	return e.Cause
}

// New 创建业务错误
func New(code ErrorCode, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装底层错误
func Wrap(code ErrorCode, message string, cause error) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// GetCode 获取错误码
func GetCode(err error) ErrorCode {
	if bizErr, ok := err.(*BusinessError); ok {
		return bizErr.Code
	}
	return ErrInternalError
}

// 错误码消息映射
var errorMessages = map[ErrorCode]string{
	ErrInvalidParam:          "无效参数",
	ErrInternalError:         "内部错误",
	ErrNotFound:              "资源不存在",
	ErrAlreadyExists:         "资源已存在",
	ErrPermissionDenied:      "权限不足",
	ErrUserNotFound:          "用户不存在",
	ErrUserAlreadyExists:     "用户已存在",
	ErrInvalidCredentials:    "用户名或密码错误",
	ErrItemNotFound:          "物品不存在",
	ErrInsufficientItem:      "物品数量不足",
	ErrInventoryFull:         "背包已满",
	ErrCardNotFound:          "卡牌不存在",
	ErrCardAlreadyExists:     "卡牌已激活",
	ErrCardMaxLevel:          "卡牌已达最大等级",
	ErrPetNotFound:           "宠物不存在",
	ErrPetMaxLevel:           "宠物已达最大等级",
	ErrGuildNotFound:         "公会不存在",
	ErrGuildAlreadyExists:    "公会名已存在",
	ErrGuildFull:             "公会人数已满",
	ErrNotGuildMember:        "不是公会成员",
	ErrAlreadyInGuild:        "已在公会中",
	ErrFriendNotFound:        "好友不存在",
	ErrAlreadyFriends:        "已经是好友",
	ErrFriendRequestNotFound: "好友请求不存在",
	ErrAlreadySigned:         "今日已签到",
	ErrSignRewardClaimed:     "奖励已领取",
	ErrInsufficientSignDays:  "签到天数不足",
	ErrTemplateNotFound:      "模板配置不存在",
	ErrConfigLoadFailed:      "配置加载失败",
}

// GetMessage 获取错误消息
func GetMessage(code ErrorCode) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
