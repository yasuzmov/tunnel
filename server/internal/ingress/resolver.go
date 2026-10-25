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
	// LogEnabled 为 false 时仍转发，但不写请求日志。
	LogEnabled bool
}

// Resolver 带缓存的域名查找表，域名增删或日志开关变化后调用 Reload。
type Resolver struct {
	mu      sync.RWMutex
	byHost  map[string]Route
	domains repository.DomainRepository
	tunnels repository.TunnelRepository
}

// NewResolver 创建解析器。日志开关来自穿透记录，和域名一起缓存。
func NewResolver(domains repository.DomainRepository, tunnels repository.TunnelRepository) *Resolver {
	return &Resolver{byHost: map[string]Route{}, domains: domains, tunnels: tunnels}
}

// Reload 从数据库重建缓存。
func (r *Resolver) Reload(ctx context.Context) error {
	list, err := r.domains.ListAll(ctx)
	if err != nil {
		return err
	}
	flags, err := r.tunnels.LoggingFlags(ctx)
	if err != nil {
		return err
	}
	r.replace(list, flags)
	return nil
}

func (r *Resolver) replace(list []model.TunnelDomain, flags map[uint]bool) {
	next := make(map[string]Route, len(list))
	for _, d := range list {
		rt := toRoute(d)
		rt.LogEnabled = flags[d.TunnelID]
		next[d.Domain] = rt
	}
	r.mu.Lock()
	r.byHost = next
	r.mu.Unlock()
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
