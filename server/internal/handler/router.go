package handler

import (
	"io/fs"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/middleware"
	"tunnel-server/internal/service"
	"tunnel-server/internal/tunnel"
)

// Deps 组装管理端路由所需的全部 handler。
type Deps struct {
	Auth       *AuthHandler
	Tunnel     *TunnelHandler
	Domain     *DomainHandler
	DomainPool *DomainPoolHandler
	Admin      *AdminHandler
	AdminSvc   *service.AdminService
	Session    *SessionHandler
	Log        *LogHandler
	Stats      *StatsHandler
	Event      *EventHandler
	TunnelWS   *tunnel.Endpoint
	JWTSecret  string
	Static     fs.FS
}

// NewEngine 创建管理端 Gin 引擎：API、隧道接入、SPA。
func NewEngine(d Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.CORS())
	r.POST("/api/auth/login", d.Auth.Login)
	authed := r.Group("/api", middleware.JWT(d.JWTSecret))
	registerAPI(authed, d)
	if d.TunnelWS != nil {
		r.GET("/tunnel/ws", d.TunnelWS.Handle)
	}
	r.NoRoute(SPA(d.Static))
	return r
}

func registerAPI(api *gin.RouterGroup, d Deps) {
	api.GET("/auth/me", d.Auth.Me)
	api.PUT("/auth/password", d.Auth.ChangePassword)
	api.GET("/stats/dashboard", requireMenu(d.AdminSvc, service.MenuDashboard), d.Stats.Dashboard)
	// SSE 被仪表盘、日志、穿透详情共用，不按菜单单独拦截。
	api.GET("/events", d.Event.Stream)
	api.GET("/logs", requireMenu(d.AdminSvc, service.MenuLogs), d.Log.List)
	api.GET("/logs/:id", requireMenu(d.AdminSvc, service.MenuLogs), d.Log.Get)
	domains := api.Group("/domains", requireMenu(d.AdminSvc, service.MenuDomains))
	domains.GET("", d.DomainPool.List)
	domains.POST("", d.DomainPool.Create)
	domains.PUT("/:id", d.DomainPool.Update)
	domains.DELETE("/:id", d.DomainPool.Delete)
	api.GET("/domains/available", requireMenu(d.AdminSvc, service.MenuTunnels), d.Domain.Available)
	api.GET("/admins", d.Admin.List)
	api.POST("/admins", d.Admin.Create)
	api.PUT("/admins/:id", d.Admin.Update)
	api.PUT("/admins/:id/password", d.Admin.ResetPassword)
	api.PUT("/admins/:id/domains", d.Admin.SetDomains)
	api.DELETE("/admins/:id", d.Admin.Delete)
	// 穿透记录按创建者隔离：CRUD 在服务层校验，子资源经 RequireAccess 统一拦截。
	t := api.Group("/tunnels", requireMenu(d.AdminSvc, service.MenuTunnels))
	t.GET("", d.Tunnel.List)
	t.POST("", d.Tunnel.Create)
	t.GET("/:id", d.Tunnel.Get)
	t.PUT("/:id", d.Tunnel.Update)
	t.DELETE("/:id", d.Tunnel.Delete)
	t.POST("/:id/reset-key", d.Tunnel.ResetKey)
	t.PUT("/:id/logging", d.Tunnel.SetLogging)
	t.GET("/:id/sessions", d.Tunnel.RequireAccess, d.Session.List)
	t.GET("/:id/logs", d.Tunnel.RequireAccess, d.Log.List)
	t.GET("/:id/domains", d.Tunnel.RequireAccess, d.Domain.List)
	t.POST("/:id/domains", d.Tunnel.RequireAccess, d.Domain.Create)
	t.DELETE("/:id/domains/:did", d.Tunnel.RequireAccess, d.Domain.Delete)
	t.GET("/:id/domains/:did/nginx-preview", d.Tunnel.RequireAccess, d.Domain.Preview)
}
