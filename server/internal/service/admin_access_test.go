package service

import (
	"context"
	"errors"
	"slices"
	"testing"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

// fakeAdminStore 内存管理员表，供权限规则单测使用。
type fakeAdminStore struct {
	users map[uint]*model.AdminUser
	menus map[uint][]string
}

func (f *fakeAdminStore) FindByUsername(_ context.Context, username string) (*model.AdminUser, error) {
	for _, u := range f.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeAdminStore) FindByID(_ context.Context, id uint) (*model.AdminUser, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (f *fakeAdminStore) Create(_ context.Context, u *model.AdminUser) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeAdminStore) UpdateProfile(_ context.Context, id uint, username, role string) error {
	u, ok := f.users[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	u.Username = username
	u.Role = role
	return nil
}

func (f *fakeAdminStore) UpdatePassword(_ context.Context, id uint, hash string) error {
	u, ok := f.users[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	u.PasswordHash = hash
	return nil
}

func (f *fakeAdminStore) Delete(_ context.Context, id uint) error {
	if _, ok := f.users[id]; !ok {
		return gorm.ErrRecordNotFound
	}
	delete(f.users, id)
	return nil
}

func (f *fakeAdminStore) List(_ context.Context) ([]model.AdminUser, error) {
	list := make([]model.AdminUser, 0, len(f.users))
	for _, u := range f.users {
		list = append(list, *u)
	}
	return list, nil
}

func (f *fakeAdminStore) CountByRole(_ context.Context, role string) (int64, error) {
	var n int64
	for _, u := range f.users {
		if u.Role == role {
			n++
		}
	}
	return n, nil
}

func (f *fakeAdminStore) ListMenus(_ context.Context, adminID uint) ([]string, error) {
	return f.menus[adminID], nil
}

func (f *fakeAdminStore) SetMenus(_ context.Context, adminID uint, menus []string) error {
	if f.menus == nil {
		f.menus = map[uint][]string{}
	}
	f.menus[adminID] = menus
	return nil
}

func TestAdminDeleteProtectsLastSuper(t *testing.T) {
	admins := &fakeAdminStore{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
	}}
	svc := NewAdminService(admins)
	ctx := context.Background()

	if err := svc.Delete(ctx, 1, 2); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("删除最后一个超级管理员应被拒绝，got %v", err)
	}
	if err := svc.Delete(ctx, 1, 1); !errors.Is(err, ErrBadInput) {
		t.Fatalf("删除自己应被拒绝，got %v", err)
	}
	if _, err := svc.Update(ctx, 1, UpdateAdminInput{Username: "root", Role: model.RoleAdmin}); !errors.Is(err, ErrLastSuperAdmin) {
		t.Fatalf("降级最后一个超级管理员应被拒绝，got %v", err)
	}
}

func TestEffectiveMenus(t *testing.T) {
	cases := []struct {
		name   string
		role   string
		stored []string
		want   []string
	}{
		{"超级管理员拥有全部菜单", model.RoleSuperAdmin, nil, AllMenus},
		{"普通管理员未配置时使用默认菜单", model.RoleAdmin, nil, DefaultAdminMenus},
		{"普通管理员按授权菜单生效", model.RoleAdmin, []string{MenuLogs, MenuDomains}, []string{MenuDomains, MenuLogs}},
		{"忽略未知与管理员菜单", model.RoleAdmin, []string{MenuTunnels, MenuAdmins, "unknown"}, []string{MenuTunnels}},
		{"只有未知菜单时回落到默认", model.RoleAdmin, []string{"unknown"}, DefaultAdminMenus},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EffectiveMenus(tc.role, tc.stored)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}
}

func TestAdminMenuPermissions(t *testing.T) {
	admins := &fakeAdminStore{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
	}}
	svc := NewAdminService(admins)
	ctx := context.Background()

	allowed, err := svc.HasMenu(ctx, 2, MenuTunnels)
	if err != nil || !allowed {
		t.Fatalf("普通管理员默认应拥有穿透记录菜单, allowed=%v err=%v", allowed, err)
	}
	allowed, err = svc.HasMenu(ctx, 2, MenuDomains)
	if err != nil || allowed {
		t.Fatalf("普通管理员默认不应拥有域名管理菜单, allowed=%v err=%v", allowed, err)
	}

	menus := []string{MenuDomains, MenuDashboard}
	if _, err := svc.Update(ctx, 2, UpdateAdminInput{Username: "ops", Menus: &menus}); err != nil {
		t.Fatalf("更新菜单权限失败: %v", err)
	}
	allowed, _ = svc.HasMenu(ctx, 2, MenuDomains)
	if !allowed {
		t.Fatal("更新后应拥有域名管理菜单")
	}
	allowed, _ = svc.HasMenu(ctx, 2, MenuTunnels)
	if allowed {
		t.Fatal("更新后不应再拥有穿透记录菜单")
	}

	empty := []string{}
	if _, err := svc.Update(ctx, 2, UpdateAdminInput{Username: "ops", Menus: &empty}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("空菜单应被拒绝, got %v", err)
	}
	unknown := []string{"unknown"}
	if _, err := svc.Update(ctx, 2, UpdateAdminInput{Username: "ops", Menus: &unknown}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("未知菜单应被拒绝, got %v", err)
	}

	allowed, err = svc.HasMenu(ctx, 1, MenuAdmins)
	if err != nil || !allowed {
		t.Fatalf("超级管理员应始终拥有管理员管理菜单, allowed=%v err=%v", allowed, err)
	}
}
