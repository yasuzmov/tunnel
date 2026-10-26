package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// fakePoolStore 内存域名池，按 created_by 过滤模拟仓储查询。
type fakePoolStore struct {
	domains []model.Domain
	nextID  uint
}

func (f *fakePoolStore) Create(_ context.Context, d *model.Domain) error {
	f.nextID++
	if d.ID == 0 {
		d.ID = 1000 + f.nextID
	}
	f.domains = append(f.domains, *d)
	return nil
}

func (f *fakePoolStore) Update(_ context.Context, id uint, domain, remark string) error {
	for i := range f.domains {
		if f.domains[i].ID == id {
			f.domains[i].Domain = domain
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
			cp := f.domains[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePoolStore) FindByDomain(_ context.Context, domain string) (*model.Domain, error) {
	for i := range f.domains {
		if f.domains[i].Domain == domain {
			cp := f.domains[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakePoolStore) List(_ context.Context, _ string, createdBy *uint) ([]model.Domain, error) {
	var list []model.Domain
	for _, d := range f.domains {
		if createdBy == nil || d.CreatedBy == *createdBy {
			list = append(list, d)
		}
	}
	return list, nil
}

func (f *fakePoolStore) ListBindings(_ context.Context) ([]repository.DomainBinding, error) {
	return nil, nil
}

// boundDomains 内存穿透域名绑定表，只实现域名池与可用域名需要的查询。
type boundDomains struct {
	noDomains
	items []model.TunnelDomain
}

func (b *boundDomains) FindByDomain(_ context.Context, domain string) (*model.TunnelDomain, error) {
	for i := range b.items {
		if b.items[i].Domain == domain {
			cp := b.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (b *boundDomains) ListAll(context.Context) ([]model.TunnelDomain, error) {
	return b.items, nil
}

const testBase = "fish.czypro.cn"

// 管理员：1 超级管理员；2、3 普通管理员。
// 域名：20 历史无归属；21、22 属于 2（22 已绑定）；23 属于 3；24 属于 1。
func poolFixture() (*fakeAdminStore, *fakePoolStore, *boundDomains) {
	admins, _ := ownershipFixture()
	pool := &fakePoolStore{domains: []model.Domain{
		{ID: 20, Domain: "legacy.example.com", CreatedBy: 0},
		{ID: 21, Domain: "ops." + testBase, CreatedBy: 2},
		{ID: 22, Domain: "ops-used." + testBase, CreatedBy: 2},
		{ID: 23, Domain: "dev." + testBase, CreatedBy: 3},
		{ID: 24, Domain: "root." + testBase, CreatedBy: 1},
	}}
	bound := &boundDomains{items: []model.TunnelDomain{
		{ID: 1, TunnelID: 11, Domain: "ops-used." + testBase},
	}}
	return admins, pool, bound
}

func newPoolSvc(admins *fakeAdminStore, pool *fakePoolStore, bound *boundDomains) *DomainPoolService {
	return NewDomainPoolService(DomainPoolDeps{Pool: pool, Bound: bound, Admins: admins, BaseDomain: testBase})
}

func viewIDs(list []DomainPoolView) []uint {
	ids := make([]uint, 0, len(list))
	for _, v := range list {
		ids = append(ids, v.ID)
	}
	slices.Sort(ids)
	return ids
}

func TestDomainPoolListIsolatedByCreator(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := newPoolSvc(admins, pool, bound)
	ctx := context.Background()

	cases := []struct {
		name    string
		adminID uint
		want    []uint
	}{
		{"超级管理员可见全部（含历史域名）", 1, []uint{20, 21, 22, 23, 24}},
		{"普通管理员只见自己创建的", 2, []uint{21, 22}},
		{"另一普通管理员只见自己的", 3, []uint{23}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := svc.List(ctx, tc.adminID)
			if err != nil {
				t.Fatal(err)
			}
			if got := viewIDs(list); !slices.Equal(got, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, got)
			}
		})
	}

	list, _ := svc.List(ctx, 1)
	for _, v := range list {
		if v.ID == 21 && (v.CreatorName != "ops" || v.Name != "ops") {
			t.Fatalf("超级管理员视角应带创建者与前缀: %+v", v)
		}
		if v.ID == 20 && v.Name != "" {
			t.Fatalf("历史非标准域名前缀应为空: %+v", v)
		}
	}
	list, _ = svc.List(ctx, 2)
	for _, v := range list {
		if v.CreatorName != "" {
			t.Fatalf("普通管理员视角不应返回创建者名称: %+v", v)
		}
	}
	if _, err := svc.List(ctx, 99); !errors.Is(err, ErrForbidden) {
		t.Fatalf("不存在的管理员应被拒绝, got %v", err)
	}
}

func TestDomainPoolOthersReturnNotFound(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := newPoolSvc(admins, pool, bound)
	ctx := context.Background()

	for _, id := range []uint{20, 23, 24, 999} {
		if _, err := svc.Get(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get %d: want ErrNotFound, got %v", id, err)
		}
		if err := svc.Update(ctx, 2, id, UpdatePoolInput{Name: "x1", Remark: "r"}); !errors.Is(err, ErrNotFound) {
			t.Errorf("Update %d: want ErrNotFound, got %v", id, err)
		}
		if err := svc.Delete(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Errorf("Delete %d: want ErrNotFound, got %v", id, err)
		}
	}
	if len(pool.domains) != 5 {
		t.Fatalf("他人域名不应被修改或删除, 剩余 %d", len(pool.domains))
	}
	if d, _ := pool.FindByID(ctx, 23); d.Domain != "dev."+testBase {
		t.Fatalf("他人域名被篡改: %s", d.Domain)
	}

	if _, err := svc.Get(ctx, 2, 21); err != nil {
		t.Fatalf("自己的域名应可读取: %v", err)
	}
	if _, err := svc.Get(ctx, 1, 23); err != nil {
		t.Fatalf("超级管理员应可读取他人域名: %v", err)
	}
	if err := svc.Update(ctx, 1, 20, UpdatePoolInput{Remark: "历史"}); err != nil {
		t.Fatalf("超级管理员应可修改历史域名备注: %v", err)
	}
	if d, _ := pool.FindByID(ctx, 20); d.Domain != "legacy.example.com" || d.Remark != "历史" {
		t.Fatalf("只改备注时不应改动历史域名: %+v", d)
	}
	if err := svc.Delete(ctx, 1, 23); err != nil {
		t.Fatalf("超级管理员应可删除他人未绑定域名: %v", err)
	}
	if err := svc.Delete(ctx, 2, 22); !errors.Is(err, ErrDomainBound) {
		t.Fatalf("已绑定域名不能删除, got %v", err)
	}
	if err := svc.Delete(ctx, 2, 21); err != nil {
		t.Fatalf("应可删除自己未绑定的域名: %v", err)
	}
}

func TestDomainPoolCreateEnforcesSuffix(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := newPoolSvc(admins, pool, bound)
	ctx := context.Background()

	d, err := svc.Create(ctx, 2, CreatePoolInput{Name: "  ABC-1 ", Remark: "测试"})
	if err != nil {
		t.Fatalf("创建失败: %v", err)
	}
	if d.Domain != "abc-1."+testBase || d.Name != "abc-1" || d.CreatedBy != 2 {
		t.Fatalf("应生成 前缀.主域名 且记录创建者: %+v", d)
	}
	if _, err := svc.Create(ctx, 2, CreatePoolInput{Domain: "full." + testBase}); err != nil {
		t.Fatalf("兼容主域名下的完整域名: %v", err)
	}

	bad := []CreatePoolInput{
		{},
		{Name: "-abc"},
		{Name: "abc-"},
		{Name: "a.b"},
		{Name: "a_b"},
		{Name: "中文"},
		{Name: "a b"},
		{Name: "www"},
		{Name: strings.Repeat("a", 64)},
		{Domain: "evil.com"},
		{Domain: "a.b." + testBase},
		{Domain: "x" + testBase},
		{Domain: testBase},
		{Domain: "x." + testBase + ".evil.com"},
	}
	for _, in := range bad {
		if _, err := svc.Create(ctx, 2, in); !errors.Is(err, ErrBadInput) {
			t.Errorf("%+v: want ErrBadInput, got %v", in, err)
		}
	}
	if _, err := svc.Create(ctx, 2, CreatePoolInput{Name: strings.Repeat("a", 63)}); err != nil {
		t.Fatalf("63 位前缀应合法: %v", err)
	}

	// 已被他人占用或已绑定的域名都提示占用，且不透露占用者。
	for _, name := range []string{"dev", "ops", "root"} {
		_, err := svc.Create(ctx, 2, CreatePoolInput{Name: name})
		if !errors.Is(err, ErrDomainPoolTaken) {
			t.Fatalf("%s: want ErrDomainPoolTaken, got %v", name, err)
		}
		if strings.Contains(err.Error(), "dev") || strings.Contains(err.Error(), "root") {
			t.Fatalf("提示不应泄露占用者: %v", err)
		}
	}
	bound.items = append(bound.items, model.TunnelDomain{ID: 2, TunnelID: 10, Domain: "orphan." + testBase})
	if _, err := svc.Create(ctx, 2, CreatePoolInput{Name: "orphan"}); !errors.Is(err, ErrDomainPoolTaken) {
		t.Fatalf("已被穿透记录使用的域名应视为占用, got %v", err)
	}

	noBase := NewDomainPoolService(DomainPoolDeps{Pool: pool, Bound: bound, Admins: admins})
	if _, err := noBase.Create(ctx, 2, CreatePoolInput{Name: "abc"}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("未配置主域名时应拒绝创建, got %v", err)
	}
}

func TestDomainPoolUpdateOnlyPrefix(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := newPoolSvc(admins, pool, bound)
	ctx := context.Background()

	if err := svc.Update(ctx, 2, 21, UpdatePoolInput{Name: "ops2", Remark: "新备注"}); err != nil {
		t.Fatalf("修改前缀失败: %v", err)
	}
	if d, _ := pool.FindByID(ctx, 21); d.Domain != "ops2."+testBase || d.Remark != "新备注" || d.CreatedBy != 2 {
		t.Fatalf("修改结果不符: %+v", d)
	}
	if err := svc.Update(ctx, 2, 21, UpdatePoolInput{Remark: "仅备注"}); err != nil {
		t.Fatalf("只改备注失败: %v", err)
	}
	if d, _ := pool.FindByID(ctx, 21); d.Domain != "ops2."+testBase {
		t.Fatalf("只改备注不应改动域名: %+v", d)
	}
	if err := svc.Update(ctx, 2, 21, UpdatePoolInput{Domain: "other.example.com"}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("不能改成其他后缀, got %v", err)
	}
	if err := svc.Update(ctx, 2, 21, UpdatePoolInput{Name: "dev"}); !errors.Is(err, ErrDomainPoolTaken) {
		t.Fatalf("不能改成他人已占用的前缀, got %v", err)
	}
	if err := svc.Update(ctx, 2, 21, UpdatePoolInput{Name: "ops2"}); err != nil {
		t.Fatalf("前缀不变时应允许保存: %v", err)
	}
	if err := svc.Update(ctx, 2, 22, UpdatePoolInput{Name: "renamed"}); !errors.Is(err, ErrBadInput) {
		t.Fatalf("已绑定域名不能改前缀, got %v", err)
	}
	if err := svc.Update(ctx, 2, 22, UpdatePoolInput{Name: "ops-used", Remark: "可改备注"}); err != nil {
		t.Fatalf("已绑定域名应可改备注: %v", err)
	}
}

func TestAvailableDomainsOwnUnused(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := &DomainService{pool: pool, admins: admins, domains: bound}
	ctx := context.Background()

	cases := []struct {
		name    string
		adminID uint
		want    []uint
	}{
		{"普通管理员只见自己创建且未使用的", 2, []uint{21}},
		{"另一普通管理员", 3, []uint{23}},
		{"超级管理员可见全部未使用的", 1, []uint{20, 21, 23, 24}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			list, err := svc.AvailableDomains(ctx, tc.adminID)
			if err != nil {
				t.Fatal(err)
			}
			ids := make([]uint, 0, len(list))
			for _, d := range list {
				ids = append(ids, d.ID)
			}
			slices.Sort(ids)
			if !slices.Equal(ids, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, ids)
			}
		})
	}
}

func TestCheckDomainAllowed(t *testing.T) {
	admins, pool, bound := poolFixture()
	svc := &DomainService{pool: pool, admins: admins, domains: bound}
	cases := []struct {
		name    string
		adminID uint
		host    string
		want    error
	}{
		{"超级管理员可用他人域名", 1, "dev." + testBase, nil},
		{"超级管理员可用历史域名", 1, "legacy.example.com", nil},
		{"超级管理员使用池外域名", 1, "other.example.com", ErrDomainNotInPool},
		{"普通管理员可用自己的域名", 2, "ops." + testBase, nil},
		{"普通管理员不可用他人域名", 2, "dev." + testBase, ErrDomainNotAllowed},
		{"普通管理员不可用超级管理员的域名", 2, "root." + testBase, ErrDomainNotAllowed},
		{"普通管理员不可用历史域名", 2, "legacy.example.com", ErrDomainNotAllowed},
		{"普通管理员使用池外域名", 2, "other.example.com", ErrDomainNotAllowed},
		{"已删除的管理员", 99, "ops." + testBase, ErrForbidden},
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

func TestDomainBindOnlyOwnUnused(t *testing.T) {
	admins, pool, bound := poolFixture()
	now := time.Now()
	tunnels := newOwnerTunnels(
		model.Tunnel{ID: 11, Name: "ops-nas", CreatedBy: 2, LastOnlineAt: &now},
	)
	svc := &DomainService{pool: pool, admins: admins, domains: bound, tunnels: tunnels}
	ctx := context.Background()

	if _, err := svc.Create(ctx, 2, 11, CreateInputDomain{Domain: "dev." + testBase}); !errors.Is(err, ErrDomainNotAllowed) {
		t.Fatalf("不能绑定他人域名, got %v", err)
	}
	if _, err := svc.Create(ctx, 2, 11, CreateInputDomain{Domain: "ops-used." + testBase}); !errors.Is(err, ErrDomainTaken) {
		t.Fatalf("不能绑定已被使用的域名, got %v", err)
	}
}
