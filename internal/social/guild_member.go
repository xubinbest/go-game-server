package social

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
)

// KickGuildMember 踢出帮派成员
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

// ChangeMemberRole 修改成员职位
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

// TransferGuildMaster 转让帮主
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

// LeaveGuild 离开帮派
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
