package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/jwtutil"
	"tunnel-server/internal/repository"
	"tunnel-server/internal/service"
)

// 以下为接口级隔离测试所需的最小内存仓储。

type testAdmins struct{ users map[uint]*model.AdminUser }

func (a *testAdmins) FindByUsername(context.Context, string) (*model.AdminUser, error) {
	return nil, gorm.ErrRecordNotFound
}
func (a *testAdmins) FindByID(_ context.Context, id uint) (*model.AdminUser, error) {
	if u, ok := a.users[id]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
}
func (a *testAdmins) Create(context.Context, *model.AdminUser) error            { return nil }
func (a *testAdmins) UpdateProfile(context.Context, uint, string, string) error { return nil }
func (a *testAdmins) UpdatePassword(context.Context, uint, string) error        { return nil }
func (a *testAdmins) Delete(context.Context, uint) error                        { return nil }
func (a *testAdmins) List(context.Context) ([]model.AdminUser, error) {
	list := make([]model.AdminUser, 0, len(a.users))
	for _, u := range a.users {
		list = append(list, *u)
	}
	return list, nil
}
func (a *testAdmins) CountByRole(context.Context, string) (int64, error) { return 0, nil }
func (a *testAdmins) ListMenus(context.Context, uint) ([]string, error)  { return nil, nil }
func (a *testAdmins) SetMenus(context.Context, uint, []string) error     { return nil }

type testTunnels struct{ items []model.Tunnel }

func (r *testTunnels) Create(context.Context, *model.Tunnel) error { return nil }
func (r *testTunnels) Update(context.Context, *model.Tunnel) error { return nil }
func (r *testTunnels) Delete(context.Context, uint) error          { return nil }
func (r *testTunnels) FindByID(_ context.Context, id uint) (*model.Tunnel, error) {
	for i := range r.items {
		if r.items[i].ID == id {
			cp := r.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *testTunnels) FindByKey(context.Context, string) (*model.Tunnel, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *testTunnels) List(_ context.Context, _, _ int, _ string, createdBy *uint) ([]model.Tunnel, int64, error) {
	var list []model.Tunnel
	for _, t := range r.items {
		if createdBy == nil || t.CreatedBy == *createdBy {
			list = append(list, t)
		}
	}
	return list, int64(len(list)), nil
}
func (r *testTunnels) OwnerOf(ctx context.Context, id uint) (uint, error) {
	t, err := r.FindByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return t.CreatedBy, nil
}
func (r *testTunnels) TouchOnline(context.Context, uint, time.Time) error { return nil }
func (r *testTunnels) LoggingFlags(context.Context) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}

type testLogs struct {
	tunnels *testTunnels
	items   []model.RequestLog
}

func (l *testLogs) CreateBatch(context.Context, []*model.RequestLog, []*model.RequestLogBody) error {
	return nil
}
func (l *testLogs) List(ctx context.Context, f repository.LogFilter) ([]model.RequestLog, int64, error) {
	var list []model.RequestLog
	for _, item := range l.items {
		if f.TunnelID != nil && item.TunnelID != *f.TunnelID {
			continue
		}
		if f.CreatedBy != nil {
			if owner, err := l.tunnels.OwnerOf(ctx, item.TunnelID); err != nil || owner != *f.CreatedBy {
				continue
			}
		}
		list = append(list, item)
	}
	return list, int64(len(list)), nil
}
func (l *testLogs) GetByID(_ context.Context, id uint) (*model.RequestLog, error) {
	for i := range l.items {
		if l.items[i].ID == id {
			cp := l.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (l *testLogs) DeleteOlderThan(context.Context, time.Time) (int64, error) { return 0, nil }

const testSecret = "handler-test-secret"

func accessEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	admins := &testAdmins{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
	}}
	tunnels := &testTunnels{items: []model.Tunnel{
		{ID: 10, Name: "legacy"},
		{ID: 11, Name: "ops-nas", CreatedBy: 2},
		{ID: 12, Name: "root-db", CreatedBy: 1},
	}}
	logs := &testLogs{tunnels: tunnels, items: []model.RequestLog{
		{ID: 100, TunnelID: 10}, {ID: 101, TunnelID: 11}, {ID: 102, TunnelID: 12},
	}}
	tunSvc := service.NewTunnelService(service.TunnelDeps{Tunnels: tunnels, Admins: admins})
	adminSvc := service.NewAdminService(admins, nil)
	return NewEngine(Deps{
		Auth:       &AuthHandler{},
		Tunnel:     NewTunnelHandler(tunSvc),
		Domain:     &DomainHandler{},
		DomainPool: &DomainPoolHandler{},
		Admin:      NewAdminHandler(adminSvc),
		AdminSvc:   adminSvc,
		Session:    &SessionHandler{},
		Log:        NewLogHandler(service.NewLogService(logs, tunnels, admins, 30)),
		Stats:      &StatsHandler{},
		Event:      &EventHandler{},
		JWTSecret:  testSecret,
	})
}

func doGet(t *testing.T, e *gin.Engine, adminID uint, path string) (int, map[string]any) {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, adminID, "u", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

func pageTotal(body map[string]any) float64 {
	data, _ := body["data"].(map[string]any)
	total, _ := data["total"].(float64)
	return total
}

func TestTunnelAPIIsolatedByCreator(t *testing.T) {
	e := accessEngine(t)

	if code, body := doGet(t, e, 1, "/api/tunnels"); code != 200 || pageTotal(body) != 3 {
		t.Fatalf("超级管理员列表: code=%d total=%v", code, pageTotal(body))
	}
	if code, body := doGet(t, e, 2, "/api/tunnels"); code != 200 || pageTotal(body) != 1 {
		t.Fatalf("普通管理员列表: code=%d total=%v", code, pageTotal(body))
	}
	if code, body := doGet(t, e, 1, "/api/logs"); code != 200 || pageTotal(body) != 3 {
		t.Fatalf("超级管理员日志: code=%d total=%v", code, pageTotal(body))
	}
	if code, body := doGet(t, e, 2, "/api/logs"); code != 200 || pageTotal(body) != 1 {
		t.Fatalf("普通管理员日志: code=%d total=%v", code, pageTotal(body))
	}

	cases := []struct {
		admin uint
		path  string
		want  int
	}{
		{2, "/api/tunnels/11", 200},
		{2, "/api/tunnels/11/logs", 200},
		{2, "/api/logs/101", 200},
		{2, "/api/tunnels/12", 404},
		{2, "/api/tunnels/10", 404},
		{2, "/api/tunnels/12/logs", 404},
		{2, "/api/tunnels/12/sessions", 404},
		{2, "/api/tunnels/12/domains", 404},
		{2, "/api/logs/102", 404},
		{2, "/api/logs/100", 404},
		{1, "/api/tunnels/10", 200},
		{1, "/api/tunnels/11/logs", 200},
		{1, "/api/logs/101", 200},
	}
	for _, tc := range cases {
		if code, _ := doGet(t, e, tc.admin, tc.path); code != tc.want {
			t.Errorf("admin %d GET %s: want %d, got %d", tc.admin, tc.path, tc.want, code)
		}
	}
}
