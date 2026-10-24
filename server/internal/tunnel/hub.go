package tunnel

import (
	"sync"
)

// Hub 保存每条穿透记录当前唯一的在线会话。
// 单实例部署下内存表即权威来源；新连接会顶替旧连接。
type Hub struct {
	mu       sync.RWMutex
	byTunnel map[uint]*Session
}

// NewHub 创建会话注册表。
func NewHub() *Hub {
	return &Hub{byTunnel: make(map[uint]*Session)}
}

// Register 登记会话。若该隧道已有连接，返回被顶替者供调用方关闭。
func (h *Hub) Register(s *Session) *Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	old := h.byTunnel[s.TunnelID]
	h.byTunnel[s.TunnelID] = s
	return old
}

// Unregister 仅在仍是当前会话时移除，避免把后来者误删。
func (h *Hub) Unregister(s *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.byTunnel[s.TunnelID]; ok && cur == s {
		delete(h.byTunnel, s.TunnelID)
	}
}

// Get 按隧道 ID 取在线会话。
func (h *Hub) Get(tunnelID uint) *Session {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.byTunnel[tunnelID]
}

// Kick 主动断开指定隧道的客户端。
func (h *Hub) Kick(tunnelID uint, reason string) {
	s := h.Get(tunnelID)
	if s != nil {
		s.Close(reason)
	}
}

// CloseAll 关闭全部在线会话并清空注册表，供进程优雅退出使用。
func (h *Hub) CloseAll(reason string) []*Session {
	h.mu.Lock()
	list := make([]*Session, 0, len(h.byTunnel))
	for id, s := range h.byTunnel {
		list = append(list, s)
		delete(h.byTunnel, id)
	}
	h.mu.Unlock()
	for _, s := range list {
		s.Close(reason)
	}
	return list
}

// NotifyConfig 向在线客户端推送域名列表。
func (h *Hub) NotifyConfig(tunnelID uint, domains []string) {
	s := h.Get(tunnelID)
	if s == nil {
		return
	}
	s.SendConfigUpdate(domains)
}

// OnlineIDs 返回当前在线隧道 ID，供排障。
func (h *Hub) OnlineIDs() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]uint, 0, len(h.byTunnel))
	for id := range h.byTunnel {
		ids = append(ids, id)
	}
	return ids
}
