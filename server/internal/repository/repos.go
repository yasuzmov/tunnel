package repository

import "gorm.io/gorm"

// Repos 聚合全部仓储，启动时一次性注入。
type Repos struct {
	Admin   AdminUserRepository
	Tunnel  TunnelRepository
	Domain  DomainRepository
	Pool    DomainPoolRepository
	Session SessionRepository
	Log     LogRepository
	Stats   StatsRepository
}

// New 基于同一 *gorm.DB 构造全部仓储。
func New(db *gorm.DB) *Repos {
	return &Repos{
		Admin:   NewAdminUserRepo(db),
		Tunnel:  NewTunnelRepo(db),
		Domain:  NewDomainRepo(db),
		Pool:    NewDomainPoolRepo(db),
		Session: NewSessionRepo(db),
		Log:     NewLogRepo(db),
		Stats:   NewStatsRepo(db),
	}
}
