package db

import (
	"context"

	"github.xubinbest.com/go-game-server/internal/db/models"
)

// GuildDatabase 接口方法实现

func (c *DatabaseClient) CreateGuild(ctx context.Context, guild *models.Guild) error {
	return c.guildDB.CreateGuild(ctx, guild)
}

func (c *DatabaseClient) CreateGuildWithMaster(ctx context.Context, guild *models.Guild, master *models.GuildMember) error {
	return c.guildDB.CreateGuildWithMaster(ctx, guild, master)
}

func (c *DatabaseClient) GetGuildByName(ctx context.Context, name string) (*models.Guild, error) {
	return c.guildDB.GetGuildByName(ctx, name)
}

func (c *DatabaseClient) GetGuild(ctx context.Context, guildID int64) (*models.Guild, error) {
	return c.guildDB.GetGuild(ctx, guildID)
}

func (c *DatabaseClient) AddGuildMember(ctx context.Context, member *models.GuildMember) error {
	return c.guildDB.AddGuildMember(ctx, member)
}

func (c *DatabaseClient) UpdateGuild(ctx context.Context, guild *models.Guild) error {
	return c.guildDB.UpdateGuild(ctx, guild)
}

func (c *DatabaseClient) DeleteGuild(ctx context.Context, guildID int64) error {
	return c.guildDB.DeleteGuild(ctx, guildID)
}

func (c *DatabaseClient) GetGuildMember(ctx context.Context, guildID, userID int64) (*models.GuildMember, error) {
	return c.guildDB.GetGuildMember(ctx, guildID, userID)
}

func (c *DatabaseClient) GetGuildMembers(ctx context.Context, guildID int64) ([]*models.GuildMember, error) {
	return c.guildDB.GetGuildMembers(ctx, guildID)
}

func (c *DatabaseClient) UpdateGuildMemberRole(ctx context.Context, guildID, userID int64, newRole int) error {
	return c.guildDB.UpdateGuildMemberRole(ctx, guildID, userID, newRole)
}

func (c *DatabaseClient) RemoveGuildMember(ctx context.Context, guildID, userID int64) error {
	return c.guildDB.RemoveGuildMember(ctx, guildID, userID)
}

func (c *DatabaseClient) CreateGuildApplication(ctx context.Context, application *models.GuildApplication) error {
	return c.guildDB.CreateGuildApplication(ctx, application)
}

func (c *DatabaseClient) GetGuildApplication(ctx context.Context, appID int64) (*models.GuildApplication, error) {
	return c.guildDB.GetGuildApplication(ctx, appID)
}

func (c *DatabaseClient) GetGuildApplications(ctx context.Context, guildID int64) ([]*models.GuildApplication, error) {
	return c.guildDB.GetGuildApplications(ctx, guildID)
}

func (c *DatabaseClient) DeleteGuildApplication(ctx context.Context, appID int64) error {
	return c.guildDB.DeleteGuildApplication(ctx, appID)
}

func (c *DatabaseClient) CreateGuildInvitation(ctx context.Context, invitation *models.GuildInvitation) error {
	return c.guildDB.CreateGuildInvitation(ctx, invitation)
}

func (c *DatabaseClient) GetGuildInvitations(ctx context.Context, guildID int64) ([]*models.GuildInvitation, error) {
	return c.guildDB.GetGuildInvitations(ctx, guildID)
}

func (c *DatabaseClient) GetUserPendingInvitations(ctx context.Context, userID int64) ([]*models.GuildInvitation, error) {
	return c.guildDB.GetUserPendingInvitations(ctx, userID)
}

func (c *DatabaseClient) GetUserGuilds(ctx context.Context, userID int64) ([]*models.Guild, error) {
	return c.guildDB.GetUserGuilds(ctx, userID)
}

func (c *DatabaseClient) GetGuildMemberCount(ctx context.Context, guildID int64) (int32, error) {
	return c.guildDB.GetGuildMemberCount(ctx, guildID)
}

func (c *DatabaseClient) GetGuildList(ctx context.Context, page, pageSize int32) ([]*models.Guild, int32, error) {
	return c.guildDB.GetGuildList(ctx, page, pageSize)
}
