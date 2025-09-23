package models

import (
	"time"
)

// GuildMemberRole 定义帮派成员角色常量
const (
	GuildRoleMaster     = 1
	GuildRoleViceMaster = 2
	GuildRoleElder      = 3
	GuildRoleElite      = 4
	GuildRoleMember     = 5
	GuildRoleApprentice = 6
)

// GuildApplicationStatus 定义帮派申请状态常量
const (
	GuildAppStatusPending  = 1
	GuildAppStatusAccepted = 2
	GuildAppStatusRejected = 3
)

// GuildInvitationStatus 定义帮派邀请状态常量
const (
	GuildInvStatusPending  = 1
	GuildInvStatusAccepted = 2
	GuildInvStatusRejected = 3
)

// Guild 公会模型
type Guild struct {
	ID           int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	Name         string    `json:"name" bson:"name" gorm:"type:varchar(50);uniqueIndex;not null"`
	Description  string    `json:"description" bson:"description" gorm:"type:text"`
	Announcement string    `json:"announcement" bson:"announcement" gorm:"type:text"`
	MasterID     int64     `json:"master_id" bson:"master_id" gorm:"type:bigint;not null;index"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at" gorm:"autoCreateTime"`
	MaxMembers   int32     `json:"max_members" bson:"max_members" gorm:"type:int;default:50;not null"`
	Version      int32     `json:"version" bson:"version" gorm:"type:int;default:1;not null"` // 用于乐观锁
}

func (Guild) TableName() string {
	return "guilds"
}

// GuildMember 公会成员模型
type GuildMember struct {
	ID        int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	GuildID   int64     `json:"guild_id" bson:"guild_id" gorm:"type:bigint;not null;index"`
	UserID    int64     `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	Role      int32     `json:"role" bson:"role" gorm:"type:int;default:5;not null"` // 1: 帮主, 2: 副帮主, 3: 长老, 4: 精英, 5: 普通成员, 6: 学徒
	JoinTime  time.Time `json:"join_time" bson:"join_time" gorm:"autoCreateTime"`
	LastLogin time.Time `json:"last_login" bson:"last_login" gorm:"autoCreateTime"`
}

func (GuildMember) TableName() string {
	return "guild_members"
}

// GuildApplication 公会申请模型
type GuildApplication struct {
	ID         int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	GuildID    int64     `json:"guild_id" bson:"guild_id" gorm:"type:bigint;not null;index"`
	UserID     int64     `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	Time       time.Time `json:"time" bson:"time" gorm:"autoCreateTime"`
	ExpireTime time.Time `json:"expire_time" bson:"expire_time" gorm:"not null"`
}

func (GuildApplication) TableName() string {
	return "guild_applications"
}

// GuildInvitation 公会邀请模型
type GuildInvitation struct {
	ID         int64     `json:"id" bson:"_id" gorm:"primaryKey;autoIncrement:false"`
	GuildID    int64     `json:"guild_id" bson:"guild_id" gorm:"type:bigint;not null;index"`
	UserID     int64     `json:"user_id" bson:"user_id" gorm:"type:bigint;not null;index"`
	InviterID  int64     `json:"inviter_id" bson:"inviter_id" gorm:"type:bigint;not null;index"`
	Time       time.Time `json:"time" bson:"time" gorm:"autoCreateTime"`
	ExpireTime time.Time `json:"expire_time" bson:"expire_time" gorm:"not null"`
}

func (GuildInvitation) TableName() string {
	return "guild_invitations"
}
