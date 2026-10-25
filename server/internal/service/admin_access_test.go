package service

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// fakeAdminStore 内存管理员表，供权限规则单测使用。
type fakeAdminStore struct {
	users map[uint]*model.AdminUser
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

// fakePoolStore 内存域名池与授权关系。
type fakePoolStore struct {
	domains []model.Domain
	perms   map[uint][]uint
}

func (f *fakePoolStore) Create(_ context.Context, d *model.Domain) error {
	f.domains = append(f.domains, *d)
	return nil
}

func (f *fakePoolStore) UpdateRemark(_ context.Context, id uint, remark string) error {
	for i := range f.domains {
		if f.domains[i].ID == id {
			f.domains[i].Remark = remark
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakePoolStore) Delete(_ context.Context, id uint) error {
	for i := range f.domains {
		if f.domains[i].ID == id {
			f.domains = append(f.domains[:i], f.domains[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakePoolStore) FindByID(_ context.Context, id uint) (*model.Domain, error) {
	for i := range f.domains {
		if f.domains[i].ID == id {
			return &f.domains[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePoolStore) FindByDomain(_ context.Context, domain string) (*model.Domain, error) {
	for i := range f.domains {
		if f.domains[i].Domain == domain {
			return &f.domains[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePoolStore) List(_ context.Context, _ string) ([]model.Domain, error) {
	return f.domains, nil
}

func (f *fakePoolStore) ListByAdmin(_ context.Context, adminID uint) ([]model.Domain, error) {
	var list []model.Domain
	for _, id := range f.perms[adminID] {
		for _, d := range f.domains {
			if d.ID == id {
				list = append(list, d)
			}
		}
	}
	return list, nil
}

func (f *fakePoolStore) ListAdminDomainIDs(_ context.Context, adminID uint) ([]uint, error) {
	return f.perms[adminID], nil
}

func (f *fakePoolStore) SetAdminDomains(_ context.Context, adminID uint, domainIDs []uint) error {
	if f.perms == nil {
		f.perms = map[uint][]uint{}
	}
	f.perms[adminID] = domainIDs
	return nil
}

func (f *fakePoolStore) ListBindings(_ context.Context) ([]repository.DomainBinding, error) {
	return nil, nil
}

func TestCheckDomainAllowed(t *testing.T) {
	admins := &fakeAdminStore{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
	}}
	pool := &fakePoolStore{
		domains: []model.Domain{
			{ID: 10, Domain: "app.example.com"},
			{ID: 11, Domain: "api.example.com"},
		},
		perms: map[uint][]uint{2: {10}},
	}
	svc := &DomainService{pool: pool, admins: admins}
	cases := []struct {
		name    string
		adminID uint
		host    string
		want    error
	}{
		{"超级管理员可用全部域名", 1, "api.example.com", nil},
		{"普通管理员可用已授权域名", 2, "app.example.com", nil},
		{"普通管理员不可用未授权域名", 2, "api.example.com", ErrDomainNotAllowed},
		{"域名不在池中", 2, "other.example.com", ErrDomainNotInPool},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.checkDomainAllowed(context.Background(), tc.adminID, tc.host)
			if tc.want == nil {
				if err != nil {
					t.Fatalf("want nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestAdminDeleteProtectsLastSuper(t *testing.T) {
	admins := &fakeAdminStore{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
	}}
	svc := NewAdminService(admins, &fakePoolStore{})
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
