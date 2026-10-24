package ingress

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"tunnel-server/internal/logwriter"
	"tunnel-server/internal/pkg/httputil"
	"tunnel-server/internal/tunnel"
)

// Server 公网流量入口。只应绑定在 localhost，由 nginx 反代进来。
type Server struct {
	Resolver        *Resolver
	Hub             *tunnel.Hub
	Logs            *logwriter.Writer
	BodyLimit       int
	ResponseTimeout time.Duration

	mu  sync.Mutex
	srv *http.Server
}

// Handler 返回标准 http.Handler，避免 gin 中间件改写 WebSocket 升级。
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(s.ServeHTTP)
}

// ServeHTTP 按 Host 找到隧道并转发；客户端不在线时返回友好 502，同时记日志。
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt, ok := s.Resolver.Lookup(r.Host)
	if !ok {
		http.NotFound(w, r)
		return
	}
	sess := s.Hub.Get(rt.TunnelID)
	if sess == nil {
		s.logOffline(r, rt)
		writeBadGateway(w, "客户端未联通，无法转发该域名")
		return
	}
	if httputil.IsWebSocketUpgrade(r.Header) {
		s.proxyWS(w, r, sess, rt)
		return
	}
	s.proxyHTTP(w, r, sess, rt)
}

func (s *Server) logOffline(r *http.Request, rt Route) {
	item := newLog(r, rt.TunnelID, rt.Domain, nil)
	item.Status = http.StatusBadGateway
	finishLog(item, "client offline")
	s.Logs.Enqueue(entryOnly(item))
	log.Printf("域名 %s 无在线客户端", rt.Domain)
}

// ListenAndServe 启动入口。
func (s *Server) ListenAndServe(addr string) error {
	hs := &http.Server{Addr: addr, Handler: s.Handler()}
	s.mu.Lock()
	s.srv = hs
	s.mu.Unlock()
	return hs.ListenAndServe()
}

// Shutdown 优雅停止入口监听。
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	hs := s.srv
	s.mu.Unlock()
	if hs == nil {
		return nil
	}
	return hs.Shutdown(ctx)
}
