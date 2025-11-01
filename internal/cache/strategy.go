package cache

import "time"

// CacheStrategy 缓存策略配置
type CacheStrategy struct {
	// 缓存键前缀
	KeyPrefix string
	// 缓存时间
	TTL time.Duration
	// 空值缓存时间
	EmptyTTL time.Duration
	// 是否使用分布式锁（防止击穿）
	UseLock bool
	// 锁超时时间
	LockTimeout time.Duration
}

// Strategies 预定义策略
var Strategies = map[string]CacheStrategy{
	// 用户信息 - 频繁读取，较少更新
	"user_info": {
		KeyPrefix:   "user:info:",
		TTL:         30 * time.Minute,
		EmptyTTL:    5 * time.Minute,
		UseLock:     true,
		LockTimeout: 10 * time.Second,
	},
	// 用户背包 - 频繁读写
	"user_inventory": {
		KeyPrefix:   "user:inventory:",
		TTL:         10 * time.Minute,
		EmptyTTL:    2 * time.Minute,
		UseLock:     true,
		LockTimeout: 5 * time.Second,
	},
	// 用户装备 - 中等频率
	"user_equipment": {
		KeyPrefix: "user:equipment:",
		TTL:       15 * time.Minute,
		EmptyTTL:  3 * time.Minute,
		UseLock:   false,
	},
	// 用户卡牌 - 较少更新
	"user_cards": {
		KeyPrefix: "user:cards:",
		TTL:       20 * time.Minute,
		EmptyTTL:  5 * time.Minute,
		UseLock:   false,
	},
	// 好友列表 - 较少更新
	"user_friends": {
		KeyPrefix: "user:friends:",
		TTL:       30 * time.Minute,
		EmptyTTL:  5 * time.Minute,
		UseLock:   false,
	},
	// 公会信息 - 较少更新
	"guild_info": {
		KeyPrefix: "guild:info:",
		TTL:       20 * time.Minute,
		EmptyTTL:  5 * time.Minute,
		UseLock:   false,
	},
	// 公会成员 - 中等频率更新
	"guild_members": {
		KeyPrefix:   "guild:members:",
		TTL:         15 * time.Minute,
		EmptyTTL:    3 * time.Minute,
		UseLock:     true,
		LockTimeout: 5 * time.Second,
	},
	// 好友申请 - 频繁更新
	"friend_requests": {
		KeyPrefix:   "friend:requests:",
		TTL:         10 * time.Minute,
		EmptyTTL:    2 * time.Minute,
		UseLock:     true,
		LockTimeout: 3 * time.Second,
	},
	// 公会申请 - 频繁更新
	"guild_applications": {
		KeyPrefix:   "guild:applications:",
		TTL:         10 * time.Minute,
		EmptyTTL:    2 * time.Minute,
		UseLock:     true,
		LockTimeout: 3 * time.Second,
	},
	// 用户宠物 - 中等频率更新
	"user_pets": {
		KeyPrefix: "pet:list:",
		TTL:       5 * time.Minute,
		EmptyTTL:  1 * time.Minute,
		UseLock:   false,
	},
	// 聊天消息 - 频繁更新
	"chat_messages": {
		KeyPrefix:   "chat:messages:",
		TTL:         5 * time.Minute,
		EmptyTTL:    1 * time.Minute,
		UseLock:     true,
		LockTimeout: 3 * time.Second,
	},
	// 月签到信息 - 频繁读取，每日更新
	"monthly_sign": {
		KeyPrefix:   "monthly:sign:",
		TTL:         1 * time.Hour,
		EmptyTTL:    5 * time.Minute,
		UseLock:     true,
		LockTimeout: 5 * time.Second,
	},
	// 月签到奖励记录 - 中等频率更新
	"monthly_sign_reward": {
		KeyPrefix:   "monthly:reward:",
		TTL:         30 * time.Minute,
		EmptyTTL:    5 * time.Minute,
		UseLock:     true,
		LockTimeout: 5 * time.Second,
	},
}
