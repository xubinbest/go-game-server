package interfaces

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// GuildDatabase 定义帮派相关的数据库操作接口
type GuildDatabase interface {
	// 创建帮派
	CreateGuild(ctx context.Context, guild *models.Guild) error

	// 创建帮派并添加帮主
	CreateGuildWithMaster(ctx context.Context, guild *models.Guild, master *models.GuildMember) error

	// 根据名称获取帮派
	GetGuildByName(ctx context.Context, name string) (*models.Guild, error)

	// 根据ID获取帮派
	GetGuild(ctx context.Context, guildID int64) (*models.Guild, error)

	// 添加帮派成员
	AddGuildMember(ctx context.Context, member *models.GuildMember) error

	// 更新帮派信息
	UpdateGuild(ctx context.Context, guild *models.Guild) error

	// 删除帮派
	DeleteGuild(ctx context.Context, guildID int64) error

	// 获取帮派成员
	GetGuildMember(ctx context.Context, guildID, userID int64) (*models.GuildMember, error)

	// 获取帮派所有成员
	GetGuildMembers(ctx context.Context, guildID int64) ([]*models.GuildMember, error)

	// 更新帮派成员角色
	UpdateGuildMemberRole(ctx context.Context, guildID, userID int64, newRole int) error

	// 移除帮派成员
	RemoveGuildMember(ctx context.Context, guildID, userID int64) error

	// 创建帮派申请
	CreateGuildApplication(ctx context.Context, application *models.GuildApplication) error

	// 获取帮派申请
	GetGuildApplication(ctx context.Context, appID int64) (*models.GuildApplication, error)

	// 获取帮派申请列表
	GetGuildApplications(ctx context.Context, guildID int64) ([]*models.GuildApplication, error)

	// 删除帮派申请
	DeleteGuildApplication(ctx context.Context, appID int64) error

	// 创建帮派邀请
	CreateGuildInvitation(ctx context.Context, invitation *models.GuildInvitation) error

	// 获取帮派邀请列表
	GetGuildInvitations(ctx context.Context, guildID int64) ([]*models.GuildInvitation, error)

	// 获取用户邀请列表
	GetUserPendingInvitations(ctx context.Context, userID int64) ([]*models.GuildInvitation, error)

	// 获取用户帮派
	GetUserGuilds(ctx context.Context, userID int64) ([]*models.Guild, error)

	// 获取帮派成员数量
	GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error)

	// 分页查询帮派列表
	GetGuildList(ctx context.Context, page, pageSize int32) ([]*models.Guild, int32, error)

	// 其他帮派相关方法...
}
