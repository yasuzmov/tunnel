package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/eventbus"
	"tunnel-server/internal/service"
)

// EventAccess SSE 推送的数据隔离：普通管理员只接收自己创建的穿透记录相关事件。
type EventAccess interface {
	Viewer(ctx context.Context, adminID uint) (service.Viewer, error)
	CanSee(ctx context.Context, v service.Viewer, tunnelID uint) bool
}

// EventHandler 把事件总线桥到 SSE，供后台自动刷新联通记录与日志。
type EventHandler struct {
	bus    *eventbus.Bus
	access EventAccess
}

// NewEventHandler 创建 SSE handler。access 为 nil 时不做隔离（仅测试使用）。
func NewEventHandler(bus *eventbus.Bus, access EventAccess) *EventHandler {
	return &EventHandler{bus: bus, access: access}
}

// Stream 长连接推送。token 走 query，因为 EventSource 不能自定义头。
func (h *EventHandler) Stream(c *gin.Context) {
	filter, ok := h.eventFilter(c)
	if !ok {
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()
	id, ch := h.bus.Subscribe()
	defer h.bus.Unsubscribe(id)
	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case ev, ok := <-ch:
			if !ok {
				return false
			}
			payload, _ := json.Marshal(ev.Payload)
			if filter != nil && !filter(payload) {
				return true
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, payload)
			return true
		}
	})
}

// eventFilter 返回 nil 表示不过滤（超级管理员）；失败时已写入响应。
func (h *EventHandler) eventFilter(c *gin.Context) (func([]byte) bool, bool) {
	if h.access == nil {
		return nil, true
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return nil, false
	}
	ctx := c.Request.Context()
	v, err := h.access.Viewer(ctx, adminID)
	if err != nil {
		writeServiceError(c, err)
		return nil, false
	}
	if v.Super {
		return nil, true
	}
	// 归属在记录生命周期内不变，按连接缓存避免每条日志事件都查库。
	cache := map[uint]bool{}
	return func(payload []byte) bool {
		tid, found := eventTunnelID(payload)
		if !found {
			return false
		}
		allowed, hit := cache[tid]
		if !hit {
			allowed = h.access.CanSee(ctx, v, tid)
			cache[tid] = allowed
		}
		return allowed
	}, true
}

// eventTunnelID 从事件载荷中取出所属穿透记录 ID，兼容 json 标签与无标签结构体。
func eventTunnelID(payload []byte) (uint, bool) {
	var probe struct {
		TunnelID  *uint `json:"tunnel_id"`
		TunnelID2 *uint `json:"TunnelID"`
	}
	if err := json.Unmarshal(payload, &probe); err != nil {
		return 0, false
	}
	if probe.TunnelID != nil {
		return *probe.TunnelID, true
	}
	if probe.TunnelID2 != nil {
		return *probe.TunnelID2, true
	}
	return 0, false
}
