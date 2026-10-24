package handler

import (
	"io/fs"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/middleware"
	"tunnel-server/internal/tunnel"
)

// Deps 组装管理端路由所需的全部 handler。
type Deps struct {
	Auth    *AuthHandler
	Tunnel  *TunnelHandler
	Domain  *DomainHandler
	Session *SessionHandler
	Log     *LogHandler
	Stats   *StatsHandler
	Event   *EventHandler
	TunnelWS *tunnel.Endpoint
	JWTSecret string
	Static    fs.FS
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
	api.GET("/stats/dashboard", d.Stats.Dashboard)
	api.GET("/events", d.Event.Stream)
	api.GET("/logs", d.Log.List)
	api.GET("/logs/:id", d.Log.Get)
	t := api.Group("/tunnels")
	t.GET("", d.Tunnel.List)
	t.POST("", d.Tunnel.Create)
	t.GET("/:id", d.Tunnel.Get)
	t.PUT("/:id", d.Tunnel.Update)
	t.DELETE("/:id", d.Tunnel.Delete)
	t.POST("/:id/reset-key", d.Tunnel.ResetKey)
	t.GET("/:id/sessions", d.Session.List)
	t.GET("/:id/logs", d.Log.List)
	t.GET("/:id/domains", d.Domain.List)
	t.POST("/:id/domains", d.Domain.Create)
	t.DELETE("/:id/domains/:did", d.Domain.Delete)
	t.GET("/:id/domains/:did/nginx-preview", d.Domain.Preview)
}
