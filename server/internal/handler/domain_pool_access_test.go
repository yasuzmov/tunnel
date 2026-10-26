package handler

import (
	"bytes"
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

// menuAdmins 给普通管理员授予「域名管理」菜单。
type menuAdmins struct{ *testAdmins }

func (menuAdmins) ListMenus(context.Context, uint) ([]string, error) {
	return []string{service.MenuTunnels, service.MenuDomains}, nil
}

type testPool struct{ items []model.Domain }

func (p *testPool) Create(_ context.Context, d *model.Domain) error {
	d.ID = uint(100 + len(p.items))
	p.items = append(p.items, *d)
	return nil
}
func (p *testPool) Update(_ context.Context, id uint, domain, remark string) error {
	for i := range p.items {
		if p.items[i].ID == id {
			p.items[i].Domain, p.items[i].Remark = domain, remark
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}
func (p *testPool) Delete(_ context.Context, id uint) error {
	for i := range p.items {
		if p.items[i].ID == id {
			p.items = append(p.items[:i], p.items[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}
func (p *testPool) FindByID(_ context.Context, id uint) (*model.Domain, error) {
	for i := range p.items {
		if p.items[i].ID == id {
			cp := p.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (p *testPool) FindByDomain(_ context.Context, host string) (*model.Domain, error) {
	for i := range p.items {
		if p.items[i].Domain == host {
			cp := p.items[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (p *testPool) List(_ context.Context, _ string, createdBy *uint) ([]model.Domain, error) {
	var list []model.Domain
	for _, d := range p.items {
		if createdBy == nil || d.CreatedBy == *createdBy {
			list = append(list, d)
		}
	}
	return list, nil
}
func (p *testPool) ListBindings(context.Context) ([]repository.DomainBinding, error) { return nil, nil }

type testBound struct{}

func (testBound) Create(context.Context, *model.TunnelDomain) error { return nil }
func (testBound) Update(context.Context, *model.TunnelDomain) error { return nil }
func (testBound) Delete(context.Context, uint) error                { return nil }
func (testBound) FindByID(context.Context, uint) (*model.TunnelDomain, error) {
	return nil, gorm.ErrRecordNotFound
}
func (testBound) FindByDomain(context.Context, string) (*model.TunnelDomain, error) {
	return nil, gorm.ErrRecordNotFound
}
func (testBound) ListByTunnel(context.Context, uint) ([]model.TunnelDomain, error) { return nil, nil }
func (testBound) ListAll(context.Context) ([]model.TunnelDomain, error)            { return nil, nil }
func (testBound) DeleteByTunnel(context.Context, uint) error                       { return nil }

func poolEngine(t *testing.T) (*gin.Engine, *testPool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	admins := menuAdmins{&testAdmins{users: map[uint]*model.AdminUser{
		1: {ID: 1, Username: "root", Role: model.RoleSuperAdmin},
		2: {ID: 2, Username: "ops", Role: model.RoleAdmin},
		3: {ID: 3, Username: "dev", Role: model.RoleAdmin},
	}}}
	pool := &testPool{items: []model.Domain{
		{ID: 1, Domain: "legacy.fish.czypro.cn"},
		{ID: 2, Domain: "ops.fish.czypro.cn", CreatedBy: 2},
		{ID: 3, Domain: "dev.fish.czypro.cn", CreatedBy: 3},
	}}
	poolSvc := service.NewDomainPoolService(service.DomainPoolDeps{
		Pool: pool, Bound: testBound{}, Admins: admins, BaseDomain: "fish.czypro.cn",
	})
	adminSvc := service.NewAdminService(admins)
	return NewEngine(Deps{
		Auth:       &AuthHandler{},
		Tunnel:     &TunnelHandler{},
		Domain:     &DomainHandler{},
		DomainPool: NewDomainPoolHandler(poolSvc),
		Admin:      NewAdminHandler(adminSvc),
		AdminSvc:   adminSvc,
		Session:    &SessionHandler{},
		Log:        &LogHandler{},
		Stats:      &StatsHandler{},
		Event:      &EventHandler{},
		JWTSecret:  testSecret,
	}), pool
}

func doJSON(t *testing.T, e *gin.Engine, adminID uint, method, path string, body any) (int, map[string]any) {
	t.Helper()
	token, err := jwtutil.Issue(testSecret, adminID, "u", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func listLen(body map[string]any) int {
	list, _ := body["data"].([]any)
	return len(list)
}

func TestDomainPoolAPIIsolatedByCreator(t *testing.T) {
	e, pool := poolEngine(t)

	if code, body := doGet(t, e, 1, "/api/domains"); code != 200 || listLen(body) != 3 {
		t.Fatalf("超级管理员应看到全部: code=%d len=%d", code, listLen(body))
	}
	if code, body := doGet(t, e, 2, "/api/domains"); code != 200 || listLen(body) != 1 {
		t.Fatalf("普通管理员只看到自己的: code=%d len=%d", code, listLen(body))
	}
	if code, body := doGet(t, e, 2, "/api/domains/config"); code != 200 {
		t.Fatalf("config code=%d", code)
	} else if data, _ := body["data"].(map[string]any); data["base_domain"] != "fish.czypro.cn" {
		t.Fatalf("主域名不符: %v", body)
	}

	cases := []struct {
		admin  uint
		method string
		path   string
		body   any
		want   int
	}{
		{2, http.MethodPut, "/api/domains/3", map[string]string{"name": "hack"}, 404},
		{2, http.MethodDelete, "/api/domains/3", nil, 404},
		{2, http.MethodPut, "/api/domains/1", map[string]string{"remark": "x"}, 404},
		{2, http.MethodDelete, "/api/domains/1", nil, 404},
		{2, http.MethodPost, "/api/domains", map[string]string{"name": "dev"}, 409},
		{2, http.MethodPost, "/api/domains", map[string]string{"domain": "evil.com"}, 400},
		{2, http.MethodPost, "/api/domains", map[string]string{"name": "a.b"}, 400},
		{2, http.MethodPost, "/api/domains", map[string]string{"name": "new-one"}, 200},
		{2, http.MethodPut, "/api/domains/2", map[string]string{"name": "ops2"}, 200},
		{1, http.MethodPut, "/api/domains/3", map[string]string{"name": "dev2"}, 200},
		{2, http.MethodPut, "/api/admins/2/domains", map[string]any{"domain_ids": []uint{3}}, 404},
	}
	for _, tc := range cases {
		if code, body := doJSON(t, e, tc.admin, tc.method, tc.path, tc.body); code != tc.want {
			t.Errorf("admin %d %s %s: want %d, got %d %v", tc.admin, tc.method, tc.path, tc.want, code, body)
		}
	}
	got := map[uint]model.Domain{}
	for _, d := range pool.items {
		got[d.ID] = d
	}
	if got[2].Domain != "ops2.fish.czypro.cn" || got[3].Domain != "dev2.fish.czypro.cn" || got[1].Domain != "legacy.fish.czypro.cn" {
		t.Fatalf("修改结果不符: %+v", pool.items)
	}
	created := pool.items[len(pool.items)-1]
	if created.Domain != "new-one.fish.czypro.cn" || created.CreatedBy != 2 {
		t.Fatalf("新建域名应归属创建者: %+v", created)
	}
}
