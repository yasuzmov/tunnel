package handler

import (
	"github.com/gin-gonic/gin"

	"tunnel-server/internal/pkg/response"
	"tunnel-server/internal/service"
)

// TunnelHandler 穿透记录 CRUD 与重置 key。
type TunnelHandler struct {
	svc *service.TunnelService
}

// NewTunnelHandler 创建穿透记录 handler。
func NewTunnelHandler(svc *service.TunnelService) *TunnelHandler {
	return &TunnelHandler{svc: svc}
}

// List 分页列表。
func (h *TunnelHandler) List(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	page := queryInt(c, "page", 1)
	size := queryInt(c, "page_size", 20)
	list, total, err := h.svc.List(c.Request.Context(), adminID, page, size, c.Query("keyword"))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.Page(c, list, total, page, size)
}

// Create 新建记录。
func (h *TunnelHandler) Create(c *gin.Context) {
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var in service.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	t, err := h.svc.Create(c.Request.Context(), adminID, in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, h.withURL(t))
}

// Get 详情，附带客户端接入地址。
func (h *TunnelHandler) Get(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	t, err := h.svc.Get(c.Request.Context(), adminID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, h.withURL(t))
}

// Update 修改名称备注。
func (h *TunnelHandler) Update(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var in service.UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	t, err := h.svc.Update(c.Request.Context(), adminID, id, in)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, h.withURL(t))
}

// Delete 删除记录。
func (h *TunnelHandler) Delete(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), adminID, id); err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, gin.H{})
}

// SetLogging 开启或关闭该记录的转发日志。
func (h *TunnelHandler) SetLogging(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var in service.SetLoggingInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "参数不合法")
		return
	}
	t, err := h.svc.SetLogging(c.Request.Context(), adminID, id, in.Enabled)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, h.withURL(t))
}

// ResetKey 轮换接入 key。
func (h *TunnelHandler) ResetKey(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		return
	}
	t, err := h.svc.ResetKey(c.Request.Context(), adminID, id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	response.OK(c, h.withURL(t))
}

// RequireAccess 用于 /tunnels/:id 下的子资源（联通记录、日志、域名）：
// 普通管理员访问他人创建的记录时直接返回 404 并中止。
func (h *TunnelHandler) RequireAccess(c *gin.Context) {
	id, ok := queryUint(c, "id")
	if !ok {
		response.BadRequest(c, "无效的记录 ID")
		c.Abort()
		return
	}
	adminID, ok := currentAdminID(c)
	if !ok {
		c.Abort()
		return
	}
	if err := h.svc.CheckAccess(c.Request.Context(), adminID, id); err != nil {
		writeServiceError(c, err)
		c.Abort()
		return
	}
	c.Next()
}

func (h *TunnelHandler) withURL(v any) gin.H {
	return gin.H{"tunnel": v, "connect_url": h.svc.ConnectURL()}
}
