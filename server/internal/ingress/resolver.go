// Package ingress 接收 nginx 反代过来的公网流量，经隧道转给客户端。
package ingress

import (
	"context"
	"sync"

	"tunnel-server/internal/model"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/repository"
)

// Route 域名到穿透记录的映射。
type Route struct {
	TunnelID     uint
	DomainID     uint
	Domain       string
	HTTPSEnabled bool
}

// Resolver 带缓存的域名查找表，域名增删后调用 Reload。
type Resolver struct {
	mu     sync.RWMutex
	byHost map[string]Route
	repo   repository.DomainRepository
}

// NewResolver 创建解析器。
func NewResolver(repo repository.DomainRepository) *Resolver {
	return &Resolver{byHost: map[string]Route{}, repo: repo}
}

// Reload 从数据库重建缓存。
func (r *Resolver) Reload(ctx context.Context) error {
	list, err := r.repo.ListAll(ctx)
	if err != nil {
		return err
	}
	next := make(map[string]Route, len(list))
	for _, d := range list {
		next[d.Domain] = toRoute(d)
	}
	r.mu.Lock()
	r.byHost = next
	r.mu.Unlock()
	return nil
}

// Lookup 按 Host 头查找，自动去掉端口。
func (r *Resolver) Lookup(host string) (Route, bool) {
	key := httputil.CanonicalHost(host)
	r.mu.RLock()
	defer r.mu.RUnlock()
	rt, ok := r.byHost[key]
	return rt, ok
}

func toRoute(d model.TunnelDomain) Route {
	return Route{
		TunnelID:     d.TunnelID,
		DomainID:     d.ID,
		Domain:       d.Domain,
		HTTPSEnabled: d.HTTPSEnabled,
	}
}
