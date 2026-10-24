// Package httputil 处理 HTTP 转发时必须剥离的逐跳头部。
// 这些头部只对当前连接有意义，透传会破坏客户端与上游的分帧。
package httputil

import (
	"net/http"
	"strings"
)

var hopByHop = map[string]struct{}{
	"connection":          {},
	"keep-alive":          {},
	"proxy-authenticate":  {},
	"proxy-authorization": {},
	"te":                  {},
	"trailer":             {},
	"transfer-encoding":   {},
	"upgrade":             {},
}

// FilterHopByHop 复制并去掉 hop-by-hop 头。
// stripUpgrade=false 时保留 Upgrade，供 WebSocket 握手使用。
func FilterHopByHop(src http.Header, stripUpgrade bool) http.Header {
	out := make(http.Header, len(src))
	skip := extraSkip(src)
	for k, vals := range src {
		lk := strings.ToLower(k)
		if _, ok := hopByHop[lk]; ok {
			if lk == "upgrade" && !stripUpgrade {
				out[k] = append([]string(nil), vals...)
			}
			continue
		}
		if _, dropped := skip[lk]; dropped {
			continue
		}
		out[k] = append([]string(nil), vals...)
	}
	return out
}

func extraSkip(src http.Header) map[string]struct{} {
	skip := map[string]struct{}{}
	for _, v := range src.Values("Connection") {
		for _, token := range strings.Split(v, ",") {
			name := strings.ToLower(strings.TrimSpace(token))
			if name != "" {
				skip[name] = struct{}{}
			}
		}
	}
	return skip
}

// HeaderToMap 转为可 JSON 序列化的 map。
func HeaderToMap(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// MapToHeader 从 JSON map 恢复 http.Header。
func MapToHeader(m map[string][]string) http.Header {
	out := make(http.Header, len(m))
	for k, v := range m {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// IsWebSocketUpgrade 判断是否为 WebSocket 握手。
func IsWebSocketUpgrade(h http.Header) bool {
	return strings.EqualFold(h.Get("Upgrade"), "websocket")
}

// IsTextContent 判断是否应按文本截断保存（否则只记类型与大小）。
func IsTextContent(contentType string) bool {
	ct := strings.ToLower(contentType)
	switch {
	case strings.HasPrefix(ct, "text/"):
		return true
	case strings.Contains(ct, "json"):
		return true
	case strings.Contains(ct, "xml"):
		return true
	case strings.Contains(ct, "javascript"):
		return true
	case strings.Contains(ct, "x-www-form-urlencoded"):
		return true
	default:
		return false
	}
}
