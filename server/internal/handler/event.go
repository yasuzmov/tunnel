package handler

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"tunnel-server/internal/eventbus"
)

// EventHandler 把事件总线桥到 SSE，供后台自动刷新联通记录与日志。
type EventHandler struct {
	bus *eventbus.Bus
}

// NewEventHandler 创建 SSE handler。
func NewEventHandler(bus *eventbus.Bus) *EventHandler {
	return &EventHandler{bus: bus}
}

// Stream 长连接推送。token 走 query，因为 EventSource 不能自定义头。
func (h *EventHandler) Stream(c *gin.Context) {
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
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, payload)
			return true
		}
	})
}
