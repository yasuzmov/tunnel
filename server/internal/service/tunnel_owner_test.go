package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// ownerTunnels 内存穿透记录表，支持按创建者过滤与软删除。
type ownerTunnels struct {
	items   map[uint]*model.Tunnel
	deleted map[uint]*model.Tunnel
	nextID  uint
}

func newOwnerTunnels(list ...model.Tunnel) *ownerTunnels {
	r := &ownerTunnels{items: map[uint]*model.Tunnel{}, deleted: map[uint]*model.Tunnel{}}
	for i := range list {
		t := list[i]
		r.items[t.ID] = &t
		if t.ID > r.nextID {
			r.nextID = t.ID
		}
	}
	return r
}

func (r *ownerTunnels) Create(_ context.Context, t *model.Tunnel) error {
	r.nextID++
	t.ID = r.nextID
	cp := *t
	r.items[t.ID] = &cp
	return nil
}

func (r *ownerTunnels) Update(_ context.Context, t *model.Tunnel) error {
	cp := *t
	r.items[t.ID] = &cp
	return nil
}

func (r *ownerTunnels) Delete(_ context.Context, id uint) error {
	if t, ok := r.items[id]; ok {
		r.deleted[id] = t
		delete(r.items, id)
	}
	return nil
}

func (r *ownerTunnels) FindByID(_ context.Context, id uint) (*model.Tunnel, error) {
	t, ok := r.items[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *t
	return &cp, nil
}

func (r *ownerTunnels) FindByKey(context.Context, string) (*model.Tunnel, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *ownerTunnels) List(_ context.Context, _, _ int, _ string, createdBy *uint) ([]model.Tunnel, int64, error) {
	var list []model.Tunnel
	for id := uint(1); id <= r.nextID; id++ {
		t, ok := r.items[id]
		if !ok {
			continue
		}
		if createdBy != nil && t.CreatedBy != *createdBy {
			continue
		}
		list = append(list, *t)
	}
	return list, int64(len(list)), nil
}

func (r *ownerTunnels) OwnerOf(_ context.Context, id uint) (uint, error) {
	if t, ok := r.items[id]; ok {
		return t.CreatedBy, nil
	}
	if t, ok := r.deleted[id]; ok {
		return t.CreatedBy, nil
	}
	return 0, gorm.ErrRecordNotFound
}

func (r *ownerTunnels) TouchOnline(context.Context, uint, time.Time) error { return nil }

func (r *ownerTunnels) LoggingFlags(context.Context) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}

// noDomains 空域名表，Delete 流程只需要它不报错。
type noDomains struct{}

func (noDomains) Create(context.Context, *model.TunnelDomain) error { return nil }
func (noDomains) Update(context.Context, *model.TunnelDomain) error { return nil }
func (noDomains) Delete(context.Context, uint) error                { return nil }
func (noDomains) FindByID(context.Context, uint) (*model.TunnelDomain, error) {
	return nil, gorm.ErrRecordNotFound
}
func (noDomains) FindByDomain(context.Context, string) (*model.TunnelDomain, error) {
	return nil, gorm.ErrRecordNotFound
}
func (noDomains) ListByTunnel(context.Context, uint) ([]model.TunnelDomain, error) { return nil, nil }
func (noDomains) ListAll(context.Context) ([]model.TunnelDomain, error)            { return nil, nil }
func (noDomains) DeleteByTunnel(context.Context, uint) error                       { return nil }

// ownerLogs 内存日志表，CreatedBy 过滤借助 tunnels.OwnerOf 模拟子查询。
type ownerLogs struct {
	tunnels *ownerTunnels
	items   []model.RequestLog
	last    repository.LogFilter
}

func (l *ownerLogs) CreateBatch(context.Context, []*model.RequestLog, []*model.RequestLogBody) error {
	return nil
}

func (l *ownerLogs) List(ctx context.Context, f repository.LogFilter) ([]model.RequestLog, int64, error) {
	l.last = f
	var list []model.RequestLog
	for _, item := range l.items {
		if f.TunnelID != nil && item.TunnelID != *f.TunnelID {
			continue
		}
		if f.CreatedBy != nil {
			owner, err := l.tunnels.OwnerOf(ctx, item.TunnelID)
			if err != nil || owner != *f.CreatedBy {
				continue
			}
		}
		list = append(list, item)
	}
	return list, int64(len(list)), nil
}

