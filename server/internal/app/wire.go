package app

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/config"
	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/handler"
	"tunnel-server/internal/ingress"
	"tunnel-server/internal/logwriter"
	"tunnel-server/internal/nginx"
	"tunnel-server/internal/repository"
	"tunnel-server/internal/service"
	"tunnel-server/internal/tunnel"
)

func wire(cfg *config.Config, repos *repository.Repos, hub *tunnel.Hub) (*gin.Engine, *ingress.Server, *logwriter.Writer) {
	bus := eventbus.New()
	resolver := ingress.NewResolver(repos.Domain, repos.Tunnel)
	_ = resolver.Reload(context.Background())
	writer := logwriter.New(repos.Log, bus, cfg.Log.ChannelSize, cfg.Log.BatchSize, cfg.Log.FlushInterval())
	runner := nginx.ExecRunner{}
	ngx := nginx.NewManager(cfg.Nginx, runner)
	issuer := nginx.NewCertbot(cfg.Certbot, cfg.Nginx, runner)

	tunSvc := service.NewTunnelService(service.TunnelDeps{
		Tunnels:    repos.Tunnel,
		Domains:    repos.Domain,
		PublicURL:  cfg.Admin.PublicURL,
		Disconnect: func(id uint, reason string) { hub.Kick(id, reason) },
		RemoveConf: func(domain string) error { return ngx.Remove(domain) },
		Reload:     resolver,
	})
	domainSvc := service.NewDomainService(service.DomainDeps{
		Tunnels: repos.Tunnel,
		Domains: repos.Domain,
		Pool:    repos.Pool,
		Admins:  repos.Admin,
		Nginx:   ngx,
		Issue:   issuer.Issue,
		Reload:  resolver,
		Notify:  hub,
		Bus:     bus,
		NginxOn: cfg.Nginx.Enabled,
		CertOn:  cfg.Certbot.Enabled,
	})
	adminSvc := service.NewAdminService(repos.Admin, repos.Pool)
	poolSvc := service.NewDomainPoolService(repos.Pool, repos.Domain)
	logSvc := service.NewLogService(repos.Log, cfg.Log.RetainDays)
	go cleanupLoop(logSvc)

	ep := &tunnel.Endpoint{
		Tunnels:  repos.Tunnel,
		Domains:  repos.Domain,
		Sessions: repos.Session,
		Hub:      hub,
		Bus:      bus,
		Idle:     cfg.Tunnel.IdleTimeout(),
	}
	ing := &ingress.Server{
		Resolver:        resolver,
		Hub:             hub,
		Logs:            writer,
		BodyLimit:       cfg.Log.EffectiveBodyLimit(),
		ResponseTimeout: cfg.Tunnel.ResponseTimeout(),
	}
	engine := handler.NewEngine(handler.Deps{
		Auth:       handler.NewAuthHandler(service.NewAuthService(repos.Admin, cfg.JWT.Secret, cfg.JWT.ExpireDuration())),
		Tunnel:     handler.NewTunnelHandler(tunSvc),
		Domain:     handler.NewDomainHandler(domainSvc),
		DomainPool: handler.NewDomainPoolHandler(poolSvc, adminSvc),
		Admin:      handler.NewAdminHandler(adminSvc),
		Session:    handler.NewSessionHandler(service.NewSessionService(repos.Session)),
		Log:        handler.NewLogHandler(logSvc),
		Stats:      handler.NewStatsHandler(service.NewStatsService(repos.Stats)),
		Event:      handler.NewEventHandler(bus),
		TunnelWS:   ep,
		JWTSecret:  cfg.JWT.Secret,
		Static:     staticFS(),
	})
	return engine, ing, writer
}

func cleanupLoop(logSvc *service.LogService) {
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()
	for {
		if n, err := logSvc.CleanupExpired(context.Background()); err != nil {
			log.Printf("清理过期日志失败: %v", err)
		} else if n > 0 {
			log.Printf("已清理 %d 条过期请求日志", n)
		}
		<-ticker.C
	}
}
