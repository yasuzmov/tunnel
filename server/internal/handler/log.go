package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/repository"
	"tunnel-server/internal/service"
)

// LogHandler 请求日志查询。
type LogHandler struct {
	svc *service.LogService
}

// NewLogHandler 创建日志 handler。
func NewLogHandler(svc *service.LogService) *LogHandler {
	return &LogHandler{svc: svc}
}

// List 支持按记录、域名、方法、状态、路径、IP、时间筛选。
func (h *LogHandler) List(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	f := repository.LogFilter{
		Domain:    c.Query("domain"),
		Method:    c.Query("method"),
		Path:      c.Query("path"),
		RemoteIP:  c.Query("remote_ip"),
		RequestID: c.Query("request_id"),
		Start:     queryTime(c, "start"),
		End:       queryTime(c, "end"),
		Status:    optionalStatus(c),
		Page:      queryInt(c, "page", 1),
		PageSize:  queryInt(c, "page_size", 20),
	}
	if tid, ok := queryUint(c, "id"); ok && c.Param("id") != "" {
		f.TunnelID = &tid
	} else if raw := c.Query("tunnel_id"); raw != "" {
		if tid, ok := queryUint(c, "tunnel_id"); ok {
			f.TunnelID = &tid
		}
	}
	list, total, err := h.svc.List(c.Request.Context(), adminID, f)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Page(c, list, total, f.Page, f.PageSize)
}

// Get 详情含截断后的请求/响应体。
func (h *LogHandler) Get(c *gin.Context) {
	id, ok := queryUint(c, "lid")
	if !ok {
		id, ok = queryUint(c, "id")
	}
	if !ok {
		response.BadRequest(c, "无效的日志 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), adminID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, item)
}
