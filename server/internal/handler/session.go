package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// SessionHandler 客户端联通记录。
type SessionHandler struct {
	svc *service.SessionService
}

// NewSessionHandler 创建联通记录 handler。
func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

// List 分页返回某穿透记录的联通历史。
func (h *SessionHandler) List(c *gin.Context) {
	tid, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 20)
	list, total, err := h.svc.List(c.Request.Context(), tid, page, size)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}
