package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
)

// NginxManager 域名配置落地。preview 模式由实现内部判断。
type NginxManager interface {
	Preview(domain string, https, certReady bool) (string, error)
	Apply(domain string, https, certReady bool) (string, error)
	Remove(domain string) error
}

// DomainReloader 刷新入口域名缓存。
type DomainReloader interface {
	Reload(ctx context.Context) error
}

// ConfigNotifier 把最新域名列表推给在线客户端。
type ConfigNotifier interface {
	NotifyConfig(tunnelID uint, domains []string)
}

// DomainService 绑定域名并驱动 nginx/certbot。
type DomainService struct {
	tunnels repository.TunnelRepository
	domains repository.DomainRepository
	pool    repository.DomainPoolRepository
	admins  repository.AdminUserRepository
	nginx   NginxManager
	issue   func(ctx context.Context, domain string) (time.Time, error)
	reload  DomainReloader
	notify  ConfigNotifier
	bus     *eventbus.Bus
	nginxOn bool
	certOn  bool
}

// DomainDeps 构造依赖，避免构造函数参数过多。
type DomainDeps struct {
	Tunnels repository.TunnelRepository
	Domains repository.DomainRepository
	Pool    repository.DomainPoolRepository
	Admins  repository.AdminUserRepository
	Nginx   NginxManager
	Issue   func(ctx context.Context, domain string) (time.Time, error)
	Reload  DomainReloader
	Notify  ConfigNotifier
	Bus     *eventbus.Bus
	NginxOn bool
	CertOn  bool
}

// NewDomainService 创建域名服务。
func NewDomainService(d DomainDeps) *DomainService {
	return &DomainService{
		tunnels: d.Tunnels,
		domains: d.Domains,
		pool:    d.Pool,
		admins:  d.Admins,
		nginx:   d.Nginx,
		issue:   d.Issue,
		reload:  d.Reload,
		notify:  d.Notify,
		bus:     d.Bus,
		nginxOn: d.NginxOn,
		certOn:  d.CertOn,
	}
}

// CreateInputDomain 新增域名。
type CreateInputDomain struct {
	Domain       string `json:"domain"`
	HTTPSEnabled bool   `json:"https_enabled"`
}

// Get 校验域名属于指定隧道。
func (s *DomainService) Get(ctx context.Context, tunnelID, domainID uint) (*model.TunnelDomain, error) {
	d, err := s.domains.FindByID(ctx, domainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if d.TunnelID != tunnelID {
		return nil, ErrDomainMismatch
	}
	return d, nil
}

// List 列出隧道下全部域名。
func (s *DomainService) List(ctx context.Context, tunnelID uint) ([]model.TunnelDomain, error) {
	return s.domains.ListByTunnel(ctx, tunnelID)
}

// Preview 返回将要写入的 nginx 配置文本。
func (s *DomainService) Preview(ctx context.Context, tunnelID, domainID uint) (string, error) {
	d, err := s.Get(ctx, tunnelID, domainID)
	if err != nil {
		return "", err
	}
	return s.nginx.Preview(d.Domain, d.HTTPSEnabled, d.CertReady())
}
