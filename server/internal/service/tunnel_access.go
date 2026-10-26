package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// Viewer 当前操作的管理员身份，用于穿透记录与请求日志的数据隔离。
// 超级管理员可见全部记录；普通管理员只能看到自己创建的记录及其日志。
// 升级前创建的历史记录 created_by = 0，不属于任何普通管理员，仅超级管理员可见。
type Viewer struct {
	AdminID uint
	Super   bool
}

// Owns 判断该管理员是否可访问创建者为 createdBy 的记录。
func (v Viewer) Owns(createdBy uint) bool {
	if v.Super {
		return true
	}
	return createdBy != 0 && createdBy == v.AdminID
}

// ownerFilter 普通管理员返回自身 ID 作为创建者过滤条件，超级管理员不过滤。
func (v Viewer) ownerFilter() *uint {
	if v.Super {
		return nil
	}
	id := v.AdminID
	return &id
}

// resolveViewer 按管理员 ID 读取角色；账号不存在（如已被删除）时拒绝访问。
func resolveViewer(ctx context.Context, admins repository.AdminUserRepository, adminID uint) (Viewer, error) {
	if admins == nil {
		return Viewer{}, ErrForbidden
	}
	u, err := admins.FindByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Viewer{}, ErrForbidden
		}
		return Viewer{}, err
	}
	return Viewer{AdminID: u.ID, Super: u.IsSuperAdmin()}, nil
}

// Viewer 解析当前管理员的数据可见范围。
func (s *TunnelService) Viewer(ctx context.Context, adminID uint) (Viewer, error) {
	return resolveViewer(ctx, s.admins, adminID)
}

// CheckAccess 校验管理员可访问指定穿透记录；无权或不存在都返回 ErrNotFound，避免泄露他人记录是否存在。
func (s *TunnelService) CheckAccess(ctx context.Context, adminID, tunnelID uint) error {
	_, err := s.Get(ctx, adminID, tunnelID)
	return err
}

// CanSee 供 SSE 推送过滤：判断该身份能否看到某条穿透记录的事件（含已删除记录）。
func (s *TunnelService) CanSee(ctx context.Context, v Viewer, tunnelID uint) bool {
	if v.Super {
		return true
	}
	owner, err := s.tunnels.OwnerOf(ctx, tunnelID)
	if err != nil {
		return false
	}
	return v.Owns(owner)
}

// fillCreatorNames 超级管理员视角下补充创建者登录名，失败时静默跳过，不影响列表。
func (s *TunnelService) fillCreatorNames(ctx context.Context, list []model.Tunnel) {
	if s.admins == nil || len(list) == 0 {
		return
	}
	users, err := s.admins.List(ctx)
	if err != nil {
		return
	}
	names := make(map[uint]string, len(users))
	for _, u := range users {
		names[u.ID] = u.Username
	}
	for i := range list {
		list[i].CreatorName = names[list[i].CreatedBy]
	}
}
