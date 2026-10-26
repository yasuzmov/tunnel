// Package repository 定义持久化接口与 Gorm 实现。
// Service 只依赖接口，便于单测替换。
package repository

import (
	"context"
	"time"

	"tunnel-server/internal/model"
)

// AdminUserRepository 管理员账号存储。
type AdminUserRepository interface {
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	FindByID(ctx context.Context, id uint) (*model.AdminUser, error)
	Create(ctx context.Context, u *model.AdminUser) error
	UpdateProfile(ctx context.Context, id uint, username, role string) error
	UpdatePassword(ctx context.Context, id uint, passwordHash string) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context) ([]model.AdminUser, error)
	CountByRole(ctx context.Context, role string) (int64, error)
	// ListMenus 返回管理员被授予的菜单权限；超级管理员不查此接口。
	ListMenus(ctx context.Context, adminID uint) ([]string, error)
	// SetMenus 覆盖式写入菜单权限。
	SetMenus(ctx context.Context, adminID uint, menus []string) error
}

// DomainBinding 域名池条目当前绑定到的穿透记录，未绑定时 TunnelID 为 0。
type DomainBinding struct {
	DomainID     uint
	TunnelID     uint
	TunnelName   string
	HTTPSEnabled bool `gorm:"column:https_enabled"`
	CertStatus   string
	NginxStatus  string
}

// DomainPoolRepository 统一域名池与管理员域名授权。
type DomainPoolRepository interface {
	Create(ctx context.Context, d *model.Domain) error
	UpdateRemark(ctx context.Context, id uint, remark string) error
	// Delete 同时删除该域名的全部管理员授权。
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Domain, error)
	FindByDomain(ctx context.Context, domain string) (*model.Domain, error)
	List(ctx context.Context, keyword string) ([]model.Domain, error)
	// ListByAdmin 列出被授权的域名；超级管理员不查此接口。
	ListByAdmin(ctx context.Context, adminID uint) ([]model.Domain, error)
	ListAdminDomainIDs(ctx context.Context, adminID uint) ([]uint, error)
	SetAdminDomains(ctx context.Context, adminID uint, domainIDs []uint) error
	ListBindings(ctx context.Context) ([]DomainBinding, error)
}

// TunnelRepository 穿透记录存储。
type TunnelRepository interface {
	Create(ctx context.Context, t *model.Tunnel) error
	Update(ctx context.Context, t *model.Tunnel) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Tunnel, error)
	FindByKey(ctx context.Context, key string) (*model.Tunnel, error)
	// List 分页列表；createdBy 非 nil 时只返回该管理员创建的记录。
	List(ctx context.Context, page, pageSize int, keyword string, createdBy *uint) ([]model.Tunnel, int64, error)
	// OwnerOf 返回记录的创建者 ID，包含已软删除的记录，供日志归属校验。
	OwnerOf(ctx context.Context, id uint) (uint, error)
	TouchOnline(ctx context.Context, id uint, at time.Time) error
	// LoggingFlags 返回全部未删除记录的日志开关，供入口缓存使用，避免每个请求查库。
	LoggingFlags(ctx context.Context) (map[uint]bool, error)
}

// DomainRepository 域名绑定存储。
type DomainRepository interface {
	Create(ctx context.Context, d *model.TunnelDomain) error
	Update(ctx context.Context, d *model.TunnelDomain) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.TunnelDomain, error)
	FindByDomain(ctx context.Context, domain string) (*model.TunnelDomain, error)
	ListByTunnel(ctx context.Context, tunnelID uint) ([]model.TunnelDomain, error)
	ListAll(ctx context.Context) ([]model.TunnelDomain, error)
	DeleteByTunnel(ctx context.Context, tunnelID uint) error
}

// SessionStats 一次联通会话的累计计数，断开时落库。
type SessionStats struct {
	RequestCount int64
	BytesIn      int64
	BytesOut     int64
}

// SessionRepository 客户端联通记录存储。
type SessionRepository interface {
	Create(ctx context.Context, s *model.ClientSession) error
	Update(ctx context.Context, s *model.ClientSession) error
	FindByID(ctx context.Context, id uint) (*model.ClientSession, error)
	ListByTunnel(ctx context.Context, tunnelID uint, page, pageSize int) ([]model.ClientSession, int64, error)
	TouchHeartbeat(ctx context.Context, id uint, at time.Time) error
	MarkOffline(ctx context.Context, id uint, at time.Time, reason string, stats SessionStats) error
	CountOnline(ctx context.Context) (int64, error)
}

// LogFilter 请求日志查询条件。
type LogFilter struct {
	TunnelID  *uint
	// CreatedBy 非 nil 时只返回该管理员创建的穿透记录产生的日志（含已删除记录）。
	CreatedBy *uint
	Domain    string
	Method    string
	Status    *int
	Path      string
	RemoteIP  string
	RequestID string
	Start     *time.Time
	End       *time.Time
	Page      int
	PageSize  int
}

// LogRepository 请求日志存储。
type LogRepository interface {
	CreateBatch(ctx context.Context, logs []*model.RequestLog, bodies []*model.RequestLogBody) error
	List(ctx context.Context, f LogFilter) ([]model.RequestLog, int64, error)
	GetByID(ctx context.Context, id uint) (*model.RequestLog, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// DashboardStats 仪表盘聚合结果。
type DashboardStats struct {
	OnlineSessions int64   `json:"online_sessions"`
	TunnelCount    int64   `json:"tunnel_count"`
	TodayRequests  int64   `json:"today_requests"`
	TodayErrors    int64   `json:"today_errors"`
	ErrorRate      float64 `json:"error_rate"`
}

// StatsRepository 统计查询。
type StatsRepository interface {
	Dashboard(ctx context.Context, todayStart time.Time) (DashboardStats, error)
}