func (l *ownerLogs) GetByID(_ context.Context, id uint) (*model.RequestLog, error) {
	for i := range l.items {
		if l.items[i].ID == id {
			cp := l.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (l *ownerLogs) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

// 管理员：1 超级管理员 root；2、3 为普通管理员。
// 记录：10 历史无归属；11 属于 2；12 属于 3；13 属于 1。
func ownershipFixture() (*fakeAdminStore, *ownerTunnels) {
	admins := &fakeAdminStore{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
		3: {ID: 3, Username: "dev", Role: model.RoleAdmin},
	}}
	tunnels := newOwnerTunnels(
		model.Tunnel{ID: 10, Name: "legacy", CreatedBy: 0},
		model.Tunnel{ID: 11, Name: "ops-nas", CreatedBy: 2},
		model.Tunnel{ID: 12, Name: "dev-web", CreatedBy: 3},
		model.Tunnel{ID: 13, Name: "root-db", CreatedBy: 1},
	)
	return admins, tunnels
}

func tunnelIDs(list []model.Tunnel) []uint {
	ids := make([]uint, 0, len(list))
	for _, t := range list {
		ids = append(ids, t.ID)
	}
	return ids
}

func equalIDs(a, b []uint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTunnelListIsolatedByCreator(t *testing.T) {
	admins, tunnels := ownershipFixture()
	svc := NewTunnelService(TunnelDeps{Tunnels: tunnels, Domains: noDomains{}, Admins: admins})
	ctx := context.Background()

	all, total, err := svc.List(ctx, 1, 1, 20, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 4 || !equalIDs(tunnelIDs(all), []uint{10, 11, 12, 13}) {
		t.Fatalf("超级管理员应看到全部记录（含历史无归属），got %v", tunnelIDs(all))
	}
	for _, item := range all {
		want := map[uint]string{10: "", 11: "ops", 12: "dev", 13: "root"}[item.ID]
		if item.CreatorName != want {
			t.Fatalf("记录 %d 创建者名称 want %q got %q", item.ID, want, item.CreatorName)
		}
	}

	own, total, err := svc.List(ctx, 2, 1, 20, "")
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || !equalIDs(tunnelIDs(own), []uint{11}) {
		t.Fatalf("普通管理员只应看到自己创建的记录，got %v", tunnelIDs(own))
	}
	if own[0].CreatorName != "" {
		t.Fatal("普通管理员视角不应填充创建者名称")
	}

	if _, _, err := svc.List(ctx, 99, 1, 20, ""); !errors.Is(err, ErrForbidden) {
		t.Fatalf("不存在的管理员应被拒绝，got %v", err)
	}
}

func TestTunnelCreateRecordsCreator(t *testing.T) {
	admins, tunnels := ownershipFixture()
	svc := NewTunnelService(TunnelDeps{Tunnels: tunnels, Domains: noDomains{}, Admins: admins})
	ctx := context.Background()

	created, err := svc.Create(ctx, 2, CreateInput{Name: "new"})
	if err != nil {
		t.Fatal(err)
	}
	if created.CreatedBy != 2 {
		t.Fatalf("创建者应为当前管理员，got %d", created.CreatedBy)
	}
	if _, err := svc.Get(ctx, 2, created.ID); err != nil {
		t.Fatalf("创建者应能查看新记录: %v", err)
	}
	if _, err := svc.Get(ctx, 3, created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("其他普通管理员不应看到新记录，got %v", err)
	}
	if _, err := svc.Get(ctx, 1, created.ID); err != nil {
		t.Fatalf("超级管理员应能查看新记录: %v", err)
	}
}

func TestTunnelOperationsForbiddenForOthers(t *testing.T) {
	admins, tunnels := ownershipFixture()
	svc := NewTunnelService(TunnelDeps{Tunnels: tunnels, Domains: noDomains{}, Admins: admins})
	ctx := context.Background()

	// 普通管理员 2 对他人（12）、超级管理员（13）与历史无归属（10）记录的任何操作都应 404。
	for _, id := range []uint{10, 12, 13} {
		if _, err := svc.Get(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Get %d: want ErrNotFound, got %v", id, err)
		}
		if _, err := svc.Update(ctx, 2, id, UpdateInput{Name: "hack"}); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Update %d: want ErrNotFound, got %v", id, err)
		}
		if _, err := svc.SetLogging(ctx, 2, id, false); !errors.Is(err, ErrNotFound) {
			t.Fatalf("SetLogging %d: want ErrNotFound, got %v", id, err)
		}
		if _, err := svc.ResetKey(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("ResetKey %d: want ErrNotFound, got %v", id, err)
		}
		if err := svc.Delete(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Delete %d: want ErrNotFound, got %v", id, err)
		}
		if err := svc.CheckAccess(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("CheckAccess %d: want ErrNotFound, got %v", id, err)
		}
		if tunnels.items[id] == nil || tunnels.items[id].Name == "hack" {
			t.Fatalf("记录 %d 不应被修改或删除", id)
		}
	}

	// 自己的记录可正常操作。
	if _, err := svc.Update(ctx, 2, 11, UpdateInput{Name: "ops-nas-2"}); err != nil {
		t.Fatalf("更新自己的记录失败: %v", err)
	}
	if err := svc.CheckAccess(ctx, 2, 11); err != nil {
		t.Fatalf("访问自己的记录失败: %v", err)
	}

	// 超级管理员可操作任意记录（含历史无归属记录）。
	for _, id := range []uint{10, 11, 12} {
		if _, err := svc.ResetKey(ctx, 1, id); err != nil {
			t.Fatalf("超级管理员重置 %d 失败: %v", id, err)
		}
	}
	if err := svc.Delete(ctx, 1, 12); err != nil {
		t.Fatalf("超级管理员删除他人记录失败: %v", err)
	}
}

func TestTunnelCanSeeForEvents(t *testing.T) {
	admins, tunnels := ownershipFixture()
	svc := NewTunnelService(TunnelDeps{Tunnels: tunnels, Domains: noDomains{}, Admins: admins})
	ctx := context.Background()
	ops, err := svc.Viewer(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	root, _ := svc.Viewer(ctx, 1)
	if !svc.CanSee(ctx, ops, 11) || svc.CanSee(ctx, ops, 12) || svc.CanSee(ctx, ops, 10) || svc.CanSee(ctx, ops, 999) {
		t.Fatal("普通管理员事件可见范围不正确")
	}
	if !svc.CanSee(ctx, root, 10) || !svc.CanSee(ctx, root, 12) {
		t.Fatal("超级管理员应可见全部事件")
	}
}

func TestLogsIsolatedByCreator(t *testing.T) {
	admins, tunnels := ownershipFixture()
	logs := &ownerLogs{tunnels: tunnels, items: []model.RequestLog{
		{ID: 100, TunnelID: 10},
		{ID: 101, TunnelID: 11},
		{ID: 102, TunnelID: 12},
		{ID: 103, TunnelID: 13},
		{ID: 104, TunnelID: 11},
	}}
	svc := NewLogService(logs, tunnels, admins, 30)
	ctx := context.Background()

	_, total, err := svc.List(ctx, 1, repository.LogFilter{})
	if err != nil || total != 5 {
		t.Fatalf("超级管理员应看到全部日志, total=%d err=%v", total, err)
	}
	if logs.last.CreatedBy != nil {
		t.Fatal("超级管理员查询不应附加创建者过滤")
	}

	list, total, err := svc.List(ctx, 2, repository.LogFilter{})
	if err != nil || total != 2 {
		t.Fatalf("普通管理员只应看到自己记录的日志, total=%d err=%v", total, err)
	}
	for _, item := range list {
		if item.TunnelID != 11 {
			t.Fatalf("混入了他人日志: %+v", item)
		}
	}

	// 伪造 CreatedBy 或指定他人记录 ID 都拿不到他人日志。
	other := uint(3)
	tid := uint(12)
	_, total, err = svc.List(ctx, 2, repository.LogFilter{CreatedBy: &other, TunnelID: &tid})
	if err != nil || total != 0 {
		t.Fatalf("普通管理员不应查到他人记录的日志, total=%d err=%v", total, err)
	}

	if _, err := svc.Get(ctx, 2, 101); err != nil {
		t.Fatalf("普通管理员应能查看自己记录的日志详情: %v", err)
	}
	for _, id := range []uint{100, 102, 103} {
		if _, err := svc.Get(ctx, 2, id); !errors.Is(err, ErrNotFound) {
			t.Fatalf("日志 %d: 普通管理员应 404, got %v", id, err)
		}
		if _, err := svc.Get(ctx, 1, id); err != nil {
			t.Fatalf("日志 %d: 超级管理员应可查看, got %v", id, err)
		}
	}

	// 记录被删除后，历史日志仍归属原创建者。
	_ = tunnels.Delete(ctx, 11)
	if _, err := svc.Get(ctx, 2, 104); err != nil {
		t.Fatalf("已删除记录的日志仍应对原创建者可见: %v", err)
	}
	if _, err := svc.Get(ctx, 3, 104); !errors.Is(err, ErrNotFound) {
		t.Fatalf("已删除记录的日志不应对他人可见, got %v", err)
	}
}
