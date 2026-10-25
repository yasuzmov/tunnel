// Package app 组装全部依赖并启动管理端与公网入口。
package app

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"time"

	"tunnel-server/internal/config"
	"tunnel-server/internal/database"
	"tunnel-server/internal/ingress"
	"tunnel-server/internal/model"
	"tunnel-server/internal/repository"
	"tunnel-server/internal/tunnel"
	"tunnel-server/web"
)

// App 进程内的服务组合。
type App struct {
	cfg        *config.Config
	admin      *http.Server
	ingressSrv *ingress.Server
	hub        *tunnel.Hub
	repos      *repository.Repos
	cancel     context.CancelFunc
	writerDone chan struct{}
}

// New 打开数据库、迁移、注入各层后返回可运行实例。
func New(cfg *config.Config) (*App, error) {
	db, err := database.Open(cfg.MySQL.DSN)
	if err != nil {
		return nil, err
	}
	if err := database.AutoMigrate(db); err != nil {
		return nil, err
	}
	if err := database.SeedAdmin(db, cfg.Bootstrap.Username, cfg.Bootstrap.Password); err != nil {
		return nil, err
	}
	if err := database.SeedDomainPool(db); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	repos := repository.New(db)
	hub := tunnel.NewHub()
	engine, ingressSrv, writer := wire(cfg, repos, hub)
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		writer.Run(ctx)
	}()
	go func() {
		log.Printf("公网入口监听 %s", cfg.Ingress.Listen)
		if err := ingressSrv.ListenAndServe(cfg.Ingress.Listen); err != nil && err != http.ErrServerClosed {
			log.Printf("公网入口退出: %v", err)
		}
	}()
	return &App{
		cfg:        cfg,
		cancel:     cancel,
		admin:      &http.Server{Addr: cfg.Admin.Listen, Handler: engine},
		ingressSrv: ingressSrv,
		hub:        hub,
		repos:      repos,
		writerDone: writerDone,
	}, nil
}

// Run 阻塞监听管理 API。
func (a *App) Run() error {
	log.Printf("管理端监听 %s", a.cfg.Admin.Listen)
	if err := a.admin.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown 停止监听、断开在线客户端并尽量刷完请求日志。
func (a *App) Shutdown(ctx context.Context) error {
	now := time.Now()
	for _, s := range a.hub.CloseAll(model.DisconnectShutdown) {
		if s.ID == 0 {
			continue
		}
		_ = a.repos.Session.MarkOffline(ctx, s.ID, now, model.DisconnectShutdown, repository.SessionStats{
			RequestCount: s.Requests.Load(),
			BytesIn:      s.BytesIn.Load(),
			BytesOut:     s.BytesOut.Load(),
		})
	}
	if a.ingressSrv != nil {
		_ = a.ingressSrv.Shutdown(ctx)
	}
	a.cancel()
	if a.admin != nil {
		_ = a.admin.Shutdown(ctx)
	}
	select {
	case <-a.writerDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func staticFS() fs.FS {
	sub, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		return web.DistFS
	}
	return sub
}
