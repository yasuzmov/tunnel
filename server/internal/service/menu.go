package service

import (
	"fmt"
	"slices"
	"strings"

	"tunnel-server/internal/model"
)

// 左侧菜单权限键。与前端 web/src/utils/menus.ts 保持一致。
const (
	MenuDashboard = "dashboard"
	MenuTunnels   = "tunnels"
	MenuDomains   = "domains"
	MenuLogs      = "logs"
	MenuAdmins    = "admins"
)

// AllMenus 超级管理员可见的全部菜单。
var AllMenus = []string{MenuDashboard, MenuTunnels, MenuDomains, MenuLogs, MenuAdmins}

// GrantableMenus 普通管理员可被授予的菜单。
// 「管理员管理」涉及创建超级管理员与重置口令，只允许超级管理员使用，不参与授权。
var GrantableMenus = []string{MenuDashboard, MenuTunnels, MenuDomains, MenuLogs}

// DefaultAdminMenus 未配置菜单权限的普通管理员默认可见菜单，保持升级前行为。
var DefaultAdminMenus = []string{MenuDashboard, MenuTunnels, MenuLogs}

// EffectiveMenus 计算管理员实际生效的菜单。
// 超级管理员始终拥有全部菜单；普通管理员未配置时回落到默认菜单。
func EffectiveMenus(role string, stored []string) []string {
	if role == model.RoleSuperAdmin {
		return slices.Clone(AllMenus)
	}
	clean := sanitizeMenus(stored)
	if len(clean) == 0 {
		return slices.Clone(DefaultAdminMenus)
	}
	return clean
}

// NormalizeMenus 校验并整理前端提交的菜单权限，按固定顺序返回并去重。
func NormalizeMenus(raw []string) ([]string, error) {
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		key := strings.TrimSpace(item)
		if !slices.Contains(GrantableMenus, key) {
			return nil, fmt.Errorf("%w: 未知的菜单权限 %q", ErrBadInput, item)
		}
		seen[key] = struct{}{}
	}
	menus := make([]string, 0, len(seen))
	for _, key := range GrantableMenus {
		if _, ok := seen[key]; ok {
			menus = append(menus, key)
		}
	}
	return menus, nil
}

// sanitizeMenus 忽略未知键，避免菜单下线后旧数据导致登录失败。
func sanitizeMenus(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		key := strings.TrimSpace(item)
		if slices.Contains(GrantableMenus, key) {
			seen[key] = struct{}{}
		}
	}
	menus := make([]string, 0, len(seen))
	for _, key := range GrantableMenus {
		if _, ok := seen[key]; ok {
			menus = append(menus, key)
		}
	}
	return menus
}
