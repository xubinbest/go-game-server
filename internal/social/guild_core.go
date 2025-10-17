package social

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// 检查用户是否在指定帮派中
func (h *Handler) isUserInGuild(ctx context.Context, guildID, userID int64) (bool, error) {
	member, err := h.dbClient.GetGuildMember(ctx, guildID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check guild membership: %v", err)
	}
	return member != nil, nil
}

// 检查用户是否在其他帮派中
func (h *Handler) isUserInOtherGuild(ctx context.Context, userID int64) (bool, error) {
	userGuilds, err := h.dbClient.GetUserGuilds(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check user guilds: %v", err)
	}
	return len(userGuilds) > 0, nil
}

// 检查帮派名是否已存在
func (h *Handler) isGuildNameExists(ctx context.Context, name string) (bool, error) {
	existingGuild, err := h.dbClient.GetGuildByName(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to check guild name: %v", err)
	}
	return existingGuild != nil, nil
}

// CreateGuild 创建帮派
func (h *Handler) CreateGuild(ctx context.Context, req *pb.CreateGuildRequest) (*pb.CreateGuildResponse, error) {
	// 1. 检查创建者是否已在其他帮派
	inOtherGuild, err := h.isUserInOtherGuild(ctx, req.CreatorId)
	if err != nil {
		return nil, err
	}
	if inOtherGuild {
		return nil, errors.New("user already in another guild")
	}

	// 2. 检查帮派名是否已存在
	nameExists, err := h.isGuildNameExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if nameExists {
		return nil, errors.New("guild name already exists")
	}

	// 创建帮派对象
	guildId, err := h.sf.NextID()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	newGuild := &models.Guild{
		ID:           guildId,
		Name:         req.Name,
		Description:  req.Description,
		Announcement: req.Announcement,
		MasterID:     req.CreatorId,
		CreatedAt:    now,
		MaxMembers:   100, // 默认最大成员数
	}

	// 自动将创建者添加为帮派成员
	memberId, err := h.sf.NextID()
	if err != nil {
		return nil, err
	}
	member := &models.GuildMember{
		ID:        memberId,
		GuildID:   newGuild.ID,
		UserID:    req.CreatorId,
		Role:      models.GuildRoleMaster,
		JoinTime:  now,
		LastLogin: now,
	}

	err = h.dbClient.CreateGuildWithMaster(ctx, newGuild, member)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildCache(ctx, newGuild.ID)
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, newGuild.ID)

	return &pb.CreateGuildResponse{
		Success: true,
		GuildId: newGuild.ID,
	}, nil
}

// GetGuildInfo 获取帮派信息
func (h *Handler) GetGuildInfo(ctx context.Context, req *pb.GetGuildInfoRequest) (*pb.GetGuildInfoResponse, error) {
	guild, err := h.cacheService.GetGuildWithCache(ctx, req.GuildId, func() (*models.Guild, error) {
		return h.dbClient.GetGuild(ctx, req.GuildId)
	})
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, fmt.Errorf("guild not found")
	}
	pbGuild := guildToPb(guild)
	return &pb.GetGuildInfoResponse{
		Guild: pbGuild,
	}, nil
}

// GetGuildMembers 获取帮派成员列表
func (h *Handler) GetGuildMembers(ctx context.Context, req *pb.GetGuildMembersRequest) (*pb.GetGuildMembersResponse, error) {
	resp, err := h.cacheService.GetGuildMembersWithCache(ctx, req.GuildId, func() (*pb.GetGuildMembersResponse, error) {
		dbMembers, err := h.dbClient.GetGuildMembers(ctx, req.GuildId)
		if err != nil {
			return nil, err
		}
		pbMembers := make([]*pb.GuildMember, len(dbMembers))
		for i, m := range dbMembers {
			pbMembers[i] = guildMemberToPb(m)
		}
		return &pb.GetGuildMembersResponse{
			Members: pbMembers,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetGuildList 获取帮派列表(分页)
func (h *Handler) GetGuildList(ctx context.Context, req *pb.GetGuildListRequest) (*pb.GetGuildListResponse, error) {
	// 从数据库获取分页数据
	guilds, total, err := h.dbClient.GetGuildList(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}

	// 转换为protobuf格式
	pbGuilds := make([]*pb.GuildInfo, len(guilds))
	for i, g := range guilds {
		pbGuilds[i] = guildToPb(g)
	}

	return &pb.GetGuildListResponse{
		Guilds:   pbGuilds,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DisbandGuild 解散帮派
func (h *Handler) DisbandGuild(ctx context.Context, req *pb.DisbandGuildRequest) (*pb.DisbandGuildResponse, error) {
	guild, err := h.dbClient.GetGuild(ctx, req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %v", err)
	}
	if guild == nil {
		return nil, errors.New("user not in guild")
	}

	// 获取分布式锁 - 锁定整个帮派
	lockKey := fmt.Sprintf("guild:%d:disband:lock", guild.ID)
	if err := h.cacheClient.Lock(ctx, lockKey, 30*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 检查操作者角色是否是帮主
	master, err := h.dbClient.GetGuildMember(ctx, guild.ID, guild.MasterID)
	if err != nil {
		return nil, err
	}
	if master == nil || master.Role != models.GuildRoleMaster {
		return nil, errors.New("only guild master can disband guild")
	}

	// 删除所有成员
	members, err := h.dbClient.GetGuildMembers(ctx, guild.ID)
	if err != nil {
		return nil, err
	}
	for _, m := range members {
		err = h.dbClient.RemoveGuildMember(ctx, guild.ID, m.UserID)
		if err != nil {
			return nil, err
		}
	}

	// 删除帮派
	err = h.dbClient.DeleteGuild(ctx, guild.ID)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildCache(ctx, guild.ID)
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, guild.ID)
	_ = h.cacheService.InvalidateGuildApplicationsCache(ctx, guild.ID)

	return &pb.DisbandGuildResponse{
		Success: true,
	}, nil
}

// 类型转换函数
func guildToPb(g *models.Guild) *pb.GuildInfo {
	if g == nil {
		return nil
	}
	return &pb.GuildInfo{
		Id:           g.ID,
		Name:         g.Name,
		Description:  g.Description,
		Announcement: g.Announcement,
		CreatedAt:    g.CreatedAt.Unix(),
		MasterId:     g.MasterID,
		MaxMembers:   g.MaxMembers,
	}
}

func guildMemberToPb(m *models.GuildMember) *pb.GuildMember {
	if m == nil {
		return nil
	}
	return &pb.GuildMember{
		UserId:         m.UserID,
		Username:       "", // 需要从用户服务获取
		Role:           pb.GuildRole(m.Role),
		JoinTime:       m.JoinTime.Unix(),
		LastActiveTime: m.LastLogin.Unix(),
	}
}

func guildApplicationToPb(a *models.GuildApplication) *pb.GuildApplication {
	if a == nil {
		return nil
	}
	return &pb.GuildApplication{
		Id:        a.ID,
		UserId:    a.UserID,
		Username:  "", // 需要从用户服务获取
		GuildId:   a.GuildID,
		ApplyTime: a.Time.Unix(),
	}
}
