package core

import (
	"fmt"
	"sync"
)

// Client 统一的客户端接口
type Client interface {
	// 认证相关
	Register(username, password, email string) (*AuthResult, error)
	Login(username, password string) (*AuthResult, error)

	// 用户功能
	GetInventory() (*InventoryResult, error)
	AddItem(itemID int64, itemName string, count int32) error
	RemoveItem(itemID int64, count int32) error
	UseItem(itemID int64, count int32) error

	// 好友功能
	GetFriendList() (*FriendListResult, error)
	SendFriendRequest(targetUserID int64) error
	GetFriendRequests() (*FriendRequestsResult, error)
	HandleFriendRequest(requestID int64, action int32) error
	DeleteFriend(friendID int64) error

	// 帮派功能
	CreateGuild(name, description, announcement string) (*GuildResult, error)
	GetGuildList() (*GuildListResult, error)
	GetGuildInfo(guildID int64) (*GuildInfoResult, error)
	GetGuildMembers(guildID int64) (*GuildMembersResult, error)
	ApplyToGuild(guildID int64) error
	InviteToGuild(targetUserID, guildID int64) error
	GetGuildApplications(guildID int64) (*GuildApplicationsResult, error)
	HandleGuildApplication(applicationID int64, action int32) error
	KickGuildMember(memberID, guildID int64) error
	ChangeMemberRole(memberID, guildID int64, role string) error
	TransferGuildMaster(newMasterID, guildID int64) error
	DisbandGuild(guildID int64) error
	LeaveGuild(guildID int64) error

	// 聊天功能
	SendChatMessage(channel int32, content string, targetID int64, extra string) error
	GetChatHistory(channel int32, targetID int64, page, pageSize int32) (*ChatHistoryResult, error)

	// 游戏功能
	JoinGame(gameID string) error
	LeaveGame(gameID string) error
	GetGameStatus(gameID string) (*GameStatusResult, error)
	PlayerAction(gameID, action string, data []byte) error

	// 连接管理
	Close() error
}

// 认证结果
type AuthResult struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
}

// 背包结果
type InventoryResult struct {
	Items []InventoryItem `json:"items"`
}

type InventoryItem struct {
	ItemID int64  `json:"item_id"`
	Name   string `json:"name"`
	Count  int32  `json:"count"`
}

// 好友列表结果
type FriendListResult struct {
	Friends []Friend `json:"friends"`
}

type Friend struct {
	UserID int64 `json:"user_id"`
}

// 好友请求结果
type FriendRequestsResult struct {
	Requests []FriendRequest `json:"requests"`
}

type FriendRequest struct {
	RequestID  int64 `json:"request_id"`
	FromUserID int64 `json:"from_user_id"`
}

// 游戏状态结果
type GameStatusResult struct {
	State   string   `json:"state"`
	Players []string `json:"players"`
}

// 帮派相关结果
type GuildResult struct {
	GuildID int64  `json:"guild_id"`
	Name    string `json:"name"`
}

type GuildListResult struct {
	Guilds []GuildInfo `json:"guilds"`
}

type GuildInfo struct {
	GuildID     int64  `json:"guild_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int32  `json:"member_count"`
	MaxMembers  int32  `json:"max_members"`
	MasterID    int64  `json:"master_id"`
	MasterName  string `json:"master_name"`
	CreatedAt   string `json:"created_at"`
}

type GuildInfoResult struct {
	Guild GuildInfo `json:"guild"`
}

type GuildMembersResult struct {
	Members []GuildMember `json:"members"`
}

type GuildMember struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	JoinTime string `json:"join_time"`
}

type GuildApplicationsResult struct {
	Applications []GuildApplication `json:"applications"`
}

type GuildApplication struct {
	ApplicationID int64  `json:"application_id"`
	UserID        int64  `json:"user_id"`
	Username      string `json:"username"`
	ApplyTime     string `json:"apply_time"`
	Message       string `json:"message"`
}

// 聊天相关结果
type ChatHistoryResult struct {
	Messages []ChatMessage `json:"messages"`
	Total    int32         `json:"total"`
	Page     int32         `json:"page"`
	PageSize int32         `json:"page_size"`
}

type ChatMessage struct {
	MessageID  int64  `json:"id"`
	Channel    int32  `json:"channel"`
	SenderID   int64  `json:"sender_id"`
	SenderName string `json:"sender_name"`
	ReceiverID int64  `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
	ExtraData  string `json:"extra_data"`
}

// 全局认证状态管理
var (
	currentUserID int64
	currentToken  string
	authMutex     sync.RWMutex
)

// SetAuth 设置认证信息
func SetAuth(token string, userID int64) {
	authMutex.Lock()
	defer authMutex.Unlock()
	currentToken = token
	currentUserID = userID
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID() int64 {
	authMutex.RLock()
	defer authMutex.RUnlock()
	return currentUserID
}

// GetCurrentToken 获取当前token
func GetCurrentToken() string {
	authMutex.RLock()
	defer authMutex.RUnlock()
	return currentToken
}

// ClearAuth 清除认证信息
func ClearAuth() {
	authMutex.Lock()
	defer authMutex.Unlock()
	currentToken = ""
	currentUserID = 0
}

func GetStringInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// 输入处理函数
func GetIntInput(prompt string) int {
	fmt.Print(prompt)
	choice := -1
	fmt.Scanln(&choice)
	return choice
}
