package social

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// ApplyToGuild 申请加入帮派
func (h *Handler) ApplyToGuild(ctx context.Context, req *pb.ApplyToGuildRequest) (*pb.ApplyToGuildResponse, error) {
	// 检查用户是否已在帮派中
	inGuild, err := h.isUserInGuild(ctx, req.GuildId, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check guild membership: %v", err)
	}
	if inGuild {
		return nil, errors.New("user already in this guild")
	}

	// 检查用户是否在其他帮派
	inOtherGuild, err := h.isUserInOtherGuild(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check user guilds: %v", err)
	}
	if inOtherGuild {
		return nil, errors.New("user already in another guild")
	}

	id, err := h.sf.NextID()
	if err != nil {
		return nil, err
	}
	app := &models.GuildApplication{
		ID:         id,
		GuildID:    req.GuildId,
		UserID:     req.UserId,
		Time:       time.Now(),
		ExpireTime: time.Now().Add(7 * 24 * time.Hour), // 7天有效期
	}
	err = h.dbClient.CreateGuildApplication(ctx, app)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildApplicationsCache(ctx, req.GuildId)

	return &pb.ApplyToGuildResponse{
		Success: true,
	}, nil
}

// InviteToGuild 邀请用户加入帮派
func (h *Handler) InviteToGuild(ctx context.Context, req *pb.InviteToGuildRequest) (*pb.InviteToGuildResponse, error) {
	// 检查邀请者是否有权限邀请
	member, err := h.dbClient.GetGuildMember(ctx, req.GuildId, req.InviterId)
	if err != nil {
		return nil, err
	}
	if member == nil || (member.Role != models.GuildRoleMaster && member.Role != models.GuildRoleViceMaster && member.Role != models.GuildRoleElder) {
		return nil, errors.New("no permission to invite")
	}

	// 检查被邀请者是否已在帮派中
	inGuild, err := h.isUserInGuild(ctx, req.GuildId, req.InviteeId)
	if err != nil {
		return nil, fmt.Errorf("failed to check invitee membership: %v", err)
	}
	if inGuild {
		return nil, errors.New("invitee already in this guild")
	}

	// 检查被邀请者是否在其他帮派
	inOtherGuild, err := h.isUserInOtherGuild(ctx, req.InviteeId)
	if err != nil {
		return nil, fmt.Errorf("failed to check invitee guilds: %v", err)
	}
	if inOtherGuild {
		return nil, errors.New("invitee already in another guild")
	}

	id, err := h.sf.NextID()
	if err != nil {
		return nil, err
	}
	inv := &models.GuildInvitation{
		ID:         id,
		GuildID:    req.GuildId,
		InviterID:  req.InviterId,
		UserID:     req.InviteeId,
		Time:       time.Now(),
		ExpireTime: time.Now().Add(7 * 24 * time.Hour), // 7天有效期
	}
	err = h.dbClient.CreateGuildInvitation(ctx, inv)
	if err != nil {
		return nil, err
	}
	return &pb.InviteToGuildResponse{
		Success: true,
	}, nil
}

// GetGuildApplications 获取帮派申请列表
func (h *Handler) GetGuildApplications(ctx context.Context, req *pb.GetGuildApplicationsRequest) (*pb.GetGuildApplicationsResponse, error) {
	applications, err := h.cacheService.GetGuildApplicationsWithCache(ctx, req.GuildId, func() ([]*models.GuildApplication, error) {
		return h.dbClient.GetGuildApplications(ctx, req.GuildId)
	})
	if err != nil {
		return nil, err
	}

	pbApps := make([]*pb.GuildApplication, len(applications))
	for i, a := range applications {
		pbApps[i] = guildApplicationToPb(a)
	}

	return &pb.GetGuildApplicationsResponse{
		Applications: pbApps,
	}, nil
}

// HandleGuildApplication 处理帮派申请
func (h *Handler) HandleGuildApplication(ctx context.Context, req *pb.HandleGuildApplicationRequest) (*pb.HandleGuildApplicationResponse, error) {
	// 获取分布式锁
	lockKey := fmt.Sprintf("guild:app:%d:lock", req.ApplicationId)
	if err := h.cacheClient.Lock(ctx, lockKey, 10*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 1. 获取申请信息
	app, err := h.dbClient.GetGuildApplication(ctx, req.ApplicationId)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, errors.New("application not found")
	}

	// 检查申请人是否已在帮派中
	inGuild, err := h.isUserInGuild(ctx, app.GuildID, app.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check applicant membership: %v", err)
	}
	if inGuild {
		return nil, errors.New("applicant already in guild")
	}

	// 检查帮派成员数是否已达上限
	count, err := h.dbClient.GetGuildMemberCount(ctx, app.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild member count: %v", err)
	}
	guild, err := h.dbClient.GetGuild(ctx, app.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild info: %v", err)
	}
	if count >= guild.MaxMembers {
		return nil, errors.New("guild is full")
	}

	// 2. 检查操作者是否有权限处理申请
	member, err := h.dbClient.GetGuildMember(ctx, app.GuildID, req.OperatorId)
	if err != nil {
		return nil, err
	}
	if member == nil || (member.Role != models.GuildRoleMaster && member.Role != models.GuildRoleViceMaster && member.Role != models.GuildRoleElder) {
		return nil, errors.New("no permission to handle application")
	}

	// 3. 更新申请状态
	switch req.Action {
	case pb.HandleGuildApplicationRequest_ACCEPT:
		id, err := h.sf.NextID()
		if err != nil {
			return nil, err
		}
		// 如果接受申请，添加用户为帮派成员
		newMember := &models.GuildMember{
			ID:        id,
			GuildID:   app.GuildID,
			UserID:    app.UserID,
			Role:      models.GuildRoleMember,
			JoinTime:  time.Now(),
			LastLogin: time.Now(),
		}
		if err := h.dbClient.AddGuildMember(ctx, newMember); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid action")
	}

	// 4. 更新申请状态
	if err := h.dbClient.DeleteGuildApplication(ctx, req.ApplicationId); err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildApplicationsCache(ctx, app.GuildID)
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, app.GuildID)

	return &pb.HandleGuildApplicationResponse{
		Success: true,
	}, nil
}
