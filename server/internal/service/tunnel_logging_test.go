package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/model"
)

// memTunnels 只实现 SetLogging 用到的方法，其余返回零值。
type memTunnels struct {
	item *model.Tunnel
}

func (m *memTunnels) Create(context.Context, *model.Tunnel) error { return nil }
func (m *memTunnels) Delete(context.Context, uint) error          { return nil }
func (m *memTunnels) FindByKey(context.Context, string) (*model.Tunnel, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *memTunnels) List(context.Context, int, int, string) ([]model.Tunnel, int64, error) {
	return nil, 0, nil
}
func (m *memTunnels) TouchOnline(context.Context, uint, time.Time) error { return nil }
func (m *memTunnels) LoggingFlags(context.Context) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}

func (m *memTunnels) FindByID(_ context.Context, id uint) (*model.Tunnel, error) {
	if m.item == nil || m.item.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	return m.item, nil
}

func (m *memTunnels) Update(_ context.Context, t *model.Tunnel) error {
	m.item = t
	return nil
}

type reloadSpy struct{ n int }

func (r *reloadSpy) Reload(context.Context) error {
	r.n++
	return nil
}

func TestSetLoggingReloadsIngressCache(t *testing.T) {
	repo := &memTunnels{item: &model.Tunnel{ID: 3, Name: "nas", LogEnabled: true}}
	spy := &reloadSpy{}
	svc := NewTunnelService(TunnelDeps{Tunnels: repo, Reload: spy})
	got, err := svc.SetLogging(context.Background(), 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.LogEnabled || repo.item.LogEnabled {
		t.Fatal("expected logging disabled")
	}
	if spy.n != 1 {
		t.Fatalf("reload count = %d", spy.n)
	}
	if _, err := svc.SetLogging(context.Background(), 3, false); err != nil {
		t.Fatal(err)
	}
	if spy.n != 1 {
		t.Fatal("unchanged flag should not reload")
	}
}

func TestSetLoggingMissing(t *testing.T) {
	svc := NewTunnelService(TunnelDeps{Tunnels: &memTunnels{}})
	_, err := svc.SetLogging(context.Background(), 9, true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}
