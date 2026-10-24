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
}

// TunnelRepository 穿透记录存储。
type TunnelRepository interface {
	Create(ctx context.Context, t *model.Tunnel) error
	Update(ctx context.Context, t *model.Tunnel) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Tunnel, error)
	FindByKey(ctx context.Context, key string) (*model.Tunnel, error)
	List(ctx context.Context, page, pageSize int, keyword string) ([]model.Tunnel, int64, error)
	TouchOnline(ctx context.Context, id uint, at time.Time) error
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
