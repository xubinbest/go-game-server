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

// 帮派相关方法
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

func (h *Handler) KickGuildMember(ctx context.Context, req *pb.KickGuildMemberRequest) (*pb.KickGuildMemberResponse, error) {
	// 获取分布式锁
	lockKey := fmt.Sprintf("guild:%d:kick:%d:lock", req.GuildId, req.MemberId)
	if err := h.cacheClient.Lock(ctx, lockKey, 10*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 检查操作者是否有权限踢人
	operator, err := h.dbClient.GetGuildMember(ctx, req.GuildId, req.OperatorId)
	if err != nil {
		return nil, err
	}
	if operator == nil || (operator.Role != models.GuildRoleMaster && operator.Role != models.GuildRoleViceMaster && operator.Role != models.GuildRoleElder) {
		return nil, errors.New("no permission to kick member")
	}

	// 检查被踢成员是否在帮派中
	inGuild, err := h.isUserInGuild(ctx, req.GuildId, req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("failed to check member status: %v", err)
	}
	if !inGuild {
		return nil, errors.New("member not in guild")
	}

	// 检查被踢成员是否是帮主(帮主不能被踢)
	member, err := h.dbClient.GetGuildMember(ctx, req.GuildId, req.MemberId)
	if err != nil {
		return nil, err
	}
	if member != nil && member.Role == models.GuildRoleMaster {
		return nil, errors.New("cannot kick guild master")
	}

	err = h.dbClient.RemoveGuildMember(ctx, req.GuildId, req.MemberId)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, req.GuildId)

	return &pb.KickGuildMemberResponse{
		Success: true,
	}, nil
}

func (h *Handler) ChangeMemberRole(ctx context.Context, req *pb.ChangeMemberRoleRequest) (*pb.ChangeMemberRoleResponse, error) {
	// 获取分布式锁
	lockKey := fmt.Sprintf("guild:%d:role:%d:lock", req.GuildId, req.MemberId)
	if err := h.cacheClient.Lock(ctx, lockKey, 10*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 检查操作者是否有权限修改成员职位
	operator, err := h.dbClient.GetGuildMember(ctx, req.GuildId, req.OperatorId)
	if err != nil {
		return nil, err
	}
	if operator == nil || (operator.Role != models.GuildRoleMaster && operator.Role != models.GuildRoleViceMaster) {
		return nil, errors.New("no permission to change member role")
	}

	// 检查成员是否在帮派中
	inGuild, err := h.isUserInGuild(ctx, req.GuildId, req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("failed to check member status: %v", err)
	}
	if !inGuild {
		return nil, errors.New("member not in guild")
	}

	// 检查成员角色是否有效
	if req.NewRole < pb.GuildRole_MEMBER || req.NewRole > pb.GuildRole_VICE_MASTER {
		return nil, errors.New("invalid role")
	}

	// 直接使用pb.GuildRole值作为int
	err = h.dbClient.UpdateGuildMemberRole(ctx, req.GuildId, req.MemberId, int(req.NewRole))
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, req.GuildId)

	return &pb.ChangeMemberRoleResponse{
		Success: true,
	}, nil
}

func (h *Handler) TransferGuildMaster(ctx context.Context, req *pb.TransferGuildMasterRequest) (*pb.TransferGuildMasterResponse, error) {
	// 获取当前用户帮派信息
	guild, err := h.dbClient.GetGuild(ctx, req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %v", err)
	}
	if guild == nil {
		return nil, errors.New("user not in guild")
	}

	// 检查当前用户是否是帮主
	currentMaster, err := h.dbClient.GetGuildMember(ctx, guild.ID, guild.MasterID)
	if err != nil {
		return nil, err
	}
	if currentMaster == nil || currentMaster.Role != models.GuildRoleMaster {
		return nil, errors.New("only guild master can transfer master")
	}

	// 获取分布式锁 - 锁定整个帮派和两个用户
	lockKey := fmt.Sprintf("guild:%d:transfer:%d:%d:lock", guild.ID, guild.MasterID, req.NewMasterId)
	if err := h.cacheClient.Lock(ctx, lockKey, 15*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 检查新帮主是否在帮派中
	inGuild, err := h.isUserInGuild(ctx, guild.ID, req.NewMasterId)
	if err != nil {
		return nil, fmt.Errorf("failed to check new master status: %v", err)
	}
	if !inGuild {
		return nil, errors.New("new master is not a guild member")
	}

	// 检查新帮主是否已经是帮主
	newMaster, err := h.dbClient.GetGuildMember(ctx, guild.ID, req.NewMasterId)
	if err != nil {
		return nil, err
	}
	if newMaster != nil && newMaster.Role == models.GuildRoleMaster {
		return nil, errors.New("new master is already the guild master")
	}

	// 更新原帮主角色
	err = h.dbClient.UpdateGuildMemberRole(ctx, guild.ID, guild.MasterID, models.GuildRoleViceMaster)
	if err != nil {
		return nil, err
	}

	// 更新新帮主角色
	err = h.dbClient.UpdateGuildMemberRole(ctx, guild.ID, req.NewMasterId, models.GuildRoleMaster)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildCache(ctx, guild.ID)
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, guild.ID)

	return &pb.TransferGuildMasterResponse{
		Success: true,
	}, nil
}

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

func (h *Handler) LeaveGuild(ctx context.Context, req *pb.LeaveGuildRequest) (*pb.LeaveGuildResponse, error) {
	// 获取当前用户帮派信息
	guild, err := h.dbClient.GetGuild(ctx, req.RoleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %v", err)
	}
	if guild == nil {
		return nil, errors.New("user not in guild")
	}

	// 获取分布式锁 - 锁定帮派和用户
	lockKey := fmt.Sprintf("guild:%d:leave:%d:lock", guild.ID, guild.MasterID)
	if err := h.cacheClient.Lock(ctx, lockKey, 10*time.Second, 5*time.Second); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %v", err)
	}
	defer h.cacheClient.Unlock(ctx, lockKey)

	// 检查成员是否在帮派中
	inGuild, err := h.isUserInGuild(ctx, guild.ID, guild.MasterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check member status: %v", err)
	}
	if !inGuild {
		return nil, errors.New("member not in guild")
	}

	// 检查是否是帮主(帮主不能直接离开，需要转让帮主)
	member, err := h.dbClient.GetGuildMember(ctx, guild.ID, guild.MasterID)
	if err != nil {
		return nil, err
	}
	if member != nil && member.Role == models.GuildRoleMaster {
		return nil, errors.New("guild master cannot leave, please transfer master first")
	}

	err = h.dbClient.RemoveGuildMember(ctx, guild.ID, guild.MasterID)
	if err != nil {
		return nil, err
	}

	// 失效相关缓存
	_ = h.cacheService.InvalidateGuildMembersCache(ctx, guild.ID)

	return &pb.LeaveGuildResponse{
		Success: true,
	}, nil
}

// 获取帮派列表(分页)
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
