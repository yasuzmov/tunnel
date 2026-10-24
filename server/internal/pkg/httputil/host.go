package httputil

import (
	"net"
	"net/http"
	"strings"
)

// CanonicalHost 去掉端口并转小写，用于域名查找。
func CanonicalHost(host string) string {
	h := strings.TrimSpace(host)
	if h == "" {
		return ""
	}
	if strings.HasPrefix(h, "[") {
		if inner, _, err := net.SplitHostPort(h); err == nil {
			return strings.ToLower(strings.Trim(inner, "[]"))
		}
	}
	if hostname, _, err := net.SplitHostPort(h); err == nil {
		return strings.ToLower(hostname)
	}
	return strings.ToLower(h)
}

// RequestScheme 优先取反代传入的 X-Forwarded-Proto。
func RequestScheme(r *http.Request) string {
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return strings.ToLower(strings.Split(proto, ",")[0])
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// ClientIP 优先取 nginx 注入的真实 IP，避免把 127.0.0.1 记进日志。
func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
